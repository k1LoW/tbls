package datasource

import (
	"os"
	"path/filepath"
	"testing"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

var tests = []struct {
	dsn           string
	schemaName    string
	tableCount    int
	relationCount int
}{
	{"my://root:mypass@localhost:33306/testdb", "testdb", 9, 6},
	{"pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable", "testdb", 11, 8},
	{"json://../testdata/testdb.json", "testdb", 7, 9},
	{"ms://SA:MSSQLServer-Passw0rd@localhost:11433/testdb", "testdb", 9, 6},
}

func TestMain(m *testing.M) {
	cPath := credentialPath()
	if _, err := os.Lstat(cPath); err == nil {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", cPath)
		bqTest := struct {
			dsn           string
			schemaName    string
			tableCount    int
			relationCount int
		}{
			"bq://bigquery-public-data/bitcoin_blockchain", "bigquery-public-data:bitcoin_blockchain", 2, 0,
		}
		tests = append(tests, bqTest)
	}
	exit := m.Run()
	if exit != 0 {
		os.Exit(exit)
	}
}

func TestAnalyzeSchema(t *testing.T) {
	for _, tt := range tests {
		schema, err := Analyze(tt.dsn)
		if err != nil {
			t.Errorf("%s", err)
		}
		want := tt.schemaName
		got := schema.Name
		if got != want {
			t.Errorf("got %v\nwant %v", got, want)
		}
	}
}

func TestAnalyzeTables(t *testing.T) {
	for _, tt := range tests {
		schema, err := Analyze(tt.dsn)
		if err != nil {
			t.Errorf("%s", err)
		}
		want := tt.tableCount
		got := len(schema.Tables)
		if got != want {
			t.Errorf("got %v\nwant %v", got, want)
		}
	}
}

func TestAnalyzeRelations(t *testing.T) {
	for _, tt := range tests {
		schema, err := Analyze(tt.dsn)
		if err != nil {
			t.Errorf("%s", err)
		}
		want := tt.relationCount
		got := len(schema.Relations)
		if got != want {
			t.Errorf("got %v\nwant %v", got, want)
		}
	}
}

func credentialPath() string {
	wd, _ := os.Getwd()
	return filepath.Join(filepath.Dir(wd), "client_secrets.json")
}
