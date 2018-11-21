package output

import (
	"io"

	"github.com/k1LoW/tbls/schema"
)

// Output is interface for output
type Output interface {
	OutputSchema(wr io.Writer, s *schema.Schema) error
	OutputTable(wr io.Writer, s *schema.Table) error
}
