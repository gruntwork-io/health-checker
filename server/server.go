package server

import (
	"net/http"
	"github.com/gruntwork-io/health-checker/options"
	"net"
	"fmt"
	"sync"
	"time"
	"github.com/gruntwork-io/docs/errors"
)

type httpResponse struct {
	StatusCode int
	Body string
}

func StartHttpServer(opts *options.Options) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp := checkTcpPorts(opts)
		err := writeHttpResponse(w, resp)
		if err != nil {
			opts.Logger.Error("Failed to send HTTP response. Exiting.")
			panic(err)
		}
	})
	err := http.ListenAndServe(opts.Listener, nil)
	if err != nil {
		return err
	}

	return nil
}

// Check that we can open a TPC connection to all the ports in opts.Ports
func checkTcpPorts(opts *options.Options) *httpResponse {
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
		return &httpResponse{ StatusCode: http.StatusOK, Body: "OK" }
	} else {
		logger.Infof("At least one health check failed. Returning HTTP 504 response.\n")
		return &httpResponse{ StatusCode: http.StatusGatewayTimeout, Body: "At least one health check failed" }
	}
}

// Attempt to open a TCP connection to the given port
func attemptTcpConnection(port int, opts *options.Options) error {
	logger := opts.Logger
	logger.Infof("Attempting to connect to port %d via TCP...", port)

	defaultTimeout := time.Second * 5

	_, err := net.DialTimeout("tcp", fmt.Sprintf("0.0.0.0:%d", port), defaultTimeout)
	if err != nil {
		return err
	}

	return nil
}

func writeHttpResponse(w http.ResponseWriter, resp *httpResponse) error {
	w.WriteHeader(resp.StatusCode)
	_, err := w.Write([]byte(resp.Body))
	if err != nil {
		return errors.WithStackTrace(err)
	}

	return nil
}

