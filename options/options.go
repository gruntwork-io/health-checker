package options

import (
	"github.com/sirupsen/logrus"
)

type Check interface {
	DoCheck(*Options) error
	ValidateCheck(*logrus.Logger)
}

// The options accepted by this CLI tool
type Options struct {
	Checks   []Check
	Listener string
	Logger   *logrus.Logger
}
