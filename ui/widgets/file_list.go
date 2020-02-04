package widgets

import (
	termui "github.com/gizak/termui/v3"
	"os"
	"path/filepath"
)

// FileList Widget
type FileList struct {
	termui.Block
	FileRows         []FileRow
	SelectedRowIndex int
	SelectedStyle    termui.Style
	HighlightedStyle termui.Style
	Style            termui.Style
	topRow           int
}

func NewFileList() *FileList {
	block := *termui.NewBlock()
	block.Border = false

	return &FileList{
		Block:         block,
		Style:         termui.NewStyle(termui.ColorWhite),
		SelectedStyle: termui.NewStyle(termui.ColorBlack, termui.ColorRed),
	}
}

func (fileList *FileList) SelectRow(rowIndex int) {
	// prevSelected := fileList.SelectedRow()
	fileList.SelectedRow().Style = fileList.Style
	fileList.SelectedRowIndex = rowIndex
	fileList.SelectedRow().Style = fileList.SelectedStyle

	// if prevSelected
}

func (fileList *FileList) PopulateRows(parentPath string, files []os.FileInfo) {
	// TODO: add link to previous path
	// if path != "/" {
	// rows = append(rows, "..")
	// }

	fileList.FileRows = nil

	for _, file := range files {
		path := filepath.Clean(parentPath + "/" + file.Name())

		fileRow := NewFileRow()
		fileRow.Path = path
		fileRow.FileInfo = file
		fileRow.Style = fileList.Style

		fileList.FileRows = append(fileList.FileRows, *fileRow)
	}

	fileList.SelectRow(0)
}

func (fileList *FileList) SelectedRow() *FileRow {
	return &fileList.FileRows[fileList.SelectedRowIndex]
}

func (self *FileList) Draw(buf *termui.Buffer) {
	self.Block.Draw(buf)
	width := self.Size().X
	maxY := self.Inner.Bounds().Max.Y
	minY := self.Inner.Bounds().Min.Y - 1
	selectedRowY := self.SelectedRowIndex
	startFrom := self.topRow

	if selectedRowY+startFrom < minY {
		startFrom = -selectedRowY
	} else if selectedRowY+startFrom > maxY {
		startFrom = maxY - selectedRowY
	}

	self.topRow = startFrom

	for i, row := range self.FileRows {
		row.SetRect(0, i+startFrom, width, i+startFrom+FileRowHeight)
		row.Draw(buf)
		// cells := termui.ParseStyles(row.Name, self.Style)

		// for x, char := range row.Name {
		// 	cell := termui.NewCell(char, self.Style)
		// 	buf.SetCell(cell, image.Pt(x+1, 0).Add(self.Min))
		// }
	}
}
