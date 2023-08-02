package cmdline

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/jimsnab/go-testutils"
)

var (
	captureStdout             = testutils.CaptureStdout
	doMapsMatch               = testutils.DoMapsMatch
	expectBool                = testutils.ExpectBool
	expectError               = testutils.ExpectError
	expectErrorContainingText = testutils.ExpectErrorContainingText
	expectPanic               = testutils.ExpectPanic
	expectPanicError          = testutils.ExpectPanicError
	expectString              = testutils.ExpectString
	expectValue               = testutils.ExpectValue
)

type testOptionTypes struct {
}

func (cl *CommandLine) summaryText() string {
	m := cl.Summary()
	text, _ := json.Marshal(m)
	return string(text)
}

func (tot *testOptionTypes) StringToAttributes(typeName string, spec string) *OptionTypeAttributes {
	if typeName == "test" {
		return &OptionTypeAttributes{Index: 0, DefaultValue: "pass"}
	} else {
		return nil
	}
}

func (tot *testOptionTypes) MakeValue(typeIndex int, inputValue string) (interface{}, error) {
	if typeIndex == 0 {
		if inputValue == "pass" || inputValue == "fail" {
			return inputValue, nil
		} else {
			return nil, fmt.Errorf("unsupported argument value \"%s\"", inputValue)
		}
	} else {
		return nil, fmt.Errorf("invalid type index %d", typeIndex)
	}
}

func (tot *testOptionTypes) NewList(typeIndex int) (interface{}, error) {
	if typeIndex == 0 {
		return []string{}, nil
	} else {
		return nil, fmt.Errorf("invalid type index %d", typeIndex)
	}
}

func (tot *testOptionTypes) AppendList(typeIndex int, list interface{}, inputValue string) (interface{}, error) {
	value, err := tot.MakeValue(typeIndex, inputValue)
	if err != nil {
		return nil, err
	}

	switch argType(typeIndex) {
	case 0:
		return append(list.([]string), value.(string)), nil

	default:
		return nil, fmt.Errorf("invalid type index %d", typeIndex)
	}
}

func TestCommandCustomOptionTypes(t *testing.T) {
	tot := OptionTypes(&testOptionTypes{})
	cl := NewCustomTypesCommandLine(tot)

	received := ""
	cl.RegisterCommand(
		func(values Values) error {
			received = values["myvar"].(string)
			return nil
		},
		"val:<test-myvar>?Tests custom option type with a command called val",
	)

	args := []string{"val:pass"}
	err := cl.Process(args)
	expectError(t, nil, err)
	expectString(t, "pass", received)

	received = ""
	args = []string{"val:fail"}
	err = cl.Process(args)
	expectError(t, nil, err)
	expectString(t, "fail", received)

	received = ""
	args = []string{"val:skipped"}
	err = cl.Process(args)
	expectError(t, fmt.Errorf("unsupported argument value \"skipped\""), err)
}

func TestCommandRequired(t *testing.T) {
	cl := NewCommandLine()

	// can't use command line without parsing rules
	args := []string{}
	expectPanic(t, func() {
		cl.Process(args)
	})

	// can't use command line with only global options
	cl.RegisterGlobalOption(
		func(values Values) error {
			return nil
		},
		"--test",
	)
	expectPanic(t, func() {
		cl.Process(args)
	})
}

func TestCommandWithProcessingContext(t *testing.T) {
	cl := NewCommandLine()

	var executed any

	cl.RegisterCommand(
		func(values Values) error {
			executed = values[""]
			return nil
		},
		"~", // unnamed single argument
	)

	args := []string{}
	err := cl.ProcessWithContext("passed thru", args)
	
	expectError(t, nil, err)
	expectString(t, "passed thru", executed.(string))
}

func TestOneGlobalFalse(t *testing.T) {
	cl := NewCommandLine()

	expectedMap := make(Values)
	expectedMap["--test"] = false
	seen := false
	executed := false

	cl.RegisterGlobalOption(
		func(values Values) error {
			seen = true
			doMapsMatch(t, expectedMap, values)
			return nil
		},
		"--test",
	)

	cl.RegisterCommand(
		func(values Values) error {
			executed = true
			return nil
		},
		"~", // unnamed single argument
	)

	args := []string{}
	err := cl.Process(args)

	expectError(t, nil, err)
	expectBool(t, false, seen)
	expectBool(t, true, executed)
}

func TestOneGlobalTrue(t *testing.T) {
	cl := NewCommandLine()

	expectedMap := make(Values)
	expectedMap["--test"] = true
	seen := false
	executed := false

	cl.RegisterGlobalOption(
		func(values Values) error {
			seen = true
			doMapsMatch(t, expectedMap, values)
			return nil
		},
		"--test",
	)

	cl.RegisterCommand(
		func(values Values) error {
			executed = true
			return nil
		},
		"~", // unnamed single argument
	)

	args := []string{"--test"}
	err := cl.Process(args)

	expectError(t, nil, err)
	expectBool(t, true, seen)
	expectBool(t, true, executed)
}

func TestUnnamedPlain(t *testing.T) {
	cl := NewCommandLine()

	executed := false
	cl.RegisterCommand(
		func(values Values) error {
			executed = true
			return nil
		},
		"~", // unnamed single argument
	)

	args := []string{}
	err := cl.Process(args)

	expectError(t, nil, err)
	expectBool(t, true, executed)
}

func TestUnnamedDuplicate(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
	)

	expectPanic(t, func() {
		cl.RegisterCommand(
			func(values Values) error {
				return nil
			},
			"~",
		)
	})
}

func TestNamedDuplicate(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test",
	)

	expectPanic(t, func() {
		cl.RegisterCommand(
			func(values Values) error {
				return nil
			},
			"test",
		)
	})
}

func TestAllowedDuplicate(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test:<string-arg>",
	)

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test2:<string-arg>",
	)

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test3",
		"-x:<string-arg>",
	)
}

func TestNamedValueDuplicate(t *testing.T) {
	cl := NewCommandLine()

	expectPanicError(
		t,
		fmt.Errorf("duplicate value spec \"dup\" in \"test:<string-dup>,<string-dup>\""),
		func() {
			cl.RegisterCommand(
				func(values Values) error {
					return nil
				},
				"test:<string-dup>,<string-dup>",
			)
		},
	)
}

func TestGlobalOptionDuplicate(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test",
	)

	expectPanic(t, func() {
		cl.RegisterGlobalOption(
			func(values Values) error {
				return nil
			},
			"test",
		)
	})

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test",
		"-x",
	)

	expectPanic(t, func() {
		cl.RegisterGlobalOption(
			func(values Values) error {
				return nil
			},
			"-x",
		)
	})

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test",
		"-x:<string-flag>",
	)

	expectPanic(t, func() {
		cl.RegisterGlobalOption(
			func(values Values) error {
				return nil
			},
			"flag",
		)
	})

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test",
		"-x:<string-opt>,<string-flag>",
	)

	expectPanic(t, func() {
		cl.RegisterGlobalOption(
			func(values Values) error {
				return nil
			},
			"flag",
		)
	})

	cl = NewCommandLine()

	cl.RegisterGlobalOption(
		func(values Values) error {
			return nil
		},
		"test",
	)

	expectPanic(t, func() {
		cl.RegisterCommand(
			func(values Values) error {
				return nil
			},
			"test",
		)
	})

	cl = NewCommandLine()

	cl.RegisterGlobalOption(
		func(values Values) error {
			return nil
		},
		"-x",
	)

	expectPanic(t, func() {
		cl.RegisterCommand(
			func(values Values) error {
				return nil
			},
			"test",
			"-x",
		)
	})

	cl = NewCommandLine()

	cl.RegisterGlobalOption(
		func(values Values) error {
			return nil
		},
		"flag",
	)

	expectPanic(t, func() {
		cl.RegisterCommand(
			func(values Values) error {
				return nil
			},
			"test",
			"-x:<string-flag>",
		)
	})

	cl = NewCommandLine()

	cl.RegisterGlobalOption(
		func(values Values) error {
			return nil
		},
		"flag",
	)

	expectPanic(t, func() {
		cl.RegisterCommand(
			func(values Values) error {
				return nil
			},
			"test",
			"-x:<string-opt>,<string-flag>",
		)
	})
}

