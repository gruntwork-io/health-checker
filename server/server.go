package server

import (
	"context"
	"fmt"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/health-checker/options"
	"net"
	"net/http"
	"os/exec"
	"sync"
	"time"
)

type httpResponse struct {
	StatusCode int
	Body       string
}

func StartHttpServer(opts *options.Options) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp := runChecks(opts)
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
func runChecks(opts *options.Options) *httpResponse {
	logger := opts.Logger
	logger.Infof("Received inbound request. Beginning health checks...")

	allChecksOk := true

	var waitGroup = sync.WaitGroup{}

	for _, port := range opts.Ports {
		waitGroup.Add(1)
		go func(port int) {
			err := attemptTcpConnection(port, opts)
			if err != nil {
				logger.Warnf("TCP connection to port %d FAILED: %s", port, err)
				allChecksOk = false
			} else {
				logger.Infof("TCP connection to port %d successful", port)
			}

			waitGroup.Done()
		}(port)
	}

	for _, script := range opts.Scripts {
		waitGroup.Add(1)
		go func(script options.Script) {

			defer waitGroup.Done()

			logger.Infof("Executing '%v' with a timeout of %v seconds...", script, opts.ScriptTimeout)

			timeout := time.Second * time.Duration(opts.ScriptTimeout)

			ctx, cancel := context.WithTimeout(context.Background(), timeout)

			defer cancel()

			cmd := exec.CommandContext(ctx, script.Name, script.Args...)

			output, err := cmd.Output()

			if err != nil {
				logger.Warnf("Script %v FAILED: %s", script.Name, err)
				logger.Warnf("Command output: %s", output)
				allChecksOk = false
			} else {
				logger.Infof("Script %v successful", script)
			}
		}(script)
	}

	waitGroup.Wait()

	if allChecksOk {
		logger.Infof("All health checks passed. Returning HTTP 200 response.\n")
		return &httpResponse{StatusCode: http.StatusOK, Body: "OK"}
	} else {
		logger.Infof("At least one health check failed. Returning HTTP 504 response.\n")
		return &httpResponse{StatusCode: http.StatusGatewayTimeout, Body: "At least one health check failed"}
	}
}

// Attempt to open a TCP connection to the given port
func attemptTcpConnection(port int, opts *options.Options) error {
	logger := opts.Logger
	logger.Infof("Attempting to connect to port %d via TCP...", port)

	defaultTimeout := time.Second * 5

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("0.0.0.0:%d", port), defaultTimeout)
	if err != nil {
		return err
	}

	defer conn.Close()

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
