package config

import (
	"github.com/goccy/go-yaml"
	"github.com/SouhlInc/tbls/schema"
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

func (f Format) MarshalYAML() ([]byte, error) {
	if len(f.HideColumnsWithoutValues) == 0 {
		s := struct {
			Adjust                   bool `yaml:"adjust,omitempty"`
			Sort                     bool `yaml:"sort,omitempty"`
			Number                   bool `yaml:"number,omitempty"`
			ShowOnlyFirstParagraph   bool `yaml:"showOnlyFirstParagraph,omitempty"`
			HideColumnsWithoutValues bool `yaml:"hideColumnsWithoutValues,omitempty"`
		}{
			Adjust:                   f.Adjust,
			Sort:                     f.Sort,
			Number:                   f.Number,
			ShowOnlyFirstParagraph:   f.ShowOnlyFirstParagraph,
			HideColumnsWithoutValues: false,
		}
		return yaml.Marshal(s)
	}
	return yaml.Marshal(f)
}

func (f *Format) UnmarshalYAML(data []byte) error {
	s := struct {
		Adjust                   bool        `yaml:"adjust,omitempty"`
		Sort                     bool        `yaml:"sort,omitempty"`
		Number                   bool        `yaml:"number,omitempty"`
		ShowOnlyFirstParagraph   bool        `yaml:"showOnlyFirstParagraph,omitempty"`
		HideColumnsWithoutValues interface{} `yaml:"hideColumnsWithoutValues,omitempty"`
	}{}
	if err := yaml.Unmarshal(data, &s); err != nil {
		return err
	}
	f.Adjust = s.Adjust
	f.Sort = s.Sort
	f.Number = s.Number
	f.ShowOnlyFirstParagraph = s.ShowOnlyFirstParagraph
	switch v := s.HideColumnsWithoutValues.(type) {
	case bool:
		if v {
			f.HideColumnsWithoutValues = schema.HideableColumns
		}
	case []interface{}:
		values := []string{}
		for _, vv := range v {
			if str, ok := vv.(string); ok {
				values = append(values, str)
			}
		}
		f.HideColumnsWithoutValues = values
	}
	return nil
}
