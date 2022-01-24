package cmdline

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

const maxLineWidth = 120
const maxRiver = 30
const riverSpaces = 2

type helpLine struct {
	str1   string
	str2   string
	indent int
	cols   int
}

type CommandLine struct {
	commands      map[string]*command
	unnamedCmd    *command
	globalOptions map[string]*globalOption
	optionTypes   OptionTypes
	printQueue    []helpLine
}

func NewCommandLine() *CommandLine {
	return newCommandLine(nil)
}

func NewCustomTypesCommandLine(optionTypes OptionTypes) *CommandLine {
	return newCommandLine(&optionTypes)
}

func newCommandLine(optionTypes *OptionTypes) *CommandLine {
	cl := CommandLine{}

	cl.commands = make(map[string]*command)
	cl.globalOptions = make(map[string]*globalOption)

	if optionTypes == nil {
		cl.optionTypes = newDefaultOptionTypes()
	} else {
		cl.optionTypes = *optionTypes
	}

	return &cl
}

func (cl *CommandLine) checkForDuplicateName(names map[string]bool, spec string) {
	_, exist := names[spec]
	if exist {
		panic(fmt.Errorf("%sunique argument \"%s\"", basePanic, spec))
	}
	names[spec] = true
}

func (cl *CommandLine) checkForDuplicateNames(newCmd *command) {
	names := make(map[string]bool)

	for _, globalOpt := range cl.globalOptions {
		cl.checkForDuplicateName(names, globalOpt.argSpec.Key)
	}

	allCommands := make([]*command, 0, len(cl.commands)+1)
	for _, cmd := range cl.commands {
		allCommands = append(allCommands, cmd)
	}
	if newCmd != nil {
		allCommands = append(allCommands, newCmd)
	}

	for _, cmd := range allCommands {
		cl.checkForDuplicateName(names, cmd.PrimaryArgSpec.Key)

		cmdNames := make(map[string]bool)
		for k,v := range names {
			cmdNames[k] = v
		}
		
		for _, valueSpec := range cmd.PrimaryArgSpec.ValueSpecs {
			cl.checkForDuplicateName(cmdNames, valueSpec.OptionName)
		}

		for _, optionSpec := range cmd.OptionSpecs {
			cl.checkForDuplicateName(cmdNames, optionSpec.Key)

			for _, valueSpec := range optionSpec.ValueSpecs {
				cl.checkForDuplicateName(cmdNames, valueSpec.OptionName)
			}
		}
	}
}

func (cl *CommandLine) RegisterCommand(handler CommandHandler, specList ...string) {
	cmd := cl.newCommand(handler, specList...)

	cl.checkForDuplicateNames(cmd)

	cl.commands[cmd.PrimaryArgSpec.Key] = cmd

	// unnamed command mode occurs with exactly one command that has name "~"
	if len(cl.commands) == 1 && cmd.PrimaryArgSpec.Unnamed {
		cl.unnamedCmd = cmd
	} else {
		cl.unnamedCmd = nil
	}
}

func (cl *CommandLine) RegisterGlobalOption(handler CommandHandler, spec string) {
	globalOpt := cl.newGlobalOption(handler, spec)

	cl.globalOptions[globalOpt.argSpec.Key] = globalOpt

	cl.checkForDuplicateNames(nil)
}

func (cl *CommandLine) shouldShow(primaryArgSpec *argSpec, optionSpecs *[]*argSpec, filter string) bool {
	filter = strings.TrimSpace(filter)
	if len(filter) == 0 {
		return true
	}

	key := primaryArgSpec.Key

	if strings.Contains(key, filter) {
		return true
	}

	help := strings.ToLower(primaryArgSpec.HelpText)
	if strings.Contains(help, filter) {
		return true
	}

	if optionSpecs != nil {
		for _, option := range *optionSpecs {
			optname := strings.ToLower(option.Key)
			if strings.Contains(optname, filter) {
				return true
			}

			help := strings.ToLower(option.HelpText)
			if strings.Contains(help, filter) {
				return true
			}
		}
	}

	return false
}

