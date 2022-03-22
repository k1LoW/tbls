package datasource

import (
	"context"
	"net/url"

	"github.com/k1LoW/tbls/drivers/mongodb"
	"github.com/k1LoW/tbls/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AnalyzeMongodb analyze `mongodb://`
func AnalyzeMongodb(urlstr string) (*schema.Schema, error) {
	s := &schema.Schema{}
	u, err := url.Parse(urlstr)
	if err != nil {
		return s, err
	}
	values := u.Query()

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(urlstr))
	if err != nil {
		return s, err
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	driver, err := mongodb.New(ctx, client, values.Get("dbName"))
	if err != nil {
		return s, err
	}

	err = driver.Analyze(s)
	if err != nil {
		return s, err
	}
	return s, nil
}
