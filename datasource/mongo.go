package datasource

import (
	"context"
	"net/url"
	"strconv"

	"github.com/k1LoW/tbls/drivers/mongodb"
	"github.com/k1LoW/tbls/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const defaultSampleSize = 1000

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

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return s, err
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	sampleSize, err := strconv.ParseInt(values.Get("sampleSize"), 10, 0)
	if err != nil {
		sampleSize = defaultSampleSize
	}
	driver, err := mongodb.New(ctx, client, values.Get("dbName"), sampleSize)
	if err != nil {
		return s, err
	}

	err = driver.Analyze(s)
	if err != nil {
		return s, err
	}
	return s, nil
}
