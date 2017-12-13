package commands

import (
	"github.com/urfave/cli"
)

// Create the CLI app with all commands, flags, and usage text configured.
func CreateCli(version string) *cli.App {
	app := cli.NewApp()

	app.Name = "health-checker"
	app.HelpName = app.Name
	app.Author = "Gruntwork <www.gruntwork.io>"
	app.Version = version
	app.Usage = "A simple HTTP server that returns a 200 OK when the given list of TCP ports all accept a connection."

	app.Action = func(cliContext *cli.Context) error {
		println("Hello Josh!")
		return nil
	}

	return app
}