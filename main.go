package main

import (
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/shkm/vagabond/ui/widgets"

	termui "github.com/gizak/termui/v3"
	termui_widgets "github.com/gizak/termui/v3/widgets"
	"github.com/pkg/sftp"
)

func openSSHConnection(host string) (*exec.Cmd, io.WriteCloser, io.Reader) {
	cmd := exec.Command("ssh", host, "-s", "sftp")

	writer, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	reader, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	// start the process
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	return cmd, writer, reader
}

func startFTPClient(reader io.Reader, writer io.WriteCloser) *sftp.Client {
	client, err := sftp.NewClientPipe(reader, writer)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func readDir(client *sftp.Client, path string) []os.FileInfo {
	files, err := client.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	return files
}

func download(client *sftp.Client, path string) {
	source_file, _ := client.Open(path)
	defer source_file.Close()

	local_path, _ := os.Getwd()
	dest_path := local_path + "/downloaded"
	var _ = dest_path
	dest_file, _ := os.Create(dest_path)

	defer dest_file.Close()

	source_file.WriteTo(dest_file)
}

func main() {
	cmd, writer, reader := openSSHConnection(os.Args[1])
	defer cmd.Wait()

	client := startFTPClient(reader, writer)
	defer client.Close()

	// Setup UI
	if err := termui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer termui.Close()

	//read a directory
	path := "/"
	files := readDir(client, path)

	// Create list
	list := termui_widgets.NewList()
	style := termui.NewStyle(termui.ColorBlack, termui.ColorWhite)
	list.SelectedRowStyle = style

	var rows []string
	for _, file := range files {
		rows = append(rows, file.Name())
	}

	list.Rows = rows

	// Create messages

	messages := widgets.NewStatusLine()
	// messages := termui_widgets.NewParagraph()

	// Position
	width, height := termui.TerminalDimensions()
	messages_height := 1
	list.Border = false
	messages.Border = false
	list.SetRect(0, 0, width, height-messages_height)
	messages.SetRect(0, height-messages_height, width, height)

	messages_style := termui.NewStyle(termui.ColorBlack, termui.ColorBlue)
	messages.Style = messages_style
	path = "/" + list.Rows[list.SelectedRow]
	messages.Text = path

	// render
	termui.Render(list, messages)
	uiEvents := termui.PollEvents()

	for {
		path = "/" + list.Rows[list.SelectedRow]

		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			os.Exit(0)
		case "j":
			if list.SelectedRow < len(list.Rows)-1 {
				list.SelectedRow += 1
			} else {
				list.SelectedRow = 0
			}

			messages.Text = path
		case "k":
			if list.SelectedRow > 0 {
				list.SelectedRow -= 1
			} else {
				list.SelectedRow = len(list.Rows) - 1
			}
		case "l", "<Enter>":
			path += "/" + list.Rows[list.SelectedRow]
			files = readDir(client, path)
			var rows []string
			for _, file := range files {
				rows = append(rows, file.Name())
			}

			list.Rows = rows
		case "y":
			download(client, path)
		}

		messages.Text = path
		termui.Render(list, messages)
	}
}
