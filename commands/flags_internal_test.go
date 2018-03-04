package commands

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/health-checker/options"
	"github.com/gruntwork-io/gruntwork-cli/logging"
	"github.com/stretchr/testify/assert"
)

func TestParseChecksFromConfigWithInvalidOrEmptyConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		config string
		expectedChecks []options.Check
		expectedErr string
	}{
		{
			``,
			nil,
			"no checks found",
		},
		{
			` `,
			nil,
			"no checks found",
		},
		{
			`there is no checks
					or even valid
					yml here? 
					-`,
			nil,
			"unmarshal error",
		},
	}

	for _, testCase := range(testCases) {
		checks, err := parseChecksFromConfigString(testCase.config)
		if testCase.expectedErr != "" && err == nil {
			t.Fatalf("Expected error to contain \"%s\" but got checks: %v", testCase.expectedErr, checks)
		}
		assert.True(t, strings.Contains(err.Error(), testCase.expectedErr))
	}
}

func parseChecksFromConfigString(configString string) ([]options.Check, error){
	logger := logging.GetLogger("TEST")
	configByteSlice := []byte(configString)

	checks, err := parseChecksFromConfig(configByteSlice, logger)
	if err != nil {
		return nil, err
	}

	return checks, nil
}