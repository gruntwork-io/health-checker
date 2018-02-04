package commands

import (
	"fmt"
	"github.com/gruntwork-io/health-checker/options"
	"github.com/gruntwork-io/gruntwork-cli/logging"
	"github.com/urfave/cli"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

const DEFAULT_LISTENER_IP_ADDRESS = "0.0.0.0"
const DEFAULT_LISTENER_PORT = 5500
const ENV_VAR_NAME_DEBUG_MODE = "HEALTH_CHECKER_DEBUG"

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

// Return true if no options at all were passed to the CLI. Note that we are specifically testing for flags, some of which
// are required, not just args.
func allCliOptionsEmpty(cliContext *cli.Context) bool {
	return cliContext.NumFlags() == 0
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
	if listener == "" {
		return nil, MissingParam(listenerFlag.Name)
	}

	return &options.Options{
		Ports:          ports,
		Listener:       listener,
		Logger:         logger,
	}, nil
}

// Some error types are simple enough that we'd rather just show the error message directly instead of vomiting out a
// whole stack trace in log output. Therefore, allow a debug mode that always shows full stack traces. Otherwise, show
// simple messages.
func isDebugMode() bool {
	envVar, _ := os.LookupEnv(ENV_VAR_NAME_DEBUG_MODE)
	envVar = strings.ToLower(envVar)
	return envVar == "true"
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