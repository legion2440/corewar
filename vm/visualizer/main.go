package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/legion2440/corewar/vm/internal/champion"
	"github.com/legion2440/corewar/vm/internal/corewar"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "corewar-visual: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("corewar-visual", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stdout, "Corewar Ebitengine visualizer")
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, "USAGE:")
		fmt.Fprintln(os.Stdout, "  corewar-visual player1.cor [player2.cor ...]")
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, "CONTROLS:")
		fmt.Fprintln(os.Stdout, "  Space       Play / pause")
		fmt.Fprintln(os.Stdout, "  S           Step one cycle while paused")
		fmt.Fprintln(os.Stdout, "  R           Reset match")
		fmt.Fprintln(os.Stdout, "  1..4        Speed: 1x / 10x / 50x / 100x")
		fmt.Fprintln(os.Stdout, "  Mouse       Hover to inspect, click to pin an address")
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	paths := fs.Args()
	if len(paths) == 0 {
		fs.Usage()
		return nil
	}
	if len(paths) > corewar.MaxPlayers {
		return fmt.Errorf("at most %d players are supported", corewar.MaxPlayers)
	}

	champions := make([]champion.Champion, 0, len(paths))
	for _, path := range paths {
		c, err := champion.Load(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		champions = append(champions, c)
	}

	g, err := newGame(champions)
	if err != nil {
		return err
	}
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Corewar // Ebitengine Visualizer")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return ebiten.RunGame(g)
}
