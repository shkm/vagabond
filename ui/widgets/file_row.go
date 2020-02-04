package widgets

import (
	termui "github.com/gizak/termui/v3"
	"image"
	"os"
)

const FileRowHeight = 1

type FileRow struct {
	termui.Block
	Path     string
	FileInfo os.FileInfo
	Style    termui.Style
}

func NewFileRow() *FileRow {
	block := *termui.NewBlock()
	block.Border = false

	return &FileRow{
		Block: block,
	}
}

func (fileRow *FileRow) Draw(buf *termui.Buffer) {
	fileRow.Block.Draw(buf)

	for x, char := range fileRow.DisplayName() {
		cell := termui.NewCell(char, fileRow.Style)
		buf.SetCell(cell, image.Pt(x+1, 0).Add(fileRow.Min))
	}
}

func (fileRow *FileRow) DisplayName() string {
	if fileRow.FileInfo.IsDir() {
		return fileRow.FileInfo.Name() + "/"
	} else {
		return fileRow.FileInfo.Name()
	}
}
