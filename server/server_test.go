package server

import (
	"github.com/gruntwork-io/gruntwork-cli/logging"
	"github.com/gruntwork-io/health-checker/options"
	"github.com/gruntwork-io/health-checker/test"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net"
	"os"
	"testing"
)

func TestParseChecksFromConfig(t *testing.T) {
	// Will *not* run parallel because we're opening random tcp ports
	// and want to avoid port clashes
	testCases := []struct {
		name           string
		numports       int
		failport       bool
		scripts        []string
		scriptTimeout  int
		expectedStatus int
	}{
		{
			"port check",
			1,
			false,
			[]string{},
			5,
			200,
		},
		{
			"multiport check",
			3,
			false,
			[]string{},
			5,
			200,
		},
		{
			"multiport check one fails",
			3,
			true,
			[]string{},
			5,
			504,
		},
		{
			"script ok",
			0,
			false,
			[]string{"echo 'hello'"},
			5,
			200,
		},
		{
			"script fail",
			0,
			false,
			[]string{"lskdf"},
			5,
			504,
		},
		{
			"multi script ok",
			0,
			false,
			[]string{"echo 'hello1'", "echo 'hello2'"},
			5,
			200,
		},
		{
			"multi script one fail",
			0,
			false,
			[]string{"echo 'hello1'", "lskdf"},
			5,
			504,
		},
		{
			"script and port",
			1,
			false,
			[]string{"echo 'hello1'"},
			5,
			200,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ports, err := test.GetFreePorts(1 + testCase.numports)

			if err != nil {
				assert.FailNow(t, "Failed to get free ports: %v", err.Error())
			}

			listenerString := test.ListenerString(test.DEFAULT_LISTENER_ADDRESS, ports[0])

			checkPorts := []int{}
			listenPorts := []int{}

			// If we're monitoring tcp ports, prepare them
			if testCase.numports > 0 {
				checkPorts = ports[1:]
				listenPorts = make([]int, len(checkPorts))
				copy(listenPorts, checkPorts)

				// If we want to fail one check, remove the first port from the listen ports
				// So the health-check cannot connect
				if testCase.failport {
					listenPorts = listenPorts[1:]
				}
			}

			listeners := []net.Listener{}

			for _, port := range listenPorts {
				t.Logf("Creating listener for port %d", port)
				l, err := net.Listen("tcp", test.ListenerString(test.DEFAULT_LISTENER_ADDRESS, port))
				if err != nil {
					t.Logf("Error creating listener for port %d: %s", port, err.Error())
					assert.FailNow(t, "Failed to start listening: %s", err.Error())
				}

				listeners = append(listeners, l)

				// Separate goroutine for the tcp listeners
				go handleRequests(t, l)
			}

			defer closeListeners(t, listeners)

			opts := createOptionsForTest(t, testCase.scriptTimeout, testCase.scripts, listenerString, checkPorts)

			// Run the checks and verify the status code
			response := runChecks(opts)
			assert.True(t, testCase.expectedStatus == response.StatusCode, "Got expected status code")
		})
	}

}

func closeListeners(t *testing.T, listeners []net.Listener) {
	for _, l := range listeners {
		err := l.Close()
		if err != nil {
			t.Fatal("Failed to close listener: ", err)
		}
	}
}

func handleRequests(t *testing.T, l net.Listener) {
	for {
		// Listen for an incoming connection.
		l.Accept()
		// We don't log these when testing because we're forcibly closing the socket
		// from the outside. If you're debugging and wish to enable the logging,
		// uncomment the lines below
		//_, err := l.Accept()
		//if err != nil {
		//	t.Logf("Error accepting: %s", err.Error())
		//}
	}
}

func createOptionsForTest(t *testing.T, scriptTimeout int, scripts []string, listener string, ports []int) *options.Options {
	logger := logging.GetLogger("health-checker")
	logger.Out = os.Stdout
	logger.Level = logrus.InfoLevel

	opts := &options.Options{}
	opts.Logger = logger
	opts.ScriptTimeout = scriptTimeout
	opts.Scripts = options.ParseScripts(scripts)
	opts.Listener = listener
	opts.Ports = ports
	return opts
}
