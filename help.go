package cmdline

import "strings"

func (cl *CommandLine) Help(err error, appName string, args []string) {

	ok := true
	if err != nil {
		_, ok = err.(*CommandLineError)
	}

	if ok {
		if len(args) > 0 && (args[0] == "help" || args[0] == "--help" || strings.HasSuffix(args[0], "?")) {
			// process help switch and filter
			filter := ""
			if args[0] == "help" || args[0] == "--help" {
				if len(args) == 2 {
					filter = args[1]
				}
			} else if len(args) == 1 {
				filter = args[0]
			}

			if strings.HasSuffix(filter, "?") {
				filter = filter[0 : len(filter)-1]
				if filter == "-" || filter == "--" {
					filter = ""
				}
			}
			cl.printCommandsWorker(filter, true)
		} else if len(args) > 0 && len(cl.PrimaryCommand(args)) > 0 {
			// command line specified a command but had an error; show help for the command
			cl.helpPrintBlanklnFirst()
			cl.helpPrintln("Syntax error.")
			cl.helpPrintBlankln()
			cl.helpPrintln("Command Help:")
			cl.helpPrintBlankln()
			cl.printCommandWorker(cl.PrimaryCommand(args))
			cl.helpPrintBlankln()
		} else {
			// show full help
			var options string
			if len(cl.globalOptions.values) == 0 {
				options = ""
			} else if len(cl.commands.values) == 1 {
				options = " <options>"
			} else {
				options = " <global options>"
			}

			cmdOptions := ""
			for _, cmd := range cl.commands.values {
				if len(cmd.OptionSpecs.values) > 0 || len(cmd.PrimaryArgSpec.ValueSpecs) > 0 {
					cmdOptions = " <options>"
					break
				}
			}
			if cmdOptions == options {
				cmdOptions = "" // remove redundancy
			}

			cmdToken := " <command>"
			if cl.unnamedCmd != nil {
				cmdToken = ""
			}

			cl.helpPrintln("Usage: " + appName + options + cmdToken + cmdOptions)
			cl.helpPrintBlankln()
			cl.printCommandsWorker("", true)

			helpLen := 0
			for _, cmd := range cl.commands.values {
				helpLen += 60 // fudge factor for each line
				helpLen += len(cmd.PrimaryArgSpec.HelpText)
				helpLen += len(cmd.PrimaryArgSpec.String())
				for _, optionSpec := range cmd.OptionSpecs.values {
					helpLen += 60 // fudge factor for each line
					helpLen += len(optionSpec.HelpText)
					helpLen += len(optionSpec.String())
				}
			}

			// only explain filter mechanism if the help text starts getting long
			if helpLen/60 >= 10 {

				// pick the first command's argument for an example
				sampleArg := ""
				if len(cl.commands.values) > 0 {
					for _, cmdName := range cl.commands.order {
						cmd := cl.commands.values[cmdName]
						sampleArg = cmd.PrimaryArgSpec.Key
						break
					}
				}

				cl.helpPrintBlankln()

				if sampleArg == "" || sampleArg == "~" {
					// unnamed primary arg
					cl.helpPrintln("Search help with: " + appName + " --help <filter text>")
				} else {
					cl.helpPrintln("Search help with " + appName + " --help <filter text>. Example: " + appName + " --help " + sampleArg)
					cl.helpPrintln("Or, put a question mark on the end. Example: " + appName + " " + sampleArg + "?")
				}

				cl.helpPrintBlankln()
			}
		}
	} else {
		// processing produced an error
		cl.helpPrintln("")
		cl.helpPrintln(err.Error())
		cl.helpPrintln("")
	}

	cl.helpRender()
}
