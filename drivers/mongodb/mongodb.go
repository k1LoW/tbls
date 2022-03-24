package mongodb

import (
	"context"
	"fmt"

	"github.com/k1LoW/tbls/dict"
	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Mongodb struct {
	ctx        context.Context
	client     *mongo.Client
	dbName     string
	sampleSize int64
}

func New(ctx context.Context, client *mongo.Client, dbName string, sampleSize int64) (*Mongodb, error) {
	return &Mongodb{
		ctx:        ctx,
		client:     client,
		dbName:     dbName,
		sampleSize: sampleSize,
	}, nil
}

func (d *Mongodb) getDatabaseNames() ([]string, error) {
	if d.dbName != "" {
		return []string{d.dbName}, nil
	} else {
		dbNames, err := d.client.ListDatabaseNames(d.ctx, bson.D{})
		if err != nil {
			return nil, err
		}
		return dbNames, nil
	}
}

func (d *Mongodb) Analyze(s *schema.Schema) error {
	drv, err := d.Info()
	if err != nil {
		return errors.WithStack(err)
	}
	s.Driver = drv

	tables := []*schema.Table{}
	dbNames, err := d.getDatabaseNames()
	for _, dbName := range dbNames {
		dbValue := d.client.Database(dbName)
		colls, err := dbValue.ListCollectionSpecifications(d.ctx, bson.D{})
		if err != nil {
			return err
		}
		for _, coll := range colls {
			colVal := dbValue.Collection(coll.Name)
			indexes, err := d.listIndexes(colVal)
			if err != nil {
				return err
			}
			estimated, err := colVal.EstimatedDocumentCount(d.ctx)
			if err != nil {
				return err
			}
			columns, err := d.listFields(colVal)
			if err != nil {
				return err
			}
			table := &schema.Table{
				Name:    fmt.Sprintf("%s.%s", dbName, coll.Name),
				Type:    coll.Type,
				Columns: columns,
				Indexes: indexes,
				Comment: fmt.Sprintf("Count of documents is %d", estimated),
			}
			tables = append(tables, table)
		}
	}
	s.Tables = tables

	return nil
}

func (d *Mongodb) listFields(collection *mongo.Collection) ([]*schema.Column, error) {
	pipeline := []bson.D{{{Key: "$sample", Value: bson.D{{Key: "size", Value: d.sampleSize}}}}}
	cursor, err := collection.Aggregate(d.ctx, pipeline)
	if err != nil {
		return nil, err
	}
	columns := []*schema.Column{}
	for cursor.Next(d.ctx) {
		var result bson.D
		if err := cursor.Decode(&result); err != nil {
			return columns, err
		}
		for key, value := range result.Map() {
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
			if !columnInColumns(column, columns) {
				columns = append(columns, column)
			}
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

func (d *Mongodb) listIndexes(collection *mongo.Collection) ([]*schema.Index, error) {
	indexes := []*schema.Index{}
	indexSpec, err := collection.Indexes().ListSpecifications(d.ctx)
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

func (d *Mongodb) Info() (*schema.Driver, error) {
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
