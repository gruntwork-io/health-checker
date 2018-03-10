package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/gruntwork-io/gruntwork-cli/logging"
	"github.com/gruntwork-io/health-checker/options"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/gruntwork-io/health-checker/server"
)

const DEFAULT_CHECKS_FILE = "health-checks.yml"
const DEFAULT_LISTENER_IP_ADDRESS = "0.0.0.0"
const DEFAULT_LISTENER_PORT = 5500
const ENV_VAR_NAME_DEBUG_MODE = "HEALTH_CHECKER_DEBUG"


var configFlag = cli.StringFlag{
	Name: "config",
	Usage: fmt.Sprintf("[Required] A YAML file containing health checks. Default: %s", DEFAULT_CHECKS_FILE),
	Value: DEFAULT_CHECKS_FILE,
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
	configFlag,
	listenerFlag,
	logLevelFlag,
}

// define structure of config file
type Checks struct {
	TcpChecks    []server.TcpCheck    `yaml:"tcp"`
	HttpChecks   []server.HttpCheck   `yaml:"http"`
	ScriptChecks []server.ScriptCheck `yaml:"script"`
}

// Parse and validate all CLI options
func parseOptions(cliContext *cli.Context) (*options.Options, error) {
	logger := logging.GetLogger("health-checker")

	// By default logrus logs to stderr. But since most output in this tool is informational, we default to stdout.
	logger.Out = os.Stdout

	logLevel := cliContext.String(logLevelFlag.Name)
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return nil, InvalidLogLevel(logLevel)
	}
	logger.SetLevel(level)

	listener := cliContext.String("listener")
	if listener == "" {
		return nil, MissingParam(listenerFlag.Name)
	}

	configFile := cliContext.String("config")
	if configFile == "" {
		return nil, MissingParam(configFlag.Name)
	}

	b, err := loadBytes(configFile)
	if err != nil {
		return nil, InvalidConfigFile(configFile)
	}

	checks, err := parseChecksFromConfig(b)
	if err != nil {
		return nil, err
	}

	return &options.Options{
		Checks:         checks,
		Listener:       listener,
		Logger:         logger,
	}, nil
}

func parseChecksFromConfig(b []byte) ([]options.Check, error) {
	var checks Checks
	var checkSlice []options.Check

	// Use UnmarshalStrict to catch any issues in the config,
	// such as misspelled keys.
	err := yaml.UnmarshalStrict(b, &checks)
	if err != nil {
		return nil, err
	}

	if len(checks.TcpChecks) + len(checks.HttpChecks) + len(checks.ScriptChecks) == 0 {
		return nil, errors.New("no checks found: must specify at least one check")
	} else {
		for n := range checks.TcpChecks {
			checkSlice = append(checkSlice, checks.TcpChecks[n])
		}
		for n := range checks.HttpChecks {
			checkSlice = append(checkSlice, checks.HttpChecks[n])
		}
		for n := range checks.ScriptChecks {
			checkSlice = append(checkSlice, checks.ScriptChecks[n])
		}
	}

	for _, check := range checkSlice {
		err := check.ValidateCheck()
		if err != nil {
			return nil, err
		}
	}

	return checkSlice, nil
}

func loadBytes(filename string) ([]byte, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil || len(b) == 0 {
		return nil, err
	}
	return b, nil
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
type InvalidConfigFile string

func (configFile InvalidConfigFile) Error() string {
	return fmt.Sprintf("Error while parsing config: %s", string(configFile))
}

type InvalidLogLevel string

func (invalidLogLevel InvalidLogLevel) Error() string {
	return fmt.Sprintf("The log-level value \"%s\" is invalid", string(invalidLogLevel))
}

type MissingParam string

func (paramName MissingParam) Error() string {
	return fmt.Sprintf("Missing required parameter --%s", string(paramName))
}

