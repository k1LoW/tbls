package datasource

import (
	"os"
	"path/filepath"
	"testing"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/k1LoW/tbls/config"
	_ "github.com/lib/pq"
)

var tests = []struct {
	dsn           config.DSN
	schemaName    string
	tableCount    int
	relationCount int
}{
	{config.DSN{URL: "my://root:mypass@localhost:33306/testdb"}, "testdb", 9, 6},
	{config.DSN{URL: "pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable"}, "testdb", 17, 12},
	{config.DSN{URL: "json://../testdata/testdb.json"}, "testdb", 11, 12},
	{config.DSN{URL: "https://raw.githubusercontent.com/k1LoW/tbls/main/testdata/testdb.json"}, "testdb", 11, 12},
	{config.DSN{URL: "ms://SA:MSSQLServer-Passw0rd@localhost:11433/testdb"}, "testdb", 11, 7},
}

func TestMain(m *testing.M) {
	cPath := credentialPath()
	if _, err := os.Lstat(cPath); err == nil {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", cPath)
		bqTest := struct {
			dsn           config.DSN
			schemaName    string
			tableCount    int
			relationCount int
		}{
			config.DSN{URL: "bq://bigquery-public-data/bitcoin_blockchain"}, "bigquery-public-data:bitcoin_blockchain", 2, 0,
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
			t.Errorf("%v: got %v\nwant %v", tt.dsn, got, want)
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

func TestAnalyzeJSONString(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(testdataDir(), "testdb.json"))
	if err != nil {
		t.Fatal(err)
	}
	s, err := AnalyzeJSONString(string(b))
	if err != nil {
		t.Fatal(err)
	}
	if want := "testdb"; s.Name != want {
		t.Errorf("got %v want %v", s.Name, want)
	}
}

func credentialPath() string {
	wd, _ := os.Getwd()
	return filepath.Join(filepath.Dir(wd), "client_secrets.json")
}

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(wd), "testdata"))
	return dir
}
