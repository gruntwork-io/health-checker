package commands

import (
	"fmt"
	"github.com/gruntwork-io/health-checker/options"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/gruntwork-cli/logging"
	"github.com/urfave/cli"
	"github.com/sirupsen/logrus"
)

var listenerFlag = cli.StringFlag{
	Name:  "listener",
	Usage: fmt.Sprintf("The IP address and port on which health-checker will listen for inbound connections"),
	Value: "0.0.0.0:5500",
}

var portFlag = cli.IntSliceFlag{
	Name: "port",
	Usage: fmt.Sprintf("Ports"),
}

var logLevelFlag = cli.StringFlag{
	Name:  "log-level",
	Usage: fmt.Sprintf("(Optional) Set the log level to `LEVEL`. Must be one of: %v", logrus.AllLevels),
	Value: logrus.InfoLevel.String(),
}

var defaultFlags = []cli.Flag{
	listenerFlag,
	portFlag,
	logLevelFlag,
}

// Parse all CLI options or flags
func parseOptions(cliContext *cli.Context) (*options.Options, error) {
	logger := logging.GetLogger("health-checker")

	logLevel := cliContext.String(logLevelFlag.Name)
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}
	logger.SetLevel(level)

	ports := cliContext.IntSlice("port")
	if len(ports) == 0 {
		return nil, errors.WithStackTrace(MissingParam(portFlag.Name))
	}

	listener := cliContext.String("listener")
	if listener == ""	 {
		return nil, errors.WithStackTrace(MissingParam(listenerFlag.Name))
	}

	return &options.Options{
		Logger:         logger,
		Ports:          ports,
		Listener:       listener,
	}, nil
}

// Custom error types

type MissingParam string

func (paramName MissingParam) Error() string {
	return fmt.Sprintf("Missing required parameter --%s", string(paramName))
}