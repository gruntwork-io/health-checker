package server

import (
	"net/http"
	"github.com/gruntwork-io/health-checker/options"
	"net"
	"fmt"
)

func StartHttpServer(opts *options.Options) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		checkTcpPorts(w, r, opts)
	})
	http.ListenAndServe(opts.Listener, nil)
}

// Check that we can open a TPC connection to all the ports in opts.Ports
func checkTcpPorts(w http.ResponseWriter, r *http.Request, opts *options.Options) {
	logger := opts.Logger
	logger.Infof("Received inbound request. Beginning health checks...")

	allPortsValid := true

	for _, port := range opts.Ports {
		err := attemptTcpConnection(port, opts)
		if err != nil {
			logger.Warnf("TCP connection to port %d FAILED!", port)
			allPortsValid = false
		} else {
			logger.Infof("TCP connection to port %d successful", port)
		}
	}

	if allPortsValid {
		logger.Infof("All health checks passed. Returning HTTP 200 response.")
		writeHttp200Response(w)
	} else {
		logger.Infof("At least one health check failed. Returning HTTP 504 response.")
		writeHttp504Response(w)
	}
}

// Attempt to open a TCP connection to the given port
func attemptTcpConnection(port int, opts *options.Options) error {
	logger := opts.Logger
	logger.Infof("Attempting to connect to port %d via TCP...", port)

	_, err := net.Dial("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return err
	}

	return nil
}

func writeHttp200Response(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func writeHttp504Response(w http.ResponseWriter) {
	w.WriteHeader(http.StatusGatewayTimeout)
	w.Write([]byte("At least one health check failed"))
}

