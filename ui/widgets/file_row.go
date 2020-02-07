package widgets

import (
	termui "github.com/gizak/termui/v3"
	"image"
	"os"
	"regexp"
)

const FileRowHeight = 1

type FileRow struct {
	termui.Block
	Path        string
	FileInfo    os.FileInfo
	Style       termui.Style
	MarkedStyle termui.Style
	MarkedText  string
}

func NewFileRow() *FileRow {
	block := *termui.NewBlock()
	block.Border = false

	return &FileRow{
		Block: block,
	}
}

func isIndexMarked(markString string, markStartIndices [][]int, index int) bool {
	for _, indices := range markStartIndices {
		if index >= indices[0] && index < indices[1] {
			return true
		}
	}

	return false
}

func (fileRow *FileRow) Draw(buf *termui.Buffer) {
	fileRow.Block.Draw(buf)
	fileRow.drawName(buf)
}

func (fileRow *FileRow) DisplayName() string {
	if fileRow.FileInfo.IsDir() {
		return fileRow.FileInfo.Name() + "/"
	} else {
		return fileRow.FileInfo.Name()
	}
}
func (fileRow *FileRow) drawName(buf *termui.Buffer) {
	name := fileRow.DisplayName()

	regex, err := regexp.Compile(fileRow.MarkedText)
	if err != nil {
		panic(err)
		// TODO: throw a proper error to UI
	}

	matchIndices := regex.FindAllStringIndex(name, -1)

	for x, char := range name {
		style := fileRow.Style
		if len(fileRow.MarkedText) > 0 && isIndexMarked(fileRow.MarkedText, matchIndices, x) {
			style = fileRow.MarkedStyle
		}

		cell := termui.NewCell(char, style)
		buf.SetCell(cell, image.Pt(x+1, 0).Add(fileRow.Min))
	}
}
