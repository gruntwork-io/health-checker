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
    {{.HelpName}} {{if .Flags}}[parameters]{{end}}
    {{end}}{{if .Commands}}
 PARAMETERS:
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
	app.Usage = "A simple HTTP server that returns a 200 OK when all given TCP ports accept inbound connections."
	app.Commands = nil
	app.Flags = defaultFlags
	app.Action = runHealthChecker

	return app
}

func runHealthChecker(cliContext *cli.Context) error {
	opts, err := parseOptions(cliContext)
	if isSimpleError(err) {
		return err
	}
	if err != nil  {
		return errors.WithStackTrace(err)
	}

	opts.Logger.Infof("The Health Check will attempt to connect to the following ports via TCP: %v", opts.Ports)
	opts.Logger.Infof("Listening on Port %s...", opts.Listener)
	server.StartHttpServer(opts)

	// When an HTTP request comes in, open a TCP health check
	return nil
}