func (cl *CommandLine) helpPrintln(text string) {
	cl.printQueue = append(cl.printQueue, helpLine{str1: text, str2: "", cols: 1})
}

func (cl *CommandLine) helpPrintf(fmtString string, args ...interface{}) {
	cl.printQueue = append(cl.printQueue, helpLine{str1: fmt.Sprintf(fmtString, args...), str2: "", cols: 1})
}

func (cl *CommandLine) helpPrintCols(indent int, argText string, description string) {
	if len(argText) == 0 {
		if len(description) > 0 {
			text := strings.Repeat("  ", indent) + description
			cl.printQueue = append(cl.printQueue, helpLine{str1: text, str2: "", cols: 2})
		}
	} else {
		cl.printQueue = append(cl.printQueue, helpLine{indent: indent, str1: argText, str2: description, cols: 2})
	}
}

func (cl *CommandLine) helpPrintBlanklnFirst() {
	if len(cl.printQueue) == 0 {
		cl.helpPrintln("")
	}
}

func (cl *CommandLine) helpPrintBlankln() {
	if len(cl.printQueue) > 0 {
		lastLine := cl.printQueue[len(cl.printQueue)-1]
		if len(lastLine.str1) > 0 {
			cl.helpPrintln("")
		}
	}
}

func (cl *CommandLine) helpRender() {
	// determine the position of the second column
	riverWidth := 0
	for _, help := range cl.printQueue {
		if help.cols > 1 {
			argText := strings.Repeat("  ", help.indent) + help.str1
			width := utf8.RuneCountInString(argText)
			if width > 0 {
				width += riverSpaces
				if width > maxRiver {
					riverWidth = maxRiver
					break
				} else if width > riverWidth {
					riverWidth = width
				}
			}
		}
	}

	// print the lines
	for _, help := range cl.printQueue {
		argText := strings.Repeat("  ", help.indent) + help.str1
		if help.cols == 1 {
			Prn.Println(argText)
		} else {
			cl.indentedPrint(argText, riverWidth, maxLineWidth, help.str2)
		}
	}

	cl.printQueue = []helpLine{}
}

func (cl *CommandLine) indentedPrint(arg string, indent int, wrap int, text string) {
	column := 0
	if len(arg) > 0 {
		Prn.BeginPrint(arg)
		column = utf8.RuneCountInString(arg)

		if len(text) == 0 {
			Prn.EndPrint("")
			return
		}

		if column >= indent {
			Prn.EndPrint("")
			column = 0
		}
	}

	lines := strings.Split(text, "\n")
	nextIndent := ""
	for _, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			if column > 0 {
				Prn.EndPrint("")
				column = 0
			} else {
				Prn.Println("")
			}
			continue
		}

		fullLine := line
		for len(fullLine) > 0 {
			if column == 0 {
				Prn.BeginPrint("")
			}

			if column < indent {
				nextIndent = strings.Repeat(" ", indent-column)
				column = indent
			}

			thisLine := fullLine
			end := column + utf8.RuneCountInString(thisLine)
			if end > wrap {
				cutPoint := -1
				for {
					nextCutPoint := indexOf(thisLine, " ", cutPoint+1)
					if nextCutPoint < 0 || nextCutPoint+column > wrap {
						break
					}
					cutPoint = nextCutPoint
				}

				if cutPoint > 0 {
					thisLine = thisLine[:cutPoint]
				}
			}

			Prn.ContinuePrint(nextIndent)
			nextIndent = ""

			Prn.EndPrint(strings.TrimSpace(thisLine))
			column = 0

			fullLine = strings.TrimSpace(fullLine[len(thisLine):])
		}
	}
}

