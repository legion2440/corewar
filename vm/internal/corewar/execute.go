package corewar

import (
	"fmt"

	"github.com/legion2440/corewar/vm/internal/arena"
)

func (vm *VM) execute(p *process) []Event {
	op := opByCode(p.opcode)
	if op == nil {
		p.pc = arena.Normalize(p.pc + 1)
		return nil
	}
	d := vm.decode(p, op)
	if !d.valid {
		oldPC := p.pc
		p.pc = arena.Normalize(p.pc + d.size)
		return []Event{{
			Kind: EventInvalid, Cycle: vm.cycle, ProcessID: p.id, PlayerID: p.playerID,
			Opcode: op.opcode, Address: oldPC, Size: d.size,
			Message: fmt.Sprintf("%s at 0x%04x: %s", op.name, oldPC, d.reason),
		}}
	}

	events := []Event{{Kind: EventInstruction, Cycle: vm.cycle, ProcessID: p.id, PlayerID: p.playerID, Opcode: op.opcode, Address: p.pc, Size: d.size, Valid: true}}
	switch op.opcode {
	case 1:
		p.lastLiveCycle = vm.cycle
		vm.livesSinceCheck++
		claimed := int(-d.args[0].raw)
		valid := claimed >= 1 && claimed <= len(vm.players)
		if valid {
			vm.lastAliveID = claimed
		}
		events = append(events, Event{Kind: EventLive, Cycle: vm.cycle, ProcessID: p.id, PlayerID: claimed, Address: p.pc, Valid: valid})
		vm.advance(p, d.size)
	case 2:
		value := vm.resolve(p, d.args[0], true)
		vm.setReg(p, d.args[1], value)
		p.carry = value == 0
		vm.advance(p, d.size)
	case 3:
		value := vm.reg(p, d.args[0])
		if d.args[1].kind == argReg {
			vm.setReg(p, d.args[1], value)
		} else {
			address := p.pc + int(d.args[1].raw)%IdxMod
			events = append(events, vm.writeInt32(p, address, value))
		}
		vm.advance(p, d.size)
	case 4:
		value := vm.reg(p, d.args[0]) + vm.reg(p, d.args[1])
		vm.setReg(p, d.args[2], value)
		p.carry = value == 0
		vm.advance(p, d.size)
	case 5:
		value := vm.reg(p, d.args[0]) - vm.reg(p, d.args[1])
		vm.setReg(p, d.args[2], value)
		p.carry = value == 0
		vm.advance(p, d.size)
	case 6, 7, 8:
		a := vm.resolve(p, d.args[0], true)
		b := vm.resolve(p, d.args[1], true)
		var value int32
		switch op.opcode {
		case 6:
			value = a & b
		case 7:
			value = a | b
		case 8:
			value = a ^ b
		}
		vm.setReg(p, d.args[2], value)
		p.carry = value == 0
		vm.advance(p, d.size)
	case 9:
		if p.carry {
			p.pc = arena.Normalize(p.pc + int(d.args[0].raw)%IdxMod)
		} else {
			vm.advance(p, d.size)
		}
	case 10:
		a := vm.resolve(p, d.args[0], true)
		b := vm.resolve(p, d.args[1], true)
		address := p.pc + int((a+b)%IdxMod)
		vm.setReg(p, d.args[2], vm.arena.ReadInt32(address))
		vm.advance(p, d.size)
	case 11:
		value := vm.reg(p, d.args[0])
		a := vm.resolve(p, d.args[1], true)
		b := vm.resolve(p, d.args[2], true)
		address := p.pc + int((a+b)%IdxMod)
		events = append(events, vm.writeInt32(p, address, value))
		vm.advance(p, d.size)
	case 12:
		childPC := p.pc + int(d.args[0].raw)%IdxMod
		events = append(events, vm.fork(p, childPC))
		vm.advance(p, d.size)
	case 13:
		value := vm.resolve(p, d.args[0], false)
		vm.setReg(p, d.args[1], value)
		p.carry = value == 0
		vm.advance(p, d.size)
	case 14:
		a := vm.resolve(p, d.args[0], false)
		b := vm.resolve(p, d.args[1], false)
		address := p.pc + int(a+b)
		value := vm.arena.ReadInt32(address)
		vm.setReg(p, d.args[2], value)
		vm.advance(p, d.size)
	case 15:
		childPC := p.pc + int(d.args[0].raw)
		events = append(events, vm.fork(p, childPC))
		vm.advance(p, d.size)
	case 16:
		vm.advance(p, d.size)
	}
	return events
}

func (vm *VM) resolve(p *process, arg decodedArg, idxMod bool) int32 {
	switch arg.kind {
	case argReg:
		return vm.reg(p, arg)
	case argDir:
		return arg.raw
	case argInd:
		offset := int(arg.raw)
		if idxMod {
			offset %= IdxMod
		}
		return vm.arena.ReadInt32(p.pc + offset)
	default:
		return 0
	}
}

func (vm *VM) reg(p *process, arg decodedArg) int32 {
	return p.registers[int(arg.raw)-1]
}

func (vm *VM) setReg(p *process, arg decodedArg, value int32) {
	p.registers[int(arg.raw)-1] = value
}

func (vm *VM) advance(p *process, size int) {
	p.pc = arena.Normalize(p.pc + size)
}

func (vm *VM) writeInt32(p *process, address int, value int32) Event {
	vm.arena.WriteInt32(address, value)
	for i := 0; i < 4; i++ {
		vm.owners[arena.Normalize(address+i)] = p.playerID
	}
	return Event{Kind: EventMemoryWrite, Cycle: vm.cycle, ProcessID: p.id, PlayerID: p.playerID, Address: arena.Normalize(address), Size: 4, Valid: true}
}

func (vm *VM) fork(parent *process, pc int) Event {
	child := *parent
	child.id = vm.nextProcessID
	vm.nextProcessID++
	child.pc = arena.Normalize(pc)
	child.opcode = 0
	child.wait = 0
	vm.processes = append(vm.processes, &child)
	return Event{Kind: EventFork, Cycle: vm.cycle, ProcessID: child.id, PlayerID: child.playerID, Address: child.pc, Valid: true}
}
