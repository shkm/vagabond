package commands

// Find Command to find files
type Find struct {
	input          string
	ui             CommandUI
	prompt         string
	onEndInput     func(Command)
	onInputChanged func(Command)
}

// NewFind returns a File command
func NewFind(args *InitCommandArgs) Command {
	return &Find{
		input:          args.Input,
		ui:             args.Ui,
		prompt:         args.Prompt,
		onEndInput:     args.OnEndInput,
		onInputChanged: args.OnInputChanged,
	}
}

func (command *Find) GetInput() string {
	return command.input
}

func (command *Find) GetFullText() string {
	return command.prompt + command.input
}

func (command *Find) ChangeInput(newInput string) {
	command.input = newInput
	command.onInputChanged(command)
}

func (command *Find) StartInput() {
	command.ui.ShowCommand(command.prompt, command.input)
}

func (command *Find) EndInput() {
	command.onEndInput(command)
}
