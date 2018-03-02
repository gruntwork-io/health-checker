package options

import (
	"github.com/sirupsen/logrus"
)

type Check interface {
	//ValidateCheck() error
	DoCheck(*Options) error
}

// The options accepted by this CLI tool
type Options struct {
	Checks   []Check
	Listener string
	Logger   *logrus.Logger
}
