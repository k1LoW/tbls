package datasource

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/drivers"
	"github.com/k1LoW/tbls/drivers/mariadb"
	"github.com/k1LoW/tbls/drivers/mssql"
	"github.com/k1LoW/tbls/drivers/mysql"
	"github.com/k1LoW/tbls/drivers/postgres"
	"github.com/k1LoW/tbls/drivers/redshift"
	"github.com/k1LoW/tbls/drivers/snowflake"
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
	if strings.Index(urlstr, "mongodb://") == 0 || strings.Index(urlstr, "mongo://") == 0 {
		return AnalyzeMongodb(urlstr)
	}
	s := &schema.Schema{}
	u, err := dburl.Parse(urlstr)
	if err != nil {
		return s, errors.WithStack(err)
	}
	splitted := strings.Split(u.Short(), "/")
	if len(splitted) < 2 {
		return s, errors.Errorf("invalid DSN: parse %s -> %#v", urlstr, u)
	}

	opts := []drivers.Option{}
	if u.Driver == "mysql" {
		values := u.Query()
		for k := range values {
			if k == "show_auto_increment" {
				opts = append(opts, mysql.ShowAutoIcrrement())
				values.Del(k)
			}
			if k == "hide_auto_increment" {
				opts = append(opts, mysql.HideAutoIcrrement())
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
	case "sqlserver":
		s.Name = splitted[1]
		driver = mssql.New(db)
	case "snowflake":
		s.Name = splitted[2]
		driver = snowflake.New(db)
	default:
		return s, errors.Errorf("unsupported driver '%s'", u.Driver)
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
