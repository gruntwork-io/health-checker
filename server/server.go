package server

import (
	"net/http"
	"github.com/gruntwork-io/health-checker/options"
	"net"
	"fmt"
	"sync"
)

func StartHttpServer(opts *options.Options) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		checkTcpPorts(w, r, opts)
	})
	err := http.ListenAndServe(opts.Listener, nil)
	if err != nil {
		return err
	}

	return nil
}

// Check that we can open a TPC connection to all the ports in opts.Ports
func checkTcpPorts(w http.ResponseWriter, r *http.Request, opts *options.Options) {
	logger := opts.Logger
	logger.Infof("Received inbound request. Beginning health checks...")

	allPortsValid := true

	var waitGroup = sync.WaitGroup{}

	for _, port := range opts.Ports {
		waitGroup.Add(1)
		go func(port int) {
			err := attemptTcpConnection(port, opts)
			if err != nil {
				logger.Warnf("TCP connection to port %d FAILED: %s", port, err)
				allPortsValid = false
			} else {
				logger.Infof("TCP connection to port %d successful", port)
			}

			waitGroup.Done()
		}(port)
	}

	waitGroup.Wait()

	if allPortsValid {
		logger.Infof("All health checks passed. Returning HTTP 200 response.\n")
		writeHttp200Response(w)
	} else {
		logger.Infof("At least one health check failed. Returning HTTP 504 response.\n")
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

