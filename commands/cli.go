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

 USAGE:
    {{.HelpName}} {{if .Flags}}[options]{{end}}
    {{if .Commands}}
 OPTIONS:
    {{range .Flags}}{{.}}
    {{end}}{{end}}{{if .Copyright }}
 COPYRIGHT:
    {{.Copyright}}
    {{end}}{{if .Version}}
 VERSION:
    {{.Version}}
    {{end}}{{if len .Authors}}
 AUTHOR(S):
    {{range .Authors}}{{ . }}{{end}}
	{{end}}
`

	app.Name = "health-checker"
	app.HelpName = app.Name
	app.Author = "Gruntwork, Inc. <www.gruntwork.io> | https://github.com/gruntwork-io/health-checker"
	app.Version = version
	app.Usage = "A simple HTTP server that will return 200 OK if the configured checks are all successful."
	app.Commands = nil
	app.Flags = defaultFlags
	app.Action = runHealthChecker

	return app
}

func runHealthChecker(cliContext *cli.Context) error {
	if allCliOptionsEmpty(cliContext) {
		cli.ShowAppHelpAndExit(cliContext, 0)
	}

	opts, err := parseOptions(cliContext)
	if isDebugMode() {
		opts.Logger.Infof("Note: To enable debug mode, set %s to \"true\"", ENV_VAR_NAME_DEBUG_MODE)
		return err
	}
	if err != nil {
		return errors.WithStackTrace(err)
	}

	opts.Logger.Infof("The Health Check will attempt to connect to the following ports via TCP: %v", opts.Ports)
	opts.Logger.Infof("Listening on Port %s...", opts.Listener)
	err = server.StartHttpServer(opts)
	if err != nil {
		return errors.WithStackTrace(err)
	}

	return nil
}
