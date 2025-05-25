package drivers

import (
	"github.com/SouhlInc/tbls/schema"
)

// Driver is the common interface for database drivers.
type Driver interface {
	Analyze(*schema.Schema) error
	Info() (*schema.Driver, error)
}

// Option is the type for change Config.
type Option func(Driver) error
