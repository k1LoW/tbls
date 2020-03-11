package dict

import "sync"

type Dict struct {
	s sync.Map
}

// NewDict return Dict
func NewDict() Dict {
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

func (d *Dict) Dump() map[string]string {
	dpd := make(map[string]string)
	d.s.Range(func(k, v interface{}) bool {
		dpd[k.(string)] = v.(string)
		return true
	})
	return dpd
}
