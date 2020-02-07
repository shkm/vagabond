package widgets

import (
	termui "github.com/gizak/termui/v3"
	"os"
	"path/filepath"
)

// FileList Widget
type FileList struct {
	termui.Block
	FileRows         []*FileRow
	SelectedRowIndex int
	SelectedStyle    termui.Style
	MarkedStyle      termui.Style
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
		MarkedStyle:   termui.NewStyle(termui.ColorBlack, termui.ColorYellow),
	}
}

func (fileList *FileList) GoToPrevMatch() {
	prev, next := fileList.buildMatchIndices()

	if len(prev) > 0 {
		fileList.SelectRow(prev[len(prev)-1])
	} else if len(next) > 0 {
		fileList.SelectRow(next[len(next)-1])
	} else {
		// set status line and return
		return
	}
}

func (fileList *FileList) GoToNextMatch() {
	prev, next := fileList.buildMatchIndices()

	if len(next) > 0 {
		fileList.SelectRow(next[0])
	} else if len(prev) > 0 {
		fileList.SelectRow(prev[0])
	} else {
		// send error to ui
		return
	}
}

func (fileList *FileList) buildMatchIndices() ([]int, []int) {
	selectedRowIndex := fileList.SelectedRowIndex
	indices := fileList.GetMarkedRowIndices()

	var previousIndices []int
	var nextIndices []int

	for _, index := range indices {
		if index == selectedRowIndex {
			continue
		}

		if index < selectedRowIndex {
			previousIndices = append(previousIndices, index)
		} else {
			nextIndices = append(nextIndices, index)
		}
	}

	return previousIndices, nextIndices
}

func (fileList *FileList) GetMarkedRowIndices() []int {
	var marked []int

	for i, row := range fileList.FileRows {
		if len(row.MarkedText) > 0 {
			marked = append(marked, i)
		}
	}

	return marked
}

func (fileList *FileList) SelectRow(rowIndex int) {
	fileList.SelectedRow().Style = fileList.Style
	fileList.SelectedRowIndex = rowIndex
	fileList.SelectedRow().Style = fileList.SelectedStyle
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
		fileRow.MarkedStyle = fileList.MarkedStyle

		fileList.FileRows = append(fileList.FileRows, fileRow)
	}

	fileList.SelectRow(0)
}

func (fileList *FileList) SelectedRow() *FileRow {
	return fileList.FileRows[fileList.SelectedRowIndex]
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
