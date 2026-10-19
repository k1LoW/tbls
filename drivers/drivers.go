package drivers

import (
	"github.com/k1LoW/tbls/schema"
)

// Driver is the common interface for database drivers.
type Driver interface {
	Analyze(*schema.Schema) error
	Info() (*schema.Driver, error)
}

// Option is the type for change Config.
type Option func(Driver) error

// TableFilterer is implemented by drivers that can apply table-name
// exclusion at the query stage as a perf optimization. The post-fetch
// schema.Filter remains authoritative; drivers that under-fetch will
// silently lose tables, so implementations must only push down patterns
// they can faithfully translate.
type TableFilterer interface {
	SetTableExcludes(patterns []string)
}
