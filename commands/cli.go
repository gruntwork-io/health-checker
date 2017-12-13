package commands

import (
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/health-checker/server"
	"github.com/urfave/cli"
)

// Create the CLI app with all commands (in this case a single one!), flags, and usage text configured.
func CreateCli(version string) *cli.App {
	app := cli.NewApp()

	app.CustomAppHelpTemplate = ` NAME:
    {{.Name}} - {{.Usage}}
    {{if len .Authors}}
 AUTHOR(S):
    {{range .Authors}}{{ . }}{{end}}

 USAGE:
    {{.HelpName}} {{if .Flags}}[options]{{end}}
    {{end}}{{if .Commands}}
 OPTIONS:
    {{range .Flags}}{{.}}
    {{end}}{{end}}{{if .Copyright }}
 COPYRIGHT:
    {{.Copyright}}
    {{end}}{{if .Version}}
 VERSION:
    {{.Version}}
    {{end}}
`

	app.Name = "health-checker"
	app.HelpName = app.Name
	app.Author = "Gruntwork <www.gruntwork.io>"
	app.Version = version
	app.Usage = "A simple HTTP server that returns a 200 OK when the given list of TCP ports all accept a connection."

	app.Commands = nil

	app.Flags = defaultFlags

	app.Action = errors.WithPanicHandling(func(cliContext *cli.Context) error {
		opts, err := parseOptions(cliContext)
		if err != nil {
			return err
		}

		opts.Logger.Infof("The Health Check will attempt to connect to the following ports via TCP: %v", opts.Ports)
		opts.Logger.Infof("Listening on Port %s...", opts.Listener)
		server.StartHttpServer(cliContext)

		// When an HTTP request comes in, open a TCP health check
		return nil
	})


	return app
}

