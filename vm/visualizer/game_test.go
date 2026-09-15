package main

import (
	"testing"

	"github.com/legion2440/corewar/vm/internal/champion"
)

func TestStepCycleMapsVMWriteEventsToFlash(t *testing.T) {
	g, err := newGame([]champion.Champion{{
		Name:        "writer",
		Description: "writes r1 into arena",
		Code:        []byte{3, 0x70, 1, 0, 20},
	}})
	if err != nil {
		t.Fatal(err)
	}

	for range 5 {
		g.stepCycle()
	}
	for addr := 20; addr < 24; addr++ {
		if g.memoryFlash[addr] != 1 {
			t.Fatalf("address %d flash=%v, want 1", addr, g.memoryFlash[addr])
		}
		if g.snapshot.Owners[addr] != 1 {
			t.Fatalf("address %d owner=%d, want player 1", addr, g.snapshot.Owners[addr])
		}
	}
}

func TestStepCycleTracksValidLiveTelemetry(t *testing.T) {
	g, err := newGame([]champion.Champion{{
		Name:        "alive",
		Description: "calls live",
		Code:        []byte{1, 0xff, 0xff, 0xff, 0xff},
	}})
	if err != nil {
		t.Fatal(err)
	}

	for range 10 {
		g.stepCycle()
	}
	if g.livesInPeriod != 1 {
		t.Fatalf("livesInPeriod=%d, want 1", g.livesInPeriod)
	}
	if g.playerLastLive[1] != 10 {
		t.Fatalf("last live cycle=%d, want 10", g.playerLastLive[1])
	}
}

func TestResetClearsPresentationState(t *testing.T) {
	g, err := newGame([]champion.Champion{{Name: "x", Description: "x", Code: []byte{16, 0x40, 1}}})
	if err != nil {
		t.Fatal(err)
	}
	g.selectedAddr = 42
	g.memoryFlash[42] = 1
	g.playerLastLive[1] = 99
	g.paused = false
	g.speed = 100

	if err := g.reset(); err != nil {
		t.Fatal(err)
	}
	if g.selectedAddr != -1 || g.memoryFlash[42] != 0 || len(g.playerLastLive) != 0 || !g.paused || g.speed != 1 {
		t.Fatalf("presentation state was not reset: %+v", g)
	}
}

func TestPresentationHelpers(t *testing.T) {
	if got := opcodeLabel(11); got != "sti" {
		t.Fatalf("opcodeLabel(11)=%q", got)
	}
	if got := truncate("abcdefghijklmnopqrstuvwxyz", 10); got != "abcdefg..." {
		t.Fatalf("truncate=%q", got)
	}
	if got := truncate("кириллица", 6); got != "кир..." {
		t.Fatalf("unicode truncate=%q", got)
	}
}
