package corewar

import "github.com/legion2440/corewar/vm/internal/arena"

type EventKind string

const (
	EventInstruction EventKind = "instruction"
	EventInvalid     EventKind = "invalid_instruction"
	EventMemoryWrite EventKind = "memory_write"
	EventFork        EventKind = "fork"
	EventLive        EventKind = "live"
	EventDeath       EventKind = "death"
	EventCycleCheck  EventKind = "cycle_check"
)

type Event struct {
	Kind      EventKind
	Cycle     int
	ProcessID int
	PlayerID  int
	Opcode    byte
	Address   int
	Size      int
	Valid     bool
	Message   string
}

type ProcessState struct {
	ID              int
	PlayerID        int
	PC              int
	Carry           bool
	CyclesRemaining int
	Opcode          byte
	LastLiveCycle   int
}

type PlayerState struct {
	ID          int
	Name        string
	Description string
	CodeSize    int
	Processes   int
}

type Snapshot struct {
	Cycle       int
	CycleToDie  int
	Memory      [arena.Size]byte
	Owners      [arena.Size]int
	Processes   []ProcessState
	Players     []PlayerState
	LastAliveID int
}
