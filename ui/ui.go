package ui

import (
	"os"
	"path/filepath"

	evbus "github.com/asaskevich/EventBus"
	termui "github.com/gizak/termui/v3"
	termui_widgets "github.com/gizak/termui/v3/widgets"
	"github.com/shkm/vagabond/ui/widgets"
)

const statusLineHeight = 1

// UI the TUI for Vagabond
type UI struct {
	StatusLine  *widgets.StatusLine
	FileManager *termui_widgets.List
	Pwd         string
	eventBus    evbus.Bus
}

func NewUI(eventBus evbus.Bus, pwd string) *UI {
	if err := termui.Init(); err != nil {
		panic(err)
	}

	ui := &UI{
		FileManager: newFileManager(),
		StatusLine:  newStatusLine(),
		Pwd:         pwd,
		eventBus:    eventBus,
	}

	eventBus.SubscribeAsync("main:directory_read", ui.enteredDirectory, true)
	eventBus.SubscribeAsync("main:downloaded_file", ui.downloadedFile, true)
	return ui
}

// Render the TUI
func (ui *UI) Render() {
	termui.Render(ui.FileManager, ui.StatusLine)
}

func (ui *UI) enterDirectory() {
	newPath := filepath.Clean(ui.Pwd + "/" + ui.FileManager.Rows[ui.FileManager.SelectedRow])

	ui.eventBus.Publish("ui:enter_directory", newPath)
	ui.eventBus.WaitAsync()
}

func (ui *UI) enteredDirectory(path string, files []os.FileInfo) {
	var rows []string

	if path != "/" {
		rows = append(rows, "..")
	}

	for _, file := range files {
		rows = append(rows, file.Name())
	}

	ui.Pwd = path
	ui.FileManager.Rows = rows
	ui.FileManager.SelectedRow = 0
	ui.StatusLine.Text = filepath.Clean(ui.Pwd + "/" + ui.FileManager.Rows[0])
}

func (ui *UI) leaveDirectory() {
	ui.eventBus.Publish("ui:leave_directory", ui.Pwd)
	ui.eventBus.WaitAsync()
}

func (ui *UI) selectNextFile() {
	if ui.FileManager.SelectedRow < len(ui.FileManager.Rows)-1 {
		ui.FileManager.SelectedRow++
	} else {
		ui.FileManager.SelectedRow = 0
	}

	ui.StatusLine.Text = filepath.Clean(ui.Pwd + "/" + ui.FileManager.Rows[ui.FileManager.SelectedRow])
}

func (ui *UI) selectPrevFile() {
	if ui.FileManager.SelectedRow > 0 {
		ui.FileManager.SelectedRow--
	} else {
		ui.FileManager.SelectedRow = len(ui.FileManager.Rows) - 1
	}

	ui.StatusLine.Text = filepath.Clean(ui.Pwd + "/" + ui.FileManager.Rows[ui.FileManager.SelectedRow])
}

func (ui *UI) downloadSelectedFile() {
	path := filepath.Clean(ui.Pwd + "/" + ui.FileManager.Rows[ui.FileManager.SelectedRow])
	ui.StatusLine.Text = "Downloading " + path + "..."
	ui.Render()

	ui.eventBus.Publish("ui:download_file", path)
	ui.eventBus.WaitAsync()
}

func (ui *UI) downloadedFile(remotePath string, localPath string) {
	ui.StatusLine.Text = "Downloaded " + remotePath + " to " + localPath
	ui.Render()
}

// Loop listens for events
func (ui *UI) Loop() {
	uiEvents := termui.PollEvents()

	for {
		e := <-uiEvents

		switch e.ID {
		case "q", "<C-c>":
			os.Exit(0)
		case "j", "<Down>", "<C-n>":
			ui.selectNextFile()
		case "k", "<Up>", "<C-p>":
			ui.selectPrevFile()
		case "l", "<Enter>":
			ui.enterDirectory()
		case "h", "<Backspace>":
			ui.leaveDirectory()
		case "y":
			ui.downloadSelectedFile()
		}

		ui.Render()
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
