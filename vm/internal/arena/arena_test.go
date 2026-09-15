package arena

import "testing"

func TestCircularReadsAndWrites(t *testing.T) {
	var a Arena
	a.WriteInt32(Size-2, 0x01020304)
	if got := a.ReadInt32(Size - 2); got != 0x01020304 {
		t.Fatalf("read=%08x", uint32(got))
	}
	if a.Byte(Size) != 0x03 || a.Byte(-1) != 0x02 {
		t.Fatal("circular addressing failed")
	}
	if Normalize(-1) != Size-1 || Normalize(Size+1) != 1 {
		t.Fatal("normalize failed")
	}
}
