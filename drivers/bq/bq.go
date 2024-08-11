package bq

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/k1LoW/errors"
	"github.com/k1LoW/tbls/ddl"
	"github.com/k1LoW/tbls/dict"
	"github.com/k1LoW/tbls/schema"
)

// Bigquery struct
type Bigquery struct {
	ctx       context.Context
	client    *bigquery.Client
	datasetID string
}

// New return new Bigquery
func New(ctx context.Context, client *bigquery.Client, datasetID string) (*Bigquery, error) {
	return &Bigquery{
		ctx:       ctx,
		client:    client,
		datasetID: datasetID,
	}, nil
}

func (b *Bigquery) Analyze(s *schema.Schema) error {
	d, err := b.Info()
	if err != nil {
		return errors.WithStack(err)
	}
	s.Driver = d

	ds := b.client.Dataset(b.datasetID)
	m, err := ds.Metadata(b.ctx)
	if err != nil {
		return err
	}
	s.Desc = m.Description
	for k, v := range m.Labels {
		s.Labels = append(s.Labels, &schema.Label{Name: fmt.Sprintf("%s:%s", k, v), Virtual: false})
	}
	sort.SliceStable(s.Labels, func(i, j int) bool { return s.Labels[i].Name < s.Labels[j].Name })

	bt := ds.Tables(b.ctx)

	// tables
	tables := []*schema.Table{}
	for {
		t, err := bt.Next()
		if err != nil {
			if err.Error() == "no more items in iterator" {
				break
			}
			return err
		}
		m, err := t.Metadata(b.ctx)
		if err != nil {
			return err
		}
		labels := schema.Labels{}
		for k, v := range m.Labels {
			labels = append(labels, &schema.Label{Name: fmt.Sprintf("%s:%s", k, v), Virtual: false})
		}
		sort.SliceStable(labels, func(i, j int) bool { return labels[i].Name < labels[j].Name })

		splitted := strings.Split(m.FullID, fmt.Sprintf("%s.", b.datasetID))

		table := &schema.Table{
			Name:    strings.Join(splitted[1:], ""),
			Comment: m.Description,
			Type:    string(m.Type),
			Def:     m.ViewQuery,
			Columns: listColumns(m.Schema, ""),
			Labels:  labels,
		}

		tables = append(tables, table)
	}
	s.Tables = tables

	// referenced tables of view
	for _, t := range s.Tables {
		if t.Type != "VIEW" {
			continue
		}
		for _, rts := range ddl.ParseReferencedTables(t.Def) {
			splitted := strings.Split(rts, ".")
			rt, err := s.FindTableByName(splitted[len(splitted)-1])
			if err != nil {
				rt = &schema.Table{
					Name:     rts,
					External: true,
				}
			}
			t.ReferencedTables = append(t.ReferencedTables, rt)
		}
	}

	return nil
}

func listColumns(s bigquery.Schema, prefix string) []*schema.Column {
	columns := []*schema.Column{}
	for _, c := range s {
		name := fmt.Sprintf("%s%s", prefix, c.Name)
		column := &schema.Column{
			Name:     name,
			Comment:  c.Description,
			Nullable: !c.Required,
			Type:     string(c.Type),
			// TODO: c.Repeated
		}
		columns = append(columns, column)
		if len(c.Schema) > 0 {
			nestedColumns := listColumns(c.Schema, fmt.Sprintf("%s.", name))
			columns = append(columns, nestedColumns...)
		}
	}
	return columns
}

func (b *Bigquery) Info() (*schema.Driver, error) {
	dct := dict.New()
	dct.Merge(map[string]string{
		"Comment": "Description",
	})

	d := &schema.Driver{
		Name:            "bigquery",
		DatabaseVersion: "",
		Meta: &schema.DriverMeta{
			Dict: &dct,
		},
	}
	return d, nil
}
