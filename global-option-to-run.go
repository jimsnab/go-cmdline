package cmdline

type globalOptionToRun struct {
	Option *globalOption
	Values map[string]any
}

func (cl *CommandLine) newGlobalOptionToRun(globalOpt *globalOption, colonValue *string, subsequentArgs []string) (*globalOptionToRun, int, error) {
	opt := globalOptionToRun{}

	opt.Option = globalOpt
	opt.Values = make(map[string]any)

	argsUsed := 0
	var err error

	if globalOpt.argSpec.ValueSpecs != nil {
		argsUsed, err = globalOpt.argSpec.Parse(&opt.Values, colonValue, subsequentArgs)
		if err != nil {
			return nil, 0, err
		}
	}

	return &opt, argsUsed, nil
}
