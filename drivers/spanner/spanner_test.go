//go:build spanner

package spanner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"cloud.google.com/go/spanner"
	cloudspanner "cloud.google.com/go/spanner"
	"github.com/SouhlInc/tbls/schema"
)

var ctx context.Context
var s *schema.Schema
var client *spanner.Client

func TestAnalyze(t *testing.T) {
	ctx, client := initClient(t)
	defer client.Close()
	driver, err := New(ctx, client)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = driver.Analyze(s)
	if err != nil {
		t.Errorf("%v", err)
	}
	table, _ := s.FindTableByName("posts")
	want := len(table.Constraints)
	if want != 2 {
		t.Errorf("got: %#v\nwant: %#v", 2, want)
	}
}

func initClient(t *testing.T) (context.Context, *cloudspanner.Client) {
	cPath := credentialPath()
	if _, err := os.Lstat(cPath); err != nil {
		t.Skipf("spanner_client_secrets.json does not exist")
	}
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", cPath)
	var err error

	projectID := os.Getenv("GCLOUD_PROJECT")
	instanceID := "test-instance"
	databaseID := "testdb"
	db := fmt.Sprintf("projects/%s/instances/%s/databases/%s", projectID, instanceID, databaseID)
	s = &schema.Schema{
		Name: db,
	}
	ctx = context.Background()
	client, err = cloudspanner.NewClient(ctx, db)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	return ctx, client
}

func credentialPath() string {
	wd, _ := os.Getwd()
	return filepath.Join(filepath.Dir(filepath.Dir(wd)), "spanner_client_secrets.json")
}
