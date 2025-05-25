package json

import (
	"encoding/json"
	"io"

	"github.com/SouhlInc/tbls/schema"
)

// JSON struct.
type JSON struct {
	inline bool
}

// New returns JSON.
func New(inline bool) *JSON {
	return &JSON{
		inline: inline,
	}
}

// OutputSchema output JSON format for full relation.
func (j *JSON) OutputSchema(wr io.Writer, s *schema.Schema) error {
	encoder := json.NewEncoder(wr)
	if !j.inline {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(s); err != nil {
		return err
	}
	return nil
}

// OutputTable output JSON format for table.
func (j *JSON) OutputTable(wr io.Writer, t *schema.Table) error {
	encoder := json.NewEncoder(wr)
	if !j.inline {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(t); err != nil {
		return err
	}
	return nil
}