func (cl *CommandLine) PrimaryCommand(args []string) string {
	filteredArgs := []string{}

	for _, arg := range args {
		argTokens := strings.Split(arg, ":")
		argToken := argTokens[0]
		_, exists := cl.globalOptions[argToken]
		if !exists {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	for _, arg := range filteredArgs {
		argTokens := strings.Split(arg, ":")
		argToken := argTokens[0]
		_, exists := cl.commands[argToken]
		if exists {
			return argToken
		}
	}

	return ""
}

func (cl *CommandLine) PrintCommand(cmdstr string) error {
	err := cl.printCommandWorker(cmdstr)
	if err != nil {
		return err
	}

	cl.helpRender()
	return nil
}

func (cl *CommandLine) printCommandWorker(cmdstr string) error {
	wantUnnamed := false
	if len(cmdstr) == 0 || cmdstr == "~" {
		wantUnnamed = true
		cmdstr = "~"
	}

	cmd, exist := cl.commands[cmdstr]
	if !exist {
		if wantUnnamed {
			return fmt.Errorf("unnamed command not found")
		} else {
			return fmt.Errorf("command \"%s\" not found", cmdstr)
		}
	}

	// no help text specified by the template
	if len(cmd.PrimaryArgSpec.HelpText) == 0 && len(cmd.OptionSpecs) == 0 {
		if wantUnnamed {
			return fmt.Errorf("help not available for the unnamed command")
		} else {
			return fmt.Errorf("help not available for the \"" + cmd.PrimaryArgSpec.Key + "\" command")
		}
	}

	optionIndent := 1
	argSpec := cmd.PrimaryArgSpec.String()
	if len(argSpec) > 0 {
		// named arg, might have help
		cl.helpPrintCols(0, argSpec, cmd.PrimaryArgSpec.HelpText)
	} else if len(cmd.PrimaryArgSpec.HelpText) > 0 {
		// unnamed arg with help
		cl.helpPrintln(cmd.PrimaryArgSpec.HelpText)
	} else {
		// unnamed arg without help
		optionIndent = 0
	}

	for _, option := range cmd.OptionSpecs {
		cl.helpPrintCols(optionIndent, option.String(), option.HelpText)
	}

	return nil
}

func optionSpecValues(m *map[string]*argSpec) []*argSpec {
	result := make([]*argSpec, len(*m))

	i := 0
	for _, v := range *m {
		result[i] = v
		i++
	}

	return result
}

func sortCompare(a string, b string) bool {
	alwr := strings.ToLower(a)
	blwr := strings.ToLower(b)

	if alwr == blwr {
		return a < b
	}

	return alwr < blwr
}

func (cl *CommandLine) PrintCommands(filter string, includeGlobal bool) {
	cl.printCommandsWorker(filter, includeGlobal)

	//
	// Print the queued help lines.
	//

	cl.helpRender()
}

func (cl *CommandLine) printCommandsWorker(filter string, includeGlobal bool) {

	//
	// Include global options if requested.
	//

	optPartial := false
	globalOptionsToPrint := []*globalOption{}
	if includeGlobal {
		for _, v := range cl.globalOptions {
			if cl.shouldShow(v.argSpec, nil, filter) {
				globalOptionsToPrint = append(globalOptionsToPrint, v)
			} else {
				optPartial = true
			}
		}
	}

	//
	// Identify which commands to print based on the filter.
	//

	filter = strings.ToLower(strings.TrimSpace(filter))

	cmdPartial := false
	commandsToPrint := []*command{}
	var singleCmd *command

	for _, v := range cl.commands {
		if singleCmd == nil {
			singleCmd = v
		} else {
			singleCmd = nil
		}

		osv := optionSpecValues(&v.OptionSpecs)
		if cl.shouldShow(v.PrimaryArgSpec, &osv, filter) {
			if !v.PrimaryArgSpec.Unnamed ||
				len(v.PrimaryArgSpec.HelpText) > 0 ||
				len(v.PrimaryArgSpec.ValueSpecs) > 0 ||
				len(v.OptionSpecs) > 0 {
				commandsToPrint = append(commandsToPrint, v)
			}
		} else {
			cmdPartial = true
		}
	}

	simpleDescription := (singleCmd != nil &&
		singleCmd.PrimaryArgSpec.Unnamed &&
		len(singleCmd.PrimaryArgSpec.HelpText) > 0 &&
		len(singleCmd.OptionSpecs) == 0 &&
		len(singleCmd.PrimaryArgSpec.ValueSpecs) == 0)

	//
	// Queue help lines for printing.
	//

	if len(globalOptionsToPrint) > 0 {
		if optPartial {
			cl.helpPrintln("Matching Global Options:")
		} else {
			cl.helpPrintln("Global Options:")
		}
		cl.helpPrintBlankln()

		sort.SliceStable(
			globalOptionsToPrint,
			func(i, j int) bool {
				return sortCompare(globalOptionsToPrint[i].String(), globalOptionsToPrint[j].String())
			},
		)

		for _, option := range globalOptionsToPrint {
			cl.helpPrintCols(1, option.argSpec.String(), option.argSpec.HelpText)
		}

		cl.helpPrintBlankln()
	}

	if len(commandsToPrint) > 0 {
		optionIndent := 2

		// which heading
		if cmdPartial {
			cl.helpPrintln("Matching Commands:")
		} else if len(cl.commands) > 1 {
			cl.helpPrintln("All Commands:")
		} else if simpleDescription {
			cl.helpPrintln("Description: " + singleCmd.PrimaryArgSpec.HelpText)
			optionIndent = 1
		} else {
			cl.helpPrintln("Command Options:")
			if singleCmd.PrimaryArgSpec.Unnamed {
				optionIndent = 1
			}
		}

		cl.helpPrintBlankln()

		// print each command and its options
		sort.SliceStable(
			commandsToPrint,
			func(i, j int) bool {
				return sortCompare(commandsToPrint[i].PrimaryArgSpec.String(), commandsToPrint[j].PrimaryArgSpec.String())
			},
		)

		for _, cmd := range commandsToPrint {
			if !simpleDescription {
				argText := cmd.PrimaryArgSpec.String()
				if len(argText) == 0 {
					if len(cmd.PrimaryArgSpec.HelpText) > 0 {
						cl.helpPrintln(cmd.PrimaryArgSpec.HelpText)
						cl.helpPrintBlankln()
					}
				} else {
					cl.helpPrintCols(optionIndent-1, argText, cmd.PrimaryArgSpec.HelpText)
				}
			}

			sorted := make([]*argSpec, 0, len(cmd.OptionSpecs))
			for _, option := range cmd.OptionSpecs {
				sorted = append(sorted, option)
			}
			sort.SliceStable(sorted, func(i, j int) bool { return sortCompare(sorted[i].String(), sorted[j].String()) })

			for _, option := range sorted {
				cl.helpPrintCols(optionIndent, option.String(), option.HelpText)
			}
		}

		cl.helpPrintBlankln()
	} else if len(globalOptionsToPrint) == 0 {
		hasOptions := false
		for _, cmd := range cl.commands {
			if len(cmd.OptionSpecs) > 0 || len(cmd.PrimaryArgSpec.ValueSpecs) > 0 {
				hasOptions = true
				break
			}
		}

		cl.helpPrintBlanklnFirst() // space for emphasis

		if len(filter) > 0 {
			cl.helpPrintf("No commands match help filter '%s'.", filter)
		} else if !hasOptions {
			cl.helpPrintln("This command has no options.")
		} else {
			cl.helpPrintln("No help is available.")
		}

		cl.helpPrintBlankln()
	}
}

func (cl *CommandLine) splitColon(arg string) (string, *string) {
	//
	// split an input argument at its colon, if any. Arguments that
	// have values separated by a space are not handled here.
	//
	delimiter := strings.IndexAny(arg, ":")
	if delimiter >= 0 {
		argVal := arg[delimiter+1:]
		return arg[:delimiter], &argVal
	} else {
		return arg, nil
	}
}

func (cl *CommandLine) Process(args []string) error {

	//
	// Enforce minimum requirements.
	//

	if len(cl.commands) == 0 {
		panic(fmt.Errorf("a command option is required"))
	}

	//
	// Extract all global args.
	//

	globalOptionsToRun := []*globalOptionToRun{}
	commandArgs := []string{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		globalArgSwitch, globalArgValue := cl.splitColon(arg)

		globalOpt, exists := cl.globalOptions[globalArgSwitch]
		if exists {
			gotr, argsUsed, err := cl.newGlobalOptionToRun(globalOpt, globalArgValue, args[i+1:])
			if err != nil {
				return err
			}
			i += argsUsed
			globalOptionsToRun = append(globalOptionsToRun, gotr)
		} else {
			commandArgs = append(commandArgs, arg)
		}
	}

	//
	// Execute the global options before processing the rest of the args.
	//

	for _, globalOptToRun := range globalOptionsToRun {
		err := globalOptToRun.Option.Handler(globalOptToRun.Values)
		if err != nil {
			return err
		}
	}

	//
	// Find the command to run.
	//

	args = commandArgs
	argBaseIndex := 1
	var cmd *command
	var primaryArgValue *string
	if len(args) == 0 {
		//
		// Invoke the single unnamed handler, if there is one.
		//

		cmd = cl.unnamedCmd

		if cmd == nil {
			return NewCommandLineError("A command is required")
		}

		argBaseIndex = 0
	} else if cl.unnamedCmd != nil {
		cmd = cl.unnamedCmd
		argBaseIndex = 0
	} else {
		//
		// Find the named arg.
		//

		var primaryArgSwitch string
		primaryArgSwitch, primaryArgValue = cl.splitColon(args[0])

		var exists bool
		cmd, exists = cl.commands[primaryArgSwitch]
		if !exists {
			return NewCommandLineError("Unrecognized command: " + primaryArgSwitch)
		}
	}

	cmdToRun, argsUsed, err := cl.newCommandToRun(cmd, primaryArgValue, args[argBaseIndex:])
	if err != nil {
		return err
	}

	//
	// Add options to the command.
	//

	requiredOptions := make(map[string]bool)

	for _, optionSpec := range cmd.OptionSpecs {
		if !optionSpec.Optional {
			requiredOptions[optionSpec.Key] = true
		}
	}

	for i := argBaseIndex + argsUsed; i < len(args); i++ {
		optionArgSwitch, optionArgValue := cl.splitColon(args[i])

		optionSpec, exists := cmd.OptionSpecs[optionArgSwitch]
		if !exists {
			return NewCommandLineError("Unrecognized command argument: " + optionArgSwitch)
		}

		cmdToRun.values[optionArgSwitch] = true
		argsUsed, err := optionSpec.Parse(&cmdToRun.values, optionArgValue, args[i+1:])
		if err != nil {
			return err
		}

		i += argsUsed

		_, exists = requiredOptions[optionArgSwitch]
		if exists {
			delete(requiredOptions, optionArgSwitch)
		}
	}

	if len(requiredOptions) > 0 {
		return NewCommandLineError("Arguments required: %s", sortedKeys(requiredOptions))
	}

	//
	// Put empty values in for all optional and unspecified options.
	//

	for _, optionSpec := range cmd.OptionSpecs {
		if optionSpec.Optional {
			cl.addDefaults(cmdToRun, optionSpec)
		}
	}

	cl.addDefaults(cmdToRun, cmd.PrimaryArgSpec)

	//
	// Execute the command.
	//

	return cmd.Handler(cmdToRun.values)
}

func (cl *CommandLine) addDefaults(cmdToRun *commandToRun, as *argSpec) {
	_, exists := cmdToRun.values[as.Key]
	if !exists {
		cmdToRun.values[as.Key] = false
	}

	for _, valueSpec := range as.ValueSpecs {
		_, exists = cmdToRun.values[valueSpec.OptionName]
		if !exists {
			cmdToRun.values[valueSpec.OptionName] = valueSpec.DefaultValue
		}
	}
}
