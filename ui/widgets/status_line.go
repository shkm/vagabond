package widgets

import (
	"image"

	termui "github.com/gizak/termui/v3"
)

const (
	// NormalMode normal mode of operation
	NormalMode = iota
	// CommandMode user is entering command
	CommandMode
	// WaitingMode when user shouldn't be able to do anything
	WaitingMode
)

// StatusLine Widget
type StatusLine struct {
	termui.Block
	Text  string
	Style termui.Style
	Mode  int
}

func NewStatusLine() *StatusLine {
	return &StatusLine{
		Block: *termui.NewBlock(),
		Style: termui.NewStyle(termui.ColorBlack, termui.ColorRed),
		Mode:  NormalMode,
	}
}

func (self *StatusLine) Draw(buf *termui.Buffer) {
	self.Block.Draw(buf)

	// cells := termui.ParseStyles(self.Text, self.Style)

	blankCell := termui.NewCell(' ', self.Style)

	// left pad
	buf.SetCell(blankCell, image.Pt(0, 0).Add(self.Min))

	for x, char := range self.Text {
		cell := termui.NewCell(char, self.Style)
		buf.SetCell(cell, image.Pt(x+1, 0).Add(self.Min))
	}

	for x := len(self.Text) + 1; x < buf.Max.X; x++ {
		buf.SetCell(blankCell, image.Pt(x, 0).Add(self.Min))
	}
}
