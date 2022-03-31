package mongodb

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/k1LoW/tbls/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// it is expected to have running https://hub.docker.com/r/weshigbee/docker-mongo-sample-datasets
func TestAnalyze(t *testing.T) {
	ctx := context.Background()
	urlstr := "mongodb://mongoadmin:secret@localhost:27017/test?authSource=admin"
	u, err := url.Parse(urlstr)
	if err != nil {
		t.Errorf("%v", err)
	}
	parsedPath := strings.Split(u.Path, "/")
	if len(parsedPath) != 2 {
		t.Error("No database name in the connection string")
	}
	dbName := parsedPath[1]
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(urlstr))
	if err != nil {
		t.Errorf("%v", err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	s := &schema.Schema{
		Name: "MongoDB local `docker-mongo-sample-datasets`",
	}
	driver, err := New(ctx, client, dbName, 10)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = driver.Analyze(s)
	if err != nil {
		t.Errorf("%v", err)
	}
	table, err := s.FindTableByName("test.restaurants")
	if err != nil {
		t.Errorf("%v", err)
	}
	want := table.Name
	if want == "" {
		t.Errorf("got not empty string.")
	}
}
