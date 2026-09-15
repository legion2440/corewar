package main

import "image/color"

var (
	colorBg          = color.RGBA{R: 7, G: 9, B: 14, A: 255}
	colorPanelBg     = color.RGBA{R: 14, G: 19, B: 29, A: 255}
	colorPanelBorder = color.RGBA{R: 28, G: 37, B: 54, A: 255}
	colorCellEmpty   = color.RGBA{R: 17, G: 23, B: 36, A: 255}
	colorTextMain    = color.RGBA{R: 203, G: 213, B: 225, A: 255}
	colorTextMuted   = color.RGBA{R: 100, G: 116, B: 139, A: 255}

	colorP1Base = color.RGBA{R: 0, G: 240, B: 255, A: 255}
	colorP1Dim  = color.RGBA{R: 8, G: 51, B: 61, A: 255}
	colorP2Base = color.RGBA{R: 16, G: 185, B: 129, A: 255}
	colorP2Dim  = color.RGBA{R: 9, G: 54, B: 38, A: 255}
	colorP3Base = color.RGBA{R: 245, G: 158, B: 11, A: 255}
	colorP3Dim  = color.RGBA{R: 61, G: 40, B: 6, A: 255}
	colorP4Base = color.RGBA{R: 244, G: 63, B: 94, A: 255}
	colorP4Dim  = color.RGBA{R: 61, G: 12, B: 23, A: 255}

	colorPCCursor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	colorHoverBox = color.RGBA{R: 148, G: 163, B: 184, A: 255}
)

func playerColor(playerID int) color.RGBA {
	switch playerID {
	case 1:
		return colorP1Base
	case 2:
		return colorP2Base
	case 3:
		return colorP3Base
	case 4:
		return colorP4Base
	default:
		return colorTextMuted
	}
}

func playerDimColor(playerID int) color.RGBA {
	switch playerID {
	case 1:
		return colorP1Dim
	case 2:
		return colorP2Dim
	case 3:
		return colorP3Dim
	case 4:
		return colorP4Dim
	default:
		return colorCellEmpty
	}
}
