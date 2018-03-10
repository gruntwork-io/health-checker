package commands

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/health-checker/options"
	"github.com/gruntwork-io/health-checker/server"
	"github.com/stretchr/testify/assert"
)

func TestParseChecksFromConfig(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name 	string
		config	string
		checks	[]options.Check
		err		string
	}{
		{
			name: "config empty",
			config: ``,
			err: "no checks found",
		},
		{
			name: "config with only whitespace",
			config: ` `,
			err: "no checks found",
		},
		{
			name: "config with invalid yaml",
			config: `there is no checks
					or even valid
					yml here? 
					-`,
			err: "unmarshal error",
		},
		{
			name: "config with an unknown key",
			config: `
http:
  - name: httpService1
    host: 127.0.0.1
    port: 8080
    success_codes: [200, 204, 301, 302]
invalidkey:
  - name: bad
    description: this should fail`,
			err: "unmarshal error",
		},
		{
			name: "config with single tcp check",
			config: `
tcp:
  - name: service1
    host: 127.0.0.1
    port: 8081`,
			checks: []options.Check{
				server.TcpCheck{
					Name: "service1",
					Host: "127.0.0.1",
					Port: 8081,
				},
			},
		},
		{
			name: "config with two tcp checks",
			config: `
tcp:
  - name: service1
    host: 127.0.0.1
    port: 8080
    timeout: 5
  - name: service2
    host: 0.0.0.0
    port: 8081`,
			checks: []options.Check{
				server.TcpCheck{
					Name: "service1",
					Host: "127.0.0.1",
					Port: 8080,
					Timeout: 5,
				},
				server.TcpCheck{
					Name: "service2",
					Host: "0.0.0.0",
					Port: 8081,
				},
			},
		},
		{
			name: "config with all check types",
			config: `
tcp:
  - name: service1
    host: 127.0.0.1
    port: 8080
    timeout: 5
http:
  - name: httpservice1
    host: localhost
    port: 80
    success_status_codes: [200, 204, 429]
    timeout: 3
  - name: httpservice2
    host: 127.0.0.1
    port: 8081
    body_regex: "test"
script:
  - name: script1
    script: /usr/local/bin/foo.sh`,
			checks: []options.Check{
				server.TcpCheck{Name: "service1", Host: "127.0.0.1", Port: 8080, Timeout: 5},
				server.HttpCheck{Name: "httpservice1", Host: "localhost", Port: 80, SuccessStatusCodes: []int{200, 204, 429}, Timeout: 3},
				server.HttpCheck{Name: "httpservice2", Host: "127.0.0.1", Port: 8081, BodyRegex: "test"},
				server.ScriptCheck{Name: "script1", Script: "/usr/local/bin/foo.sh"},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			b := []byte(tc.config)

			checks, err := parseChecksFromConfig(b)
			if err != nil && tc.err != "" {
				assert.True(t, strings.Contains(err.Error(), tc.err))
			} else if err != nil {
				t.Fatalf("unexpected error, got %v", err.Error())
			}

			assert.Equal(t, tc.checks, checks)
		})
	}
}

