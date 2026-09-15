package main

const (
	screenWidth  = 1600
	screenHeight = 960

	cellSize    = 11
	cellPadding = 1
	arenaCols   = 64
	arenaRows   = 64
	arenaWidth  = arenaCols * cellSize
	arenaHeight = arenaRows * cellSize
	arenaX      = 24
	arenaY      = 64

	inspectorX = 24
	inspectorY = 784
	inspectorW = arenaWidth
	inspectorH = screenHeight - inspectorY - 24

	sidebarX = arenaX + arenaWidth + 24
	sidebarY = 64
	sidebarW = screenWidth - sidebarX - 24
	sidebarH = screenHeight - sidebarY - 24
)
