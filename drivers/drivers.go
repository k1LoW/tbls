package drivers

import (
	"github.com/k1LoW/tbls/schema"
)

// Driver is the common interface for database drivers
type Driver interface {
	Analyze(*schema.Schema) error
	Info() (*schema.Driver, error)
}
