package cmdline

import (
	"fmt"
	"strings"

	"github.com/jimsnab/go-simpleutils"
)

type argValueSpec struct {
	ArgIndex     int
	OptionName   string
	Optional     bool
	Multi        bool
	DefaultValue any
}

type argSpec struct {
	CmdLine     *CommandLine
	Key         string
	Unnamed     bool
	Optional    bool
	ValuesDelim rune // the delimiter between value name and list of values
	ValueDelim  rune // the delimiter between values in a list
	ValueSpecs  []*argValueSpec
	MultiValue  bool
	HelpText    string
}

func indexOf(str string, substr string, pos int) int {
	index := strings.Index(str[pos:], substr)
	if index >= 0 {
		index += pos
	}
	return index
}

func parseError(expected string, orgSpec string, specRemaining string, parsePos int) error {
	if len(specRemaining) <= parsePos || orgSpec == specRemaining[parsePos:] {
		return fmt.Errorf("%s%s in \"%s\"", basePanic, expected, orgSpec)
	} else {
		return fmt.Errorf("%s%s at \"%s\" of \"%s\"", basePanic, expected, specRemaining[parsePos:], orgSpec)
	}
}

func (cl *CommandLine) newArgSpec(spec string, primaryArg bool) *argSpec {
	orgSpec := spec

	//
	// Syntaxes:
	//
	//      -arg
	//      -arg:<value>
	//      -arg:<value>,<value> ...
	//      -arg[:<value>]
	//      -arg:<value>[,<value>] ...
	//      [-arg]
	//      [-arg:<value>]
	//      [-arg:<value>,<value>] ...
	//      [-arg[:<value>]]
	//      [-arg:<value>[,<value>]] ...
	//      -arg [<value>]
	//      -arg [<value>] [<value>] ...
	//      -arg <value>[,<value>] ...
	//      -arg <value> [<value>] ...
	//      [-arg <value>]
	//      [-arg <value>,<value>] ...
	//      [-arg <value> <value>] ...
	//      [-arg <value> [<value>]] ...
	//
	// And the entire spec can be prefixed with asterisk (*) for a value list.
	// Example:
	//
	//      *-arg:<value>
	//
	// A value that can potentially contain commas can only appear as the only value. Example:
	//
	//      -a:<string-a> -b:<string-b>     # values a and b can have a comma
	//      -ab:<string-a>,<string-b>       # values a and b cannot have a comma
	//
	// Help text is added to the end with a question mark separator. Example:
	//
	//      [-t:<string-text>]?Specifies the text to save
	//

	as := argSpec{}
	as.CmdLine = cl

	as.HelpText = ""
	helpCutPoint := strings.LastIndex(spec, "?")
	if helpCutPoint >= 0 {
		as.HelpText = spec[helpCutPoint+1:]
		spec = spec[:helpCutPoint]
	}

	as.ValueSpecs = []*argValueSpec{}

	if strings.HasPrefix(spec, "*") {
		spec = spec[1:]
		as.MultiValue = true
	}

	if strings.HasPrefix(spec, "[") && strings.HasSuffix(spec, "]") {
		spec = spec[1 : len(spec)-1]
		as.Optional = true
	}

	argDelimiter := strings.IndexAny(spec, ": ")
	if argDelimiter < 0 {
		as.Key = spec
	} else {
		as.Key = spec[:argDelimiter]
		as.ValuesDelim = rune(spec[argDelimiter])

		suffix := simpleutils.WhichSuffix(as.Key, " [", "[")
		if suffix != nil {
			as.Key = as.Key[:argDelimiter-1]
			if *suffix == " [" {
				*suffix = "[ "
			}
			spec = *suffix + spec[argDelimiter+1:]
		} else {
			spec = spec[argDelimiter+1:]
		}

		parsePos := 0
		if len(spec) == 0 {
			panic(parseError("value spec", orgSpec, spec, parsePos))
		}

		for parsePos < len(spec) {
			avs := argValueSpec{}
			avs.Optional = false

			var optionType string

			c := spec[parsePos]
			if c == '[' {
				avs.Optional = true
				parsePos++
				c = spec[parsePos]
			} else if parsePos+1 < len(spec) && c == ' ' && spec[parsePos+1] == '[' {
				avs.Optional = true
				parsePos++
			}

			if (c == ',' || c == ' ') && len(as.ValueSpecs) > 0 {
				if as.ValueDelim == 0 {
					if c == ' ' && as.ValuesDelim == ':' {
						panic(parseError("comma-separated value spec list", orgSpec, spec, parsePos))
					}
					as.ValueDelim = rune(c)
				} else if as.ValueDelim != rune(c) {
					panic(parseError("uniform value delimiter", orgSpec, spec, parsePos))
				}
				parsePos++
				c = spec[parsePos]
			}

			if c == '*' {
				avs.Multi = true
				parsePos++
				c = spec[parsePos]
			}

			if c != '<' {
				panic(parseError("'<'", orgSpec, spec, parsePos))
			}

			parsePos++
			dashPos := indexOf(spec, "-", parsePos)
			if dashPos < 0 {
				panic(parseError("'-'", orgSpec, spec, parsePos))
			}

			optionType = spec[parsePos:dashPos]
			ot := cl.optionTypes.StringToAttributes(optionType, orgSpec)

			// defensive
			if ot == nil {
				panic(parseError("valid option type", orgSpec, spec, parsePos))
			}

			parsePos = dashPos + 1

			closeBracket := indexOf(spec, ">", parsePos)
			if closeBracket < 0 {
				panic(parseError("'>'", orgSpec, spec, parsePos))
			}

			avs.OptionName = spec[parsePos:closeBracket]
			if !simpleutils.IsTokenName(avs.OptionName) {
				panic(parseError("valid option name", orgSpec, spec, parsePos))
			}

			parsePos = closeBracket + 1

			if avs.Optional {
				if parsePos >= len(spec) || spec[parsePos] != ']' {
					panic(parseError("']'", orgSpec, spec, parsePos))
				}
				parsePos++
			}

			attribs := cl.optionTypes.StringToAttributes(optionType, orgSpec)

			avs.ArgIndex = attribs.Index
			avs.DefaultValue = attribs.DefaultValue

			// check for a dup
			for _, arg := range as.ValueSpecs {
				if avs.OptionName == arg.OptionName {
					panic(fmt.Errorf("duplicate value spec \"%s\" in \"%s\"", arg.OptionName, orgSpec))
				}
			}

			as.ValueSpecs = append(as.ValueSpecs, &avs)
		} // for parsePos
	}

	if len(as.Key) == 0 {
		panic(parseError("argument name", orgSpec, spec, 0))
	}

	if as.Key == "~" {
		as.Unnamed = true
	}

	// remove leading dash or dash-dash
	trimmedKey := strings.TrimPrefix(as.Key, "-")
	trimmedKey = strings.TrimPrefix(trimmedKey, "-")

	if !simpleutils.IsTokenNameWithMiddleChars(trimmedKey, "-") && !as.Unnamed {
		panic(parseError("a valid argument token", orgSpec, spec, 0))
	}

	if primaryArg {
		if as.Optional {
			panic(parseError("non-optional primary argument", orgSpec, spec, 0))
		}
		if as.MultiValue {
			panic(parseError("single-value primary argument", orgSpec, spec, 0))
		}
		if as.Unnamed && as.ValuesDelim == ':' {
			panic(parseError("unnamed argument without a value spec", orgSpec, spec, 0))
		}
	} else {
		if as.Unnamed {
			panic(parseError("named seconary argument", orgSpec, spec, 0))
		}
	}

	return &as
}

