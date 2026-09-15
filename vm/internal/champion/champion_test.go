package champion

import (
	"encoding/binary"
	"testing"
)

func validBinary(code []byte) []byte {
	b := make([]byte, HeaderSize+len(code))
	binary.BigEndian.PutUint32(b[:4], Magic)
	copy(b[4:], "test")
	sz := 4 + NameLength + 4
	binary.BigEndian.PutUint32(b[sz:sz+4], uint32(len(code)))
	copy(b[sz+4:], "description")
	copy(b[HeaderSize:], code)
	return b
}

func TestParse(t *testing.T) {
	c, err := Parse(validBinary([]byte{1, 2, 3}))
	if err != nil {
		t.Fatal(err)
	}
	if c.Name != "test" || c.Description != "description" || len(c.Code) != 3 {
		t.Fatalf("unexpected champion: %+v", c)
	}
}

func TestRejectsCorruption(t *testing.T) {
	cases := [][]byte{
		{},
		validBinary([]byte{1}),
		validBinary(make([]byte, MaxCodeSize+1)),
	}
	cases[1][0] = 1
	for i, data := range cases {
		if _, err := Parse(data); err == nil {
			t.Fatalf("case %d should fail", i)
		}
	}
	badSize := validBinary([]byte{1, 2})
	binary.BigEndian.PutUint32(badSize[4+NameLength+4:4+NameLength+8], 1)
	if _, err := Parse(badSize); err == nil {
		t.Fatal("size mismatch should fail")
	}
}
