package commands

import (
	"fmt"
	"github.com/gruntwork-io/health-checker/options"
	"github.com/gruntwork-io/gruntwork-cli/logging"
	"github.com/urfave/cli"
	"github.com/sirupsen/logrus"
)

const DEFAULT_LISTENER_IP_ADDRESS = "0.0.0.0"
const DEFAULT_LISTENER_PORT = 5500

var portFlag = cli.IntSliceFlag{
	Name: "port",
	Usage: fmt.Sprintf("[Required] The port number on which a TCP connection will be attempted. Specify one or more times. Example: 8000"),
}

var listenerFlag = cli.StringFlag{
	Name:  "listener",
	Usage: fmt.Sprintf("[Optional] The IP address and port on which inbound HTTP connections will be accepted."),
	Value: fmt.Sprintf("%s:%d", DEFAULT_LISTENER_IP_ADDRESS, DEFAULT_LISTENER_PORT),
}

var logLevelFlag = cli.StringFlag{
	Name:  "log-level",
	Usage: fmt.Sprintf("[Optional] Set the log level to `LEVEL`. Must be one of: %v", logrus.AllLevels),
	Value: logrus.InfoLevel.String(),
}

var defaultFlags = []cli.Flag{
	portFlag,
	listenerFlag,
	logLevelFlag,
}

// Parse and validate all CLI options
func parseOptions(cliContext *cli.Context) (*options.Options, error) {
	logger := logging.GetLogger("health-checker")

	logLevel := cliContext.String(logLevelFlag.Name)
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return nil, InvalidLogLevel(logLevel)
	}
	logger.SetLevel(level)

	ports := cliContext.IntSlice("port")
	if len(ports) == 0 {
		return nil, MissingParam(portFlag.Name)
	}

	listener := cliContext.String("listener")

	return &options.Options{
		Ports:          ports,
		Listener:       listener,
		Logger:         logger,
	}, nil
}

// Some error types are simple enough that we'd rather just show the error message directly instead vomiting out a
// whole stack trace in log output
func isSimpleError(err error) bool {
	_, isInvalidLogLevelErr := err.(InvalidLogLevel)
	_, isMissingParam := err.(MissingParam)

	return isInvalidLogLevelErr || isMissingParam
}

// Custom error types

type InvalidLogLevel string

func (invalidLogLevel InvalidLogLevel) Error() string {
	return fmt.Sprintf("The log-level value \"%s\" is invalid", string(invalidLogLevel))
}

type MissingParam string

func (paramName MissingParam) Error() string {
	return fmt.Sprintf("Missing required parameter --%s", string(paramName))
}