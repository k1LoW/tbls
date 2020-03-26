package config

import (
	"github.com/goccy/go-yaml"
)

func (d DSN) MarshalYAML() ([]byte, error) {
	if len(d.Headers) == 0 {
		dsn := d.URL
		return yaml.Marshal(dsn)
	}
	return yaml.Marshal(d)
}

func (d *DSN) UnmarshalYAML(data []byte) error {
	var dsn interface{}
	if err := yaml.Unmarshal(data, &dsn); err != nil {
		return err
	}
	switch raw := dsn.(type) {
	case string:
		d.URL = raw
	case interface{}:
		if err := yaml.Unmarshal(data, d); err != nil {
			return err
		}
	}
	return nil
}
