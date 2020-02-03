package ui

import (
	"github.com/shkm/vagabond/ui/commands"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	evbus "github.com/asaskevich/EventBus"
	termui "github.com/gizak/termui/v3"
	termui_widgets "github.com/gizak/termui/v3/widgets"
	"github.com/shkm/vagabond/ui/widgets"
)

const (
	statusLineHeight = 1
	// NormalMode normal mode of operation
	NormalMode = iota
	// CommandMode user is entering command
	CommandMode
	// WaitingMode when user shouldn't be able to do anything
	WaitingMode
)

// UI the TUI for Vagabond
type UI struct {
	StatusLine     *widgets.StatusLine
	FileManager    *termui_widgets.List
	Pwd            string
	localPwd       string
	eventBus       evbus.Bus
	mode           int
	currentCommand commands.Command
}

func (ui *UI) ShowCommand(content string) {
	ui.mode = CommandMode
	ui.StatusLine.Text = content
	ui.Render()
}

func NewUI(eventBus evbus.Bus, localPwd string, pwd string) *UI {
	if err := termui.Init(); err != nil {
		panic(err)
	}

	ui := &UI{
		FileManager: newFileManager(),
		StatusLine:  newStatusLine(),
		localPwd:    localPwd,
		Pwd:         pwd,
		eventBus:    eventBus,
		mode:        NormalMode,
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
	selectedDir := ui.FileManager.Rows[ui.FileManager.SelectedRow]

	// TODO: ugly, we should really check if it's actually a dir
	if strings.HasSuffix(selectedDir, "/") || selectedDir == ".." {
		newPath := filepath.Clean(ui.Pwd + "/" + selectedDir)
		ui.eventBus.Publish("ui:enter_directory", newPath)
		ui.eventBus.WaitAsync()
	} else {
		// TODO: throw UI error, can't enter a file
	}
}

func (ui *UI) enteredDirectory(path string, files []os.FileInfo) {
	var rows []string

	if path != "/" {
		rows = append(rows, "..")
	}

	for _, file := range files {
		name := file.Name()
		if file.IsDir() {
			name += "/"
		}
		rows = append(rows, name)
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

func (ui *UI) selectedFileName() string {
	return ui.FileManager.Rows[ui.FileManager.SelectedRow]
}

func (ui *UI) downloadSelectedFile() {
	localPath := filepath.Clean(ui.localPwd + "/" + ui.selectedFileName())
	commandArgs := &commands.InitCommandArgs{
		Input:          localPath,
		Ui:             ui,
		Prompt:         "Download to: ",
		OnEndInput:     ui.handleConfirmedDownloadLocation,
		OnInputChanged: ui.handleCommandInputChanged,
	}

	ui.currentCommand = commands.NewDownloadFile(commandArgs)
	ui.currentCommand.StartInput()
}

func (ui *UI) handleConfirmedDownloadLocation(command commands.Command) {
	remotePath := filepath.Clean(ui.Pwd + "/" + ui.FileManager.Rows[ui.FileManager.SelectedRow])
	localPath := filepath.Clean(command.GetInput())
	ui.mode = WaitingMode
	ui.StatusLine.Text = "Downloading…"
	ui.Render()

	ui.eventBus.Publish("ui:download_file", remotePath, localPath)
}

func (ui *UI) handleCommandInputChanged(command commands.Command) {
	ui.StatusLine.Text = command.GetFullText() // not enough, need prompt too
	ui.Render()
}

func (ui *UI) downloadedFile(remotePath string, localPath string) {
	ui.mode = NormalMode
	ui.StatusLine.Text = "Downloaded " + remotePath + " to " + localPath
	ui.Render()
}

// Loop listens for events
func (ui *UI) Loop() {
	uiEvents := termui.PollEvents()

	for {
		e := <-uiEvents

		switch ui.mode {
		case NormalMode:
			switch e.ID {
			case "q", "<C-c>":
				return
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

		case CommandMode:
			if e.ID == "<C-c>" {
				ui.currentCommand = nil
				ui.mode = NormalMode
				ui.Render()
			} else {
				input := ui.currentCommand.GetInput()

				switch e.ID {
				case "<C-c>":
				case "<Enter>":
					ui.currentCommand.EndInput()
				case "<Backspace>":
					r, size := utf8.DecodeLastRuneInString(input)
					if r == utf8.RuneError && (size == 0 || size == 1) {
						size = 0
					}

					input = input[:len(input)-size]
				case "<Space>":
					input += " "
				default:
					if len(e.ID) == 1 {
						input += e.ID
					}
				}

				ui.currentCommand.ChangeInput(input)
			}
		}
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
