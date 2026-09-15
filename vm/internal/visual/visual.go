package visual

import (
	"fmt"
	"io"

	"github.com/legion2440/corewar/vm/internal/arena"
	"github.com/legion2440/corewar/vm/internal/corewar"
)

// Terminal is a dependency-free diagnostic visualizer. The final graphical renderer
// can implement the same Snapshot/Event boundary without coupling rendering to VM logic.
type Terminal struct {
	Out   io.Writer
	Every int
}

func (t Terminal) Render(state corewar.Snapshot, events []corewar.Event, force bool) {
	every := t.Every
	if every <= 0 {
		every = 50
	}
	if !force && state.Cycle%every != 0 {
		return
	}
	out := t.Out
	if out == nil {
		return
	}

	active := make(map[int]int, len(state.Processes))
	for _, p := range state.Processes {
		active[p.PC] = p.PlayerID
	}
	fmt.Fprint(out, "\x1b[2J\x1b[H")
	fmt.Fprintf(out, "Corewar  cycle=%d  cycle_to_die=%d  processes=%d\n", state.Cycle, state.CycleToDie, len(state.Processes))
	for _, p := range state.Players {
		fmt.Fprintf(out, "P%d %-16s processes=%d code=%d bytes\n", p.ID, p.Name, p.Processes, p.CodeSize)
	}
	fmt.Fprintln(out)
	for row := 0; row < 64; row++ {
		for col := 0; col < 64; col++ {
			addr := row*64 + col
			if id, ok := active[addr]; ok {
				fmt.Fprintf(out, "\x1b[1;3%dm@\x1b[0m", color(id))
				continue
			}
			owner := state.Owners[addr]
			if owner > 0 {
				fmt.Fprintf(out, "\x1b[3%dm%d\x1b[0m", color(owner), owner)
			} else if state.Memory[addr] != 0 {
				fmt.Fprint(out, "+")
			} else {
				fmt.Fprint(out, "·")
			}
		}
		fmt.Fprintln(out)
	}
	if len(events) > 0 {
		last := events[len(events)-1]
		fmt.Fprintf(out, "\nlast event: %s p=%d addr=0x%04x\n", last.Kind, last.ProcessID, arena.Normalize(last.Address))
	}
}

func color(player int) int {
	return 1 + (player-1)%4
}
