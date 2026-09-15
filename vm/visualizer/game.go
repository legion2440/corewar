package main

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/legion2440/corewar/vm/internal/arena"
	"github.com/legion2440/corewar/vm/internal/champion"
	"github.com/legion2440/corewar/vm/internal/corewar"
)

type game struct {
	champions      []champion.Champion
	vm             *corewar.VM
	snapshot       corewar.Snapshot
	memoryFlash    [arena.Size]float32
	hoverAddress   int
	selectedAddr   int
	paused         bool
	speed          int
	lastCheckCycle int
	livesInPeriod  int
	playerLastLive map[int]int
	lastEvents     []corewar.Event
}

func newGame(champions []champion.Champion) (*game, error) {
	vm, err := corewar.New(champions)
	if err != nil {
		return nil, err
	}
	g := &game{
		champions:      append([]champion.Champion(nil), champions...),
		vm:             vm,
		snapshot:       vm.Snapshot(),
		hoverAddress:   -1,
		selectedAddr:   -1,
		paused:         true,
		speed:          1,
		playerLastLive: make(map[int]int, len(champions)),
	}
	return g, nil
}

func (g *game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.paused = !g.paused
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) && g.paused && g.vm.Alive() {
		g.stepCycle()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		if err := g.reset(); err != nil {
			return err
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDigit1) {
		g.speed = 1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDigit2) {
		g.speed = 10
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDigit3) {
		g.speed = 50
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDigit4) {
		g.speed = 100
	}

	mx, my := ebiten.CursorPosition()
	if mx >= arenaX && mx < arenaX+arenaWidth && my >= arenaY && my < arenaY+arenaHeight {
		col := (mx - arenaX) / cellSize
		row := (my - arenaY) / cellSize
		g.hoverAddress = row*arenaCols + col
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.selectedAddr = g.hoverAddress
		}
	} else {
		g.hoverAddress = -1
	}

	for i := range g.memoryFlash {
		if g.memoryFlash[i] <= 0 {
			continue
		}
		g.memoryFlash[i] -= 0.05
		if g.memoryFlash[i] < 0 {
			g.memoryFlash[i] = 0
		}
	}

	if !g.paused && g.vm.Alive() {
		for i := 0; i < g.speed && g.vm.Alive(); i++ {
			g.stepCycle()
		}
	}
	return nil
}

func (g *game) stepCycle() {
	events := g.vm.Step()
	g.lastEvents = append(g.lastEvents[:0], events...)
	for _, event := range events {
		switch event.Kind {
		case corewar.EventMemoryWrite:
			size := event.Size
			if size <= 0 {
				size = 1
			}
			for off := 0; off < size; off++ {
				g.memoryFlash[arena.Normalize(event.Address+off)] = 1
			}
		case corewar.EventLive:
			if event.Valid {
				g.livesInPeriod++
				g.playerLastLive[event.PlayerID] = event.Cycle
			}
		case corewar.EventCycleCheck:
			g.lastCheckCycle = event.Cycle
			g.livesInPeriod = 0
		}
	}
	g.snapshot = g.vm.Snapshot()
}

func (g *game) reset() error {
	vm, err := corewar.New(g.champions)
	if err != nil {
		return err
	}
	g.vm = vm
	g.snapshot = vm.Snapshot()
	g.hoverAddress = -1
	g.selectedAddr = -1
	g.paused = true
	g.speed = 1
	g.lastCheckCycle = 0
	g.livesInPeriod = 0
	clear(g.playerLastLive)
	g.lastEvents = nil
	for i := range g.memoryFlash {
		g.memoryFlash[i] = 0
	}
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(colorBg)
	g.drawHeader(screen)
	g.drawArena(screen)
	g.drawInspector(screen)
	g.drawSidebar(screen)
}

