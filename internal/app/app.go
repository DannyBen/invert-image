package app

import (
	"fmt"
	"io"
)

const usage = `Usage: invert-image FILE [OUTFILE]

Invert the colors of a PNG or JPEG image.
Without OUTFILE, replace the original file.

  --help       show help
  --version    show version
`

func Run(args []string, version string, out io.Writer) error {
	if len(args) == 1 {
		switch args[0] {
		case "--help", "-h":
			_, err := fmt.Fprint(out, usage)
			return err
		case "--version":
			_, err := fmt.Fprintln(out, "invert-image", version)
			return err
		}
	}
	if len(args) < 1 || len(args) > 2 {
		return fmt.Errorf("usage: invert-image FILE [OUTFILE]")
	}
	output := ""
	if len(args) == 2 {
		output = args[1]
		if output == "" {
			return fmt.Errorf("output filename is empty")
		}
	}
	return invertFile(args[0], output)
}
