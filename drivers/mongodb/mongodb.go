package mongodb

import (
	"context"
	"fmt"

	"github.com/k1LoW/tbls/dict"
	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Mongodb struct {
	ctx    context.Context
	client *mongo.Client
	dbName string
}

func New(ctx context.Context, client *mongo.Client, dbName string) (*Mongodb, error) {
	return &Mongodb{
		ctx:    ctx,
		client: client,
		dbName: dbName,
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
			table := &schema.Table{
				Name:    fmt.Sprintf("%s.%s", dbName, coll.Name),
				Type:    coll.Type,
				Indexes: indexes,
				Comment: fmt.Sprintf("Estimated count of documents is %d", estimated),
			}
			tables = append(tables, table)
		}
	}
	s.Tables = tables

	return nil
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
