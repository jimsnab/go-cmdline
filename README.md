# Command Line Parser

## Description
This package provides an easy to use command line parser. By describing
your command line with string descriptors, you get:

* Execution of a command handler function, selected according to the process arguments
* Ability to describe the command values close to the way the command is used
* Auto-generated help
* Parsing of options available to all command handlers (global options)
* Conversion into the basic types: `string`, `bool`, `int`, `float64`, `path`
* Support for simple position-oriented parameters
* Support for optional parameters
* Support for repeated parameters
* Ability to extend the supported types with your own conversion code
* Detailed descriptor errors to speed fixing development mistakes

## Introduction Examples
Let's look at a process that we name `myexample` that has no command line arguments:

<details>
	<summary>Code</summary>

```go
import (
  "github.com/jimsnab/go-cmdline"
  "fmt"
  "os"
)

func main() {
	cl := cmdline.NewCommandLine()

	cl.RegisterCommand(singleCommand, "~")

	args := os.Args[1:] // exclude executable name in os.Args[0]
	err := cl.Process(args)
	if err != nil {
		cl.Help(err, "myexample", args)
	}
}

func singleCommand(args cmdline.Values) error {
	fmt.Println("Hello, world!")
	return nil
}
```

</details>

In the code above, after creating the command line struct, the command handler function `singleCommand` is registered as command named `~`, which means "unnamed". Then the command line arguments are sent in to be parsed. If something goes wrong, help is printed.

Let's build and run it.

<details><summary>Build and Run</summary>

```bash
$ # build myexample
$ go build

$ # run it
$ ./myexample
Hello, world!

$ ./myexample myarg
Usage: myexample

This command has no options.

$ ./myexample --help

This command has no options.

```
</details>
<br/>

Fine. Let's add some help. Add `?This is an example` to the code above.

<details><summary>Code</summary>

```go
	cl.RegisterCommand(singleCommand, "~?This is an example")
```
</details>

<details><summary>Build and Run</summary>

```bash
$ go build
$ ./myexample --help
Description: This is an example

$ ./myexample foo
Usage: myexample

Description: This is an example

```
</details>
<br/>

The main use of this package is to parse command line options, so let's add
one. Update the same line and change the `singleCommand` handler to use the argument, as follows:

<details><summary>Code</summary>

```go
	cl.RegisterCommand(singleCommand, "~ <string-name>?This is an example")

	...

	fmt.Printf("Hello, %s!\n", args["name"].(string))

```

</details>

Now the program requires an argument, called `name`.

<details><summary>Build and Run</summary>


```bash
$ go build
$ ./myexample
Usage: myexample <command>

Command Options:

<name>  This is an example

$ ./myexample fido
Hello, fido!
```

</details>
<br/>

One of the major use cases supported by the `cmdline` package is a tool that
provides several commands. Here's an example.

<details><summary>Code</summary>

```go
func main() {
	cl := cmdline.NewCommandLine()

	cl.RegisterCommand(
		func (args cmdline.Values) error {
			fmt.Println("74 degrees F and partly cloudy with 6 MPH winds")
			return nil
		},
		"now?Provides the current weather",
	)
	cl.RegisterCommand(
		func (args cmdline.Values) error {
			fmt.Println("Today will be sunny and 82 degrees F")
			return nil
		},
		"forecast?Provides a forecast of today's weather",
	)
	cl.RegisterCommand(
		func (args cmdline.Values) error {
			fmt.Println("0.2 in of precipitation so far this month")
			return nil
		},
		"precip?Print the precipitation statistics",
	)

	args := os.Args[1:] // exclude executable name in os.Args[0]
	err := cl.Process(args)
	if err != nil {
		cl.Help(err, "myexample", args)
	}
}
```
</details>
<details><summary>Build and Run</summary>

