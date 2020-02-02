package main

import (
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/shkm/vagabond/ui"
	"github.com/shkm/vagabond/ui/widgets"

	termui "github.com/gizak/termui/v3"
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

func download(client *sftp.Client, statusLine *widgets.StatusLine, path string) {
	sourceFile, _ := client.Open(path)
	defer sourceFile.Close()

	localPath, _ := os.Getwd()
	destPath := localPath + "/downloaded"
	var _ = destPath
	destFile, _ := os.Create(destPath)

	defer destFile.Close()

	sourceFile.WriteTo(destFile)
}

func main() {
	cmd, writer, reader := openSSHConnection(os.Args[1])
	defer cmd.Wait()

	client := startFTPClient(reader, writer)
	defer client.Close()

	// Setup UI
	vagabondUI := ui.NewUI()
	defer termui.Close()

	// populate dir
	path := "/"
	files := readDir(client, path)

	var rows []string
	for _, file := range files {
		rows = append(rows, file.Name())
	}
	vagabondUI.FileManager.Rows = rows

	// Populate status line
	path = "/" + vagabondUI.FileManager.Rows[vagabondUI.FileManager.SelectedRow]
	vagabondUI.StatusLine.Text = path

	// render
	vagabondUI.Render()
	vagabondUI.Loop()
}
