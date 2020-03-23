package dict

import "testing"

func TestDelete(t *testing.T) {
	tests := []struct {
		in   map[string]string
		want int
	}{
		{
			in:   map[string]string{"a": "1", "b": "1", "c": "1"},
			want: 2,
		},
		{
			in:   map[string]string{"b": "1", "c": "1", "d": "1"},
			want: 3,
		},
	}

	for _, tt := range tests {
		a := New()
		for k, v := range tt.in {
			a.Store(k, v)
		}
		a.Delete("a")

		if len(a.Dump()) != tt.want {
			t.Errorf("got %v\nwant %v", len(a.Dump()), tt.want)
		}
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		ain  map[string]string
		bin  map[string]string
		want int
	}{
		{
			ain:  map[string]string{"a": "1"},
			bin:  map[string]string{"b": "2"},
			want: 2,
		},
		{
			ain:  map[string]string{"a": "1", "b": "1"},
			bin:  map[string]string{"b": "2"},
			want: 2,
		},
		{
			ain:  map[string]string{"a": "1", "b": "1", "c": "1"},
			bin:  map[string]string{"b": "2"},
			want: 3,
		},
	}

	for _, tt := range tests {
		a := New()
		b := New()
		for k, v := range tt.ain {
			a.Store(k, v)
		}
		for k, v := range tt.bin {
			b.Store(k, v)
		}
		a.Merge(b.Dump())

		dumped := a.Dump()

		if len(dumped) != tt.want {
			t.Errorf("got %v\nwant %v", len(dumped), tt.want)
		}

		if want := "1"; a.Lookup("a") != want {
			t.Errorf("got %v\nwant %v", a.Lookup("a"), want)
		}

		if want := "2"; a.Lookup("b") != want {
			t.Errorf("got %v\nwant %v", a.Lookup("b"), want)
		}
	}
}

func TestMergeIfNotPresent(t *testing.T) {
	tests := []struct {
		ain  map[string]string
		bin  map[string]string
		want int
	}{
		{
			ain:  map[string]string{"a": "1", "b": "1"},
			bin:  map[string]string{"b": "2"},
			want: 2,
		},
		{
			ain:  map[string]string{"a": "1", "b": "1", "c": "1"},
			bin:  map[string]string{"b": "2"},
			want: 3,
		},
	}

	for _, tt := range tests {
		a := New()
		b := New()
		for k, v := range tt.ain {
			a.Store(k, v)
		}
		for k, v := range tt.bin {
			b.Store(k, v)
		}
		a.MergeIfNotPresent(b.Dump())

		dumped := a.Dump()

		if len(dumped) != tt.want {
			t.Errorf("got %v\nwant %v", len(dumped), tt.want)
		}

		if want := "1"; a.Lookup("a") != want {
			t.Errorf("got %v\nwant %v", a.Lookup("a"), want)
		}

		if want := "1"; a.Lookup("b") != want {
			t.Errorf("got %v\nwant %v", a.Lookup("b"), want)
		}
	}
}
