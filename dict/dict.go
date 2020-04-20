package dict

import (
	"encoding/json"
	"sync"

	"github.com/goccy/go-yaml"
)

type Dict struct {
	s sync.Map `json:"-" yaml:"-"`
}

// New return Dict
func New() Dict {
	return Dict{}
}

func (d *Dict) Lookup(k string) string {
	if v, ok := d.s.Load(k); ok {
		return v.(string)
	}
	return k
}

func (d *Dict) Store(k, v string) {
	d.s.Store(k, v)
}

func (d *Dict) Delete(k string) {
	d.s.Delete(k)
}

func (d *Dict) Range(f func(key, value interface{}) bool) {
	d.s.Range(f)
}

func (d *Dict) Merge(in map[string]string) {
	for k, v := range in {
		d.s.Store(k, v)
	}
}

func (d *Dict) MergeIfNotPresent(in map[string]string) {
	for k, v := range in {
		_, _ = d.s.LoadOrStore(k, v)
	}
}

func (d *Dict) Dump() map[string]string {
	dpd := make(map[string]string)
	d.s.Range(func(k, v interface{}) bool {
		dpd[k.(string)] = v.(string)
		return true
	})
	return dpd
}

func (d *Dict) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Dump())
}

func (d *Dict) UnmarshalJSON(data []byte) error {
	m := map[string]string{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	for k, v := range m {
		d.s.Store(k, v)
	}
	return nil
}

func (d *Dict) MarchalYAML() ([]byte, error) {
	return yaml.Marshal(d.Dump())
}

func (d *Dict) UnmarshalYAML(data []byte) error {
	m := map[string]string{}
	err := yaml.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	for k, v := range m {
		d.s.Store(k, v)
	}
	return nil
}
