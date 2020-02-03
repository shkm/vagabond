package commands

type CommandUI interface {
	ShowCommand(string)
}

type Command interface {
	ChangeInput(string)
	StartInput()
	EndInput()
	GetInput() string
	GetFullText() string
}

type InitCommandArgs struct {
	Input          string
	Ui             CommandUI
	Prompt         string
	OnEndInput     func(Command)
	OnInputChanged func(Command)
}