func TestNamedSyntaxError(t *testing.T) {
	cl := NewCommandLine()

	expectPanicError(
		t,
		fmt.Errorf("command line template syntax error! expected value spec in \"test:\""),
		func() {
			cl.RegisterCommand(
				func(values Values) error {
					return nil
				},
				"test:",
			)
		},
	)

	expectPanicError(
		t,
		fmt.Errorf("command line template syntax error! expected value spec in \"test \""),
		func() {
			cl.RegisterCommand(
				func(values Values) error {
					return nil
				},
				"test ",
			)
		},
	)

	expectPanicError(
		t,
		fmt.Errorf("command line template syntax error! expected a valid argument token in \"test$\""),
		func() {
			cl.RegisterCommand(
				func(values Values) error {
					return nil
				},
				"test$",
			)
		},
	)

	expectPanicError(
		t,
		fmt.Errorf("command line template syntax error! expected a valid argument token in \"test<string-value>\""),
		func() {
			cl.RegisterCommand(
				func(values Values) error {
					return nil
				},
				"test<string-value>",
			)
		},
	)

	expectPanicError(
		t,
		fmt.Errorf("command line template syntax error! expected valid arg type strings in test:<strings-value>"),
		func() {
			cl.RegisterCommand(
				func(values Values) error {
					return nil
				},
				"test:<strings-value>",
			)
		},
	)

	expectPanicError(
		t,
		fmt.Errorf("command line template syntax error! expected valid option name at \"value$>\" of \"test:<string-value$>\""),
		func() {
			cl.RegisterCommand(
				func(values Values) error {
					return nil
				},
				"test:<string-value$>",
			)
		},
	)

	expectPanicError(
		t,
		fmt.Errorf("command line template syntax error! expected '<' at \"string-value\" of \"test:string-value\""),
		func() {
			cl.RegisterCommand(
				func(values Values) error {
					return nil
				},
				"test:string-value",
			)
		},
	)

	expectPanicError(
		t,
		fmt.Errorf("command line template syntax error! expected '-' at \"string=value>\" of \"test:<string=value>\""),
		func() {
			cl.RegisterCommand(
				func(values Values) error {
					return nil
				},
				"test:<string=value>",
			)
		},
	)

	expectPanicError(
		t,
		fmt.Errorf("command line template syntax error! expected '>' at \"value\" of \"test:<string-value\""),
		func() {
			cl.RegisterCommand(
				func(values Values) error {
					return nil
				},
				"test:<string-value",
			)
		},
	)

	expectPanicError(
		t,
		fmt.Errorf("command line template syntax error! expected ']' in \"test:[<string-value>\""),
		func() {
			cl.RegisterCommand(
				func(values Values) error {
					return nil
				},
				"test:[<string-value>",
			)
		},
	)
}

func TestUnnamedUnexpectedArg(t *testing.T) {
	cl := NewCommandLine()

	executed := false
	cl.RegisterCommand(
		func(values Values) error {
			executed = true
			return nil
		},
		"~", // unnamed single argument
	)

	args := []string{"unexpected"}
	err := cl.Process(args)

	expectError(t, NewCommandLineError("Unrecognized command argument: unexpected"), err)
	expectBool(t, false, executed)
}

func TestNamedUnexpectedArg(t *testing.T) {
	cl := NewCommandLine()

	executed := false
	cl.RegisterCommand(
		func(values Values) error {
			executed = true
			return nil
		},
		"test",
	)

	args := []string{"unexpected"}
	err := cl.Process(args)

	expectError(t, NewCommandLineError("Unrecognized command: unexpected"), err)
	expectBool(t, false, executed)
}

func TestUnnamedPlainWithValueSpec(t *testing.T) {
	cl := NewCommandLine()

	expectPanic(t, func() {
		cl.RegisterCommand(
			func(values Values) error {
				return nil
			},
			"~:<string-test>", // unnamed single argument can't have a value spec
		)
	})

	expectPanic(t, func() {
		cl.RegisterCommand(
			func(values Values) error {
				return nil
			},
			"~:*<string-test>",
		)
	})

	expectPanic(t, func() {
		cl.RegisterCommand(
			func(values Values) error {
				return nil
			},
			"~:[<string-test>]",
		)
	})

	expectPanic(t, func() {
		cl.RegisterCommand(
			func(values Values) error {
				return nil
			},
			"~:*[<string-test>]",
		)
	})
}

func TestUnnamedPlainWithTokenSpec(t *testing.T) {
	cl := NewCommandLine()

	executed := false
	value := ""
	cl.RegisterCommand(
		func(values Values) error {
			executed = true
			value = values["test"].(string)
			return nil
		},
		"~ <string-test>", // unnamed single argument can specify position-based tokens
	)

	args := []string{"expected"}
	err := cl.Process(args)

	expectError(t, nil, err)
	expectBool(t, true, executed)
	expectString(t, "expected", value)
}

func TestPrimaryCommandUnnamed(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
	)

	args := []string{}
	primary := cl.PrimaryCommand(args)

	expectString(t, "", primary)

	expectString(t, "{\"unnamed\":{\"primary\":{\"\":\"\"}}}", cl.summaryText())

	args = []string{"test"}
	primary = cl.PrimaryCommand(args)

	expectString(t, "", primary)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~ <string-arg>",
	)

	args = []string{}
	primary = cl.PrimaryCommand(args)

	expectString(t, "", primary)

	expectString(t, "{\"unnamed\":{\"primary\":{\"\\u003carg\\u003e\":\"\"}}}", cl.summaryText())

	args = []string{"test"}
	primary = cl.PrimaryCommand(args)

	expectString(t, "", primary)
}

func TestPrintCommandUnnamed(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
	)

	err := cl.PrintCommand("")
	expectError(t, fmt.Errorf("help not available for the unnamed command"), err)

	expectString(t, "{\"unnamed\":{\"primary\":{\"\":\"\"}}}", cl.summaryText())

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~?Test",
	)

	output := captureStdout(
		t,
		func() {
			err := cl.PrintCommand("")
			expectError(t, nil, err)
		},
	)

	expectString(t, "Test\n", output)

	expectString(t, "{\"unnamed\":{\"primary\":{\"\":\"Test\"}}}", cl.summaryText())

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~ <string-val>?Test",
	)

	output = captureStdout(
		t,
		func() {
			err := cl.PrintCommand("")
			expectError(t, nil, err)
		},
	)

	expectString(t, "<val>  Test\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~ <string-val> <string-val2>?Test",
	)

	output = captureStdout(
		t,
		func() {
			err := cl.PrintCommand("")
			expectError(t, nil, err)
		},
	)

	expectString(t, "<val> <val2>  Test\n", output)

	expectString(t, "{\"unnamed\":{\"primary\":{\"\\u003cval\\u003e \\u003cval2\\u003e\":\"Test\"}}}", cl.summaryText())

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~ <string-val>,<string-val2>?Test",
	)

	output = captureStdout(
		t,
		func() {
			err := cl.PrintCommand("")
			expectError(t, nil, err)
		},
	)

	expectString(t, "<val>,<val2>  Test\n", output)
}

