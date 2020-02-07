package widgets

import (
	"image"

	termui "github.com/gizak/termui/v3"
)

// StatusLine Widget
type StatusLine struct {
	termui.Block
	StaticText string
	Text       string
	Style      termui.Style
}

func NewStatusLine() *StatusLine {
	return &StatusLine{
		Block: *termui.NewBlock(),
		Style: termui.NewStyle(termui.ColorBlack, termui.ColorRed),
	}
}

func (statusLine *StatusLine) getFullText() string {
	return statusLine.StaticText + statusLine.Text
}

func (self *StatusLine) Draw(buf *termui.Buffer) {
	self.Block.Draw(buf)

	blankCell := termui.NewCell(' ', self.Style)

	text := self.getFullText()

	// left pad
	buf.SetCell(blankCell, image.Pt(0, 0).Add(self.Min))

	for x, char := range text {
		cell := termui.NewCell(char, self.Style)
		buf.SetCell(cell, image.Pt(x+1, 0).Add(self.Min))
	}

	for x := len(text) + 1; x < buf.Max.X; x++ {
		buf.SetCell(blankCell, image.Pt(x, 0).Add(self.Min))
	}
}
