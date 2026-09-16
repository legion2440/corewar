package corewar

import (
	"testing"

	"github.com/legion2440/corewar/vm/internal/champion"
)

func champ(name string, code []byte) champion.Champion {
	return champion.Champion{Name: name, Description: name, Code: append([]byte(nil), code...)}
}

func TestLastPlayerExecutesFirst(t *testing.T) {
	vm, err := New([]champion.Champion{champ("one", []byte{16, 0x40, 1}), champ("two", []byte{16, 0x40, 1})})
	if err != nil {
		t.Fatal(err)
	}
	if events := vm.Step(); len(events) != 0 {
		t.Fatalf("cycle 1 events=%v", events)
	}
	events := vm.Step()
	if len(events) != 2 {
		t.Fatalf("cycle 2 events=%v", events)
	}
	if events[0].ProcessID != 2 || events[1].ProcessID != 1 {
		t.Fatalf("execution order = %d,%d", events[0].ProcessID, events[1].ProcessID)
	}
}

func TestForkedProcessStartsNextCycleAndRunsFirst(t *testing.T) {
	code := []byte{12, 0, 3, 16, 0x40, 1}
	vm, _ := New([]champion.Champion{champ("forker", code)})
	parent := vm.processes[0]
	parent.opcode = 12
	parent.wait = 1
	events := vm.Step()
	if len(vm.processes) != 2 {
		t.Fatalf("processes=%d", len(vm.processes))
	}
	if events[len(events)-1].Kind != EventFork {
		t.Fatalf("events=%v", events)
	}
	if vm.processes[1].pc != 3 {
		t.Fatalf("child pc=%d", vm.processes[1].pc)
	}
	vm.Step()
	events = vm.Step()
	if len(events) < 2 {
		t.Fatalf("events=%v", events)
	}
	if events[0].ProcessID != 2 || events[1].ProcessID != 1 {
		t.Fatalf("fork order=%v", events)
	}
}

func TestInvalidPcodeAdvancesByEncodedParameterSizes(t *testing.T) {
	vm, _ := New([]champion.Champion{champ("bad", []byte{2, 0x50, 1, 2, 16, 0x40, 1})})
	for i := 0; i < 4; i++ {
		vm.Step()
	}
	events := vm.Step()
	if len(events) != 1 || events[0].Kind != EventInvalid {
		t.Fatalf("events=%v", events)
	}
	if vm.processes[0].pc != 4 {
		t.Fatalf("pc=%d", vm.processes[0].pc)
	}
}

func TestLifeCheckBoundaryMatchesReferenceCountdown(t *testing.T) {
	vm, _ := New([]champion.Champion{champ("x", []byte{0})})

	for vm.Cycle() < CycleToDie {
		events := vm.Step()
		for _, event := range events {
			if event.Kind == EventCycleCheck {
				t.Fatalf("life check occurred early at cycle %d", vm.Cycle())
			}
		}
	}
	if !vm.Alive() {
		t.Fatal("process died before reference life-check boundary")
	}

	events := vm.Step()
	if vm.Cycle() != CycleToDie+1 {
		t.Fatalf("cycle=%d, want %d", vm.Cycle(), CycleToDie+1)
	}
	foundCheck := false
	for _, event := range events {
		if event.Kind == EventCycleCheck {
			foundCheck = true
		}
	}
	if !foundCheck {
		t.Fatalf("missing life check at cycle %d", vm.Cycle())
	}
	if vm.Alive() {
		t.Fatal("process without live should die at first life check")
	}
}

func TestLifeChecksAndCycleToDieReduction(t *testing.T) {
	vm, _ := New([]champion.Champion{champ("x", []byte{0})})
	p := vm.processes[0]
	vm.cycle = CycleToDie + 1
	p.lastLiveCycle = 1
	vm.livesSinceCheck = NbrLive
	vm.checkProcesses()
	if !vm.Alive() {
		t.Fatal("process that lived at the interval boundary was killed")
	}
	if vm.cycleToDie != CycleToDie-CycleDelta {
		t.Fatalf("cycleToDie=%d", vm.cycleToDie)
	}

	vm.cycle += vm.cycleToDie + 1
	vm.checkProcesses()
	if vm.Alive() {
		t.Fatal("stale process should be killed")
	}
}

func TestCycleToDieReductionWaitsUntilAfterMaxChecks(t *testing.T) {
	vm, _ := New([]champion.Champion{champ("x", []byte{0})})
	p := vm.processes[0]

	for check := 1; check <= MaxChecks; check++ {
		p.lastLiveCycle = vm.cycle
		vm.checkProcesses()
		if vm.cycleToDie != CycleToDie {
			t.Fatalf("cycleToDie decreased on check %d: got %d, want %d", check, vm.cycleToDie, CycleToDie)
		}
	}

	p.lastLiveCycle = vm.cycle
	vm.checkProcesses()
	if vm.cycleToDie != CycleToDie-CycleDelta {
		t.Fatalf("cycleToDie=%d after check %d, want %d", vm.cycleToDie, MaxChecks+1, CycleToDie-CycleDelta)
	}
}

func TestLdAndArithmeticCarry(t *testing.T) {
	code := []byte{
		2, 0x90, 0, 0, 0, 0, 2,
		4, 0x54, 2, 2, 3,
		16, 0x40, 1,
	}
	vm, _ := New([]champion.Champion{champ("math", code)})
	for i := 0; i < 5; i++ {
		vm.Step()
	}
	p := vm.processes[0]
	if !p.carry || p.registers[1] != 0 || p.pc != 7 {
		t.Fatalf("after ld: %+v", p)
	}
	for i := 0; i < 10; i++ {
		vm.Step()
	}
	if !p.carry || p.registers[2] != 0 || p.pc != 12 {
		t.Fatalf("after add: %+v", p)
	}
}

func TestTerminatorBeatsAmebaInEitherPosition(t *testing.T) {
	ameba := champ("ameba", []byte{
		0x0b, 0x68, 0x01, 0x00, 0x0f, 0x00, 0x01,
		0x06, 0x64, 0x01, 0x00, 0x00, 0x00, 0x00, 0x01,
		0x01, 0x00, 0x00, 0x00, 0x01,
		0x09, 0xff, 0xfb,
	})
	terminator := champ("terminator", []byte{
		0x0b, 0x68, 0x01, 0x00, 0x0f, 0x00, 0x01,
		0x06, 0x64, 0x01, 0x00, 0x00, 0x00, 0x00, 0x02,
		0x01, 0x00, 0x00, 0x00, 0x00,
		0x0f, 0x07, 0xec,
		0x09, 0xff, 0xf8,
	})

	for _, order := range [][]champion.Champion{{terminator, ameba}, {ameba, terminator}} {
		vm, err := New(order)
		if err != nil {
			t.Fatal(err)
		}
		for vm.Alive() && vm.Cycle() < 200000 {
			vm.Step()
		}
		if vm.Alive() {
			t.Fatal("match did not terminate")
		}
		id, name, ok := vm.Winner()
		if !ok || name != "terminator" {
			t.Fatalf("winner id=%d name=%q ok=%v cycle=%d", id, name, ok, vm.Cycle())
		}
		if vm.Cycle() != 24398 {
			t.Fatalf("match ended at cycle %d, want reference cycle 24398", vm.Cycle())
		}
	}
}
