package main

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/legion2440/corewar/vm/internal/champion"
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

func TestRunAcceptsDumpFlagAroundPlayerArguments(t *testing.T) {
	player1 := writeTestChampion(t, "alpha")
	player2 := writeTestChampion(t, "beta")

	tests := []struct {
		name string
		args []string
	}{
		{name: "before players", args: []string{"-d", "5", player1}},
		{name: "after player", args: []string{player1, "-d", "5"}},
		{name: "between players", args: []string{player1, "-d", "5", player2}},
		{name: "after players", args: []string{player1, player2, "-d", "5"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if err := run(tt.args, &stdout, &stderr); err != nil {
				t.Fatalf("run() error = %v, stderr = %q", err, stderr.String())
			}
			if !strings.HasSuffix(stdout.String(), "cycle 5: Nobody wins!\n") {
				t.Fatalf("missing dump result line; tail = %q", tail(stdout.String(), 120))
			}
		})
	}
}

func TestRunPrintsResultAtCycleZeroDump(t *testing.T) {
	player := writeTestChampion(t, "zero")

	var stdout, stderr bytes.Buffer
	if err := run([]string{player, "-d", "0"}, &stdout, &stderr); err != nil {
		t.Fatalf("run() error = %v, stderr = %q", err, stderr.String())
	}
	if !strings.HasSuffix(stdout.String(), "cycle 0: Nobody wins!\n") {
		t.Fatalf("missing cycle-zero result line; tail = %q", tail(stdout.String(), 120))
	}
}

func TestRunDumpsFinalStateWhenRequestedCycleOutlivesMatch(t *testing.T) {
	player := writeTestChampion(t, "short")

	var stdout, stderr bytes.Buffer
	if err := run([]string{player, "-d", "999999"}, &stdout, &stderr); err != nil {
		t.Fatalf("run() error = %v, stderr = %q", err, stderr.String())
	}

	lines := strings.Split(strings.TrimSuffix(stdout.String(), "\n"), "\n")
	dumpRows := 0
	for _, line := range lines {
		if len(line) >= 10 && line[8:10] == "  " {
			dumpRows++
		}
	}
	if dumpRows != 128 {
		t.Fatalf("got %d final dump rows, want 128", dumpRows)
	}
	if !strings.HasSuffix(stdout.String(), "cycle 1536: Nobody wins!\n") {
		t.Fatalf("unexpected final result; tail = %q", tail(stdout.String(), 120))
	}
}

func writeTestChampion(t *testing.T, name string) string {
	t.Helper()

	data := make([]byte, champion.HeaderSize)
	binary.BigEndian.PutUint32(data[:4], champion.Magic)
	copy(data[4:4+champion.NameLength], name)
	descOffset := 4 + champion.NameLength + 4 + 4
	copy(data[descOffset:descOffset+champion.DescriptionLength], "test champion")

	path := filepath.Join(t.TempDir(), name+".cor")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write test champion: %v", err)
	}
	return path
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