func (g *game) drawHeader(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, screenWidth, 48, colorPanelBg, false)
	vector.StrokeLine(screen, 0, 48, screenWidth, 48, 1, colorPanelBorder, false)

	status := "RUNNING"
	if !g.vm.Alive() {
		status = "TERMINATED"
	} else if g.paused {
		status = "PAUSED"
	}

	elapsed := g.snapshot.Cycle - g.lastCheckCycle
	info := fmt.Sprintf("COREWAR // EBITENGINE | STATUS: %s | SPEED: %dx | CYCLE: %d | CTD: %d/%d",
		status, g.speed, g.snapshot.Cycle, elapsed, g.snapshot.CycleToDie)
	ebitenutil.DebugPrintAt(screen, info, 24, 16)

	controls := "[SPACE] Play/Pause  [S] Step  [R] Reset  [1..4] Speed"
	ebitenutil.DebugPrintAt(screen, controls, screenWidth-470, 16)
}

func (g *game) drawArena(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, arenaX-4, arenaY-4, arenaWidth+8, arenaHeight+8, colorPanelBg, false)
	vector.StrokeRect(screen, arenaX-4, arenaY-4, arenaWidth+8, arenaHeight+8, 1, colorPanelBorder, false)

	for row := 0; row < arenaRows; row++ {
		for col := 0; col < arenaCols; col++ {
			addr := row*arenaCols + col
			x := float32(arenaX + col*cellSize)
			y := float32(arenaY + row*cellSize)

			cellColor := colorCellEmpty
			if owner := g.snapshot.Owners[addr]; owner > 0 {
				cellColor = playerDimColor(owner)
			}
			vector.DrawFilledRect(screen, x, y, cellSize-cellPadding, cellSize-cellPadding, cellColor, false)

			if g.memoryFlash[addr] > 0 {
				alpha := uint8(g.memoryFlash[addr] * 255)
				vector.DrawFilledRect(screen, x, y, cellSize-cellPadding, cellSize-cellPadding,
					color.RGBA{R: 255, G: 255, B: 255, A: alpha}, false)
			}
		}
	}

	for _, proc := range g.snapshot.Processes {
		pc := arena.Normalize(proc.PC)
		col := pc % arenaCols
		row := pc / arenaCols
		x := float32(arenaX + col*cellSize)
		y := float32(arenaY + row*cellSize)

		vector.DrawFilledRect(screen, x, y, cellSize, cellSize, colorPCCursor, false)
		vector.DrawFilledRect(screen, x+2, y+2, cellSize-4, cellSize-4, playerColor(proc.PlayerID), false)
	}

	target := g.hoverAddress
	if target == -1 {
		target = g.selectedAddr
	}
	if target >= 0 && target < arena.Size {
		col := target % arenaCols
		row := target / arenaCols
		x := float32(arenaX + col*cellSize)
		y := float32(arenaY + row*cellSize)
		vector.StrokeRect(screen, x-1, y-1, cellSize+2, cellSize+2, 1, colorHoverBox, false)
	}
}

func (g *game) drawInspector(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, inspectorX, inspectorY, inspectorW, inspectorH, colorPanelBg, false)
	vector.StrokeRect(screen, inspectorX, inspectorY, inspectorW, inspectorH, 1, colorPanelBorder, false)

	target := g.selectedAddr
	if target == -1 {
		target = g.hoverAddress
	}
	ebitenutil.DebugPrintAt(screen, "BYTE & PROCESS INSPECTOR [Hover or click an arena cell]", inspectorX+16, inspectorY+12)

	if target < 0 || target >= arena.Size {
		ebitenutil.DebugPrintAt(screen, "Move the cursor over the 64x64 arena to inspect memory in real time.", inspectorX+16, inspectorY+48)
		return
	}

	value := g.snapshot.Memory[target]
	owner := g.snapshot.Owners[target]
	opcode := opcodeLabel(value)
	details := fmt.Sprintf(
		"Address: 0x%04X (%d) | Col: %d, Row: %d\nByte Value: 0x%02X (%d) | Owner: Player %d\nOpcode Mapped: %s",
		target, target, target%arenaCols, target/arenaCols, value, value, owner, opcode,
	)
	ebitenutil.DebugPrintAt(screen, details, inspectorX+16, inspectorY+36)

	for _, proc := range g.snapshot.Processes {
		if arena.Normalize(proc.PC) != target {
			continue
		}
		procInfo := fmt.Sprintf("Process PID: #%d (Player %d) | Carry: %t | Wait: %d cycles | Latched opcode: %d",
			proc.ID, proc.PlayerID, proc.Carry, proc.CyclesRemaining, proc.Opcode)
		ebitenutil.DebugPrintAt(screen, procInfo, inspectorX+16, inspectorY+90)
		return
	}
	ebitenutil.DebugPrintAt(screen, "No active process currently executing at this address.", inspectorX+16, inspectorY+90)
}

