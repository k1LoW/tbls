package datasource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"cloud.google.com/go/bigquery"
	cloudspanner "cloud.google.com/go/spanner"
	"github.com/k1LoW/tbls/drivers"
	"github.com/k1LoW/tbls/drivers/bq"
	"github.com/k1LoW/tbls/drivers/mssql"
	"github.com/k1LoW/tbls/drivers/mysql"
	"github.com/k1LoW/tbls/drivers/postgres"
	"github.com/k1LoW/tbls/drivers/redshift"
	"github.com/k1LoW/tbls/drivers/spanner"
	"github.com/k1LoW/tbls/drivers/sqlite"
	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
	"github.com/xo/dburl"
)

// Analyze database
func Analyze(urlstr string) (*schema.Schema, error) {
	if strings.Index(urlstr, "json://") == 0 {
		return AnalizeJSON(urlstr)
	}
	if strings.Index(urlstr, "bq://") == 0 || strings.Index(urlstr, "bigquery://") == 0 {
		return AnalizeBigquery(urlstr)
	}
	if strings.Index(urlstr, "span://") == 0 || strings.Index(urlstr, "spanner://") == 0 {
		return AnalizeSpanner(urlstr)
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

	db, err := dburl.Open(urlstr)
	defer db.Close()
	if err != nil {
		return s, errors.WithStack(err)
	}
	if err = db.Ping(); err != nil {
		return s, errors.WithStack(err)
	}

	var driver drivers.Driver

	switch u.Driver {
	case "postgres":
		s.Name = splitted[1]
		if u.Scheme == "rs" || u.Scheme == "redshift" {
			driver = redshift.NewRedshift(db)
		} else {
			driver = postgres.NewPostgres(db)
		}
	case "mysql":
		s.Name = splitted[1]
		driver = mysql.NewMysql(db)
	case "sqlite3":
		s.Name = splitted[len(splitted)-1]
		driver = sqlite.NewSqlite(db)
	case "mssql":
		s.Name = splitted[1]
		driver = mssql.NewMssql(db)
	default:
		return s, errors.WithStack(fmt.Errorf("unsupported driver '%s'", u.Driver))
	}
	d, err := driver.Info()
	if err != nil {
		return s, err
	}
	s.Driver = d
	err = driver.Analyze(s)
	if err != nil {
		return s, err
	}
	return s, nil
}

// AnalizeJSON analyze `json://`
func AnalizeJSON(urlstr string) (*schema.Schema, error) {
	s := &schema.Schema{}
	splitted := strings.Split(urlstr, "json://")
	file, err := os.Open(splitted[1])
	if err != nil {
		return s, errors.WithStack(err)
	}
	dec := json.NewDecoder(file)
	err = dec.Decode(s)
	if err != nil {
		return s, errors.WithStack(err)
	}
	err = s.Repair()
	if err != nil {
		return s, errors.WithStack(err)
	}
	return s, nil
}

// AnalizeBigquery analyze `bq://`
func AnalizeBigquery(urlstr string) (*schema.Schema, error) {
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
	datasetID := splitted[1]

	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return s, err
	}
	defer client.Close()

	s.Name = fmt.Sprintf("%s:%s", projectID, datasetID)
	driver, err := bq.NewBigquery(ctx, client, datasetID)
	if err != nil {
		return s, err
	}
	d, err := driver.Info()
	if err != nil {
		return s, err
	}
	s.Driver = d
	err = driver.Analyze(s)
	if err != nil {
		return s, err
	}
	return s, nil
}

// AnalizeSpanner analyze `spanner://`
func AnalizeSpanner(urlstr string) (*schema.Schema, error) {
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

	driver, err := spanner.NewSpanner(ctx, client)
	if err != nil {
		return s, err
	}
	d, err := driver.Info()
	if err != nil {
		return s, err
	}
	s.Driver = d
	err = driver.Analyze(s)
	if err != nil {
		return s, err
	}
	return s, nil
}

// AnalizeDynamodb analyze `dynamodb://`
func AnalyzeDynamodb(urlstr string) (*schema.Schema, error) {
	s := &schema.Schema{}

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
