package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/legion2440/corewar/vm/internal/champion"
	"github.com/legion2440/corewar/vm/internal/corewar"
	"github.com/legion2440/corewar/vm/internal/visual"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "corewar: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("corewar", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dump := fs.Int("d", -1, "dump memory after N cycles")
	verbose := fs.Bool("v", false, "print VM execution events")
	visualMode := fs.Bool("visual", false, "run the terminal visualizer")
	visualEvery := fs.Int("visual-every", 50, "visualizer refresh interval in cycles")
	fs.Usage = func() { printHelp(stdout) }
	if len(args) == 0 {
		printHelp(stdout)
		return nil
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	paths := fs.Args()
	if len(paths) == 0 {
		printHelp(stdout)
		return nil
	}
	if len(paths) > corewar.MaxPlayers {
		return fmt.Errorf("at most %d players are supported", corewar.MaxPlayers)
	}
	if *dump < -1 {
		return fmt.Errorf("-d must be zero or a positive cycle number")
	}

	champions := make([]champion.Champion, 0, len(paths))
	for _, path := range paths {
		c, err := champion.Load(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		champions = append(champions, c)
	}
	vm, err := corewar.New(champions)
	if err != nil {
		return err
	}
	printPlayers(stdout, vm.Players())

	renderer := visual.Terminal{Out: stdout, Every: *visualEvery}
	if *visualMode {
		renderer.Render(vm.Snapshot(), nil, true)
	}
	if *dump == 0 {
		dumpMemory(stdout, vm.Snapshot().Memory)
		return nil
	}

	for vm.Alive() {
		events := vm.Step()
		for _, event := range events {
			if event.Kind == corewar.EventInvalid {
				fmt.Fprintf(stderr, "cycle %d: process %d: %s\n", event.Cycle, event.ProcessID, event.Message)
			}
			if *verbose {
				fmt.Fprintf(stderr, "cycle=%d event=%s process=%d player=%d opcode=%d address=0x%04x valid=%v\n",
					event.Cycle, event.Kind, event.ProcessID, event.PlayerID, event.Opcode, event.Address, event.Valid)
			}
		}
		if *visualMode {
			renderer.Render(vm.Snapshot(), events, false)
		}
		if *dump >= 0 && vm.Cycle() == *dump {
			if *visualMode {
				renderer.Render(vm.Snapshot(), events, true)
			}
			dumpMemory(stdout, vm.Snapshot().Memory)
			return nil
		}
	}
	if *visualMode {
		renderer.Render(vm.Snapshot(), nil, true)
	}
	if id, name, ok := vm.Winner(); ok {
		fmt.Fprintf(stdout, "cycle %d: The winner is player %d: %s!\n", vm.Cycle(), id, name)
	} else {
		fmt.Fprintf(stdout, "cycle %d: Nobody wins!\n", vm.Cycle())
	}
	return nil
}

func printPlayers(w io.Writer, players []corewar.PlayerState) {
	fmt.Fprintln(w, "For this match the players will be:")
	for _, p := range players {
		fmt.Fprintf(w, "Player %d (%d bytes): %s (%s)\n", p.ID, p.CodeSize, p.Name, p.Description)
	}
}

func dumpMemory(w io.Writer, memory [4096]byte) {
	for i := 0; i < len(memory); i += 32 {
		fmt.Fprintf(w, "%08x ", i)
		for j := 0; j < 32; j++ {
			fmt.Fprintf(w, " %02x", memory[i+j])
		}
		fmt.Fprintln(w)
	}
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "Corewar virtual machine")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "USAGE:")
	fmt.Fprintln(w, "  corewar [-d NB_CYCLES] [-v] [--visual] player1.cor [player2.cor ...]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "OPTIONS:")
	fmt.Fprintln(w, "  -d N             Stop at cycle N and dump arena memory, 32 bytes per row")
	fmt.Fprintln(w, "  -v               Print deterministic execution events to stderr")
	fmt.Fprintln(w, "  --visual          Run the dependency-free terminal visualizer")
	fmt.Fprintln(w, "  --visual-every N  Refresh visualizer every N cycles (default 50)")
}
