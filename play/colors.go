package play

import (
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

var AvailableColors = []color.Attribute{
	color.FgGreen,
	color.FgYellow,
	color.FgBlue,
	color.FgMagenta,
	color.FgCyan,
}

func init() {
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		color.NoColor = true
	}
}
