package component

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type LogrusKey struct{}

type LogrusStdOut struct {
	initialized bool
	logger      *logrus.Logger
}

func Logrus(reg *Registry) (LogrusComponent, error) {
	if reg.Logrus == nil {
		return nil, errors.New("logrus is not set")
	}
	return reg.Logrus, nil
}

// NewLog creates a new component
func NewLogrusStdOut() *LogrusStdOut {
	return &LogrusStdOut{initialized: false}
}

func (c *LogrusStdOut) Priority() int8 {
	return 10
}

func (c *LogrusStdOut) Key() interface{} {
	return LogrusKey{}
}

func (c *LogrusStdOut) Name() string {
	return "shared.log"
}

func (c *LogrusStdOut) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "loglevel",
			Value:   "info",
			Usage:   "Logrus log level default 'info', {panic,fatal,error,warn,info,debug,trace} available",
			EnvVars: []string{"LOG_LEVEL"},
		},
	}
}

func (c *LogrusStdOut) Initialized() bool {
	return c.initialized
}

func (c *LogrusStdOut) Init(registry *Registry, cli *cli.Context) error {
	if c.initialized {
		return nil
	}

	lvl, err := logrus.ParseLevel(cli.String("loglevel"))
	if err != nil {
		return err
	}

	c.logger = logrus.New()
	c.logger.Out = os.Stdout
	c.logger.Level = lvl

	// Hack to set the Logger in the registry for LogrusFromRegistry
	registry.Logrus = c

	c.initialized = true
	return nil
}

func (c *LogrusStdOut) Health(context context.Context) (string, bool) {
	if !c.Initialized() {
		return "Not initialized", true
	}

	return "All fine", false
}

func (c *LogrusStdOut) Logger() *logrus.Logger {
	return c.logger
}

// WithFunc creates a logger with pkgpath and function
func (c *LogrusStdOut) WithFunc(pkgPath string, function string) *logrus.Entry {
	return c.logger.WithFields(
		logrus.Fields{
			"func": fmt.Sprintf("(%s).%s", pkgPath, function),
		})
}

// WithClassFunc creates a logger with pkgpath and function
func (c *LogrusStdOut) WithClassFunc(pkgPath, class, function string) *logrus.Entry {
	return c.logger.WithFields(
		logrus.Fields{
			"func": fmt.Sprintf("(%s.%s).%s", pkgPath, class, function),
		})
}
