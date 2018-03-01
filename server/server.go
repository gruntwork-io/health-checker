package server

import (
	"net/http"
	"net"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gruntwork-io/health-checker/options"
	"github.com/gruntwork-io/gruntwork-cli/errors"
)

type httpResponse struct {
	StatusCode int
	Body string
}

func StartHttpServer(opts *options.Options) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp := checkHealthChecks(opts)
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

func checkHealthChecks(opts *options.Options) *httpResponse {
	logger := opts.Logger
	logger.Infof("Received inbound request. Beginning health checks...")

	// initialize failedChecks to 0
	var failedChecks uint64
	var waitGroup = sync.WaitGroup{}

	for _, tcpCheck := range opts.Checks.TcpChecks {
		name := tcpCheck.Name
		host := tcpCheck.Host
		port := tcpCheck.Port
		waitGroup.Add(1)
		go func(port int) {
			err := attemptTcpConnection(tcpCheck.Host, tcpCheck.Port,opts)
			if err != nil {
				logger.Warnf("TCP connection to %s at %s:%d FAILED: %s", name, host, port, err)
				atomic.AddUint64(&failedChecks, 1)
			} else {
				logger.Infof("TCP connection to %s at %s:%d successful", name, host, port)
			}
			waitGroup.Done()
		}(port)
	}

	waitGroup.Wait()

	failedChecksFinal := atomic.LoadUint64(&failedChecks)
	if failedChecksFinal > 0 {
		logger.Infof("At least one health check failed. Returning HTTP 504 response.\n")
		return &httpResponse{ StatusCode: http.StatusGatewayTimeout, Body: "At least one health check failed" }
	} else {
		logger.Infof("All health checks passed. Returning HTTP 200 response.\n")
		return &httpResponse{ StatusCode: http.StatusOK, Body: "OK" }
	}
}

// Attempt to open a TCP connection to the given port
func attemptTcpConnection(host string, port int, opts *options.Options) error {
	logger := opts.Logger
	logger.Infof("Attempting to connect to port %d via TCP...", port)

	defaultTimeout := time.Second * 5

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), defaultTimeout)
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

