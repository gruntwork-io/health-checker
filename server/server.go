package server

import (
	"context"
	"fmt"
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
				logger.Warnf(err.Error())
				atomic.AddUint64(&failedChecks, 1)
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

func (c TcpCheck) ValidateCheck () error {
	if c.Name == "" {
		return &InvalidCheck{name: "tcp", key: "name"}
	}
	if c.Host == "" {
		return &InvalidCheck{name: c.Name, key: "host"}
	}
	if c.Port == 0 {
		return &InvalidCheck{name: c.Name, key: "port"}
	}
	return nil
}

func (c TcpCheck) DoCheck (opts *options.Options) error {
	logger := opts.Logger
	logger.Infof("Checking %s at %s:%d via TCP...", c.Name, c.Host, c.Port)

	timeout := time.Second * DEFAULT_CHECK_TIMEOUT
	if c.Timeout != 0 {
		// override default with user defined timeout
		timeout = time.Second * time.Duration(c.Timeout)
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port), timeout)
	if err != nil {
		return &CheckFail{name: c.Name, reason: err.Error()}
	}

	defer conn.Close()
	logger.Infof("Check SUCCESS: %s", c.Name)
	return nil
}

func (c HttpCheck) ValidateCheck () error {
	if c.Name == "" {
		return &InvalidCheck{name: "http", key: "name"}
	}
	if c.Host == "" {
		return &InvalidCheck{name: c.Name, key: "host"}
	}
	if c.Port == 0 {
		return &InvalidCheck{name: c.Name, key: "port"}
	}
	return nil
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
		return &CheckFail{name: c.Name, reason: err.Error()}
	}

	switch {
	case len(c.SuccessStatusCodes) > 0:
		if !contains(c.SuccessStatusCodes, resp.StatusCode) {
			return &CheckFail{name: c.Name, reason: fmt.Sprintf("wanted one of %v, got %d", c.SuccessStatusCodes, resp.Status)}
		}
	case c.BodyRegex != "":
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return &CheckFail{name: c.Name, reason: "failed reading body"}
		}

		if !strings.Contains(string(body), c.BodyRegex) {
			return &CheckFail{name: c.Name, reason: fmt.Sprintf("wanted %s in http body, got %s", c.BodyRegex, body)}
		}
	default:
		// no success_codes or body_regex defined, only pass on 200
		if resp.StatusCode != http.StatusOK {
			return &CheckFail{name: c.Name, reason: fmt.Sprintf("wanted status code 200, got %d", resp.StatusCode)}
		}
	}

	logger.Infof("Check SUCCESS: %s", c.Name)
	return nil
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

func (c ScriptCheck) ValidateCheck () error {
	if c.Name == "" {
		return &InvalidCheck{name: "script", key: "name"}
	}
	if c.Script == "" {
		return &InvalidCheck{name: c.Name, key: "script"}
	}
	return nil
}

func (c ScriptCheck) DoCheck (opts *options.Options) error {
	logger := opts.Logger
	logger.Infof("Checking %s at %s...", c.Name, c.Script)

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
		return &CheckTimeout{name: c.Name, timeout: int(timeout)}
	}
	if err != nil {
		return &CheckFail{name: c.Name, reason: "non-zero exit code"}
	}
	logger.Infof("Check SUCCESS: %s", c.Name)
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

// custom error types
type InvalidCheck struct {
	name, key string
}

func (e *InvalidCheck) Error() string {
	return fmt.Sprintf("Invalid check: %s, missing key: %s", string(e.name), string(e.key))
}

type CheckFail struct {
	name, reason	string
}

func (e *CheckFail) Error() string {
	return fmt.Sprintf("Check FAILED: %s reason: %s", string(e.name), string(e.reason))
}

type CheckTimeout struct {
	name	string
	timeout int
}

func (e *CheckTimeout) Error() string {
	return fmt.Sprintf("Check TIMEOUT: %s took longer than configured timeout: %ds", string(e.name), int(e.timeout))
}
