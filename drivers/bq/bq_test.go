//go:build bq

package bq

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/SouhlInc/tbls/schema"
)

var projectID = "bigquery-public-data"
var ctx context.Context
var client *bigquery.Client

func TestAnalyze(t *testing.T) {
	ctx, client := initClient(t)
	s := &schema.Schema{
		Name: projectID,
	}
	defer client.Close()
	driver, err := New(ctx, client, "crypto_bitcoin")
	if err != nil {
		t.Errorf("%v", err)
	}
	err = driver.Analyze(s)
	if err != nil {
		t.Errorf("%v", err)
	}
	table, _ := s.FindTableByName("inputs")
	want := table.Def
	if want == "" {
		t.Errorf("got not empty string.")
	}
}

func initClient(t *testing.T) (context.Context, *bigquery.Client) {
	cPath := credentialPath()
	if _, err := os.Lstat(cPath); err != nil {
		t.Skipf("client_secrets.json does not exist")
	}
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", cPath)
	var err error
	ctx = context.Background()
	client, err = bigquery.NewClient(ctx, projectID)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	return ctx, client
}

func credentialPath() string {
	wd, _ := os.Getwd()
	return filepath.Join(filepath.Dir(filepath.Dir(wd)), "client_secrets.json")
}
