package corewar

import (
	"testing"

	"github.com/legion2440/corewar/vm/internal/champion"
)

func executeOnce(t *testing.T, code []byte, prepare func(*VM, *process)) (*VM, *process, []Event) {
	t.Helper()
	vm, err := New([]champion.Champion{champ("test", code)})
	if err != nil {
		t.Fatal(err)
	}
	p := vm.processes[0]
	if prepare != nil {
		prepare(vm, p)
	}
	p.opcode = code[0]
	p.wait = 1
	events := vm.Step()
	return vm, p, events
}

func TestStoreRegisterAndIndirect(t *testing.T) {
	_, p, _ := executeOnce(t, []byte{3, 0x50, 1, 2}, nil)
	if p.registers[1] != -1 {
		t.Fatalf("st register: r2=%d", p.registers[1])
	}

	vm, _, events := executeOnce(t, []byte{3, 0x70, 1, 0, 20}, nil)
	if got := vm.arena.ReadInt32(20); got != -1 {
		t.Fatalf("st indirect wrote %d", got)
	}
	if len(events) < 2 || events[1].Kind != EventMemoryWrite {
		t.Fatalf("missing write event: %#v", events)
	}
}

func TestSubAndBitwiseOperations(t *testing.T) {
	_, p, _ := executeOnce(t, []byte{5, 0x54, 2, 3, 4}, func(_ *VM, p *process) {
		p.registers[1] = 40
		p.registers[2] = 7
	})
	if p.registers[3] != 33 || p.carry {
		t.Fatalf("sub result=%d carry=%v", p.registers[3], p.carry)
	}

	cases := []struct {
		opcode byte
		want   int32
	}{
		{6, 0x0f},
		{7, 0xff},
		{8, 0xf0},
	}
	for _, tc := range cases {
		_, p, _ := executeOnce(t, []byte{tc.opcode, 0xa4, 0, 0, 0, 0xff, 0, 0, 0, 0x0f, 2}, nil)
		if p.registers[1] != tc.want {
			t.Fatalf("opcode %d result=%#x want=%#x", tc.opcode, p.registers[1], tc.want)
		}
	}
}

func TestZjmpUsesCarryAndIdxMod(t *testing.T) {
	_, p, _ := executeOnce(t, []byte{9, 0x02, 0x58}, func(_ *VM, p *process) {
		p.carry = true
	})
	if p.pc != 88 { // 600 % 512
		t.Fatalf("taken zjmp pc=%d", p.pc)
	}

	_, p, _ = executeOnce(t, []byte{9, 0x02, 0x58}, func(_ *VM, p *process) {
		p.carry = false
	})
	if p.pc != 3 {
		t.Fatalf("untaken zjmp pc=%d", p.pc)
	}
}

func TestLoadAddressingLongVariants(t *testing.T) {
	ldCode := []byte{2, 0xd0, 0x02, 0x58, 2}
	_, p, _ := executeOnce(t, ldCode, func(vm *VM, _ *process) {
		vm.arena.WriteInt32(88, 111)
		vm.arena.WriteInt32(600, 222)
	})
	if p.registers[1] != 111 {
		t.Fatalf("ld indirect=%d", p.registers[1])
	}

	lldCode := []byte{13, 0xd0, 0x02, 0x58, 2}
	_, p, _ = executeOnce(t, lldCode, func(vm *VM, _ *process) {
		vm.arena.WriteInt32(88, 111)
		vm.arena.WriteInt32(600, 222)
	})
	if p.registers[1] != 222 {
		t.Fatalf("lld indirect=%d", p.registers[1])
	}
}

func TestLdiAndLldiAddressing(t *testing.T) {
	ldi := []byte{10, 0xa4, 0x01, 0x2c, 0x01, 0x2c, 2} // 300 + 300
	_, p, _ := executeOnce(t, ldi, func(vm *VM, _ *process) {
		vm.arena.WriteInt32(88, 111)
		vm.arena.WriteInt32(600, 222)
	})
	if p.registers[1] != 111 {
		t.Fatalf("ldi=%d", p.registers[1])
	}

	lldi := []byte{14, 0xa4, 0x01, 0x2c, 0x01, 0x2c, 2}
	_, p, _ = executeOnce(t, lldi, func(vm *VM, _ *process) {
		vm.arena.WriteInt32(88, 111)
		vm.arena.WriteInt32(600, 222)
	})
	if p.registers[1] != 222 {
		t.Fatalf("lldi=%d", p.registers[1])
	}
}

func TestLongLoadCarrySemantics(t *testing.T) {
	_, p, _ := executeOnce(t, []byte{13, 0x90, 0, 0, 0, 0, 2}, nil)
	if !p.carry || p.registers[1] != 0 {
		t.Fatalf("lld carry=%v value=%d", p.carry, p.registers[1])
	}

	_, p, _ = executeOnce(t, []byte{14, 0xa4, 0, 100, 0, 100, 2}, nil)
	if p.carry || p.registers[1] != 0 {
		t.Fatalf("lldi changed carry=%v value=%d", p.carry, p.registers[1])
	}
}

func TestStiWritesAtSummedAddress(t *testing.T) {
	code := []byte{11, 0x68, 2, 0, 10, 0, 7}
	vm, _, events := executeOnce(t, code, func(_ *VM, p *process) {
		p.registers[1] = 0x10203040
	})
	if got := vm.arena.ReadInt32(17); got != 0x10203040 {
		t.Fatalf("sti wrote %#x", uint32(got))
	}
	if len(events) < 2 || events[1].Kind != EventMemoryWrite {
		t.Fatalf("missing sti write event: %#v", events)
	}
}

func TestForkAndLforkAddressing(t *testing.T) {
	vm, _, events := executeOnce(t, []byte{12, 0x02, 0x58}, nil)
	if len(vm.processes) != 2 || vm.processes[1].pc != 88 {
		t.Fatalf("fork child=%+v", vm.processes)
	}
	if events[len(events)-1].Kind != EventFork {
		t.Fatalf("fork event=%#v", events)
	}

	vm, _, _ = executeOnce(t, []byte{15, 0x02, 0x58}, nil)
	if len(vm.processes) != 2 || vm.processes[1].pc != 600 {
		t.Fatalf("lfork child pc=%d", vm.processes[1].pc)
	}
}

func TestLiveAndNop(t *testing.T) {
	vm, p, events := executeOnce(t, []byte{1, 0xff, 0xff, 0xff, 0xff}, nil)
	if p.lastLiveCycle != 1 || vm.lastAliveID != 1 {
		t.Fatalf("live state: last cycle=%d player=%d", p.lastLiveCycle, vm.lastAliveID)
	}
	if events[len(events)-1].Kind != EventLive || !events[len(events)-1].Valid {
		t.Fatalf("live event=%#v", events)
	}

	_, p, _ = executeOnce(t, []byte{16, 0x40, 1}, nil)
	if p.pc != 3 {
		t.Fatalf("nop pc=%d", p.pc)
	}
}
