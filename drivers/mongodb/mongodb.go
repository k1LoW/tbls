package mongodb

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/k1LoW/errors"
	"github.com/SouhlInc/tbls/dict"
	"github.com/SouhlInc/tbls/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const columnTypeSeparator = ","

type Mongodb struct {
	ctx               context.Context
	client            *mongo.Client
	dbName            string
	sampleSize        int64
	multipleFieldType bool
}

func New(ctx context.Context, client *mongo.Client, dbName string, sampleSize int64, multipleFieldType bool) (*Mongodb, error) {
	return &Mongodb{
		ctx:               ctx,
		client:            client,
		dbName:            dbName,
		sampleSize:        sampleSize,
		multipleFieldType: multipleFieldType,
	}, nil
}

func (m *Mongodb) Analyze(s *schema.Schema) error {
	drv, err := m.Info()
	if err != nil {
		return errors.WithStack(err)
	}
	s.Driver = drv

	tables := []*schema.Table{}
	dbValue := m.client.Database(m.dbName)
	colls, err := dbValue.ListCollectionSpecifications(m.ctx, bson.D{})
	if err != nil {
		return err
	}
	for _, coll := range colls {
		colVal := dbValue.Collection(coll.Name)
		indexes, err := m.listIndexes(colVal)
		if err != nil {
			return err
		}
		estimated, err := colVal.EstimatedDocumentCount(m.ctx)
		if err != nil {
			return err
		}
		columns, err := m.listFields(colVal)
		if err != nil {
			return err
		}
		sort.Slice(columns, func(i, j int) bool {
			return columns[i].Name < columns[j].Name
		})
		table := &schema.Table{
			Name:    fmt.Sprintf("%s.%s", m.dbName, coll.Name),
			Type:    coll.Type,
			Columns: columns,
			Indexes: indexes,
			Comment: fmt.Sprintf("Count of documents is %d", estimated),
		}
		tables = append(tables, table)
	}
	s.Tables = tables

	return nil
}

func (m *Mongodb) listFields(collection *mongo.Collection) ([]*schema.Column, error) {
	pipeline := []bson.D{{{Key: "$sample", Value: bson.D{{Key: "size", Value: m.sampleSize}}}}}
	cursor, err := collection.Aggregate(m.ctx, pipeline)
	if err != nil {
		return nil, err
	}
	columns := []*schema.Column{}
	total := 0.0
	occurrences := map[string]float64{}
	for cursor.Next(m.ctx) {
		var result bson.D
		if err := cursor.Decode(&result); err != nil {
			return columns, err
		}
		total++
		for _, entry := range result {
			key, value := entry.Key, entry.Value
			var valueType string
			switch value.(type) {
			case string:
				valueType = "string"
			case int64:
				valueType = "int64"
			case primitive.D:
				valueType = "document"
			case primitive.DateTime:
				valueType = "datetime"
			case primitive.A:
				valueType = "array"
			case primitive.ObjectID:
				valueType = "objectId"
			default:
				valueType = fmt.Sprintf("%T", value)
			}
			column := &schema.Column{
				Name:     key,
				Type:     valueType,
				Nullable: false,
			}
			if val, ok := occurrences[key]; ok {
				occurrences[key] = val + 1
			} else {
				occurrences[key] = 1
			}
			if !columnInColumns(column, columns) {
				columns = append(columns, column)
			}

			if m.multipleFieldType {
				columns = addColumnType(columns, key, valueType)
			}
		}
	}
	for _, col := range columns {
		if stat, ok := occurrences[col.Name]; ok {
			col.Occurrences = sql.NullInt32{Int32: int32(stat), Valid: true}
			col.Percents = sql.NullFloat64{Float64: stat / total * 100, Valid: true}
		} else {
			return columns, fmt.Errorf("not able find %s in occurancies", col.Name)
		}
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return columns, nil
}

func columnInColumns(a *schema.Column, list []*schema.Column) bool {
	for _, b := range list {
		if b.Name == a.Name {
			return true
		}
	}
	return false
}

// addColumnType adds a new type to the specified column.
func addColumnType(list []*schema.Column, columnName, valueType string) []*schema.Column {
	columns := make([]*schema.Column, len(list))

	for i, col := range list {
		column := *col
		if column.Name == columnName {
			types := append(strings.Split(column.Type, columnTypeSeparator), valueType)
			slices.Sort(types)
			uniqTypes := slices.Compact(types)
			column.Type = strings.Join(uniqTypes, columnTypeSeparator)
		}

		columns[i] = &column
	}

	return columns
}

func (m *Mongodb) listIndexes(collection *mongo.Collection) ([]*schema.Index, error) {
	indexes := []*schema.Index{}
	indexSpec, err := collection.Indexes().ListSpecifications(m.ctx)
	if err != nil {
		return nil, err
	}
	for _, spec := range indexSpec {
		var isUnique string
		if spec.Unique != nil && *spec.Unique {
			isUnique = "Unique"
		} else {
			isUnique = "Non-unique"
		}
		comment := fmt.Sprintf("%s, Version %d", isUnique, spec.Version)
		Index := &schema.Index{
			Name:    spec.Name,
			Def:     spec.KeysDocument.String(),
			Comment: comment,
		}
		indexes = append(indexes, Index)
	}
	return indexes, nil
}

func (m *Mongodb) Info() (*schema.Driver, error) {
	dct := dict.New()
	dct.Merge(map[string]string{
		"Column":  "Attribute",
		"Columns": "Attributes",
		"Indexes": "Indexes",
	})

	driver := &schema.Driver{
		Name:            "mongodb",
		DatabaseVersion: "",
		Meta: &schema.DriverMeta{
			Dict: &dct,
		},
	}
	return driver, nil
}
