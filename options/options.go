package options

import (
	"github.com/sirupsen/logrus"
)

type Tcp struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port int `yaml:"port"`
}

type Http struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port int `yaml:"port"`
	SuccessStatusCodes []int `yaml:"success_status_codes"`
	BodyRegex string `yaml:"body_regex"`
}

type Script struct {
	Script string `yaml:"script"`
	SuccessExitCodes []int `yaml:"success_exit_codes"`
}

type Checks struct {
	TcpChecks    []Tcp	  `yaml:"tcp"`
	HttpChecks   []Http	  `yaml:"http"`
	ScriptChecks []Script `yaml:"scripts"`
}

// The options accepted by this CLI tool
type Options struct {
	Checks   *Checks
	Listener string
	Logger   *logrus.Logger
}
