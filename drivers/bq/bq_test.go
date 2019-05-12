package bq

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/k1LoW/tbls/schema"
)

var ctx context.Context
var s *schema.Schema
var client *bigquery.Client

func TestAnalyzeView(t *testing.T) {
	ctx, client := initClient(t)
	defer client.Close()
	driver, err := NewBigquery(ctx, client, "crypto_bitcoin")
	if err != nil {
		t.Errorf("%v", err)
	}
	err = driver.Analyze(s)
	if err != nil {
		t.Errorf("%v", err)
	}
	view, _ := s.FindTableByName("inputs")
	expected := view.Def
	if expected == "" {
		t.Errorf("actual not empty string.")
	}
}

func initClient(t *testing.T) (context.Context, *bigquery.Client) {
	cPath := credentialPath()
	if _, err := os.Lstat(cPath); err != nil {
		t.Skipf("client_secrets.json does not exist")
	}
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", cPath)
	var err error
	projectID := "bigquery-public-data"
	s = &schema.Schema{
		Name: projectID,
	}
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
	return filepath.Join(filepath.Dir(wd), "client_secrets.json")
}