func (g *game) drawSidebar(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, sidebarX, sidebarY, sidebarW, sidebarH, colorPanelBg, false)
	vector.StrokeRect(screen, sidebarX, sidebarY, sidebarW, sidebarH, 1, colorPanelBorder, false)

	ebitenutil.DebugPrintAt(screen, "COREWAR ARENA TELEMETRY", sidebarX+20, sidebarY+16)

	elapsed := g.snapshot.Cycle - g.lastCheckCycle
	progress := float32(0)
	if g.snapshot.CycleToDie > 0 {
		progress = float32(elapsed) / float32(g.snapshot.CycleToDie)
		if progress > 1 {
			progress = 1
		}
	}

	barW := float32(sidebarW - 40)
	vector.DrawFilledRect(screen, sidebarX+20, sidebarY+42, barW, 8, colorCellEmpty, false)
	barColor := colorP1Base
	if progress > 0.8 {
		barColor = colorP4Base
	}
	vector.DrawFilledRect(screen, sidebarX+20, sidebarY+42, barW*progress, 8, barColor, false)

	stats := fmt.Sprintf("Current Cycle: %d | Active Processes: %d | CYCLE_TO_DIE: %d | Period Lives: %d",
		g.snapshot.Cycle, len(g.snapshot.Processes), g.snapshot.CycleToDie, g.livesInPeriod)
	ebitenutil.DebugPrintAt(screen, stats, sidebarX+20, sidebarY+58)

	cardY := float32(sidebarY + 90)
	for _, player := range g.snapshot.Players {
		vector.DrawFilledRect(screen, sidebarX+20, cardY, barW, 64, colorCellEmpty, false)
		vector.StrokeRect(screen, sidebarX+20, cardY, barW, 64, 1, playerColor(player.ID), false)

		status := "DEAD"
		if player.Processes > 0 {
			status = "ALIVE"
		}
		text := fmt.Sprintf("PLAYER %d: %s\n\"%s\"\nProcs: %d | Last Live Cycle: %d | Status: %s",
			player.ID, truncate(player.Name, 38), truncate(player.Description, 72), player.Processes,
			g.playerLastLive[player.ID], status)
		ebitenutil.DebugPrintAt(screen, text, int(sidebarX+32), int(cardY+10))
		cardY += 76
	}

	if !g.vm.Alive() {
		boxY := cardY + 20
		vector.DrawFilledRect(screen, sidebarX+20, boxY, barW, 54, colorP2Dim, false)
		vector.StrokeRect(screen, sidebarX+20, boxY, barW, 54, 1, colorP2Base, false)
		message := "MATCH TERMINATED // NOBODY WINS"
		if id, name, ok := g.vm.Winner(); ok {
			message = fmt.Sprintf("MATCH TERMINATED // WINNER: PLAYER %d\n%s", id, name)
		}
		ebitenutil.DebugPrintAt(screen, message, int(sidebarX+32), int(boxY+12))
	}
}

func (g *game) Layout(_, _ int) (int, int) {
	return screenWidth, screenHeight
}

func truncate(value string, max int) string {
	value = strings.ReplaceAll(value, "\n", " ")
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	if max <= 3 {
		return string(runes[:max])
	}
	return string(runes[:max-3]) + "..."
}

func opcodeLabel(opcode byte) string {
	names := [...]string{
		"",
		"live", "ld", "st", "add", "sub", "and", "or", "xor",
		"zjmp", "ldi", "sti", "fork", "lld", "lldi", "lfork", "nop",
	}
	if int(opcode) < len(names) && names[opcode] != "" {
		return names[opcode]
	}
	return "-"
}
