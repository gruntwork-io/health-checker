package options

import "github.com/sirupsen/logrus"

// The common options that apply to all CLI commands
type Options struct {
	Ports    []int
	Listener string
	Logger   *logrus.Logger
}
