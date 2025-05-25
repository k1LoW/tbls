package datasource

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/SouhlInc/tbls/drivers/mongodb"
	"github.com/SouhlInc/tbls/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	defaultSampleSize        = 1000
	defaultMultipleFieldType = false
)

// AnalyzeMongodb analyze `mongodb://`
func AnalyzeMongodb(urlstr string) (*schema.Schema, error) {
	s := &schema.Schema{}
	u, err := url.Parse(urlstr)
	if err != nil {
		return s, err
	}
	values := u.Query()
	parsedPath := strings.Split(u.Path, "/")
	if len(parsedPath) != 2 {
		return nil, errors.New("no database name in the connection string")
	}
	dbName := parsedPath[1]

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
	multipleFieldType, err := strconv.ParseBool(values.Get("multipleFieldType"))
	if err != nil {
		multipleFieldType = defaultMultipleFieldType
	}
	driver, err := mongodb.New(ctx, client, dbName, sampleSize, multipleFieldType)
	if err != nil {
		return s, err
	}

	err = driver.Analyze(s)
	if err != nil {
		return s, err
	}
	return s, nil
}
