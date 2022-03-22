package datasource

import (
	"context"

	"github.com/k1LoW/tbls/drivers/mongodb"
	"github.com/k1LoW/tbls/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AnalyzeMongodb analyze `mongodb://`
func AnalyzeMongodb(urlstr string) (*schema.Schema, error) {
	s := &schema.Schema{}

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(urlstr))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	driver, err := mongodb.New(ctx, client)
	if err != nil {
		return s, err
	}

	err = driver.Analyze(s)
	if err != nil {
		return s, err
	}
	return s, nil
}
