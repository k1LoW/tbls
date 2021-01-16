package datasource

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	cloudspanner "cloud.google.com/go/spanner"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/drivers"
	"github.com/k1LoW/tbls/drivers/bq"
	"github.com/k1LoW/tbls/drivers/dynamo"
	"github.com/k1LoW/tbls/drivers/mariadb"
	"github.com/k1LoW/tbls/drivers/mssql"
	"github.com/k1LoW/tbls/drivers/mysql"
	"github.com/k1LoW/tbls/drivers/postgres"
	"github.com/k1LoW/tbls/drivers/redshift"
	"github.com/k1LoW/tbls/drivers/snowflake"
	"github.com/k1LoW/tbls/drivers/spanner"
	"github.com/k1LoW/tbls/drivers/sqlite"
	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
	"github.com/xo/dburl"
)

// Analyze database
func Analyze(dsn config.DSN) (*schema.Schema, error) {
	urlstr := dsn.URL
	if strings.Index(urlstr, "https://") == 0 || strings.Index(urlstr, "http://") == 0 {
		return AnalyzeHTTPResource(dsn)
	}
	if strings.Index(urlstr, "json://") == 0 {
		return AnalyzeJSON(urlstr)
	}
	if strings.Index(urlstr, "bq://") == 0 || strings.Index(urlstr, "bigquery://") == 0 {
		return AnalyzeBigquery(urlstr)
	}
	if strings.Index(urlstr, "span://") == 0 || strings.Index(urlstr, "spanner://") == 0 {
		return AnalyzeSpanner(urlstr)
	}
	if strings.Index(urlstr, "dynamodb://") == 0 || strings.Index(urlstr, "dynamo://") == 0 {
		return AnalyzeDynamodb(urlstr)
	}
	s := &schema.Schema{}
	u, err := dburl.Parse(urlstr)
	if err != nil {
		return s, errors.WithStack(err)
	}
	splitted := strings.Split(u.Short(), "/")
	if len(splitted) < 2 {
		return s, errors.WithStack(fmt.Errorf("invalid DSN: parse %s -> %#v", urlstr, u))
	}

	opts := []drivers.Option{}
	if u.Driver == "mysql" {
		values := u.Query()
		for k := range values {
			if k == "show_auto_increment" {
				opts = append(opts, mysql.ShowAutoIcrrement())
				values.Del(k)
			}
		}
		u.RawQuery = values.Encode()
		urlstr = u.String()
	}

	db, err := dburl.Open(urlstr)
	defer db.Close()
	if err != nil {
		return s, errors.WithStack(err)
	}
	if err := db.Ping(); err != nil {
		return s, errors.WithStack(err)
	}

	var driver drivers.Driver

	switch u.Driver {
	case "postgres":
		s.Name = splitted[1]
		if u.Scheme == "rs" || u.Scheme == "redshift" {
			driver = redshift.New(db)
		} else {
			driver = postgres.New(db)
		}
	case "mysql":
		s.Name = splitted[1]
		if u.Scheme == "maria" || u.Scheme == "mariadb" {
			driver, err = mariadb.New(db, opts...)
		} else {
			driver, err = mysql.New(db, opts...)
		}
		if err != nil {
			return s, err
		}
	case "sqlite3":
		s.Name = splitted[len(splitted)-1]
		driver = sqlite.New(db)
	case "mssql":
		s.Name = splitted[1]
		driver = mssql.New(db)
	case "snowflake":
		s.Name = splitted[2]
		driver = snowflake.New(db)
	default:
		return s, errors.WithStack(fmt.Errorf("unsupported driver '%s'", u.Driver))
	}
	err = driver.Analyze(s)
	if err != nil {
		return s, err
	}
	return s, nil
}

