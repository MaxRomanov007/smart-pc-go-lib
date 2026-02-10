package commands

type CommandError struct {
	Message string
}

func (e *CommandError) Error() string {
	return e.Message
}

func Error(message string) error {
	return &CommandError{message}
}
