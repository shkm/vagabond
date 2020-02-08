package widgets

import (
	termui "github.com/gizak/termui/v3"
	"image"
)

// LineCharacter the character to use when drawing the line
const LineCharacter = '│'

type VerticalSeparator struct {
	termui.Block
	Style termui.Style
}

func NewVerticalSeparator() *VerticalSeparator {
	block := *termui.NewBlock()
	block.Border = false

	return &VerticalSeparator{
		Block: block,
		Style: termui.NewStyle(termui.ColorWhite),
	}
}

func (verticalSeparator *VerticalSeparator) Draw(buf *termui.Buffer) {

	for y := 0; y < verticalSeparator.Inner.Bounds().Max.Y; y++ {
		cell := termui.NewCell(LineCharacter, verticalSeparator.Style)
		buf.SetCell(cell, image.Pt(0, y).Add(verticalSeparator.Min))
	}
}
