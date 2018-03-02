package server

import (
	"fmt"
	gerrors "errors"
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

type TcpCheck struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port int `yaml:"port"`
}

type HttpCheck struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port int `yaml:"port"`
	SuccessStatusCodes []int `yaml:"success_status_codes"`
	BodyRegex string `yaml:"body_regex"`
}

type ScriptCheck struct {
	Script string `yaml:"script"`
	SuccessExitCodes []int `yaml:"success_exit_codes"`
}

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
	fmt.Printf("%v", opts.Checks)

	// initialize failedChecks to 0, used as atomic counter for goroutines below
	var failedChecks uint64
	var waitGroup = sync.WaitGroup{}

	for _, check := range opts.Checks {
		waitGroup.Add(1)
		go func(check options.Check) {
			defer waitGroup.Done()
			err := check.DoCheck(opts)
			if err != nil {
				logger.Warnf("Check for %s FAILED: %s", check, err)
				atomic.AddUint64(&failedChecks, 1)
			} else {
				logger.Infof("Check for %s successful", check)
			}
		}(check)
	}

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

func (c TcpCheck) DoCheck (opts *options.Options) error {
	logger := opts.Logger
	logger.Infof("Attempting to connect to %s at %s:%d via TCP...", c.Name, c.Host, c.Port)

	defaultTimeout := time.Second * 5

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port), defaultTimeout)
	if err != nil {
		return err
	}

	defer conn.Close()

	return nil
}

func (c HttpCheck) DoCheck (opts *options.Options) error {
	logger := opts.Logger
	logger.Infof("Checking %s at %s:%d via HTTP...", c.Name, c.Host, c.Port)

	defaultTimeout := time.Second * 5
	client := http.Client{
		Timeout: defaultTimeout,
	}
	resp, err := client.Get(fmt.Sprintf("http://%s:%d", c.Host, c.Port))
	if err != nil {
		return err
	}

	if len(c.SuccessStatusCodes) > 0 {
		// when success_codes is defined we only need to check this
		if contains(c.SuccessStatusCodes, resp.StatusCode) {
			// Success! response has one of the success_codes
			return nil
		} else {
			return gerrors.New(fmt.Sprintf("http status code %s was not one of %v", resp.StatusCode, c.SuccessStatusCodes))
		}
	} else {
		// since no success_codes defined we compare body with body_regex
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if strings.Contains(string(body), c.BodyRegex) {
			// Success! resp body has expected string
			return nil
		} else {
			return gerrors.New(fmt.Sprintf("expected %s in http body: %s", c.BodyRegex, body))
		}
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

func (c ScriptCheck) DoCheck (opts *options.Options) error {
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