func (as *argSpec) storeArg(effectiveArgs *map[string]any, spec *argValueSpec, input string) error {
	if as.MultiValue || spec.Multi {
		//
		// The very first arg will exist in effectiveArgs map with nil; convert it to a list.
		// Subsequent args of the same option will be added to the list.
		//
		var err error
		var list any
		optVal := (*effectiveArgs)[spec.OptionName]
		if optVal == nil {
			list, err = as.CmdLine.optionTypes.NewList(spec.ArgIndex)

			// defensive
			if err != nil {
				return err
			}
		} else {
			list = optVal
		}

		list, err = as.CmdLine.optionTypes.AppendList(spec.ArgIndex, list, input)
		if err != nil {
			return err
		}
		(*effectiveArgs)[spec.OptionName] = list
	} else {
		value, err := as.CmdLine.optionTypes.MakeValue(spec.ArgIndex, input)
		if err != nil {
			return err
		}
		(*effectiveArgs)[spec.OptionName] = value
	}

	return nil
}

func (as *argSpec) Parse(effectiveArgs *map[string]any, colonValue *string, subsequentArgs []string) (int, error) {

	argsUsed := 0
	input := colonValue

	if input == nil && as.ValuesDelim == ' ' {
		if len(subsequentArgs) > 0 && !strings.HasPrefix(subsequentArgs[0], "-") {
			input = &subsequentArgs[0]
			argsUsed = 1
		}
	}

	if input == nil {
		if len(as.ValueSpecs) > 0 && !as.ValueSpecs[0].Optional {
			return 0, NewCommandLineError("Required value %s is missing", as.ValueSpecs[0].OptionName)
		}

		if len(as.ValueSpecs) > 0 {
			for _, valueSpec := range as.ValueSpecs {
				(*effectiveArgs)[valueSpec.OptionName] = valueSpec.DefaultValue
			}
		}
	} else if len(as.ValueSpecs) == 0 {
		return 0, NewCommandLineError("Unexpected command argument: %s", *input)
	} else if len(as.ValueSpecs) == 1 {
		err := as.storeArg(effectiveArgs, as.ValueSpecs[0], *input)
		if err != nil {
			return 0, err
		}

		if as.ValueSpecs[0].Multi && as.ValuesDelim == ' ' {
			for {
				if argsUsed >= len(subsequentArgs) || strings.HasPrefix(subsequentArgs[argsUsed], "-") {
					break
				}

				err := as.storeArg(effectiveArgs, as.ValueSpecs[0], subsequentArgs[argsUsed])
				if err != nil {
					return 0, err
				}
				argsUsed++
			}
		}
	} else {
		var values []string

		if as.ValueDelim == ',' {
			values = strings.Split(*input, ",")
		} else {
			// Because of syntax enforcement, argsUsed == 1 and *input == subsequentArgs[0]
			argsUsed = 0

			for i := 0; i < len(as.ValueSpecs); i++ {
				if argsUsed >= len(subsequentArgs) {
					break
				}
				if strings.HasPrefix(subsequentArgs[argsUsed], "-") {
					break
				}
				values = append(values, subsequentArgs[argsUsed])
				argsUsed++

				if as.ValueSpecs[i].Multi {
					for argsUsed < len(subsequentArgs) && !strings.HasPrefix(subsequentArgs[argsUsed], "-") {
						values = append(values, subsequentArgs[argsUsed])
						argsUsed++
					}
				}
			}
		}

		for i, valueSpec := range as.ValueSpecs {
			if i >= len(values) {
				if as.ValueDelim == ',' {
					// For comma-separated list, use the last value as a default when too few args are provided
					err := as.storeArg(effectiveArgs, as.ValueSpecs[i], values[len(values)-1])

					// defensive
					if err != nil {
						return 0, err
					}
				} else if valueSpec.Optional {
					break
				} else {
					return 0, NewCommandLineError("Required value %s is missing", valueSpec.OptionName)
				}
			} else {
				err := as.storeArg(effectiveArgs, as.ValueSpecs[i], values[i])
				if err != nil {
					return 0, err
				}

				if valueSpec.Multi && as.ValuesDelim == ' ' {
					for {
						if i+1 >= len(values) {
							break
						}

						value := values[i+1]
						values = append(values[0:i], values[i+2:]...)

						err := as.storeArg(effectiveArgs, as.ValueSpecs[i], value)
						if err != nil {
							return 0, err
						}
					}
				}
			}
		}
	}

	(*effectiveArgs)[as.Key] = true

	return argsUsed, nil
}

func (as *argSpec) String() string {
	var sb strings.Builder
	if as.MultiValue {
		sb.WriteString("*")
	}
	if as.Optional {
		sb.WriteString("[")
	}

	if !as.Unnamed {
		sb.WriteString(as.Key)
	}

	first := true
	optionalValues := 0

	for _, valueSpec := range as.ValueSpecs {
		chars := make([]rune, 0, 2)
		if valueSpec.Optional {
			chars = append(chars, '[')
			optionalValues++
		}
		if first {
			if !as.Unnamed {
				chars = append(chars, as.ValuesDelim) // note the 's' on ValuesDelim
			}
			first = false
		} else {
			chars = append(chars, as.ValueDelim)
		}

		s := string(chars)
		if s == "[ " {
			s = " [" // flip for readability
		}
		sb.WriteString(s)

		sb.WriteString("<")
		sb.WriteString(valueSpec.OptionName)
		sb.WriteString(">")
	}

	for optionalValues > 0 {
		optionalValues--
		sb.WriteString("]")
	}

	if as.Optional {
		sb.WriteString("]")
	}

	return sb.String()
}
