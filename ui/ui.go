package ui

import (
	"github.com/shkm/vagabond/ui/commands"
	"os"
	"path/filepath"
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
	FileManager    *widgets.FileList
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

	fileManager := widgets.NewFileList()
	width, height := termui.TerminalDimensions()
	fileManager.SetRect(0, 0, width, height-statusLineHeight)

	ui := &UI{
		FileManager: fileManager,
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
	// selectedDir := ui.FileManager.Rows[ui.FileManager.SelectedRow]

	selectedRow := ui.FileManager.SelectedRow()
	if selectedRow.FileInfo.IsDir() {
		ui.eventBus.Publish("ui:enter_directory", filepath.Clean(selectedRow.Path))
		ui.eventBus.WaitAsync()
	} else {
		// TODO: throw UI error, can't enter a file
	}
}

func (ui *UI) enteredDirectory(path string, files []os.FileInfo) {
	ui.FileManager.PopulateRows(path, files)

	ui.Pwd = path
	ui.StatusLine.Text = filepath.Clean(ui.FileManager.SelectedRow().FileInfo.Name())
	ui.Render()
}

func (ui *UI) leaveDirectory() {
	ui.eventBus.Publish("ui:leave_directory", ui.Pwd)
	ui.eventBus.WaitAsync()
}

func (ui *UI) selectNextFile() {
	if ui.FileManager.SelectedRowIndex < len(ui.FileManager.FileRows)-1 {
		ui.FileManager.SelectRow(ui.FileManager.SelectedRowIndex + 1)
	} else {
		ui.FileManager.SelectRow(0)
	}

	ui.StatusLine.Text = ui.FileManager.SelectedRow().FileInfo.Name()
}

func (ui *UI) selectPrevFile() {
	if ui.FileManager.SelectedRowIndex > 0 {
		ui.FileManager.SelectRow(ui.FileManager.SelectedRowIndex - 1)
	} else {
		ui.FileManager.SelectRow(len(ui.FileManager.FileRows) - 1)
	}

	ui.StatusLine.Text = ui.FileManager.SelectedRow().FileInfo.Name()
}

// func (ui *UI) startFinding() {
// 	commandArgs := &commands.InitCommandArgs{
// 		Ui:             ui,
// 		Prompt:         "/",
// 		OnEndInput:     ui.handleFinishedFinding,
// 		// OnInputChanged: ui.handleFindChanged,
// 	}
// }

// func (ui *UI) handleFinishedFinding(command commands.Command) {

// }

// func (ui *UI) handleFindChanged(command commands.Command) {
// 	for _, row := range ui.FileManager.FileRows {
// 		if strings.Contains(row, command.GetInput()) {

// 		}
// 	}
// 	ui.FileManager.Rows
// }

func (ui *UI) selectedFileName() string {
	return ui.FileManager.SelectedRow().FileInfo.Name()
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
	remotePath := filepath.Clean(ui.Pwd + "/" + ui.FileManager.SelectedRow().FileInfo.Name())
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
			// case "d":
			// 	ui.FileManager.ScrollDown()
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
				// case "f":
				// 	ui.startFinding()
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
