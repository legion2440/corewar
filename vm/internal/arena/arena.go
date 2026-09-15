package arena

const Size = 4096

type Arena struct {
	memory [Size]byte
}

func Normalize(address int) int {
	address %= Size
	if address < 0 {
		address += Size
	}
	return address
}

func (a *Arena) Byte(address int) byte {
	return a.memory[Normalize(address)]
}

func (a *Arena) ReadInt16(address int) int16 {
	return int16(uint16(a.Byte(address))<<8 | uint16(a.Byte(address+1)))
}

func (a *Arena) ReadInt32(address int) int32 {
	return int32(uint32(a.Byte(address))<<24 |
		uint32(a.Byte(address+1))<<16 |
		uint32(a.Byte(address+2))<<8 |
		uint32(a.Byte(address+3)))
}

func (a *Arena) SetByte(address int, value byte) {
	a.memory[Normalize(address)] = value
}

func (a *Arena) WriteInt32(address int, value int32) {
	u := uint32(value)
	a.SetByte(address, byte(u>>24))
	a.SetByte(address+1, byte(u>>16))
	a.SetByte(address+2, byte(u>>8))
	a.SetByte(address+3, byte(u))
}

func (a *Arena) Write(address int, data []byte) {
	for i, b := range data {
		a.SetByte(address+i, b)
	}
}

func (a *Arena) Bytes() [Size]byte {
	return a.memory
}