func TestPrimaryCommandNamed(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test",
	)

	args := []string{}
	primary := cl.PrimaryCommand(args)

	expectString(t, "", primary)

	args = []string{"test"}
	primary = cl.PrimaryCommand(args)

	expectString(t, "test", primary)

	args = []string{"text"}
	primary = cl.PrimaryCommand(args)

	expectString(t, "", primary)

	expectString(t, "{\"named\":[{\"primary\":{\"test\":\"\"}}]}", cl.summaryText())
}

func TestPrintCommandNamed(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test",
	)

	err := cl.PrintCommand("test")
	expectError(t, fmt.Errorf("help not available for the \"test\" command"), err)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test?",
	)

	err = cl.PrintCommand("test")
	expectError(t, fmt.Errorf("help not available for the \"test\" command"), err)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test?Test",
	)

	output := captureStdout(
		t,
		func() {
			err := cl.PrintCommand("test")
			expectError(t, nil, err)
		},
	)

	expectString(t, "test  Test\n", output)

	expectString(t, "{\"named\":[{\"primary\":{\"test\":\"Test\"}}]}", cl.summaryText())

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test?Test",
		"--option:<bool-opt>",
	)

	output = captureStdout(
		t,
		func() {
			err := cl.PrintCommand("test")
			expectError(t, nil, err)
		},
	)

	expectString(t, "test              Test\n  --option:<opt>\n", output)

	expectString(t, "{\"named\":[{\"options\":{\"--option:\\u003copt\\u003e\":\"\"},\"primary\":{\"test\":\"Test\"}}]}", cl.summaryText())

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test?Test",
		"--option:<bool-opt>?This option has help",
	)

	output = captureStdout(
		t,
		func() {
			err := cl.PrintCommand("test")
			expectError(t, nil, err)
		},
	)

	expectString(t, "test              Test\n  --option:<opt>  This option has help\n", output)

	expectString(t, "{\"named\":[{\"options\":{\"--option:\\u003copt\\u003e\":\"This option has help\"},\"primary\":{\"test\":\"Test\"}}]}", cl.summaryText())
}

func TestPrintCommandsBase(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
	)

	output := captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "\nThis command has no options.\n\n", output)

	output = captureStdout(t, func() { cl.PrintCommands("test", true) })
	expectString(t, "\nNo commands match help filter 'test'.\n\n", output)

	cl = NewCommandLine()

	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-x")

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
	)

	output = captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "Global Options:\n\n  -x\n\n", output)

	output = captureStdout(t, func() { cl.PrintCommands("-x", true) })
	expectString(t, "Global Options:\n\n  -x\n\n", output)

	output = captureStdout(t, func() { cl.PrintCommands("x", true) })
	expectString(t, "Global Options:\n\n  -x\n\n", output)

	output = captureStdout(t, func() { cl.PrintCommands("test", true) })
	expectString(t, "\nNo commands match help filter 'test'.\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~?",
	)

	output = captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "\nThis command has no options.\n\n", output)

	output = captureStdout(t, func() { cl.PrintCommands("test", true) })
	expectString(t, "\nNo commands match help filter 'test'.\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~?Test help",
	)

	output = captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "Description: Test help\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~?Test help",
	)

	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-x")

	output = captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "Global Options:\n\n  -x\n\nDescription: Test help\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~?Test help",
	)

	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-x?An option")

	output = captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "Global Options:\n\n  -x  An option\n\nDescription: Test help\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~?Test help",
		"-x?Required option",
	)

	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-y")

	output = captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "Global Options:\n\n  -y\n\nCommand Options:\n\nTest help\n\n  -x  Required option\n\n", output)
}

func TestPrintCommandsGlobalOptions(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
	)

	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-x")
	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-y")

	output := captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "Global Options:\n\n  -x\n  -y\n\n", output)

	cl = NewCommandLine()

	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-x?With help")
	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-y?")

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
	)

	output = captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "Global Options:\n\n  -x  With help\n  -y\n\n", output)

	cl = NewCommandLine()

	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-x?With help")
	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-y?Also with help")
	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-z")
	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-Z")

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
	)

	output = captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "Global Options:\n\n  -x  With help\n  -y  Also with help\n  -Z\n  -z\n\n", output)

	cl = NewCommandLine()

	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-Z")
	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-x")
	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-y")
	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-z")

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
	)

	output = captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "Global Options:\n\n  -x\n  -y\n  -Z\n  -z\n\n", output)

	cl = NewCommandLine()

	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-z")
	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-x")
	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-y")
	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-Z")

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
	)

	output = captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "Global Options:\n\n  -x\n  -y\n  -Z\n  -z\n\n", output)

	output = captureStdout(t, func() { cl.PrintCommands("z", true) })
	expectString(t, "Matching Global Options:\n\n  -z\n\n", output)

	output = captureStdout(t, func() { cl.PrintCommands("Z", true) })
	expectString(t, "Matching Global Options:\n\n  -Z\n\n", output)

	cl = NewCommandLine()

	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-cat")
	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-dog?Test of a filter")
	cl.RegisterGlobalOption(func(values Values) error { return nil }, "-cow?Not a cat")

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
	)

	output = captureStdout(t, func() { cl.PrintCommands("cat", true) })
	expectString(t, "Matching Global Options:\n\n  -cat\n  -cow  Not a cat\n\n", output)

	output = captureStdout(t, func() { cl.PrintCommands("-cat", true) })
	expectString(t, "Matching Global Options:\n\n  -cat\n\n", output)

	output = captureStdout(t, func() { cl.PrintCommands("filter", true) })
	expectString(t, "Matching Global Options:\n\n  -dog  Test of a filter\n\n", output)
}

func TestPrintCommandCantFind(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"named",
	)

	err := cl.PrintCommand("")
	expectError(t, fmt.Errorf("unnamed command not found"), err)

	err = cl.PrintCommand("~")
	expectError(t, fmt.Errorf("unnamed command not found"), err)

	err = cl.PrintCommand("other")
	expectError(t, fmt.Errorf("command \"other\" not found"), err)
}

func TestPrintCommandsTwoCommands(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(func(values Values) error { return nil }, "first?1")
	cl.RegisterCommand(func(values Values) error { return nil }, "second?2")

	output := captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "All Commands:\n\n  first   1\n  second  2\n\n", output)

	output = captureStdout(t, func() { cl.PrintCommands("1", true) })
	expectString(t, "Matching Commands:\n\n  first  1\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(func(values Values) error { return nil }, "first?1", "-x?option")
	cl.RegisterCommand(func(values Values) error { return nil }, "second?2")

	output = captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "All Commands:\n\n  first   1\n    -x    option\n  second  2\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(func(values Values) error { return nil }, "first?1", "-x?option 1", "-abc?option 2")
	cl.RegisterCommand(func(values Values) error { return nil }, "second?2")

	output = captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "All Commands:\n\n  first   1\n    -x    option 1\n    -abc  option 2\n  second  2\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(func(values Values) error { return nil }, "first?1", "-longoptionname?test")
	cl.RegisterCommand(func(values Values) error { return nil }, "second?2")

	output = captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "All Commands:\n\n  first              1\n    -longoptionname  test\n  second             2\n\n", output)

	output = captureStdout(t, func() { cl.PrintCommands("long", true) })
	expectString(t, "Matching Commands:\n\n  first              1\n    -longoptionname  test\n\n", output)

	output = captureStdout(t, func() { cl.PrintCommands("EST", true) })
	expectString(t, "Matching Commands:\n\n  first              1\n    -longoptionname  test\n\n", output)
}

