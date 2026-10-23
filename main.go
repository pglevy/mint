// Command mint queries and exports Sailwind design tokens.
package main

import (
	"os"

	"github.com/pglevy/mint/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], cli.OSEnv()))
}