```bash
$ go build
$ ./myexample
Usage: myexample <command>

All Commands:

  forecast  Provides a forecast of today's weather
  now       Provides the current weather
  precip    Print the precipitation statistics

$ ./myexample now
74 degrees F and partly cloudy with 6 MPH winds
```

</details>

<br/>

# Reference
## Command Handler Function
A command handler receives a map of command line arguments and returns an `error`.

```go
func myHandler (args cmdline.Values) error {
  // your code
	return err
}
```

The `args` is filled with:

* A `bool` `true` for the primary command, exactly as the command is named.
* A `bool` for each option of the command, `true` if the option was specified.
* Each value for parameters of each switch, if any.

For example, consider the following command.

<details><summary>Code</summary>

```go
package main

import (
	"fmt"
	"os"
	"github.com/jimsnab/go-cmdline"
)

func main() {
	cl := cmdline.NewCommandLine()

	cl.RegisterCommand(
		myHandler,
		"format?Formats the storage",
		"-i <path-initFile>?Specifies the path to the initialization descriptor file",
		"[--force]?Performs the format even if the storage has been formatted",
		"[--dynamic[ <int-blockSize>]]?Formats for dynamic sizing, with optional blockSize",
	)
	
	args := os.Args[1:] // exclude executable name in os.Args[0]
	err := cl.Process(args)
	if err != nil {
		cl.Help(err, "myexample", args)
	}
}

func myHandler(args cmdline.Values) error {
	fmt.Println(args)
	return nil
}
```

</details>
<details><summary>Build and Run</summary>

```bash
$ go build
$ ./myexample
Usage: myexample <command> <options>

Command Options:

  format                      Formats the storage
    -i <initFile>             Specifies the path to the initialization descriptor file
    [--dynamic [<blockSize>]] Formats for dynamic sizing, with optional blockSize
    [--force]                 Performs the format even if the storage has been formatted

$ ./myexample format -i /tmp/example.cfg
map[--dynamic:false --force:false -i:true blockSize:0 format:true initFile:/tmp/example.cfg]
```
</details>
<br/>

The optional values that are not specified on the command line will have the default value
for the type.

Values in the map are typed, so the handler code can use type assertions.

<details><summary>Code</summary>

```go
func myHandler(args cmdline.Values) error {
	blockSize := args["blockSize"].(int)
	fmt.Println(blockSize)
	return nil
}
```

</details>
<details><summary>Build and Run</summary>

```bash
$ go build
$ ./myexample format -i /tmp/example.cfg
0
```
</details>
<br/>

## Auto-generated Help

As shown above, the standard pattern for showing help is:

```go
	args := os.Args[1:] // exclude executable name in os.Args[0]
	err := cl.Process(args)
	if err != nil {
		cl.Help(err, "myexample", args)
	}
```

The `Process()` function will parse the command line arguments and invoke
the corresponding handler, or, return an `error` if something went wrong. A
command line error is of type `cmdline.CommandLineError`. This type can be used
to distinguish between command line syntax errors and runtime errors.

The `Help()` function generates help according to the command line definition.
It also handles `help` and `--help` switches.

`Help()` will show help for a command when the command argument ends in a question mark.
In the "format" example above:

<details><summary>Help Example</summary>

```bash
$ ./myexample format?
Command Options:

  format                      Formats the storage
    -i <initFile>             Specifies the path to the initialization descriptor file
    [--dynamic [<blockSize>]] Formats for dynamic sizing, with optional blockSize
    [--force]                 Performs the format even if the storage has been formatted

```

</details>
<br/>

`Help()` supports a `filter` argument. The filter can be specified with `--help <filter>`,
or simply `<filter>`:

<details><summary>Help Example</summary>

```bash
$ ./myexample form
Usage: myexample <command> <options>

Command Options:

  format                      Formats the storage
    -i <initFile>             Specifies the path to the initialization descriptor file
    [--dynamic [<blockSize>]] Formats for dynamic sizing, with optional blockSize
    [--force]                 Performs the format even if the storage has been formatted
```
</details>
<br/>

