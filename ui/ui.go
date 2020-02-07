package ui

import (
	"github.com/shkm/vagabond/ui/commands"
	"os"
	"path/filepath"
	"strconv"
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
	FileList       *widgets.FileList
	Pwd            string
	localPwd       string
	eventBus       evbus.Bus
	mode           int
	currentCommand commands.Command
}

func (ui *UI) ShowCommand(staticText string, inputText string) {
	ui.mode = CommandMode
	ui.StatusLine.StaticText = staticText
	ui.StatusLine.Text = inputText
	ui.Render()
}

func NewUI(eventBus evbus.Bus, localPwd string, pwd string) *UI {
	if err := termui.Init(); err != nil {
		panic(err)
	}

	fileList := widgets.NewFileList()
	width, height := termui.TerminalDimensions()
	fileList.SetRect(0, 0, width, height-statusLineHeight)

	ui := &UI{
		FileList:   fileList,
		StatusLine: newStatusLine(),
		localPwd:   localPwd,
		Pwd:        pwd,
		eventBus:   eventBus,
		mode:       NormalMode,
	}

	eventBus.SubscribeAsync("main:directory_read", ui.enteredDirectory, true)
	eventBus.SubscribeAsync("main:downloaded_file", ui.downloadedFile, true)
	return ui
}

// Render the TUI
func (ui *UI) Render() {
	termui.Render(ui.FileList, ui.StatusLine)
}

func (ui *UI) goToPrevMatch() {
	prev, next := ui.buildMatchIndices()

	if len(prev) > 0 {
		ui.FileList.SelectRow(prev[len(prev)-1])
	} else if len(next) > 0 {
		ui.FileList.SelectRow(next[len(next)-1])
	} else {
		// set status line and return
		return
	}

	ui.Render()
}

func (ui *UI) goToNextMatch() {
	prev, next := ui.buildMatchIndices()

	if len(next) > 0 {
		ui.FileList.SelectRow(next[0])
	} else if len(prev) > 0 {
		ui.FileList.SelectRow(prev[0])
	} else {
		// set status line and return
		return
	}

	ui.Render()
}

func (ui *UI) buildMatchIndices() ([]int, []int) {
	selectedRowIndex := ui.FileList.SelectedRowIndex
	indices := ui.FileList.GetMarkedRowIndices()

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

func (ui *UI) enterDirectory() {
	// selectedDir := ui.FileList.Rows[ui.FileList.SelectedRow]

	selectedRow := ui.FileList.SelectedRow()
	if selectedRow.FileInfo.IsDir() {
		ui.eventBus.Publish("ui:enter_directory", filepath.Clean(selectedRow.Path))
		ui.eventBus.WaitAsync()
	} else {
		// TODO: throw UI error, can't enter a file
	}
}

func (ui *UI) enteredDirectory(path string, files []os.FileInfo) {
	ui.FileList.PopulateRows(path, files)

	ui.Pwd = path
	ui.StatusLine.Text = filepath.Clean(ui.FileList.SelectedRow().FileInfo.Name())
	ui.Render()
}

func (ui *UI) leaveDirectory() {
	ui.eventBus.Publish("ui:leave_directory", ui.Pwd)
	ui.eventBus.WaitAsync()
}

func (ui *UI) selectNextFile() {
	if ui.FileList.SelectedRowIndex < len(ui.FileList.FileRows)-1 {
		ui.FileList.SelectRow(ui.FileList.SelectedRowIndex + 1)
	} else {
		ui.FileList.SelectRow(0)
	}

	ui.StatusLine.Text = ui.FileList.SelectedRow().FileInfo.Name()
}

func (ui *UI) selectPrevFile() {
	if ui.FileList.SelectedRowIndex > 0 {
		ui.FileList.SelectRow(ui.FileList.SelectedRowIndex - 1)
	} else {
		ui.FileList.SelectRow(len(ui.FileList.FileRows) - 1)
	}

	ui.StatusLine.Text = ui.FileList.SelectedRow().FileInfo.Name()
}

func (ui *UI) startFinding() {
	commandArgs := &commands.InitCommandArgs{
		Ui:             ui,
		Prompt:         "/",
		OnEndInput:     ui.handleFinishedFinding,
		OnInputChanged: ui.handleFindChanged,
	}

	// clear all rows first
	for _, row := range ui.FileList.FileRows {
		row.MarkedText = ""
	}

	ui.currentCommand = commands.NewFind(commandArgs)
	ui.currentCommand.StartInput()
}

func (ui *UI) handleFinishedFinding(command commands.Command) {
	count := strconv.Itoa(len(ui.FileList.GetMarkedRowIndices()))
	ui.exitCommandMode("Found " + count + " match(es).")
	ui.Render()
}

func (ui *UI) handleFindChanged(command commands.Command) {
	input := command.GetInput()
	ui.StatusLine.Text = input

	for _, row := range ui.FileList.FileRows {
		if strings.Contains(row.DisplayName(), input) {
			row.MarkedText = input
		} else {
			row.MarkedText = ""
		}
	}

	ui.Render()
}

func (ui *UI) selectedFileName() string {
	return ui.FileList.SelectedRow().FileInfo.Name()
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
	remotePath := filepath.Clean(ui.Pwd + "/" + ui.FileList.SelectedRow().FileInfo.Name())
	localPath := filepath.Clean(command.GetInput())
	ui.mode = WaitingMode
	ui.StatusLine.Text = "Downloading…"
	ui.Render()

	ui.eventBus.Publish("ui:download_file", remotePath, localPath)
}

func (ui *UI) handleCommandInputChanged(command commands.Command) {
	ui.StatusLine.Text = command.GetFullText()
	ui.Render()
}

func (ui *UI) downloadedFile(remotePath string, localPath string) {
	ui.exitCommandMode("Downloaded " + remotePath + " to " + localPath)
	ui.Render()
}

func (ui *UI) exitCommandMode(newText string) {
	ui.mode = NormalMode
	ui.StatusLine.StaticText = ""
	ui.StatusLine.Text = newText
	ui.currentCommand = nil
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
			// 	ui.FileList.ScrollDown()
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
			case "/":
				ui.startFinding()
			case "n":
				ui.goToNextMatch()
			case "N":
				ui.goToPrevMatch()
			}

			ui.Render()

		case CommandMode:
			if e.ID == "<C-c>" || e.ID == "<Escape>" {
				ui.exitCommandMode(filepath.Clean(ui.FileList.SelectedRow().FileInfo.Name()))
				ui.Render()
			} else if e.ID == "<C-c>" || e.ID == "<Enter>" {
				ui.currentCommand.EndInput()
			} else {

				input := ui.currentCommand.GetInput()

				switch e.ID {
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

func newFileList() *termui_widgets.List {
	fileList := termui_widgets.NewList()
	style := termui.NewStyle(termui.ColorBlack, termui.ColorWhite)
	fileList.SelectedRowStyle = style
	fileList.Border = false

	width, height := termui.TerminalDimensions()
	fileList.SetRect(0, 0, width, height-statusLineHeight)

	return fileList
}

func newStatusLine() *widgets.StatusLine {
	statusLine := widgets.NewStatusLine()
	statusLine.Border = false

	width, height := termui.TerminalDimensions()
	statusLine.SetRect(0, height-statusLineHeight, width, height)

	return statusLine
}