func TestIndent(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(func(values Values) error { return nil }, "longcommandname:<string-begin>,<string-end>,<string-maxcount>?Long command example")

	output := captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "Command Options:\n\n  longcommandname:<begin>,<end>,<maxcount>\n                              Long command example\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(func(values Values) error { return nil }, "mycmd?This is an example help message that requires word wrap because of its long length. The test must pass and should not fail.")

	output = captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "Command Options:\n\n  mycmd  This is an example help message that requires word wrap because of its long length. The test must pass and\n         should not fail.\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error { return nil },
		"longcommandname:<string-begin>,<string-end>,<string-maxcount>?This is an example help message that requires word wrap because of its long length. The test must pass and should not fail.",
	)

	output = captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "Command Options:\n\n  longcommandname:<begin>,<end>,<maxcount>\n                              This is an example help message that requires word wrap because of its long length. The\n                              test must pass and should not fail.\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(func(values Values) error { return nil }, "mycmd?Help is allowed\n\nmultiple\n\nlines.")

	output = captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "Command Options:\n\n  mycmd  Help is allowed\n\n         multiple\n\n         lines.\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(func(values Values) error { return nil }, "mycmd?Help is allowed\n  \nmultiple\n  \nlines.")

	output = captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "Command Options:\n\n  mycmd  Help is allowed\n\n         multiple\n\n         lines.\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(func(values Values) error { return nil }, "mycmd?Help text can include a very long description if necessary. It is a common case. The description can be listed     on\n  \nmultiple\n  \nlines.")

	output = captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "Command Options:\n\n  mycmd  Help text can include a very long description if necessary. It is a common case. The description can be listed\n         on\n\n         multiple\n\n         lines.\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(func(values Values) error { return nil }, "mycmd?\nHelp can be forced down a line.")

	output = captureStdout(t, func() { cl.PrintCommands("", true) })
	expectString(t, "Command Options:\n\n  mycmd\n         Help can be forced down a line.\n\n", output)

}

func TestUnexpectedArg(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(func(values Values) error { return nil }, "test")

	args := []string{"test", "something"}
	err := cl.Process(args)

	expectError(t, NewCommandLineError("Unrecognized command argument: something"), err)

	cl = NewCommandLine()

	cl.RegisterCommand(func(values Values) error { return nil }, "~")
	cl.RegisterGlobalOption(func(values Values) error { return nil }, "--opt")

	args = []string{"--opt:something"}
	err = cl.Process(args)
	expectErrorContainingText(t, "something", err)

	args = []string{"--opt", "something"}
	err = cl.Process(args)
	expectErrorContainingText(t, "something", err)
}

func TestMissingUnnamedCommand(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(func(values Values) error { return nil }, "test")

	args := []string{}
	err := cl.Process(args)

	expectError(t, NewCommandLineError("A command is required"), err)
}

func TestMissingOptionalValue(t *testing.T) {
	cl := NewCommandLine()

	hasFlag := true

	cl.RegisterCommand(
		func(values Values) error {
			hasFlag = values["--flag"].(bool)
			return nil
		},
		"test",
		"[--flag]",
	)

	args := []string{"test"}
	err := cl.Process(args)
	expectError(t, nil, err)
	expectBool(t, false, hasFlag)

	expectString(t,  "{\"named\":[{\"options\":{\"[--flag]\":\"\"},\"primary\":{\"test\":\"\"}}]}", cl.summaryText())
}

func TestMissingRequiredValue(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test",
		"--flag",
	)

	args := []string{"test"}
	err := cl.Process(args)
	expectError(t, NewCommandLineError("Arguments required: [--flag]"), err)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test",
		"--flag1",
		"--flag2",
	)

	args = []string{"test"}
	err = cl.Process(args)
	expectError(t, NewCommandLineError("Arguments required: [--flag1 --flag2]"), err)
}

func TestDefaultedValue(t *testing.T) {
	cl := NewCommandLine()

	hasFlag1 := false
	hasFlag2 := false
	v1 := true
	v2 := true

	cl.RegisterCommand(
		func(values Values) error {
			v1, hasFlag1 = values["v1"].(bool)
			v2, hasFlag2 = values["v2"].(bool)
			return nil
		},
		"test",
		"-x[:<bool-v1>][,<bool-v2>]",
	)

	args := []string{"test", "-x"}
	err := cl.Process(args)
	expectError(t, nil, err)
	expectBool(t, true, hasFlag1)
	expectBool(t, false, v1)
	expectBool(t, true, hasFlag2)
	expectBool(t, false, v2)

	expectString(t, "{\"named\":[{\"options\":{\"-x[:\\u003cv1\\u003e[,\\u003cv2\\u003e]]\":\"\"},\"primary\":{\"test\":\"\"}}]}", cl.summaryText())

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			v1, hasFlag1 = values["v1"].(bool)
			v2, hasFlag2 = values["v2"].(bool)
			return nil
		},
		"test",
		"[-x[:<bool-v1>][,<bool-v2>]]",
	)

	hasFlag1 = false
	hasFlag2 = false
	v1 = true
	v2 = true
	args = []string{"test"}
	err = cl.Process(args)
	expectError(t, nil, err)
	expectBool(t, true, hasFlag1)
	expectBool(t, false, v1)
	expectBool(t, true, hasFlag2)
	expectBool(t, false, v2)

	hasFlag1 = false
	v1 = true
	hasFlag2 = false
	v2 = true
	args = []string{"test", "-x:false"}
	err = cl.Process(args)
	expectError(t, nil, err)
	expectBool(t, true, hasFlag1)
	expectBool(t, false, v1)
	expectBool(t, true, hasFlag2) // value of v1 becomes value of v2 when not specified
	expectBool(t, false, v2)

	hasFlag1 = false
	v1 = true
	hasFlag2 = false
	v2 = true
	args = []string{"test", "-x:false,false"}
	err = cl.Process(args)
	expectError(t, nil, err)
	expectBool(t, true, hasFlag1)
	expectBool(t, false, v1)
	expectBool(t, true, hasFlag2)
	expectBool(t, false, v2)
}

func TestParseError(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test",
		"-x:<bool-v1>",
	)

	args := []string{"test", "-x:invalid"}
	err := cl.Process(args)
	_, numError := strconv.ParseBool("invalid")
	expectError(t, numError, err)
}

func TestHandlerError(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return fmt.Errorf("command handler error")
		},
		"test",
	)

	cl.RegisterGlobalOption(
		func(values Values) error {
			return fmt.Errorf("option handler error")
		},
		"--fail",
	)

	args := []string{"test"}
	err := cl.Process(args)
	expectError(t, fmt.Errorf("command handler error"), err)

	args = []string{"test", "--fail"}
	err = cl.Process(args)
	expectError(t, fmt.Errorf("option handler error"), err)
}

