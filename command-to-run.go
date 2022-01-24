package cmdline

type commandToRun struct {
	cmd    *command
	values map[string]interface{}
}

func (cl *CommandLine) newCommandToRun(cmd *command, primaryArgValue *string, subsequentArgs []string) (*commandToRun, int, error) {
	cmdToRun := commandToRun{}

	cmdToRun.cmd = cmd
	cmdToRun.values = make(map[string]interface{})
	argsUsed, err := cmdToRun.cmd.PrimaryArgSpec.Parse(&cmdToRun.values, primaryArgValue, subsequentArgs)

	if err != nil {
		return nil, 0, err
	}

	return &cmdToRun, argsUsed, nil
}
