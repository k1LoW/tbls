package drivers

import (
	"database/sql"

	"github.com/k1LoW/tbls/schema"
)

// Driver is the common interface for database drivers
type Driver interface {
	Analyze(*sql.DB, *schema.Schema) error
	Info(*sql.DB) (*schema.Driver, error)
}
