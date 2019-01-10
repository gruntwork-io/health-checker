package commands

import (
	"flag"
	"fmt"
	"github.com/gruntwork-io/health-checker/options"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
	"strings"
	"testing"
)

func TestParseChecksFromConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		args            []string
		expectedOptions *options.Options
		expectedErr     string
	}{
		{
			"no options",
			[]string{},
			nil,
			"Missing required parameter, one of",
		},
		{
			"invalid log-level",
			[]string{"--log-level", "notreally"},
			nil,
			"The log-level value",
		},
		{
			"invalid listener",
			[]string{"--listener"},
			optionsForTest(t, DEFAULT_SCRIPT_TIMEOUT_SEC, []string{}, defaultListener(), []int{8080}),
			"Missing required parameter --listener",
		},
		{
			"invalid listener",
			[]string{"--listener", "1234", "--port", "4321"},
			optionsForTest(t, DEFAULT_SCRIPT_TIMEOUT_SEC, []string{}, listenerString(DEFAULT_LISTENER_IP_ADDRESS, 1234), []int{4321}),
			"",
		},
		{
			"single port",
			[]string{"--port", "8080"},
			optionsForTest(t, DEFAULT_SCRIPT_TIMEOUT_SEC, []string{}, defaultListener(), []int{8080}),
			"",
		},
		{
			"multiple ports",
			[]string{"--port", "8080", "--port", "8081"},
			optionsForTest(t, DEFAULT_SCRIPT_TIMEOUT_SEC, []string{}, defaultListener(), []int{8080, 8081}),
			"",
		},
		{
			"both port and script",
			[]string{"--port", "8080", "--script", "\"/usr/local/bin/check.sh 1234\""},
			optionsForTest(t, DEFAULT_SCRIPT_TIMEOUT_SEC, []string{"\"/usr/local/bin/check.sh 1234\""}, defaultListener(), []int{8080}),
			"",
		},
		{
			"single script",
			[]string{"--script", "/usr/local/bin/check.sh"},
			optionsForTest(t, DEFAULT_SCRIPT_TIMEOUT_SEC, []string{"/usr/local/bin/check.sh"}, defaultListener(), []int{}),
			"",
		},
		{
			"single script with custom timeout",
			[]string{"--script", "/usr/local/bin/check.sh", "--script-timeout", "11"},
			optionsForTest(t, 11, []string{"/usr/local/bin/check.sh"}, defaultListener(), []int{}),
			"",
		},
		{
			"multiple scripts",
			[]string{"--script", "/usr/local/bin/check1.sh", "--script", "/usr/local/bin/check2.sh"},
			optionsForTest(t, DEFAULT_SCRIPT_TIMEOUT_SEC, []string{"/usr/local/bin/check1.sh", "/usr/local/bin/check2.sh"}, defaultListener(), []int{}),
			"",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			context := getContextForTesting(testCase.args)

			actualOptions, actualErr := parseOptions(context)

			if testCase.expectedErr != "" {
				if actualErr == nil {
					assert.FailNow(t, "Expected error %v but got nothing.", testCase.expectedErr)
				}
				assert.True(t, strings.Contains(actualErr.Error(), testCase.expectedErr), "Expected error %v but got error %v", testCase.expectedErr, actualErr)
			} else {
				assert.Nil(t, actualErr, "Unexpected error: %v", actualErr)
				assertOptionsEqual(t, *testCase.expectedOptions, *actualOptions, "For args %v", testCase.args)
			}
		})
	}

}

func getContextForTesting(args []string) *cli.Context {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	c := CreateCli("0.0.0")
	ctx := cli.NewContext(c, flagSet, nil)
	for _, f := range c.Flags {
		f.Apply(flagSet)
	}
	flagSet.Parse(args)
	return ctx
}

func optionsForTest(t *testing.T, scriptTimeout int, scripts []string, listener string, ports []int) *options.Options {
	opts := &options.Options{}
	opts.ScriptTimeout = scriptTimeout
	opts.Scripts = scripts
	opts.Listener = listener
	opts.Ports = ports
	return opts
}

func assertOptionsEqual(t *testing.T, expected options.Options, actual options.Options, msgAndArgs ...interface{}) {
	assert.Equal(t, expected.ScriptTimeout, actual.ScriptTimeout, msgAndArgs...)
	assert.Equal(t, expected.Scripts, actual.Scripts, msgAndArgs...)
	assert.Equal(t, expected.Listener, actual.Listener, msgAndArgs...)
	assert.Equal(t, expected.Ports, actual.Ports, msgAndArgs...)
}

func defaultListener() string {
	return listenerString(DEFAULT_LISTENER_IP_ADDRESS, DEFAULT_LISTENER_PORT)
}

func listenerString(address string, port int) string {
	return fmt.Sprintf("%s:%d", address, port)
}