If the filter text is found somewhere in the command help, the help for the entire
command will be printed. This is better than piping help to `grep`.

Your code can print a specific command with `cl.PrintCommand()`, or print the help
without "Usage" or filter help text by using `cl.PrintCommands()`.

## Global Options

A program with several commands can benefit from global options that are available
across all of the commands. When a global option is specified, an option handler
is invoked, with a map containing the global options.

Global options must be provided on the command line before the command.

<details>
 <summary>Code</summary>

```go
package main

import (
	"fmt"
	"os"
	"github.com/jimsnab/go-cmdline"
)

func main() {
	cl := cmdline.NewCommandLine()

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
		func(values cmdline.Values) error { 
			fmt.Println("global", values)
			return nil 
		},
		"--env:<string-env>",
	)

	args := os.Args[1:] // exclude executable name in os.Args[0]
	err := cl.Process(args)
	if err != nil {
		cl.Help(err, "myexample", args)
	}
}

func exampleHandler(args cmdline.Values) error {
	fmt.Println("command", args)
	return nil
}
```
</details>
<details>
	<summary>Build and Run</summary>

```bash
$ go build
$ ./myexample
Usage: myexample <global options> <command> <options>

Global Options:

  --env:<env>

All Commands:

  groups                      Performs operations on user groups
    [--create <createGroup>]  Creates a user group
    [--delete <deleteGroup>]  Deletes a user group
    [--describe <describeGroup>]
                              Describe details of a user group
    [--list]                  List user groups
  users                       Performs operations on a user
    [--create <createUser>]   Creates a user
    [--delete <deleteUser>]   Deletes a user
    [--list]                  List users

Search help with myexample --help <filter text>. Example: myexample --help users
Or, put a question mark on the end. Example: myexample users?

$ ./myexample --env:prod users --list
global map[--env:true env:prod]
command map[--create:false --delete:false --list:true createUser: deleteUser: users:true]
```
</details>
<br/>

NOTE: The example above needs improvement. Adding mutually exclusive secondary
arguments is in the backlog.

## Types Supported

An argument value is specified as `<` _type_ `-` _variable name_ `>`, where _type_ can be one of:

* `string` - an ordinary text string
* `bool` - `true` or `false` text
* `int` - a 32-bit integer
* `float64` - a floating point value
* `path` - a string holding a path in its canonical (absolute) form

## Simple Position-Oriented Parameters
A command can have optional arguments based on their position. Only a single list of
position-based arguments can be specified. A list of multiple values can be specified
at the last position using an asterisk `*`.

<details><summary>Syntax</summary>

```go
	// unnamed single command
	cl.RegisterCommand(
		myHandler,
		"~ <string-posarg1> <string-posarg2> <string-posarg3>",
	)
```

or

```go
    // named command (multiple named commands supported)
	cl.RegisterCommand(
		myHandler,
		"mycommand <string-posarg1> <string-posarg2> <string-posarg3>",
	)
```

or

```go
    // one or more values combined into an array
	cl.RegisterCommand(
		myHandler,
		"mycommand *<string-multiPosArg>",
	)
```

</details>
<br/>

The right side arguments can be optional.

<details><summary>Syntax</summary>

```go
	cl.RegisterCommand(
		myHandler,
		"~ <string-posarg1> [<string-posarg2>] [<string-posarg3>]",
	)
```

</details>
<br/>

Position-oriented parameters cannot have values that start with a dash, as
that is used to match named parameters.

It is possible to register named command handlers along with a position-oriented 
handler. Named command handlers have priority.

For example, it is possible to add a catch-all handler as shown here:

<details><summary>Syntax</summary>

```go
	cl.RegisterCommand(
		namedHandler,
		"named <string-arg>",
	)

	cl.RegisterCommand(
		defaultHandler,
		"~ *<string-args>",
	)
```

