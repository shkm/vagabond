package ui

import (
	"os"

	termui "github.com/gizak/termui/v3"
	termui_widgets "github.com/gizak/termui/v3/widgets"
	"github.com/shkm/vagabond/ui/widgets"
)

const statusLineHeight = 1

// UI the TUI for Vagabond
type UI struct {
	StatusLine  *widgets.StatusLine
	FileManager *termui_widgets.List
}

// Render the TUI
func (ui *UI) Render() {
	termui.Render(ui.FileManager, ui.StatusLine)
}

// Loop listens for events
func (ui *UI) Loop() {
	uiEvents := termui.PollEvents()
	for {
		path := "/" + ui.FileManager.Rows[ui.FileManager.SelectedRow]

		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			os.Exit(0)
		case "j":
			if ui.FileManager.SelectedRow < len(ui.FileManager.Rows)-1 {
				ui.FileManager.SelectedRow++
			} else {
				ui.FileManager.SelectedRow = 0
			}
			path = "/" + ui.FileManager.Rows[ui.FileManager.SelectedRow]
			ui.StatusLine.Text = path
		case "k":
			if ui.FileManager.SelectedRow > 0 {
				ui.FileManager.SelectedRow--
			} else {
				ui.FileManager.SelectedRow = len(ui.FileManager.Rows) - 1
			}
			path = "/" + ui.FileManager.Rows[ui.FileManager.SelectedRow]
		case "l", "<Enter>":
			// files = readDir(client, path)
			// var rows []string
			// for _, file := range files {
			// 	rows = append(rows, file.Name())
			// }

			// ui.FileManager.Rows = rows
			// case "y":
			// 	download(client, statusLine, path)
		}

		ui.StatusLine.Text = path
		ui.Render()
	}
}

// NewUI sets up the UI
func NewUI() *UI {
	if err := termui.Init(); err != nil {
		panic(err)
	}

	return &UI{
		FileManager: newFileManager(),
		StatusLine:  newStatusLine(),
	}
}

func newFileManager() *termui_widgets.List {
	fileManager := termui_widgets.NewList()
	style := termui.NewStyle(termui.ColorBlack, termui.ColorWhite)
	fileManager.SelectedRowStyle = style
	fileManager.Border = false

	width, height := termui.TerminalDimensions()
	fileManager.SetRect(0, 0, width, height-statusLineHeight)

	return fileManager
}

func newStatusLine() *widgets.StatusLine {
	statusLine := widgets.NewStatusLine()
	statusLine.Border = false

	width, height := termui.TerminalDimensions()
	statusLine.SetRect(0, height-statusLineHeight, width, height)

	return statusLine
}
