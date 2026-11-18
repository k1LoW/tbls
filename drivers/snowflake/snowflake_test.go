package snowflake

import (
	"database/sql"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/tbls/schema"
	"github.com/mattn/go-sqlite3"
)

var registerMetadataTestDriver sync.Once

func TestAnalyzeColumnMetadata(t *testing.T) {
	tests := []struct {
		column                 schema.Column
		characterMaximumLength any
		numericPrecision       any
		numericScale           any
		isIdentity             any
		identityStart          any
		identityIncrement      any
		wantType               string
		wantExtraDef           string
	}{
		{
			column: schema.Column{
				Name:    "AMOUNT",
				Type:    "NUMBER",
				Default: sql.NullString{String: "0", Valid: true},
				Comment: "Amount with precision and scale",
			},
			numericPrecision: 10,
			numericScale:     5,
			isIdentity:       "NO",
			wantType:         "NUMBER(10,5)",
		},
		{
			column:           schema.Column{Name: "DEFAULT_NUMBER", Type: "NUMBER", Nullable: true},
			numericPrecision: 38,
			numericScale:     0,
			wantType:         "NUMBER(38,0)",
		},
		{
			column:                 schema.Column{Name: "LABEL", Type: "TEXT", Nullable: true},
			characterMaximumLength: 20,
			wantType:               "TEXT(20)",
		},
		{
			column:                 schema.Column{Name: "DEFAULT_TEXT", Type: "TEXT", Nullable: true},
			characterMaximumLength: 16777216,
			wantType:               "TEXT(16777216)",
		},
		{
			column:   schema.Column{Name: "NUMBER_NULL_METADATA", Type: "NUMBER", Nullable: true},
			wantType: "NUMBER",
		},
		{
			column:       schema.Column{Name: "NUMBER_NULL_PRECISION", Type: "NUMBER", Nullable: true},
			numericScale: 5,
			wantType:     "NUMBER",
		},
		{
			column:           schema.Column{Name: "NUMBER_NULL_SCALE", Type: "NUMBER", Nullable: true},
			numericPrecision: 10,
			wantType:         "NUMBER",
		},
		{
			column:   schema.Column{Name: "TEXT_NULL_LENGTH", Type: "TEXT", Nullable: true},
			wantType: "TEXT",
		},
		{
			column:           schema.Column{Name: "APPROXIMATE", Type: "FLOAT", Nullable: true},
			numericPrecision: 53,
			numericScale:     0,
			wantType:         "FLOAT",
		},
		{
			column:   schema.Column{Name: "FLAG", Type: "BOOLEAN", Nullable: true},
			wantType: "BOOLEAN",
		},
		{
			column:   schema.Column{Name: "DAY", Type: "DATE", Nullable: true},
			wantType: "DATE",
		},
		{
			column:                 schema.Column{Name: "BYTES", Type: "BINARY", Nullable: true},
			characterMaximumLength: 20,
			wantType:               "BINARY",
		},
		{
			column:   schema.Column{Name: "DOCUMENT", Type: "VARIANT", Nullable: true},
			wantType: "VARIANT",
		},
		{
			column:   schema.Column{Name: "CREATED_AT", Type: "TIMESTAMP_NTZ", Nullable: true},
			wantType: "TIMESTAMP_NTZ",
		},
		{
			column:            schema.Column{Name: "ID", Type: "NUMBER", Comment: "Generated identifier"},
			numericPrecision:  38,
			numericScale:      0,
			isIdentity:        "YES",
			identityStart:     1000,
			identityIncrement: 1,
			wantType:          "NUMBER(38,0)",
			wantExtraDef:      "IDENTITY START 1000 INCREMENT 1",
		},
		{
			column:            schema.Column{Name: "DEFAULT_IDENTITY", Type: "NUMBER"},
			numericPrecision:  38,
			numericScale:      0,
			isIdentity:        "YES",
			identityStart:     "1",
			identityIncrement: "1",
			wantType:          "NUMBER(38,0)",
			wantExtraDef:      "IDENTITY START 1 INCREMENT 1",
		},
		{
			column:            schema.Column{Name: "DESCENDING_IDENTITY", Type: "NUMBER"},
			numericPrecision:  38,
			numericScale:      0,
			isIdentity:        "YES",
			identityStart:     "1000",
			identityIncrement: "-1",
			wantType:          "NUMBER(38,0)",
			wantExtraDef:      "IDENTITY START 1000 INCREMENT -1",
		},
		{
			column: schema.Column{
				Name:    "SEQUENCE_DEFAULT",
				Type:    "NUMBER",
				Default: sql.NullString{String: "SEQUENCE_NAME.NEXTVAL", Valid: true},
			},
			numericPrecision: 38,
			numericScale:     0,
			isIdentity:       "NO",
			wantType:         "NUMBER(38,0)",
		},
		{
			column: schema.Column{
				Name:     "EMPTY_DEFAULT",
				Type:     "TEXT",
				Nullable: true,
				Default:  sql.NullString{String: "", Valid: true},
			},
			wantType: "TEXT",
		},
		{
			column:       schema.Column{Name: "IDENTITY_NULL_METADATA", Type: "NUMBER"},
			isIdentity:   "YES",
			wantType:     "NUMBER",
			wantExtraDef: "IDENTITY",
		},
		{
			column:            schema.Column{Name: "IDENTITY_NULL_START", Type: "NUMBER"},
			isIdentity:        "YES",
			identityIncrement: 1,
			wantType:          "NUMBER",
			wantExtraDef:      "IDENTITY",
		},
		{
			column:        schema.Column{Name: "IDENTITY_NULL_INCREMENT", Type: "NUMBER"},
			isIdentity:    "YES",
			identityStart: 1000,
			wantType:      "NUMBER",
			wantExtraDef:  "IDENTITY",
		},
		{
			column:            schema.Column{Name: "NOT_IDENTITY", Type: "NUMBER", Nullable: true},
			isIdentity:        "NO",
			identityStart:     1000,
			identityIncrement: 1,
			wantType:          "NUMBER",
		},
		{
			column:            schema.Column{Name: "NULL_IDENTITY", Type: "NUMBER", Nullable: true},
			identityStart:     1000,
			identityIncrement: 1,
			wantType:          "NUMBER",
		},
	}

	db := newMetadataTestDB(t)
	if _, err := db.Exec(`INSERT INTO information_schema.tables VALUES ('PUBLIC', 'METADATA', 'BASE TABLE', 'Column metadata')`); err != nil {
		t.Fatal(err)
	}
	// Insert out of order to verify that Analyze preserves ordinal_position.
	for i := len(tests) - 1; i >= 0; i-- {
		tt := tests[i]
		nullable := "NO"
		if tt.column.Nullable {
			nullable = "YES"
		}
		if _, err := db.Exec(`INSERT INTO information_schema.columns (
table_schema, table_name, ordinal_position, column_name, column_default,
is_nullable, data_type, comment, character_maximum_length, numeric_precision,
numeric_scale, is_identity, identity_start, identity_increment
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			"PUBLIC", "METADATA", i+1, tt.column.Name, tt.column.Default,
			nullable, tt.column.Type, sql.NullString{String: tt.column.Comment, Valid: tt.column.Comment != ""},
			tt.characterMaximumLength, tt.numericPrecision, tt.numericScale,
			tt.isIdentity, tt.identityStart, tt.identityIncrement,
		); err != nil {
			t.Fatal(err)
		}
	}
	// Rows from another schema or table must not enter the analyzed columns.
	if _, err := db.Exec(`INSERT INTO information_schema.columns (table_schema, table_name, ordinal_position, column_name, is_nullable, data_type)
VALUES ('OTHER_SCHEMA', 'METADATA', 1, 'OTHER_SCHEMA_COLUMN', 'YES', 'BOOLEAN'),
       ('PUBLIC', 'OTHER_TABLE', 1, 'OTHER_TABLE_COLUMN', 'YES', 'BOOLEAN')`); err != nil {
		t.Fatal(err)
	}

	s := &schema.Schema{Name: "PUBLIC"}
	if err := New(db).Analyze(s); err != nil {
		t.Fatal(err)
	}
	if len(s.Tables) != 1 {
		t.Fatalf("got %d tables, want 1", len(s.Tables))
	}
	table := s.Tables[0]
	if len(table.Columns) != len(tests) {
		t.Fatalf("got %d columns, want %d", len(table.Columns), len(tests))
	}
	for i, tt := range tests {
		t.Run(tt.column.Name, func(t *testing.T) {
			want := tt.column
			want.Type = tt.wantType
			want.ExtraDef = tt.wantExtraDef
			if diff := cmp.Diff(&want, table.Columns[i]); diff != "" {
				t.Errorf("column metadata mismatch (-want +got):\n%s", diff)
			}
		})
	}
	if !table.ShowColumn(schema.ColumnExtraDef, nil) {
		t.Error("identity columns should show Extra Definition")
	}
	plainTable := &schema.Table{}
	for i, tt := range tests {
		if tt.wantExtraDef == "" {
			plainTable.Columns = append(plainTable.Columns, table.Columns[i])
		}
	}
	if plainTable.ShowColumn(schema.ColumnExtraDef, nil) {
		t.Error("columns without identity should not show Extra Definition")
	}
}

// newMetadataTestDB supplies synthetic Snowflake metadata, not a Snowflake server.
func newMetadataTestDB(t *testing.T) *sql.DB {
	t.Helper()
	const driverName = "snowflake-metadata-test"
	registerMetadataTestDriver.Do(func() {
		sql.Register(driverName, &sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				if err := conn.RegisterFunc("CURRENT_VERSION", func() string { return "test-version" }, true); err != nil {
					return err
				}
				if err := conn.RegisterFunc("GET_DDL", func(_, _ string) string { return "CREATE TABLE METADATA (...)" }, true); err != nil {
					return err
				}
				// Analyze opens nested queries while table rows are still active.
				// Every connection therefore needs the same attached in-memory schema.
				_, err := conn.Exec(`ATTACH DATABASE 'file:tbls-snowflake-metadata?mode=memory&cache=shared' AS information_schema`, nil)
				return err
			},
		})
	})
	db, err := sql.Open(driverName, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if _, err := db.Exec(`CREATE TABLE information_schema.tables (
table_schema TEXT, table_name TEXT, table_type TEXT, comment TEXT
);
CREATE TABLE information_schema.columns (
table_schema TEXT, table_name TEXT, ordinal_position INTEGER,
column_name TEXT, column_default TEXT, is_nullable TEXT, data_type TEXT, comment TEXT,
character_maximum_length INTEGER, numeric_precision INTEGER, numeric_scale INTEGER,
is_identity TEXT, identity_start TEXT, identity_increment TEXT
)`); err != nil {
		t.Fatal(err)
	}
	return db
}