func TestMultiValueBool(t *testing.T) {
	cl := NewCommandLine()

	flags := []bool{}

	cl.RegisterCommand(
		func(values Values) error {
			flags = values["tflag"].([]bool)
			return nil
		},
		"~",
		"*[-t:<bool-tflag>]",
	)

	args := []string{"-t:true", "-t:false"}
	err := cl.Process(args)
	expectError(t, nil, err)

	if len(flags) != 2 {
		t.Errorf("Unexpected multi-arg length: %d", len(flags))
	}

	expectBool(t, true, flags[0])
	expectBool(t, false, flags[1])
}

func TestMultiValueString(t *testing.T) {
	cl := NewCommandLine()

	flags := []string{}

	cl.RegisterCommand(
		func(values Values) error {
			flags = values["tflag"].([]string)
			return nil
		},
		"~",
		"*[-t:<string-tflag>]",
	)

	args := []string{"-t:one", "-t:two"}
	err := cl.Process(args)
	expectError(t, nil, err)

	if len(flags) != 2 {
		t.Errorf("Unexpected multi-arg length: %d", len(flags))
	}

	expectString(t, "one", flags[0])
	expectString(t, "two", flags[1])

	expectString(t, "{\"unnamed\":{\"options\":{\"*[-t:\\u003ctflag\\u003e]\":\"\"},\"primary\":{\"\":\"\"}}}", cl.summaryText())
}

func TestMultiValueInt(t *testing.T) {
	cl := NewCommandLine()

	flags := []int{}

	cl.RegisterCommand(
		func(values Values) error {
			flags = values["tflag"].([]int)
			return nil
		},
		"~",
		"*[-t:<int-tflag>]",
	)

	args := []string{"-t:1", "-t:2"}
	err := cl.Process(args)
	expectError(t, nil, err)

	if len(flags) != 2 {
		t.Errorf("Unexpected multi-arg length: %d", len(flags))
	}

	expectValue(t, 1, flags[0])
	expectValue(t, 2, flags[1])
}

func TestMultiValueFloat(t *testing.T) {
	cl := NewCommandLine()

	flags := []float64{}

	cl.RegisterCommand(
		func(values Values) error {
			flags = values["tflag"].([]float64)
			return nil
		},
		"~",
		"*[-t:<float64-tflag>]",
	)

	args := []string{"-t:1.5", "-t:2.5"}
	err := cl.Process(args)
	expectError(t, nil, err)

	if len(flags) != 2 {
		t.Errorf("Unexpected multi-arg length: %d", len(flags))
	}

	expectValue(t, 1.5, flags[0])
	expectValue(t, 2.5, flags[1])
}

func TestMultiValuePath(t *testing.T) {
	cl := NewCommandLine()

	flags := []string{}

	cl.RegisterCommand(
		func(values Values) error {
			flags = values["tflag"].([]string)
			return nil
		},
		"~",
		"*[-t:<path-tflag>]",
	)

	args := []string{"-t:.", "-t:./testpath"}
	err := cl.Process(args)
	expectError(t, nil, err)

	if len(flags) != 2 {
		t.Errorf("Unexpected multi-arg length: %d", len(flags))
	}

	dir, err := os.Getwd()
	if err != nil {
		t.Error(err)
		return
	}

	expectValue(t, dir, flags[0])
	expectValue(t, path.Join(dir, "testpath"), flags[1])
}

func TestMultiValueInvalid(t *testing.T) {
	cl := NewCommandLine()

	expectPanic(t, func() {
		cl.RegisterCommand(
			func(values Values) error {
				return nil
			},
			"~",
			"*[-t:<invalid-tflag>]",
		)
	})

	cl = NewCommandLine()

	// bool
	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
		"*[-t:<bool-tflag>]",
	)

	args := []string{"-t:one", "-t:cat"}
	err := cl.Process(args)

	_, boolErr := strconv.ParseBool("one")
	expectError(t, boolErr, err)

	// int
	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
		"*[-t:<int-tflag>]",
	)

	args = []string{"-t:one", "-t:cat"}
	err = cl.Process(args)

	_, intErr := strconv.Atoi("one")
	expectError(t, intErr, err)

	// float64
	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
		"*[-t:<float64-tflag>]",
	)

	args = []string{"-t:one", "-t:cat"}
	err = cl.Process(args)

	_, floatErr := strconv.ParseFloat("one", 64)
	expectError(t, floatErr, err)
}

func TestNonUniformValueDelimiter(t *testing.T) {
	cl := NewCommandLine()

	expectPanic(t, func() {
		cl.RegisterCommand(
			func(values Values) error {
				return nil
			},
			"~",
			"-t:<string-flag1>,<string-flag2> <string-flag3>",
		)
	})

	expectPanic(t, func() {
		cl.RegisterCommand(
			func(values Values) error {
				return nil
			},
			"~",
			"-t:<string-flag1> <string-flag2>,<string-flag3>",
		)
	})
}

func TestInvalidArgName(t *testing.T) {
	cl := NewCommandLine()

	expectPanic(t, func() {
		cl.RegisterCommand(
			func(values Values) error {
				return nil
			},
			"~",
			":<string-flag>",
		)
	})

	expectPanic(t, func() {
		cl.RegisterCommand(
			func(values Values) error {
				return nil
			},
			"~",
			"-t:<string->",
		)
	})
}

func TestInvalidCmdName(t *testing.T) {
	cl := NewCommandLine()

	expectPanic(t, func() {
		cl.RegisterCommand(
			func(values Values) error {
				return nil
			},
			"",
		)
	})

	expectPanic(t, func() {
		cl.RegisterCommand(
			func(values Values) error {
				return nil
			},
			"[cmd]",
		)
	})

	expectPanic(t, func() {
		cl.RegisterCommand(
			func(values Values) error {
				return nil
			},
			"*cmd",
		)
	})

	expectPanic(t, func() {
		cl.RegisterCommand(
			func(values Values) error {
				return nil
			},
			"~ <string:~>",
		)
	})

	expectPanic(t, func() {
		cl.RegisterCommand(
			func(values Values) error {
				return nil
			},
			"~",
			"~",
		)
	})
}

func TestMissingRequiredArg(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test:<string-flag>",
	)

	args := []string{"test"}
	err := cl.Process(args)

	expectError(t, NewCommandLineError("Required value flag is missing"), err)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test <string-flag1> <string-flag2>",
	)

	args = []string{"test"}
	err = cl.Process(args)

	expectError(t, NewCommandLineError("Required value flag1 is missing"), err)
}

func TestUnexpectedCmdArg(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test",
		"-t",
	)

	args := []string{"test", "-t:data"}
	err := cl.Process(args)

	expectError(t, NewCommandLineError("Unexpected command argument: data"), err)
}

func TestTemplateMistakes(t *testing.T) {
	cl := NewCommandLine()

	expectPanic(
		t,
		func() {
			cl.RegisterCommand(func(values Values) error { return nil }, "named <string-named>")
		},
	)

	expectPanic(
		t,
		func() {
			cl.RegisterCommand(func(values Values) error { return nil }, "named", "-t:<string-named>")
		},
	)

	expectPanic(
		t,
		func() {
			cl.RegisterCommand(func(values Values) error { return nil }, "~ <string-named>", "-t:<string-named>")
		},
	)

	expectPanic(
		t,
		func() {
			cl.RegisterCommand(func(values Values) error { return nil })
		},
	)
}

