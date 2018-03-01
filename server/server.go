package server

import (
	gerrors "errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/health-checker/options"
)

type httpResponse struct {
	StatusCode int
	Body       string
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

	// initialize failedChecks to 0, used as atomic counter for goroutines below
	var failedChecks uint64
	var waitGroup = sync.WaitGroup{}

	for _, tcpCheck := range opts.Checks.TcpChecks {
		name := tcpCheck.Name
		host := tcpCheck.Host
		port := tcpCheck.Port
		waitGroup.Add(1)
		go func(name string, host string, port int) {
			err := checkTcpConnection(name, host, port, opts)
			if err != nil {
				logger.Warnf("TCP connection to %s at %s:%d FAILED: %s", name, host, port, err)
				atomic.AddUint64(&failedChecks, 1)
			} else {
				logger.Infof("TCP connection to %s at %s:%d successful", name, host, port)
			}
			defer waitGroup.Done()
		}(name, host, port)
	}

	for _, httpCheck := range opts.Checks.HttpChecks {
		name := httpCheck.Name
		host := httpCheck.Host
		port := httpCheck.Port
		successCodes := httpCheck.SuccessStatusCodes
		expected := httpCheck.BodyRegex
		waitGroup.Add(1)
		go func(name string, host string, port int, successCodes []int, expected string) {
			if len(successCodes) > 0 {
				err := checkHttpResponse(name, host, port, successCodes, opts)
				if err != nil {
					logger.Warnf("HTTP Status check to %s at %s:%d FAILED: %s", name, host, port, err)
					atomic.AddUint64(&failedChecks, 1)
				} else {
					logger.Infof("HTTP Status check to %s at %s:%d successful", name, host, port)
				}
			} else if len(expected) > 0 {
				err := checkHttpResponseBody(name, host, port, expected, opts)
				if err != nil {
					logger.Warnf("HTTP Body check to %s at %s:%d FAILED: %s", name, host, port, err)
					atomic.AddUint64(&failedChecks, 1)
				} else {
					logger.Infof("HTTP Body check to %s at %s:%d successful", name, host, port)
				}
			} else {
				logger.Warnf("FAILED: At least one of success_codes or body_regex not specified for %s", name)
				atomic.AddUint64(&failedChecks, 1)
			}
			defer waitGroup.Done()
		}(name, host, port, successCodes, expected)
	}

	// TODO: implement scriptCheck logic

	waitGroup.Wait()

	failedChecksFinal := atomic.LoadUint64(&failedChecks)
	if failedChecksFinal > 0 {
		logger.Infof("At least one health check failed. Returning HTTP 504 response.\n")
		return &httpResponse{StatusCode: http.StatusGatewayTimeout, Body: "At least one health check failed"}
	} else {
		logger.Infof("All health checks passed. Returning HTTP 200 response.\n")
		return &httpResponse{StatusCode: http.StatusOK, Body: "OK"}
	}
}

func checkTcpConnection(name string, host string, port int, opts *options.Options) error {
	logger := opts.Logger
	logger.Infof("Attempting to connect to %s at %s:%d via TCP...", name, host, port)

	defaultTimeout := time.Second * 5

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), defaultTimeout)
	if err != nil {
		return err
	}

	defer conn.Close()

	return nil
}

func checkHttpResponse(name string, host string, port int, successCodes []int, opts *options.Options) error {
	logger := opts.Logger
	logger.Infof("Checking %s at %s:%d via HTTP...", name, host, port)

	defaultTimeout := time.Second * 5
	client := http.Client{
		Timeout: defaultTimeout,
	}
	resp, err := client.Get(fmt.Sprintf("http://%s:%d", host, port))
	if err != nil {
		return err
	}

	if contains(successCodes, resp.StatusCode) {
		// Success! resp has one of the success_codes
		return nil
	} else {
		return gerrors.New(fmt.Sprintf("http status code %s was not one of %v", resp.StatusCode, successCodes))
	}
}

// TODO: move into helpers
func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func checkHttpResponseBody(name string, host string, port int, expected string, opts *options.Options) error {
	logger := opts.Logger
	logger.Infof("Checking HTTP response body for %s at %s:%d...", name, host, port)

	defaultTimeout := time.Second * 5
	client := http.Client{
		Timeout: defaultTimeout,
	}
	resp, err := client.Get(fmt.Sprintf("http://%s:%d", host, port))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if strings.Contains(string(body), expected) {
		// Success! resp body has expected string
		return nil
	} else {
		return gerrors.New(fmt.Sprintf("expected %s in http body: %s", expected, body))
	}
}

//func checkScript(script string, expectedExitStatus int, opts *options.Options) error {
//	logger := opts.Logger
//	logger.Infof("Checking script %s for exit status %d...", script, expectedExitStatus)

//defaultTimeout := time.Second * 5

// TODO: add code here
//}

func writeHttpResponse(w http.ResponseWriter, resp *httpResponse) error {
	w.WriteHeader(resp.StatusCode)
	_, err := w.Write([]byte(resp.Body))
	if err != nil {
		return errors.WithStackTrace(err)
	}

	return nil
}
