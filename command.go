package cmdline

import (
	"fmt"
)

const basePanic = "command line template syntax error! expected "

type Values map[string]any
type CommandHandler func(values Values) error

type command struct {
	Handler        CommandHandler
	PrimaryArgSpec *argSpec
	OptionSpecs    *orderedArgSpecMap
}

func (cl *CommandLine) newCommand(handler CommandHandler, specList ...string) *command {
	cmd := command{}

	cmd.Handler = handler

	if len(specList) == 0 {
		panic(fmt.Errorf("argument error: specList is required"))
	}

	spec := cl.newArgSpec(specList[0], true)

	cmd.PrimaryArgSpec = spec

	cmd.OptionSpecs = newOrderedArgSpecMap()
	for i := 1; i < len(specList); i++ {
		spec := cl.newArgSpec(specList[i], false)
		cmd.OptionSpecs.add(spec.Key, spec)
	}

	return &cmd
}
