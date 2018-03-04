package server

import (
	"context"
	"fmt"
	gerrors "errors"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/health-checker/options"
	"github.com/sirupsen/logrus"
)

const DEFAULT_CHECK_TIMEOUT = 5

type TcpCheck struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port int `yaml:"port"`
	Timeout int `yaml:"timeout"`
}

type HttpCheck struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port int `yaml:"port"`
	SuccessStatusCodes []int `yaml:"success_status_codes"`
	BodyRegex string `yaml:"body_regex"`
	Timeout int `yaml:"timeout"`
}

type ScriptCheck struct {
	Name string `yaml:"name"`
	Script string `yaml:"script"`
	Timeout int `yaml:"timeout"`
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

func (c TcpCheck) ValidateCheck (logger *logrus.Logger) {
	if c.Name == "" {
		missingRequiredKey("tcp","name", logger)
	}
	if c.Host == "" {
		missingRequiredKey("tcp","host", logger)
	}
	if c.Port == 0 {
		missingRequiredKey("tcp","port", logger)
	}
}

func (c TcpCheck) DoCheck (opts *options.Options) error {
	logger := opts.Logger
	logger.Infof("Attempting to connect to %s at %s:%d via TCP...", c.Name, c.Host, c.Port)

	timeout := time.Second * DEFAULT_CHECK_TIMEOUT
	if c.Timeout != 0 {
		// override default with user defined timeout
		timeout = time.Second * time.Duration(c.Timeout)
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port), timeout)
	if err != nil {
		return err
	}

	defer conn.Close()

	return nil
}

func (c HttpCheck) ValidateCheck (logger *logrus.Logger) {
	if c.Name == "" {
		missingRequiredKey("http","name", logger)
	}
	if c.Host == "" {
		missingRequiredKey("http","host", logger)
	}
	if c.Port == 0 {
		missingRequiredKey("http","port", logger)
	}
}

func (c HttpCheck) DoCheck (opts *options.Options) error {
	logger := opts.Logger
	logger.Infof("Checking %s at %s:%d via HTTP...", c.Name, c.Host, c.Port)

	timeout := time.Second * DEFAULT_CHECK_TIMEOUT
	if c.Timeout != 0 {
		// override default with user defined timeout
		timeout = time.Second * time.Duration(c.Timeout)
	}

	client := http.Client{
		Timeout: timeout,
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
			return gerrors.New(fmt.Sprintf("http check %s wanted one of %v got %d", c.Name, c.SuccessStatusCodes, resp.Status))
		}
	} else if c.BodyRegex != ""{
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
			return gerrors.New(fmt.Sprintf("http check %s wanted %s in http body got %s", c.Name, c.BodyRegex, body))
		}
	} else {
s_codes or body_regex defined, only pass on 200
		if resp.StatusCode == http.StatusOK {
			return nil
		} else {
			return gerrors.New(fmt.Sprintf("http check %s wanted status code 200 got %d", c.Name, resp.StatusCode))
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

func missingRequiredKey(check string, key string, logger *logrus.Logger) {
	logger.Fatalf("Failed to parse YAML: %s check missing required key: %s", check, key)
}

func (c ScriptCheck) ValidateCheck (logger *logrus.Logger) {
	if c.Name == "" {
		missingRequiredKey("script","name", logger)
	}
	if c.Script == "" {
		missingRequiredKey("script","script", logger)
	}
}

func (c ScriptCheck) DoCheck (opts *options.Options) error {
	timeout := time.Second * DEFAULT_CHECK_TIMEOUT
	if c.Timeout != 0 {
		// override default with user defined timeout
		timeout = time.Second * time.Duration(c.Timeout)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.Script)
	_, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		// script timed out
		return gerrors.New(fmt.Sprintf("check %s at %s FAILED to complete within %ds", c.Name, c.Script, timeout))
	}
	if err != nil {
		return gerrors.New(fmt.Sprintf("check %s at %s FAILED with a non-zero exit code", c.Name, c.Script))
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
