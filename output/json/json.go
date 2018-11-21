package json

import (
	"encoding/json"
	"io"

	"github.com/k1LoW/tbls/schema"
)

// JSON struct
type JSON struct{}

// OutputSchema output JSON format for full relation.
func (j *JSON) OutputSchema(wr io.Writer, s *schema.Schema) error {
	encoder := json.NewEncoder(wr)
	encoder.SetIndent("", "  ")
	encoder.Encode(s)
	return nil
}

// OutputTable output dot format for table.
func (j *JSON) OutputTable(wr io.Writer, t *schema.Table) error {
	encoder := json.NewEncoder(wr)
	encoder.SetIndent("", "  ")
	encoder.Encode(t)
	return nil
}
