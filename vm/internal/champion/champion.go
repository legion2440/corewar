package champion

import (
	"encoding/binary"
	"fmt"
	"os"
)

const (
	Magic             = uint32(0x00ea83f3)
	NameLength        = 128
	DescriptionLength = 2048
	HeaderSize        = 4 + NameLength + 4 + 4 + DescriptionLength + 4
	MaxCodeSize       = 4096 / 6
)

type Champion struct {
	Name        string
	Description string
	Code        []byte
}

func Load(path string) (Champion, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Champion{}, err
	}
	return Parse(data)
}

func Parse(data []byte) (Champion, error) {
	if len(data) < HeaderSize {
		return Champion{}, fmt.Errorf("file is too small: %d bytes, minimum is %d", len(data), HeaderSize)
	}
	magic := binary.BigEndian.Uint32(data[:4])
	if magic != Magic {
		return Champion{}, fmt.Errorf("wrong signature: 0x%08x", magic)
	}
	sizeOffset := 4 + NameLength + 4
	declared := int(binary.BigEndian.Uint32(data[sizeOffset : sizeOffset+4]))
	if declared > MaxCodeSize {
		return Champion{}, fmt.Errorf("program size %d exceeds maximum %d", declared, MaxCodeSize)
	}
	actual := len(data) - HeaderSize
	if declared != actual {
		return Champion{}, fmt.Errorf("declared program size %d does not match actual size %d", declared, actual)
	}
	descOffset := sizeOffset + 4
	return Champion{
		Name:        cString(data[4 : 4+NameLength]),
		Description: cString(data[descOffset : descOffset+DescriptionLength]),
		Code:        append([]byte(nil), data[HeaderSize:]...),
	}, nil
}

func cString(data []byte) string {
	for i, b := range data {
		if b == 0 {
			return string(data[:i])
		}
	}
	return string(data)
}