func TestArgsWithSpace(t *testing.T) {
	cl := NewCommandLine()

	flag1 := ""
	flag2 := ""

	cl.RegisterCommand(
		func(values Values) error {
			flag1 = values["flag1"].(string)
			flag2 = values["flag2"].(string)
			return nil
		},
		"test <string-flag1> <string-flag2>",
	)

	args := []string{"test", "first", "second"}
	err := cl.Process(args)
	expectError(t, nil, err)
	expectString(t, "first", flag1)
	expectString(t, "second", flag2)

	flag1 = ""
	flag2 = ""
	args = []string{"test", "first"}
	err = cl.Process(args)
	expectError(t, NewCommandLineError("Required value flag2 is missing"), err)

	flag1 = ""
	flag2 = ""
	args = []string{"test", "first", "--second"}
	err = cl.Process(args)
	expectError(t, NewCommandLineError("Required value flag2 is missing"), err)
}

func TestArgsWithSpaceBools(t *testing.T) {
	cl := NewCommandLine()

	flag1 := true
	flag2 := true

	cl.RegisterCommand(
		func(values Values) error {
			flag1 = values["flag1"].(bool)
			flag2 = values["flag2"].(bool)
			return nil
		},
		"test <bool-flag1> <bool-flag2>",
	)

	args := []string{"test", "false", "false"}
	err := cl.Process(args)
	expectError(t, nil, err)
	expectBool(t, false, flag1)
	expectBool(t, false, flag2)

	flag1 = true
	flag2 = true
	args = []string{"test", "invalid"}
	err = cl.Process(args)
	_, invalidBool := strconv.ParseBool("invalid")
	expectError(t, invalidBool, err)
}

func TestArgsWithSpaceOptional(t *testing.T) {
	cl := NewCommandLine()

	hasFlag1 := false
	hasFlag2 := false
	flag1 := true
	flag2 := true

	cl.RegisterCommand(
		func(values Values) error {
			flag1, hasFlag1 = values["flag1"].(bool)
			flag2, hasFlag2 = values["flag2"].(bool)
			return nil
		},
		"test <bool-flag1> [<bool-flag2>]",
	)

	args := []string{"test", "false"}
	err := cl.Process(args)
	expectError(t, nil, err)
	expectBool(t, true, hasFlag1)
	expectBool(t, false, flag1)
	expectBool(t, true, hasFlag2)
	expectBool(t, false, flag2)

	hasFlag1 = false
	hasFlag2 = false
	flag1 = true
	flag2 = true
	args = []string{"test", "false", "false"}
	err = cl.Process(args)
	expectError(t, nil, err)
	expectBool(t, true, hasFlag1)
	expectBool(t, false, flag1)
	expectBool(t, true, hasFlag2)
	expectBool(t, false, flag2)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			flag1, hasFlag1 = values["flag1"].(bool)
			flag2, hasFlag2 = values["flag2"].(bool)
			return nil
		},
		"test [<bool-flag1>] [<bool-flag2>]",
	)

	hasFlag1 = false
	hasFlag2 = false
	flag1 = false
	flag2 = false
	args = []string{"test"}
	err = cl.Process(args)
	expectError(t, nil, err)
	expectBool(t, true, hasFlag1)
	expectBool(t, false, flag1)
	expectBool(t, true, hasFlag2)
	expectBool(t, false, flag2)

	hasFlag1 = false
	hasFlag2 = false
	flag1 = true
	flag2 = true
	args = []string{"test", "false"}
	err = cl.Process(args)
	expectError(t, nil, err)
	expectBool(t, true, hasFlag1)
	expectBool(t, false, flag1)
	expectBool(t, true, hasFlag2)
	expectBool(t, false, flag2)

	hasFlag1 = false
	hasFlag2 = false
	flag1 = true
	flag2 = true
	args = []string{"test", "false", "false"}
	err = cl.Process(args)
	expectError(t, nil, err)
	expectBool(t, true, hasFlag1)
	expectBool(t, false, flag1)
	expectBool(t, true, hasFlag2)
	expectBool(t, false, flag2)
}

func TestArgsWithComma(t *testing.T) {
	cl := NewCommandLine()

	flag1 := ""
	flag2 := ""

	cl.RegisterCommand(
		func(values Values) error {
			flag1 = values["flag1"].(string)
			flag2 = values["flag2"].(string)
			return nil
		},
		"test <string-flag1>,<string-flag2>",
	)

	args := []string{"test", "first,second"}
	err := cl.Process(args)
	expectError(t, nil, err)
	expectString(t, "first", flag1)
	expectString(t, "second", flag2)

	flag1 = ""
	flag2 = ""
	args = []string{"test", "first"}
	err = cl.Process(args)
	expectError(t, nil, err)
	expectString(t, "first", flag1)
	expectString(t, "first", flag2)
}

func TestPrintCommandMulti(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
		"*-t <string-val>",
	)

	output := captureStdout(
		t,
		func() {
			err := cl.PrintCommand("")
			expectError(t, nil, err)
		},
	)

	expectString(t, "*-t <val>\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
		"*-t <string-val>?Test option",
	)

	output = captureStdout(
		t,
		func() {
			err := cl.PrintCommand("")
			expectError(t, nil, err)
		},
	)

	expectString(t, "*-t <val>  Test option\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~?The command",
		"*-t <string-val>?Test option",
	)

	output = captureStdout(
		t,
		func() {
			err := cl.PrintCommand("")
			expectError(t, nil, err)
		},
	)

	expectString(t, "The command\n  *-t <val>  Test option\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
		"[-t <string-val>]?Test option",
	)

	output = captureStdout(
		t,
		func() {
			err := cl.PrintCommand("")
			expectError(t, nil, err)
		},
	)

	expectString(t, "[-t <val>]  Test option\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
		"*[-t <string-val>]?Test option",
	)

	output = captureStdout(
		t,
		func() {
			err := cl.PrintCommand("")
			expectError(t, nil, err)
		},
	)

	expectString(t, "*[-t <val>]  Test option\n", output)
}

func TestPrintCommandsMulti(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
		"*-t <string-val>",
	)

	output := captureStdout(
		t,
		func() {
			cl.PrintCommands("", true)
		},
	)

	expectString(t, "Command Options:\n\n  *-t <val>\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
		"*-t <string-val>?Test option",
	)

	output = captureStdout(
		t,
		func() {
			cl.PrintCommands("", true)
		},
	)

	expectString(t, "Command Options:\n\n  *-t <val>  Test option\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~?The command",
		"*-t <string-val>?Test option",
	)

	output = captureStdout(
		t,
		func() {
			cl.PrintCommands("", true)
		},
	)

	expectString(t, "Command Options:\n\nThe command\n\n  *-t <val>  Test option\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
		"[-t <string-val>]?Test option",
	)

	output = captureStdout(
		t,
		func() {
			cl.PrintCommands("", true)
		},
	)

	expectString(t, "Command Options:\n\n  [-t <val>]  Test option\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
		"*[-t <string-val>]?Test option",
	)

	output = captureStdout(
		t,
		func() {
			cl.PrintCommands("", true)
		},
	)

	expectString(t, "Command Options:\n\n  *[-t <val>]  Test option\n\n", output)
}

