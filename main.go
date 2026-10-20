package main

import (
	"fmt"
	"os"

	"github.com/dannyben/invert-image/internal/app"
)

var Version = "0.0.0-dev"

func main() {
	if err := app.Run(os.Args[1:], Version, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
