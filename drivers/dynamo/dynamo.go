package dynamo

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/k1LoW/errors"
	"github.com/k1LoW/tbls/dict"
	"github.com/k1LoW/tbls/schema"
)

var re = regexp.MustCompile(`(?s)\n\s*`)

type Dynamodb struct {
	ctx    context.Context
	client *dynamodb.DynamoDB
}

func New(ctx context.Context, client *dynamodb.DynamoDB) (*Dynamodb, error) {
	return &Dynamodb{
		ctx:    ctx,
		client: client,
	}, nil
}

func (d *Dynamodb) Analyze(s *schema.Schema) error {
	drv, err := d.Info()
	if err != nil {
		return errors.WithStack(err)
	}
	s.Driver = drv

	input := &dynamodb.ListTablesInput{}

	// tables
	tables := []*schema.Table{}
	tableType := "BASIC TABLE"
	for {
		list, err := d.client.ListTablesWithContext(d.ctx, input)
		if err != nil {
			return err
		}

		for _, t := range list.TableNames {
			input := &dynamodb.DescribeTableInput{
				TableName: t,
			}
			desc, err := d.client.DescribeTableWithContext(d.ctx, input)
			if err != nil {
				return err
			}
			table := &schema.Table{
				Name:        *desc.Table.TableName,
				Type:        tableType,
				Columns:     listColumns(desc.Table),
				Constraints: listConstraints(desc.Table),
				Indexes:     listIndexes(desc.Table),
			}
			tables = append(tables, table)
		}

		input.ExclusiveStartTableName = list.LastEvaluatedTableName

		if list.LastEvaluatedTableName == nil {
			break
		}
	}

	s.Tables = tables

	return nil
}

func listColumns(td *dynamodb.TableDescription) []*schema.Column {
	columns := []*schema.Column{}
	for _, ad := range td.AttributeDefinitions {
		column := &schema.Column{
			Name:     *ad.AttributeName,
			Type:     *ad.AttributeType,
			Nullable: false,
		}
		columns = append(columns, column)
	}
	return columns
}

func listConstraints(td *dynamodb.TableDescription) []*schema.Constraint {
	constraints := []*schema.Constraint{}
	switch {
	case len(td.KeySchema) == 2:
		columns := []string{}
		for _, k := range td.KeySchema {
			columns = append(columns, *k.AttributeName)
		}
		def := re.ReplaceAllString(fmt.Sprintf("%v", td.KeySchema), " ")
		constraint := &schema.Constraint{
			Name:    "Primary Key",
			Type:    "Partition key and sort key",
			Def:     def,
			Columns: columns,
		}
		constraints = append(constraints, constraint)
	case len(td.KeySchema) == 1:
		columns := []string{}
		for _, k := range td.KeySchema {
			columns = append(columns, *k.AttributeName)
		}
		def := re.ReplaceAllString(fmt.Sprintf("%v", td.KeySchema), " ")
		constraint := &schema.Constraint{
			Name:    "Primary Key",
			Type:    "Partition key",
			Def:     def,
			Columns: columns,
		}
		constraints = append(constraints, constraint)
	}
	return constraints
}

func listIndexes(td *dynamodb.TableDescription) []*schema.Index {
	indexes := []*schema.Index{}
	for _, lsi := range td.LocalSecondaryIndexes {
		def := re.ReplaceAllString(fmt.Sprintf("LocalSecondaryIndex { %s, %s }", lsi.KeySchema, lsi.Projection.String()), " ")
		Index := &schema.Index{
			Name: *lsi.IndexName,
			Def:  def,
		}
		indexes = append(indexes, Index)
	}
	for _, gsi := range td.GlobalSecondaryIndexes {
		def := re.ReplaceAllString(fmt.Sprintf("GlobalSecondaryIndex { %s, %s }", gsi.KeySchema, gsi.Projection.String()), " ")
		Index := &schema.Index{
			Name: *gsi.IndexName,
			Def:  def,
		}
		indexes = append(indexes, Index)
	}
	return indexes
}

func (d *Dynamodb) Info() (*schema.Driver, error) {
	dct := dict.New()
	dct.Merge(map[string]string{
		"Column":      "Attribute",
		"Columns":     "Attributes",
		"Constraints": "Primary Key",
		"Indexes":     "Secondary Indexes",
	})

	driver := &schema.Driver{
		Name:            "dynamodb",
		DatabaseVersion: "",
		Meta: &schema.DriverMeta{
			Dict: &dct,
		},
	}
	return driver, nil
}