// AnalyzeHTTPResource analyze `https://` or `http://`
func AnalyzeHTTPResource(dsn config.DSN) (*schema.Schema, error) {
	s := &schema.Schema{}
	req, err := http.NewRequest("GET", dsn.URL, nil)
	if err != nil {
		return s, errors.WithStack(err)
	}
	for k, v := range dsn.Headers {
		req.Header.Add(k, v)
	}
	client := &http.Client{Timeout: time.Duration(10) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return s, errors.WithStack(err)
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(s); err != nil {
		return s, errors.WithStack(err)
	}
	if err := s.Repair(); err != nil {
		return s, errors.WithStack(err)
	}
	return s, nil
}

// AnalyzeJSON analyze `json://`
func AnalyzeJSON(urlstr string) (*schema.Schema, error) {
	s := &schema.Schema{}
	splitted := strings.Split(urlstr, "json://")
	file, err := os.Open(splitted[1])
	if err != nil {
		return s, errors.WithStack(err)
	}
	dec := json.NewDecoder(file)
	if err := dec.Decode(s); err != nil {
		return s, errors.WithStack(err)
	}
	if err := s.Repair(); err != nil {
		return s, errors.WithStack(err)
	}
	return s, nil
}

// Deprecated
func AnalyzeJSONString(str string) (*schema.Schema, error) {
	return AnalyzeJSONStringOrFile(str)
}

// AnalyzeJSONStringOrFile analyze JSON string or JSON file
func AnalyzeJSONStringOrFile(strOrPath string) (s *schema.Schema, err error) {
	s = &schema.Schema{}
	var buf io.Reader
	if strings.HasPrefix(strOrPath, "{") {
		buf = bytes.NewBufferString(strOrPath)
	} else {
		buf, err = os.Open(filepath.Clean(strOrPath))
		if err != nil {
			return s, errors.WithStack(err)
		}
	}
	dec := json.NewDecoder(buf)
	if err := dec.Decode(s); err != nil {
		return s, errors.WithStack(err)
	}
	if err := s.Repair(); err != nil {
		return s, errors.WithStack(err)
	}
	return s, nil
}

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
	values := u.Query()
	err = setEnvGoogleApplicationCredentials(values)
	if err != nil {
		return nil, "", "", err
	}

	splitted := strings.Split(u.Path, "/")

	projectID := u.Host
	datasetID := splitted[1]

	client, err := bigquery.NewClient(ctx, projectID)
	return client, projectID, datasetID, err
}

// AnalyzeSpanner analyze `spanner://`
func AnalyzeSpanner(urlstr string) (*schema.Schema, error) {
	s := &schema.Schema{}
	u, err := url.Parse(urlstr)
	if err != nil {
		return s, err
	}

	values := u.Query()
	err = setEnvGoogleApplicationCredentials(values)
	if err != nil {
		return s, err
	}

	splitted := strings.Split(u.Path, "/")
	projectID := u.Host
	instanceID := splitted[1]
	databaseID := splitted[2]

	db := fmt.Sprintf("projects/%s/instances/%s/databases/%s", projectID, instanceID, databaseID)
	ctx := context.Background()
	client, err := cloudspanner.NewClient(ctx, db)
	if err != nil {
		return s, err
	}
	defer client.Close()
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

// AnalizeDynamodb analyze `dynamodb://`
func AnalyzeDynamodb(urlstr string) (*schema.Schema, error) {
	s := &schema.Schema{}
	u, err := url.Parse(urlstr)
	if err != nil {
		return s, err
	}

	values := u.Query()
	err = setEnvAWSCredentials(values)
	if err != nil {
		return s, err
	}

	region := u.Host

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	config := aws.NewConfig().WithRegion(region)
	if os.Getenv("AWS_ENDPOINT_URL") != "" {
		config = config.WithEndpoint(os.Getenv("AWS_ENDPOINT_URL"))
	}

	client := dynamodb.New(sess, config)
	ctx := context.Background()

	driver, err := dynamo.New(ctx, client)
	if err != nil {
		return s, err
	}

	s.Name = fmt.Sprintf("Amazon DynamoDB (%s)", region)
	err = driver.Analyze(s)
	if err != nil {
		return s, err
	}
	return s, nil
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

func setEnvAWSCredentials(values url.Values) error {
	for k := range values {
		if strings.HasPrefix(k, "aws_") {
			return os.Setenv(strings.ToUpper(k), values.Get(k))
		}
	}
	return nil
}
