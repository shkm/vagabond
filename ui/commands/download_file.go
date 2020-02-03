package commands

// DownloadFile Command to download a file
type DownloadFile struct {
	input          string
	ui             CommandUI
	prompt         string
	onEndInput     func(Command)
	onInputChanged func(Command)
}

// NewDownloadFile returns a DownloadFile command
func NewDownloadFile(args *InitCommandArgs) Command {
	return &DownloadFile{
		input:          args.Input,
		ui:             args.Ui,
		prompt:         args.Prompt,
		onEndInput:     args.OnEndInput,
		onInputChanged: args.OnInputChanged,
	}
}

func (command *DownloadFile) GetInput() string {
	return command.input
}

func (command *DownloadFile) GetFullText() string {
	return command.prompt + command.input
}

func (command *DownloadFile) ChangeInput(newInput string) {
	command.input = newInput
	command.onInputChanged(command)
}

func (command *DownloadFile) StartInput() {
	command.ui.ShowCommand(command.GetFullText())
}

func (command *DownloadFile) EndInput() {
	command.onEndInput(command)
}