func TestPrintCommandsNamedMulti(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test",
		"*-t <string-val>",
	)

	output := captureStdout(
		t,
		func() {
			cl.PrintCommands("", true)
		},
	)

	expectString(t, "Command Options:\n\n  test\n    *-t <val>\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test",
		"*-t <string-val>?Test option",
	)

	output = captureStdout(
		t,
		func() {
			cl.PrintCommands("", true)
		},
	)

	expectString(t, "Command Options:\n\n  test\n    *-t <val>  Test option\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test?The command",
		"*-t <string-val>?Test option",
	)

	output = captureStdout(
		t,
		func() {
			cl.PrintCommands("", true)
		},
	)

	expectString(t, "Command Options:\n\n  test         The command\n    *-t <val>  Test option\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test",
		"[-t <string-val>]?Test option",
	)

	output = captureStdout(
		t,
		func() {
			cl.PrintCommands("", true)
		},
	)

	expectString(t, "Command Options:\n\n  test\n    [-t <val>]  Test option\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test",
		"*[-t <string-val>]?Test option",
	)

	output = captureStdout(
		t,
		func() {
			cl.PrintCommands("", true)
		},
	)

	expectString(t, "Command Options:\n\n  test\n    *[-t <val>]  Test option\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test",
		"*[-t [<string-val>]]?Test option",
	)

	output = captureStdout(
		t,
		func() {
			cl.PrintCommands("", true)
		},
	)

	expectString(t, "Command Options:\n\n  test\n    *[-t [<val>]]  Test option\n\n", output)
}

func TestPrintHelp(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test",
		"*-t <string-val>",
	)

	output := captureStdout(
		t,
		func() {
			args := []string{}
			err := cl.Process(args)
			cl.Help(err, "unit-test", args)
		},
	)

	expectString(t, "Usage: unit-test <command> <options>\n\nCommand Options:\n\n  test\n    *-t <val>\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test?This is help for the test option",
		"*-t <string-val>",
	)

	output = captureStdout(
		t,
		func() {
			args := []string{}
			err := cl.Process(args)
			cl.Help(err, "unit-test", args)
		},
	)

	expectString(t, "Usage: unit-test <command> <options>\n\nCommand Options:\n\n  test         This is help for the test option\n    *-t <val>\n\n", output)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test?This is help",
		"-t1 <string-value1>",
		"-t2 <string-value2>",
		"-t3 <string-value3>",
		"-t4 <string-value4>",
		"-t5 <string-value5>",
		"-t6 <string-value6>",
		"-t7 <string-value7>",
		"-t8 <string-value8>",
		"-t9 <string-value9>",
		"-t10 <string-value10>",
		"-t11 <string-value11>",
		"-t12 <string-value12>",
		"-t13 <string-value13>",
	)

	output = captureStdout(
		t,
		func() {
			args := []string{}
			err := cl.Process(args)
			cl.Help(err, "unit-test", args)
		},
	)

	expectString(
		t,
		"Usage: unit-test <command> <options>\n\nCommand Options:\n\n  test              "+
			"This is help\n    -t1 <value1>\n    -t2 <value2>\n    -t3 <value3>\n    -t4 <value4>\n"+
			"    -t5 <value5>\n    -t6 <value6>\n    -t7 <value7>\n    -t8 <value8>\n    -t9 <value9>\n"+
			"    -t10 <value10>\n    -t11 <value11>\n    -t12 <value12>\n    -t13 <value13>\n\n"+
			"Search help with unit-test --help <filter text>. Example: unit-test --help test\n"+
			"Or, put a question mark on the end. Example: unit-test test?\n\n",
		output,
	)

	cl = NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
		"-t1 <string-value1>",
		"-t2 <string-value2>",
		"-t3 <string-value3>",
		"-t4 <string-value4>",
		"-t5 <string-value5>",
		"-t6 <string-value6>",
		"-t7 <string-value7>",
		"-t8 <string-value8>",
		"-t9 <string-value9>",
		"-t10 <string-value10>",
		"-t11 <string-value11>",
		"-t12 <string-value12>",
		"-t13 <string-value13>",
	)

	output = captureStdout(
		t,
		func() {
			args := []string{}
			err := cl.Process(args)
			cl.Help(err, "unit-test", args)
		},
	)

	expectString(
		t,
		"Usage: unit-test <options>\n\nCommand Options:\n\n  -t1 <value1>\n  -t2 <value2>\n  -t3 <value3>\n"+
			"  -t4 <value4>\n  -t5 <value5>\n  -t6 <value6>\n  -t7 <value7>\n  -t8 <value8>\n  -t9 <value9>\n"+
			"  -t10 <value10>\n  -t11 <value11>\n  -t12 <value12>\n  -t13 <value13>\n\n"+
			"Search help with: unit-test --help <filter text>\n\n",
		output,
	)
}

func TestPrintHelpError(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test",
	)

	output := captureStdout(
		t,
		func() {
			args := []string{"invalid"}
			err := cl.Process(args)
			cl.Help(err, "unit-test", args)
		},
	)

	expectString(t, "Usage: unit-test <command>\n\nCommand Options:\n\n  test\n\n", output)

	output = captureStdout(
		t,
		func() {
			args := []string{"invalid"}
			cl.Help(fmt.Errorf("test"), "unit-test", args)
		},
	)

	expectString(t, "\ntest\n\n", output)
}

func TestPrintHelpSwitch(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test?Give me help",
	)

	output := captureStdout(
		t,
		func() {
			args := []string{"help"}
			cl.Help(nil, "unit-test", args)
		},
	)

	expectString(t, "Command Options:\n\n  test  Give me help\n\n", output)

	output = captureStdout(
		t,
		func() {
			args := []string{"--help"}
			cl.Help(nil, "unit-test", args)
		},
	)

	expectString(t, "Command Options:\n\n  test  Give me help\n\n", output)
}

func TestPrintHelpSwitchTwoCommands(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test?Give me help",
	)

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"dog?Fido is his name",
	)

	output := captureStdout(
		t,
		func() {
			args := []string{"help", "dog"}
			cl.Help(nil, "unit-test", args)
		},
	)

	expectString(t, "Matching Commands:\n\n  dog  Fido is his name\n\n", output)

	output = captureStdout(
		t,
		func() {
			args := []string{"--help", "test"}
			cl.Help(nil, "unit-test", args)
		},
	)

	expectString(t, "Matching Commands:\n\n  test  Give me help\n\n", output)

	output = captureStdout(
		t,
		func() {
			args := []string{"invalid"}
			cl.Help(nil, "unit-test", args)
		},
	)

	expectString(t, "Usage: unit-test <command>\n\nAll Commands:\n\n  dog   Fido is his name\n  test  Give me help\n\n", output)

	output = captureStdout(
		t,
		func() {
			args := []string{"test?"}
			cl.Help(nil, "unit-test", args)
		},
	)

	expectString(t, "Matching Commands:\n\n  test  Give me help\n\n", output)

	output = captureStdout(
		t,
		func() {
			args := []string{"-"}
			cl.Help(nil, "unit-test", args)
		},
	)

	expectString(t, "Usage: unit-test <command>\n\nAll Commands:\n\n  dog   Fido is his name\n  test  Give me help\n\n", output)

	output = captureStdout(
		t,
		func() {
			args := []string{"--"}
			cl.Help(nil, "unit-test", args)
		},
	)

	expectString(t, "Usage: unit-test <command>\n\nAll Commands:\n\n  dog   Fido is his name\n  test  Give me help\n\n", output)

	output = captureStdout(
		t,
		func() {
			args := []string{"-?"}
			cl.Help(nil, "unit-test", args)
		},
	)

	expectString(t, "All Commands:\n\n  dog   Fido is his name\n  test  Give me help\n\n", output)
}

