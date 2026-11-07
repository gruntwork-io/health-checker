package main

import (
	"fmt"
	"os"

	"github.com/gruntwork-io/health-checker/commands"
)

// This variable is set at build time using -ldflags parameters. For example, .github/workflows/ci.yml sets this flag
// to the Git tag when building release binaries:
//
// go build -ldflags "-X main.VERSION=v0.1.0" .
//
// For more info, see: http://stackoverflow.com/a/11355611/483528
var VERSION string

// This is the main entry point for the app.
func main() {
	app := commands.CreateCli(VERSION)
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}