</details>
<br/>

## Colon and Comma Delimeters

A single argument can be divided into values by using a colon to delimit the
first value, and commas to delimit subsequent values.

<details><summary>Code</summary>

```go
package main

import (
	"fmt"
	"os"
	"github.com/jimsnab/go-cmdline"
)

func main() {
	cl := cmdline.NewCommandLine()

	cl.RegisterCommand(
		exampleHandler,
		"--first:<int-begin>,<int-end>",
	)

	cl.RegisterCommand(
		exampleHandler,
		"rangeA",
		"--second:<int-begin>,<int-end>",
	)
	
	cl.RegisterCommand(
		exampleHandler,
		"rangeB",
		"--third:<int-begin>[,<int-end>]",
	)

	args := os.Args[1:] // exclude executable name in os.Args[0]
	err := cl.Process(args)
	if err != nil {
		cl.Help(err, "myexample", args)
	}
}

func exampleHandler(args cmdline.Values) error {
	fmt.Println("command", args)
	return nil
}
```
</details>
<br/>

In the code above, the `--third` switch has an optional `end` argument. When using comma-separated optional arguments, the last value specified on the command line is
used as the default for each missing optional value.

<details><summary>Example Run</summary>

```bash
$ ./myexample
Usage: myexample <command> <options>

All Commands:

  --first:<begin>,<end>
  rangeA
    --second:<begin>,<end>
  rangeB
    --third:<begin>[,<end>]

$ ./myexample rangeB --third:10
command map[--third:true begin:10 end:10 rangeB:true]
```

</details>

## Subcommands
It is possible to register two or more tokens as the "primary command".

Example: suppose you wanted a command `view` with different handlers for each kind of view, such as `table` `row` and `cell`. Easy! Just use `+` to indicate a space in the primary command. Expand below for more details.

<details><summary>Code</summary>

```go
package main

import (
	"fmt"
	"os"
	"github.com/jimsnab/go-cmdline"
)

func main() {
	cl := cmdline.NewCommandLine()

	cl.RegisterCommand(
		func(args cmdline.Values) error {
			fmt.Println("view table...")
		},
		"view+table",
	)

	cl.RegisterCommand(
		func(args cmdline.Values) error {
			fmt.Println("view row...")
		},
		"view+row",
	)
	
	cl.RegisterCommand(
		func(args cmdline.Values) error {
			fmt.Println("view cell...")
		},
		"view+cell",
	)

	args := os.Args[1:] // exclude executable name in os.Args[0]
	err := cl.Process(args)
	if err != nil {
		cl.Help(err, "myexample", args)
	}
}
```
</details>

<details><summary>Example Run</summary>

```bash
$ ./myexample view table
view table...
```

</details>

## Repeated Parameters
To allow a command line switch to be used more than once, it can be marked
with an asterisk (`*`), and the same switch can be specified more than once.
The command handler will get an array value.

<details><summary>Code</summary>

package main

import (
	"fmt"
	"os"
	"github.com/jimsnab/go-cmdline"
)

func main() {
	cl := cmdline.NewCommandLine()

	cl.RegisterCommand(
		exampleHandler,
		"~",
		"*-f:<string-text>",
	)

	args := os.Args[1:] // exclude executable name in os.Args[0]
	err := cl.Process(args)
	if err != nil {
		cl.Help(err, "myexample", args)
	}
}

func exampleHandler(args cmdline.Values) error {
	fmt.Println("command", args)
	return nil
}

</details>

<details><summary>Output</summary>

```bash
$ ./myexample
Usage: myexample <options>

Command Options:

  *-f:<text>

$ ./myexample -f:one -f:two -f:three
command map[-f:true text:[one two three] ~:true]
```

</details>
<br/>

To support zero or more multiple switches, make the argument optional with the asterisk first, e.g., `*[-f:<string-text>]`.

