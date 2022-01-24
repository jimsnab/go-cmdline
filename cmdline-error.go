package cmdline

import "fmt"

type CommandLineError struct {
	reason string
}

func (e *CommandLineError) Error() string {
	return e.reason
}

func NewCommandLineError(format string, args ...interface{}) error {
	err := new(CommandLineError)
	err.reason = fmt.Sprintf(format, args...)

	return err
}
