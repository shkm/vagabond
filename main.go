package main

import (
	evbus "github.com/asaskevich/EventBus"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shkm/vagabond/ui"

	termui "github.com/gizak/termui/v3"
	"github.com/pkg/sftp"
)

var sshConnection *exec.Cmd
var sftpClient *sftp.Client
var vagabondUI *ui.UI
var eventBus evbus.Bus

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

func startFTPClient(reader io.Reader, writer io.WriteCloser) (*sftp.Client, string) {
	client, err := sftp.NewClientPipe(reader, writer)
	if err != nil {
		panic(err)
	}

	pwd, err := client.Getwd()
	if err != nil {
		panic(err)
	}

	return client, filepath.Clean(pwd)
}

func leaveDirectory(path string) {
	newPath := filepath.Clean(path + "/..")
	readDir(newPath)
}

func readDir(path string) {
	files, err := sftpClient.ReadDir(path)
	if err != nil {
		panic(err)
	}

	eventBus.Publish("main:directory_read", path, files)
}

func downloadFile(path string) {
	sourceFile, err := sftpClient.Open(path)
	if err != nil {
		// TODO: error event
		panic(err)
	}
	defer sourceFile.Close()

	localPath, err := os.Getwd()
	if err != nil {
		// TODO: error event
		panic(err)
	}

	// TODO: proper path
	destPath := localPath + "/downloaded"
	destFile, err := os.Create(destPath)
	if err != nil {
		// TODO: error event
		panic(err)
	}
	defer destFile.Close()

	sourceFile.WriteTo(destFile)

	eventBus.Publish("main:downloaded_file", path, destPath)
}

func main() {
	// Setup Event Bus
	eventBus = evbus.New()
	eventBus.SubscribeAsync("ui:enter_directory", readDir, true)
	eventBus.SubscribeAsync("ui:download_file", downloadFile, true)
	eventBus.SubscribeAsync("ui:leave_directory", leaveDirectory, true)

	sshConnection, writer, reader := openSSHConnection(os.Args[1])
	defer sshConnection.Wait()

	var pwd string
	sftpClient, pwd = startFTPClient(reader, writer)
	defer sftpClient.Close()

	// Setup UI
	vagabondUI = ui.NewUI(eventBus, pwd)
	defer termui.Close()

	readDir(pwd)

	// render
	vagabondUI.Render()
	vagabondUI.Loop()
}
