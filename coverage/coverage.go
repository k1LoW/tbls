package coverage

import (
	"math"

	"github.com/k1LoW/tbls/output/config"
	"github.com/k1LoW/tbls/schema"
)

type Coverage struct {
	Name     string           `json:"name"`
	Coverage float64          `json:"coverage"`
	Tables   []*TableCoverage `json:"tables"`
	Covered  int              `json:"-"`
	Total    int              `json:"-"`
}

type TableCoverage struct {
	Name     string  `json:"name"`
	Coverage float64 `json:"coverage"`
	Covered  int     `json:"-"`
	Total    int     `json:"-"`
}

// Measure coverage
func Measure(s *schema.Schema) *Coverage {
	cover := &Coverage{
		Name: s.Name,
	}
	// schema
	cover.Total += 1
	if s.Desc != "" {
		cover.Covered += 1
	}

	// tables
	for _, t := range s.Tables {
		tcover := &TableCoverage{
			Name: t.Name,
		}
		cover.Tables = append(cover.Tables, tcover)

		cover.Total += 1
		tcover.Total += 1
		if t.Comment != "" && t.Comment != config.NoTableComment {
			cover.Covered += 1
			tcover.Covered += 1
		}

		for _, c := range t.Columns {
			cover.Total += 1
			tcover.Total += 1
			if c.Comment != "" && c.Comment != config.NoColumnComment {
				cover.Covered += 1
				tcover.Covered += 1
			}
		}

		for _, i := range t.Indexes {
			cover.Total += 1
			tcover.Total += 1
			if i.Comment != "" {
				cover.Covered += 1
				tcover.Covered += 1
			}
		}

		for _, c := range t.Constraints {
			cover.Total += 1
			tcover.Total += 1
			if c.Comment != "" {
				cover.Covered += 1
				tcover.Covered += 1
			}
		}

		for _, trig := range t.Triggers {
			cover.Total += 1
			tcover.Total += 1
			if trig.Comment != "" {
				cover.Covered += 1
				tcover.Covered += 1
			}
		}

		tcover.Coverage = round(float64(tcover.Covered) / float64(tcover.Total) * 100)
	}

	cover.Coverage = round(float64(cover.Covered) / float64(cover.Total) * 100)
	return cover
}

func round(f float64) float64 {
	places := 1
	shift := math.Pow(10, float64(places))
	return math.Floor(f*shift+.5) / shift
}
