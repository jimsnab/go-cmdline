package cmdline

import (
	"github.com/jimsnab/go-toolprinter"
)

var Prn = toolprinter.NewToolPrinter()

func SetPrinter(prn toolprinter.ToolPrinter) toolprinter.ToolPrinter {
	prior := prn
	Prn = prn
	return prior
}
