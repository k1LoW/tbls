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
	"github.com/k1LoW/duration"
	"github.com/SouhlInc/tbls/drivers/bq"
	"github.com/SouhlInc/tbls/drivers/spanner"
	"github.com/SouhlInc/tbls/schema"
	"golang.org/x/oauth2"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"
)

const defaultImpersonateServiceAccountLifetimeStr = "300sec"

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

// NewBigqueryClient returns new bigquery.Client.
func NewBigqueryClient(ctx context.Context, urlstr string) (*bigquery.Client, string, string, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, "", "", err
	}

	splitted := strings.Split(u.Path, "/")
	projectID := u.Host
	datasetID := splitted[1]

	var options []option.ClientOption

	// Setup credential
	values := u.Query()
	if err := setEnvGoogleApplicationCredentials(values); err != nil {
		return nil, "", "", err
	}

	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" && os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON") != "" {
		options = append(options, option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON"))))
	}

	// Setup impersonate service account configuration
	impersonateServiceAccount := getImpersonateServiceAccount()
	if impersonateServiceAccount != "" {
		lifetime, err := getImpersonateServiceAccountLifetime()
		if err != nil {
			return nil, "", "", err
		}
		ts, err := createImpersonationTokenSource(ctx, impersonateServiceAccount, lifetime)
		if err != nil {
			return nil, "", "", err
		}
		options = append(options, option.WithTokenSource(ts))
	}

	client, err := bigquery.NewClient(ctx, projectID, options...)
	if err != nil {
		return nil, "", "", err
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

// NewSpannerClient returns new cloudspanner.Client.
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

	var options []option.ClientOption

	// Setup credential
	values := u.Query()
	if err := setEnvGoogleApplicationCredentials(values); err != nil {
		return nil, "", err
	}

	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" && os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON") != "" {
		options = append(options, option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON"))))
	}

	// Setup impersonate service account configuration
	impersonateServiceAccount := getImpersonateServiceAccount()
	if impersonateServiceAccount != "" {
		lifetime, err := getImpersonateServiceAccountLifetime()
		if err != nil {
			return nil, "", err
		}
		ts, err := createImpersonationTokenSource(ctx, impersonateServiceAccount, lifetime)
		if err != nil {
			return nil, "", err
		}
		options = append(options, option.WithTokenSource(ts))
	}

	client, err := cloudspanner.NewClient(ctx, db, options...)
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

func createImpersonationTokenSource(ctx context.Context, impersonateServiceAccount string, impersonateServiceAccountLifetime time.Duration) (oauth2.TokenSource, error) {
	return impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
		TargetPrincipal: impersonateServiceAccount,
		Scopes:          []string{"https://www.googleapis.com/auth/cloud-platform"},
		Lifetime:        impersonateServiceAccountLifetime,
	})
}

func getImpersonateServiceAccount() string {
	return os.Getenv("GOOGLE_IMPERSONATE_SERVICE_ACCOUNT")
}

func getImpersonateServiceAccountLifetime() (time.Duration, error) {
	durationStr := os.Getenv("GOOGLE_IMPERSONATE_SERVICE_ACCOUNT_LIFETIME")
	if durationStr == "" {
		durationStr = defaultImpersonateServiceAccountLifetimeStr
	}
	d, err := duration.Parse(durationStr)
	if err != nil {
		return 0, err
	}
	return d, nil
}
