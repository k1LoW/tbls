package dburlscheme

import (
	"testing"

	sf "github.com/snowflakedb/gosnowflake/v2"
	"github.com/xo/dburl"
)

func TestDburlSchemes(t *testing.T) {
	tests := []struct {
		dsn        string
		wantDriver string
		wantDSN    string
	}{
		{"pg://dbuser:dbpass@hostname:5432/dbname?sslmode=disable", "postgres", "dbname=dbname host=hostname password=dbpass port=5432 sslmode=disable user=dbuser"},
		{"postgres://dbuser:dbpass@hostname:5432/dbname", "postgres", "dbname=dbname host=hostname password=dbpass port=5432 user=dbuser"},
		{"postgresql://u:p@h/db", "postgres", "dbname=db host=h password=p user=u"},
		{"rs://dbuser:dbpass@hostname:5439/dbname", "postgres", "postgres://dbuser:dbpass@hostname:5439/dbname"},
		{"redshift://dbuser:dbpass@hostname:5432/dbname?sslmode=require", "postgres", "postgres://dbuser:dbpass@hostname:5432/dbname?sslmode=require"},
		{"my://dbuser:dbpass@hostname:3306/dbname", "mysql", "dbuser:dbpass@tcp(hostname:3306)/dbname"},
		{"mysql://dbuser:dbpass@hostname:3306/dbname?hide_auto_increment", "mysql", "dbuser:dbpass@tcp(hostname:3306)/dbname?hide_auto_increment="},
		{"maria://dbuser:dbpass@hostname:3306/dbname", "mysql", "dbuser:dbpass@tcp(hostname:3306)/dbname"},
		{"mariadb://dbuser:dbpass@hostname:3306/dbname?parseTime=true", "mysql", "dbuser:dbpass@tcp(hostname:3306)/dbname?parseTime=true"},
		{"clickhouse://default@localhost/testdb", "clickhouse", "clickhouse://default@localhost:9000/testdb"},
		{"ch://u:p@h:9000/db", "clickhouse", "clickhouse://u:p@h:9000/db"},
		{"ch+http://u:p@h:8123/db", "clickhouse", "http://u:p@h:8123/db"},
		{"ch+https://u:p@h/db", "clickhouse", "https://u:p@h/db"},
		{"rs://u:p@h:5439/db?options=-c%20a%3D1&options=-c%20b%3D2", "postgres", "postgres://u:p@h:5439/db?options=-c+a%3D1+-c+b%3D2"},
		{"cockroachdb://u:p@h:26257/db?sslmode=disable", "postgres", "postgres://u:p@h:26257/db?sslmode=disable"},
		{"cr://u@h/db?options=-c%20a%3D1&options=-c%20b%3D2", "postgres", "postgres://u@h:26257/db?options=-c+a%3D1+-c+b%3D2&sslmode=disable"},
		{"memsql://u:p@h/db", "mysql", "u:p@tcp(h)/db"},
		{"tidb://u:p@h/db", "mysql", "u:p@tcp(h)/db"},
		{"ti://u:p@h:4000/db?parseTime=false", "mysql", "u:p@tcp(h:4000)/db?parseTime=false"},
		{"vitess://u:p@h/db", "mysql", "u:p@tcp(h)/db"},
		{"vt://u:p@h:3306/db", "mysql", "u:p@tcp(h:3306)/db"},
		{"ms://u:p@h/db", "sqlserver", "sqlserver://u:p@h/?database=db"},
		{"ms://u:p@h/db?fedauth=ActiveDirectoryDefault", "azuresql", "sqlserver://u:p@h/?database=db&fedauth=ActiveDirectoryDefault"},
		{"azuresql://cid:secret@host/mydb", "azuresql", "sqlserver://cid:secret@host/?database=mydb"},
		{"sf://alice:secret@myaccount/mydb/myschema", "snowflake", "alice:secret@myaccount/mydb/myschema"},
		{"sf://alice@myaccount/mydb/myschema?authenticator=externalbrowser", "snowflake", "alice@myaccount/mydb/myschema?authenticator=externalbrowser"},
		{"snowflake://alice@myaccount/mydb/myschema?authenticator=SNOWFLAKE_JWT&privateKey=abc", "snowflake", "alice@myaccount/mydb/myschema?authenticator=SNOWFLAKE_JWT&privateKey=abc"},
	}
	for _, tt := range tests {
		t.Run(tt.dsn, func(t *testing.T) {
			u, err := dburl.Parse(tt.dsn)
			if err != nil {
				t.Fatal(err)
			}
			if u.Driver != tt.wantDriver {
				t.Errorf("got driver %q, want %q", u.Driver, tt.wantDriver)
			}
			if u.DSN != tt.wantDSN {
				t.Errorf("got DSN %q, want %q", u.DSN, tt.wantDSN)
			}
		})
	}
}

func TestDburlSnowflakeCredentials(t *testing.T) {
	tests := []struct {
		dsn          string
		wantUser     string
		wantPassword string
	}{
		{"sf://alice:secret@myaccount/mydb/myschema", "alice", "secret"},
		{"sf://alice@myaccount/mydb/myschema?authenticator=externalbrowser", "alice", ""},
	}
	for _, tt := range tests {
		t.Run(tt.dsn, func(t *testing.T) {
			u, err := dburl.Parse(tt.dsn)
			if err != nil {
				t.Fatal(err)
			}
			c, err := sf.ParseDSN(u.DSN)
			if err != nil {
				t.Fatal(err)
			}
			if c.User != tt.wantUser || c.Password != tt.wantPassword {
				t.Errorf("got user %q password %q, want user %q password %q", c.User, c.Password, tt.wantUser, tt.wantPassword)
			}
		})
	}
}
