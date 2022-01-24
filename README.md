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

```
import (
  "cmdline/cmdline"
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

```
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

```
	cl.RegisterCommand(singleCommand, "~?This is an example")
```
</details>

<details><summary>Build and Run</summary>

```
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

```
	cl.RegisterCommand(singleCommand, "~ <string-name>?This is an example")

	...

	fmt.Printf("Hello, %s!\n", args["name"].(string))

```

</details>

Now the program requires an argument, called `name`.

<details><summary>Build and Run</summary>


```
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

```
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

```
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

```
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

```
package main

import (
	"fmt"
	"os"
	"cmdline/cmdline"
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

```
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

```
func myHandler(args cmdline.Values) error {
	blockSize := args["blockSize"].(int)
	fmt.Println(blockSize)
	return nil
}
```

</details>
<details><summary>Build and Run</summary>

```
$ go build
$ ./myexample format -i /tmp/example.cfg
0
```
</details>
<br/>

## Auto-generated Help

As shown above, the standard pattern for showing help is:

```
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

```
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

```
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

```
package main

import (
	"fmt"
	"os"
	"cmdline/cmdline"
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

```
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
global map[env:prod]
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
position-based arguments can be specified.

<details><summary>Syntax</summary>

```
	// unnamed single command
	cl.RegisterCommand(
		myHandler,
		"~ <string-posarg1> <string-posarg2> <string-posarg3>",
	)
```

or

```
  // named command (multiple named commands supported)
	cl.RegisterCommand(
		myHandler,
		"mycommand <string-posarg1> <string-posarg2> <string-posarg3>",
	)
```

</details>
<br/>

The right side arguments can be optional.

<details><summary>Syntax</summary>

```
	cl.RegisterCommand(
		myHandler,
		"~ <string-posarg1> [<string-posarg2>] [<string-posarg3>]",
	)
```

</details>
<br/>

## Colon and Comma Delimeters

A single argument can be divided into values by using a colon to delimit the
first value, and commas to delimit subsequent values.

<details><summary>Code</summary>

```
package main

import (
	"fmt"
	"os"
	"cmdline/cmdline"
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

```
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

## Repeated Parameters
To allow a command line switch to be used more than once, it can be marked
with an asterisk (`*`), and the same switch can be specified more than once.
The command handler will get an array value.

<details><summary>Code</summary>

package main

import (
	"fmt"
	"os"
	"cmdline/cmdline"
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

```
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

## Extending Types

You can write your own `cmdline.OptionTypes` interface to convert arguments to your own
types and structs. Construct the command line object with:

```
	cl := cmdline.NewCustomTypesCommandLine(myType)
```

where `myType` fulfills the following interface:

```
type OptionTypes interface {
	StringToAttributes(typeName string, spec string) *OptionTypeAttributes
	MakeValue(typeIndex int, inputValue string) (interface{}, error)
	NewList(typeIndex int) (interface{}, error)
	AppendList(typeIndex int, list interface{}, inputValue string) (interface{}, error)
}
```

Your implementation determines valid values for `typeIndex`. Typically it is an integer
enumeration.

* `StringToAttributes` converts type string `spec` to the corresponding index and typed default value (the two members of `cmdline.OptionTypeAttributes`)
* `MakeValue` coverts command line input `inputValue` into the corresponding typed value
* `NewList` allocates a new typed array (see repeated values above)
* `AppendList` appends a value to the typed array provided by `NewList`

## Descriptor errors

If your command or global option registration is malformed, the registration API will
invoke `panic` with a message explaining the error. This helps quickly spot typos and
unsupported syntax.

It is not advised to try to `recover` from a registration api panic.
