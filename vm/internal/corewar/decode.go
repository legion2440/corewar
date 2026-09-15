package corewar

type decodedArg struct {
	kind byte
	raw  int32
}

type decodedInstruction struct {
	args   []decodedArg
	size   int
	valid  bool
	reason string
}

func (vm *VM) decode(p *process, op *opSpec) decodedInstruction {
	d := decodedInstruction{valid: true, size: 1, args: make([]decodedArg, 0, len(op.allowed))}
	cursor := p.pc + 1
	var kinds []byte
	if op.pcode {
		pcode := vm.arena.Byte(cursor)
		cursor++
		d.size++
		kinds = make([]byte, len(op.allowed))
		for i := range op.allowed {
			bits := (pcode >> (6 - 2*i)) & 0x03
			switch bits {
			case 1:
				kinds[i] = argReg
			case 2:
				kinds[i] = argDir
			case 3:
				kinds[i] = argInd
			default:
				d.valid = false
				if d.reason == "" {
					d.reason = "missing argument type in pcode"
				}
			}
			if kinds[i] != 0 && op.allowed[i]&kinds[i] == 0 {
				d.valid = false
				if d.reason == "" {
					d.reason = "argument type is not allowed"
				}
			}
		}
		unusedPairs := 4 - len(op.allowed)
		if unusedPairs > 0 {
			mask := byte((1 << (unusedPairs * 2)) - 1)
			if pcode&mask != 0 {
				d.valid = false
				if d.reason == "" {
					d.reason = "non-zero unused pcode bits"
				}
			}
		}
	} else {
		kinds = make([]byte, len(op.allowed))
		for i, mask := range op.allowed {
			kinds[i] = mask
		}
	}

	for i, kind := range kinds {
		size := argSize(kind, op.idx)
		d.size += size
		arg := decodedArg{kind: kind}
		switch kind {
		case argReg:
			arg.raw = int32(vm.arena.Byte(cursor))
			if arg.raw < 1 || arg.raw > RegNumber {
				d.valid = false
				if d.reason == "" {
					d.reason = "register number out of range"
				}
			}
		case argDir:
			if op.idx {
				arg.raw = int32(vm.arena.ReadInt16(cursor))
			} else {
				arg.raw = vm.arena.ReadInt32(cursor)
			}
		case argInd:
			arg.raw = int32(vm.arena.ReadInt16(cursor))
		default:
			// Invalid 00 pcode consumes no argument bytes.
		}
		cursor += size
		if i < len(op.allowed) && kind != 0 && op.allowed[i]&kind == 0 {
			d.valid = false
		}
		d.args = append(d.args, arg)
	}
	return d
}
