package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestDumpMemoryFormat(t *testing.T) {
	var memory [4096]byte
	memory[0] = 0x01
	memory[31] = 0xff
	memory[32] = 0x7a

	var out bytes.Buffer
	dumpMemory(&out, memory)
	lines := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	if len(lines) != 128 {
		t.Fatalf("got %d dump rows, want 128", len(lines))
	}
	if !strings.HasPrefix(lines[0], "00000000  01 ") {
		t.Fatalf("unexpected first row: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "00000020  7a ") {
		t.Fatalf("unexpected second row: %q", lines[1])
	}
	if fields := strings.Fields(lines[0]); len(fields) != 33 {
		t.Fatalf("first row has %d fields, want address + 32 bytes", len(fields))
	}
}
