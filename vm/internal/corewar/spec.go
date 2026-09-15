package corewar

const (
	MaxPlayers      = 4
	IdxMod          = 4096 / 8
	RegNumber       = 16
	CycleToDie      = 1536
	CycleDelta      = 50
	NbrLive         = 21
	MaxChecks       = 10
	argReg     byte = 1
	argDir     byte = 2
	argInd     byte = 4
)

type opSpec struct {
	name    string
	opcode  byte
	cycles  int
	pcode   bool
	idx     bool
	allowed []byte
}

var ops = [16]opSpec{
	{name: "live", opcode: 1, cycles: 10, allowed: []byte{argDir}},
	{name: "ld", opcode: 2, cycles: 5, pcode: true, allowed: []byte{argInd | argDir, argReg}},
	{name: "st", opcode: 3, cycles: 5, pcode: true, allowed: []byte{argReg, argReg | argInd}},
	{name: "add", opcode: 4, cycles: 10, pcode: true, allowed: []byte{argReg, argReg, argReg}},
	{name: "sub", opcode: 5, cycles: 10, pcode: true, allowed: []byte{argReg, argReg, argReg}},
	{name: "and", opcode: 6, cycles: 6, pcode: true, allowed: []byte{argReg | argInd | argDir, argReg | argInd | argDir, argReg}},
	{name: "or", opcode: 7, cycles: 6, pcode: true, allowed: []byte{argReg | argInd | argDir, argReg | argInd | argDir, argReg}},
	{name: "xor", opcode: 8, cycles: 6, pcode: true, allowed: []byte{argReg | argInd | argDir, argReg | argInd | argDir, argReg}},
	{name: "zjmp", opcode: 9, cycles: 20, idx: true, allowed: []byte{argDir}},
	{name: "ldi", opcode: 10, cycles: 25, pcode: true, idx: true, allowed: []byte{argReg | argInd | argDir, argReg | argDir, argReg}},
	{name: "sti", opcode: 11, cycles: 25, pcode: true, idx: true, allowed: []byte{argReg, argReg | argInd | argDir, argReg | argDir}},
	{name: "fork", opcode: 12, cycles: 800, idx: true, allowed: []byte{argDir}},
	{name: "lld", opcode: 13, cycles: 10, pcode: true, allowed: []byte{argInd | argDir, argReg}},
	{name: "lldi", opcode: 14, cycles: 50, pcode: true, idx: true, allowed: []byte{argReg | argInd | argDir, argReg | argDir, argReg}},
	{name: "lfork", opcode: 15, cycles: 1000, idx: true, allowed: []byte{argDir}},
	{name: "nop", opcode: 16, cycles: 2, pcode: true, allowed: []byte{argReg}},
}

func opByCode(code byte) *opSpec {
	if code < 1 || code > byte(len(ops)) {
		return nil
	}
	return &ops[code-1]
}

func argSize(kind byte, idx bool) int {
	switch kind {
	case argReg:
		return 1
	case argInd:
		return 2
	case argDir:
		if idx {
			return 2
		}
		return 4
	default:
		return 0
	}
}