func TestPrintHelpSyntaxError(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test:<bool-flag>?Give me help",
	)

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"cat:<int-num>?Morris is his name",
	)

	output := captureStdout(
		t,
		func() {
			args := []string{"test:1"}
			err := cl.Process(args)
			cl.Help(err, "unit-test", args)
		},
	)

	expectString(t, "\nSyntax error.\n\nCommand Help:\n\ntest:<flag>  Give me help\n\n", output)

	output = captureStdout(
		t,
		func() {
			args := []string{"cat:false"}
			err := cl.Process(args)
			cl.Help(err, "unit-test", args)
		},
	)

	expectString(t, "\nstrconv.Atoi: parsing \"false\": invalid syntax\n\n", output)
}

func TestPrintHelpGlobalOptions(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test:<bool-flag>?Give me help",
	)

	cl.RegisterGlobalOption(
		func(values Values) error {
			return nil
		},
		"cat:<int-num>?Morris is his name",
	)

	output := captureStdout(
		t,
		func() {
			args := []string{""}
			err := cl.Process(args)
			cl.Help(err, "unit-test", args)
		},
	)

	expectString(
		t,
		"Usage: unit-test <options> <command>\n\n"+
			"Global Options:\n\n"+
			"  cat:<num>    Morris is his name\n\n"+
			"Command Options:\n\n"+
			"  test:<flag>  Give me help\n\n",
		output,
	)

	cl.RegisterGlobalOption(
		func(values Values) error {
			return nil
		},
		"dog?Fido is his name",
	)

	output = captureStdout(
		t,
		func() {
			args := []string{}
			err := cl.Process(args)
			cl.Help(err, "unit-test", args)
		},
	)

	expectString(
		t,
		"Usage: unit-test <options> <command>\n\n"+
			"Global Options:\n\n"+
			"  cat:<num>    Morris is his name\n"+
			"  dog          Fido is his name\n\n"+
			"Command Options:\n\n"+
			"  test:<flag>  Give me help\n\n",
		output,
	)
}

func TestPrintHelpGlobalOptionsMultipleCommands(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"test:<bool-flag>?Give me help",
	)

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"run?A second command",
	)

	cl.RegisterGlobalOption(
		func(values Values) error {
			return nil
		},
		"cat:<int-num>?Morris is his name",
	)

	output := captureStdout(
		t,
		func() {
			args := []string{""}
			err := cl.Process(args)
			cl.Help(err, "unit-test", args)
		},
	)

	expectString(
		t,
		"Usage: unit-test <global options> <command> <options>\n\n"+
			"Global Options:\n\n"+
			"  cat:<num>    Morris is his name\n\n"+
			"All Commands:\n\n"+
			"  run          A second command\n  test:<flag>  Give me help\n\n",
		output,
	)

	cl.RegisterGlobalOption(
		func(values Values) error {
			return nil
		},
		"dog?Fido is his name",
	)

	output = captureStdout(
		t,
		func() {
			args := []string{}
			err := cl.Process(args)
			cl.Help(err, "unit-test", args)
		},
	)

	expectString(
		t,
		"Usage: unit-test <global options> <command> <options>\n\n"+
			"Global Options:\n\n"+
			"  cat:<num>    Morris is his name\n"+
			"  dog          Fido is his name\n\n"+
			"All Commands:\n\n"+
			"  run          A second command\n"+
			"  test:<flag>  Give me help\n\n",
		output,
	)
}

func TestInvalidOptionTypes(t *testing.T) {
	dot := defaultOptionTypes{}

	expectPanic(t, func() { dot.StringToAttributes("foo", "spec") })
	expectPanic(t, func() { dot.MakeValue(-1, "spec") })
	expectPanic(t, func() { dot.NewList(-1) })
	expectPanic(t, func() { dot.AppendList(-1, nil, "") })
}

func TestUseCaseNoOptions(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~",
	)

	output := captureStdout(
		t,
		func() {
			args := []string{"arg"}
			err := cl.Process(args)
			cl.Help(err, "tester", args)
		},
	)

	expectString(t, "Usage: tester\n\nThis command has no options.\n\n", output)

	output = captureStdout(
		t,
		func() {
			args := []string{"--help"}
			err := cl.Process(args)
			cl.Help(err, "tester", args)
		},
	)

	expectString(t, "\nThis command has no options.\n\n", output)
}

func TestUseCaseNoOptionsWithHelp(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~?Help me",
	)

	output := captureStdout(
		t,
		func() {
			args := []string{"arg"}
			err := cl.Process(args)
			cl.Help(err, "tester", args)
		},
	)

	expectString(t, "Usage: tester\n\nDescription: Help me\n\n", output)

	output = captureStdout(
		t,
		func() {
			args := []string{"--help"}
			err := cl.Process(args)
			cl.Help(err, "tester", args)
		},
	)

	expectString(t, "Description: Help me\n\n", output)
}

func TestUseCaseOneOptionWithHelp(t *testing.T) {
	cl := NewCommandLine()

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~ <string-flag>?Help me",
	)

	output := captureStdout(
		t,
		func() {
			args := []string{"arg"}
			err := cl.Process(args)
			cl.Help(err, "tester", args)
		},
	)

	expectString(t, "Usage: tester <options>\n\nCommand Options:\n\n<flag>  Help me\n\n", output)
}

func TestUseCaseUserTool(t *testing.T) {

	exampleHandler := func(args Values) error {
		fmt.Println(args)
		return nil
	}

	cl := NewCommandLine()

	cl.RegisterCommand(
		exampleHandler,
		"users?Performs operations on a user",
		"[--create <string-createUser>]?Creates a user",
		"[--delete <string-deleteUser>]?Deletes a user",
		"[--list]?List users",
	)

	cl.RegisterCommand(
		exampleHandler,
		"groups?Performs operations on user groups",
		"[--create <string-createGroup>]?Creates a user group",
		"[--delete <string-deleteGroup>]?Deletes a user group",
		"[--list]?List user groups",
		"[--describe <string-describeGroup>]?Describe details of a user group",
	)

	cl.RegisterGlobalOption(
		func(values Values) error { return nil },
		"--env:<string-env>",
	)

	args := []string{"--env:prod", "users", "--list"}
	output := captureStdout(
		t,
		func() {
			err := cl.Process(args)
			expectError(t, err, nil)
		},
	)

	expectString(t, "map[:<nil> --create:false --delete:false --list:true createUser: deleteUser: users:true]\n", output)
}

func TestOptionWithDash(t *testing.T) {
	cl := NewCommandLine()

	opt := ""

	cl.RegisterCommand(
		func(values Values) error {
			opt = values["opt"].(string)
			return nil
		},
		"~?Help me",
		"--my-option <string-opt>",
	)

	args := []string{"--my-option", "test"}
	err := cl.Process(args)
	expectError(t, nil, err)

	expectString(t, "test", opt)
}

func TestGlobalVariableMatchesSwitch(t *testing.T) {
	cl := NewCommandLine()

	optSwitch := false

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~?Help me",
	)

	cl.RegisterGlobalOption(
		func(values Values) (err error) {
			optSwitch = values["--opt"].(bool)
			return
		},
		"--opt",
	)

	args := []string{"--opt"}
	err := cl.Process(args)
	expectError(t, nil, err)

	expectBool(t, true, optSwitch)

	cl = NewCommandLine()

	optSwitch = false

	cl.RegisterCommand(
		func(values Values) error {
			return nil
		},
		"~?Help me",
	)

	cl.RegisterGlobalOption(
		func(values Values) (err error) {
			optSwitch = values["--opt"].(bool)
			return
		},
		"--opt <string-opt>",
	)

	args = []string{"--opt", "test"}
	err = cl.Process(args)
	expectError(t, nil, err)

	expectBool(t, true, optSwitch)
}
