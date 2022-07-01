package datasource

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	cloudspanner "cloud.google.com/go/spanner"
	"github.com/k1LoW/tbls/drivers/bq"
	"github.com/k1LoW/tbls/drivers/spanner"
	"github.com/k1LoW/tbls/schema"
	"golang.org/x/oauth2"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"
)

// AnalyzeBigquery analyze `bq://`
func AnalyzeBigquery(urlstr string) (*schema.Schema, error) {
	s := &schema.Schema{}
	ctx := context.Background()
	client, projectID, datasetID, err := NewBigqueryClient(ctx, urlstr)
	if err != nil {
		return s, err
	}
	defer client.Close()

	s.Name = fmt.Sprintf("%s:%s", projectID, datasetID)
	driver, err := bq.New(ctx, client, datasetID)
	if err != nil {
		return s, err
	}
	err = driver.Analyze(s)
	if err != nil {
		return s, err
	}
	return s, nil
}

// NewBigqueryClient returns new bigquery.Client
func NewBigqueryClient(ctx context.Context, urlstr string) (*bigquery.Client, string, string, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, "", "", err
	}

	splitted := strings.Split(u.Path, "/")
	projectID := u.Host
	datasetID := splitted[1]

	values := u.Query()
	if err := setEnvGoogleApplicationCredentials(values); err != nil {
		return nil, "", "", err
	}
	var client *bigquery.Client
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" && os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON") != "" {
		client, err = bigquery.NewClient(ctx, projectID, option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON"))))
	} else {
		client, err = bigquery.NewClient(ctx, projectID)
	}
	return client, projectID, datasetID, err
}

// AnalyzeSpanner analyze `spanner://`
func AnalyzeSpanner(urlstr string) (*schema.Schema, error) {
	s := &schema.Schema{}
	ctx := context.Background()
	client, db, err := NewSpannerClient(ctx, urlstr)
	if err != nil {
		return s, err
	}
	defer client.Close() //nolint

	s.Name = db
	driver, err := spanner.New(ctx, client)
	if err != nil {
		return s, err
	}
	err = driver.Analyze(s)
	if err != nil {
		return s, err
	}
	return s, nil
}

// NewSpannerClient returns new cloudspanner.Client
func NewSpannerClient(ctx context.Context, urlstr string) (*cloudspanner.Client, string, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, "", err
	}

	splitted := strings.Split(u.Path, "/")
	projectID := u.Host
	instanceID := splitted[1]
	databaseID := splitted[2]
	db := fmt.Sprintf("projects/%s/instances/%s/databases/%s", projectID, instanceID, databaseID)

	values := u.Query()
	if err := setEnvGoogleApplicationCredentials(values); err != nil {
		return nil, "", err
	}
	var client *cloudspanner.Client
	options := []option.ClientOption{}

	ts, err := getImpersonationTokenSource(ctx, values)
	if err != nil {
		return nil, "", err
	}
	if ts != nil {
		options = append(options, option.WithTokenSource(ts))
	}

	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" && os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON") != "" {
		options = append(options, option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON"))))
	}

	client, err = cloudspanner.NewClient(ctx, db, options...)
	if err != nil {
		return nil, "", err
	}
	return client, db, nil
}

func setEnvGoogleApplicationCredentials(values url.Values) error {
	keys := []string{
		"google_application_credentials",
		"credentials",
		"creds",
	}
	for _, k := range keys {
		if values.Get(k) != "" {
			return os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", values.Get(k))
		}
	}
	return nil
}

func getImpersonationTokenSource(ctx context.Context, values url.Values) (oauth2.TokenSource, error) {
	impersonateServiceAccount := values.Get("impersonate_service_account")
	if impersonateServiceAccount == "" {
		return nil, nil
	}
	// Setting up options for service account impersonation
	durationStr := values.Get("impersonate_service_account_duration")
	if durationStr == "" {
		durationStr = "300s"
	}
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return nil, err
	}
	return impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
		TargetPrincipal: impersonateServiceAccount,
		Scopes:          []string{"https://www.googleapis.com/auth/cloud-platform"},
		Lifetime:        duration,
	})
}
