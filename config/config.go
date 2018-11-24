package config

import (
	"os"

	"github.com/pkg/errors"
)

// Config ...
type Config struct {
	DSN     string
	DocPath string
}

// NewConfig return Config
func NewConfig() (*Config, error) {
	docPath := os.Getenv("TBLS_DOC_PATH")
	if docPath == "" {
		docPath = "."
	}

	c := Config{
		DSN:     os.Getenv("TBLS_DSN"),
		DocPath: docPath,
	}
	return &c, nil
}

// LoadArgs load args slice
func (c *Config) LoadArgs(args []string) error {
	if len(args) == 2 {
		c.DSN = args[0]
		c.DocPath = args[1]
	}
	if len(args) > 2 {
		return errors.WithStack(errors.New("requires two args"))
	}
	if len(args) == 1 {
		if c.DSN == "" {
			c.DSN = args[0]
		} else {
			c.DocPath = args[0]
		}
	}
	return nil
}