## Primary Command

Your program can use the parser to extract the primary command.

```go
	primaryCmd := cl.PrimaryCommand(os.Args[1:])
```

This might be useful to find the primary command specified when global options are possible.

An empty string is returned if the command line arguments do not map to a command.

## Extending Types

You can write your own `cmdline.OptionTypes` interface to convert arguments to your own
types and structs. It takes a moment to understand this interface, but it ultimately
pretty simple.

Construct the command line object with:

```go
	cl := cmdline.NewCustomTypesCommandLine(myType)
```

where `myType` fulfills the following interface:

```go
type OptionTypes interface {
	StringToAttributes(typeName string, spec string) *OptionTypeAttributes
	MakeValue(typeIndex int, inputValue string) (any, error)
	NewList(typeIndex int) (any, error)
	AppendList(typeIndex int, list any, inputValue string) (any, error)
}
```

Your implementation determines valid values for `typeIndex`. Typically it is an integer
enumeration.

* `StringToAttributes` converts type string `spec` to the corresponding index and typed default value (the two members of `cmdline.OptionTypeAttributes`)
* `MakeValue` converts command line input `inputValue` into the corresponding typed value
* `NewList` allocates a new typed array (see repeated values above)
* `AppendList` appends a value to the typed array provided by `NewList`

A custom types handler owns supporting all the `spec` types used in your command line.
It is often desired to retain the default types (`bool`, `string`, `int`, `float64`,
`path`). This can be achieved via `NewDefaultOptionTypes()`, which provides the
default interface, so `myType` can fall back to the default implementation on unknown
`spec` types or unknown `typeIndex` values.

<details><summary>Example</summary>

```
type (
	cmdLineTypes struct {
		dot   *cmdline.DefaultOptionTypes
		index int
	}
)

func newCmdLineTypes() *cmdLineTypes {
	ut := cmdLineTypes{}
	ut.dot, ut.index = cmdline.NewDefaultOptionTypes()
	return &ut
}

func (t *cmdLineTypes) StringToAttributes(typeName string, spec string) (a *cmdline.OptionTypeAttributes) {
	if typeName == "uint64" {
		a = &cmdline.OptionTypeAttributes{
			DefaultValue: uint64(0),
			Index:        t.index,
		}
	} else {
		a = t.dot.StringToAttributes(typeName, spec)
	}
	return
}

func (t *cmdLineTypes) MakeValue(typeIndex int, inputValue string) (v any, err error) {
	if typeIndex == t.index {
		var n uint64
		if inputValue != "" {
			n, err = strconv.ParseUint(inputValue, 10, 64)
			if err != nil {
				return
			}
		}

		v = n
	} else {
		v, err = t.dot.MakeValue(typeIndex, inputValue)
	}
	return
}

func (t *cmdLineTypes) NewList(typeIndex int) (v any, err error) {
	if typeIndex == t.index {
		v = []uint64{}
	} else {
		v, err = t.dot.NewList(typeIndex)
	}
	return
}

func (t *cmdLineTypes) AppendList(typeIndex int, list any, inputValue string) (v any, err error) {
	if typeIndex == t.index {
		var n any
		n, err = t.MakeValue(typeIndex, inputValue)
		if err != nil {
			return
		}
		l := list.([]uint64)
		v = append(l, n.(uint64))
	} else {
		v, err = t.dot.AppendList(typeIndex, list, inputValue)
	}
	return
}
```

</details>

## Descriptor errors

If your command or global option registration is malformed, the registration API will
invoke `panic` with a message explaining the error. This helps quickly spot typos and
unsupported syntax.

It is not advised to try to `recover` from a registration api panic.

## Console Printer

The command line parser uses [toolprinter](https://github.com/jimsnab/go-toolprinter) to print to stdout.
You can provide your own implementation of this interface by calling `SetPrinter()`, if you want
to render help on something other than a shell terminal.
