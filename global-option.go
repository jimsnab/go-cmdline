package cmdline

type globalOption struct {
	Handler CommandHandler
	argSpec *argSpec
}

func (cl *CommandLine) newGlobalOption(handler CommandHandler, spec string) *globalOption {
	globalOpt := globalOption{}

	globalOpt.Handler = handler

	argSpec := cl.newArgSpec(spec, false)

	globalOpt.argSpec = argSpec

	return &globalOpt
}

func (glopt *globalOption) String() string {
	return glopt.argSpec.String()
}
