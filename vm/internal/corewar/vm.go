package corewar

import (
	"fmt"

	"github.com/legion2440/corewar/vm/internal/arena"
	"github.com/legion2440/corewar/vm/internal/champion"
)

type player struct {
	id   int
	name string
	desc string
	size int
}

type process struct {
	id            int
	playerID      int
	pc            int
	registers     [RegNumber]int32
	carry         bool
	lastLiveCycle int
	opcode        byte
	wait          int
}

type VM struct {
	arena               arena.Arena
	owners              [arena.Size]int
	players             []player
	processes           []*process
	cycle               int
	cycleToDie          int
	lastCheckCycle      int
	livesSinceCheck     int
	checksSinceDecrease int
	lastAliveID         int
	nextProcessID       int
}

func New(champions []champion.Champion) (*VM, error) {
	if len(champions) == 0 || len(champions) > MaxPlayers {
		return nil, fmt.Errorf("expected 1 to %d players, got %d", MaxPlayers, len(champions))
	}
	vm := &VM{cycleToDie: CycleToDie, nextProcessID: 1}
	spacing := arena.Size / len(champions)
	for i, c := range champions {
		id := i + 1
		start := i * spacing
		vm.players = append(vm.players, player{id: id, name: c.Name, desc: c.Description, size: len(c.Code)})
		vm.arena.Write(start, c.Code)
		for j := range c.Code {
			vm.owners[arena.Normalize(start+j)] = id
		}
		p := &process{id: vm.nextProcessID, playerID: id, pc: start}
		vm.nextProcessID++
		p.registers[0] = int32(-id)
		vm.processes = append(vm.processes, p)
	}
	return vm, nil
}

func (vm *VM) Cycle() int           { return vm.cycle }
func (vm *VM) Alive() bool          { return len(vm.processes) > 0 }
func (vm *VM) CycleToDieValue() int { return vm.cycleToDie }

func (vm *VM) Step() []Event {
	if len(vm.processes) == 0 {
		return nil
	}
	vm.cycle++
	events := make([]Event, 0, 8)
	countAtCycleStart := len(vm.processes)
	for i := countAtCycleStart - 1; i >= 0; i-- {
		p := vm.processes[i]
		if p.wait == 0 {
			code := vm.arena.Byte(p.pc)
			op := opByCode(code)
			if op == nil {
				p.pc = arena.Normalize(p.pc + 1)
				continue
			}
			p.opcode = code
			p.wait = op.cycles
		}
		p.wait--
		if p.wait == 0 && p.opcode != 0 {
			events = append(events, vm.execute(p)...)
			p.opcode = 0
		}
	}

	if vm.cycle-vm.lastCheckCycle > vm.cycleToDie {
		events = append(events, vm.checkProcesses()...)
	}
	return events
}

func (vm *VM) Run() []Event {
	var all []Event
	for vm.Alive() {
		all = append(all, vm.Step()...)
	}
	return all
}

func (vm *VM) Winner() (id int, name string, ok bool) {
	if vm.lastAliveID < 1 || vm.lastAliveID > len(vm.players) {
		return 0, "", false
	}
	p := vm.players[vm.lastAliveID-1]
	return p.id, p.name, true
}

func (vm *VM) Players() []PlayerState {
	return vm.Snapshot().Players
}

func (vm *VM) Snapshot() Snapshot {
	counts := make([]int, len(vm.players)+1)
	processes := make([]ProcessState, 0, len(vm.processes))
	for _, p := range vm.processes {
		counts[p.playerID]++
		processes = append(processes, ProcessState{
			ID: p.id, PlayerID: p.playerID, PC: p.pc, Carry: p.carry,
			CyclesRemaining: p.wait, Opcode: p.opcode, LastLiveCycle: p.lastLiveCycle,
		})
	}
	players := make([]PlayerState, 0, len(vm.players))
	for _, p := range vm.players {
		players = append(players, PlayerState{ID: p.id, Name: p.name, Description: p.desc, CodeSize: p.size, Processes: counts[p.id]})
	}
	return Snapshot{
		Cycle: vm.cycle, CycleToDie: vm.cycleToDie, Memory: vm.arena.Bytes(), Owners: vm.owners,
		Processes: processes, Players: players, LastAliveID: vm.lastAliveID,
	}
}

func (vm *VM) checkProcesses() []Event {
	events := []Event{{Kind: EventCycleCheck, Cycle: vm.cycle, Message: "life check"}}
	alive := vm.processes[:0]
	for _, p := range vm.processes {
		if vm.cycle-p.lastLiveCycle > vm.cycleToDie {
			events = append(events, Event{Kind: EventDeath, Cycle: vm.cycle, ProcessID: p.id, PlayerID: p.playerID, Address: p.pc})
			continue
		}
		alive = append(alive, p)
	}
	vm.processes = alive
	vm.lastCheckCycle = vm.cycle

	if vm.livesSinceCheck >= NbrLive {
		vm.cycleToDie -= CycleDelta
		vm.checksSinceDecrease = 0
	} else {
		vm.checksSinceDecrease++
		if vm.checksSinceDecrease >= MaxChecks {
			vm.cycleToDie -= CycleDelta
			vm.checksSinceDecrease = 0
		}
	}
	vm.livesSinceCheck = 0
	return events
}
