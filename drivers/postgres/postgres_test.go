//go:build postgres

package postgres

import (
	"database/sql"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/tbls/schema"
	_ "github.com/lib/pq"
	"github.com/xo/dburl"
)

var s *schema.Schema
var db *sql.DB

func TestMain(m *testing.M) {
	s = &schema.Schema{
		Name: "testdb",
	}
	db, _ = dburl.Open("pg://postgres:pgpass@localhost:55413/testdb?sslmode=disable")
	defer db.Close()
	exit := m.Run()
	if exit != 0 {
		os.Exit(exit)
	}
}

func TestAnalyzeView(t *testing.T) {
	driver := New(db)
	err := driver.Analyze(s)
	if err != nil {
		t.Errorf("%+v", err)
	}
	view, _ := s.FindTableByName("post_comments")
	want := view.Def
	if want == "" {
		t.Errorf("got not empty string.")
	}
}

func TestExtraDef(t *testing.T) {
	driver := New(db)
	if err := driver.Analyze(s); err != nil {
		t.Fatal(err)
	}
	tbl, _ := s.FindTableByName("comments")
	{
		c, _ := tbl.FindColumnByName("post_id_desc")
		got := c.ExtraDef
		if want := "GENERATED ALWAYS AS (post_id * '-1'::integer) STORED"; got != want {
			t.Errorf("got %v\nwant %v", got, want)
		}
	}
}

func TestInfo(t *testing.T) {
	driver := New(db)
	d, err := driver.Info()
	if err != nil {
		t.Errorf("%v", err)
	}
	if d.Name != "postgres" {
		t.Errorf("got %v\nwant %v", d.Name, "postgres")
	}
	if d.DatabaseVersion == "" {
		t.Errorf("got not empty string.")
	}
}

func TestTriggersOrder(t *testing.T) {
	if _, err := db.Exec(`DROP TABLE IF EXISTS tbls_trigger_order`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE tbls_trigger_order (a int)`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.Exec(`DROP TABLE IF EXISTS tbls_trigger_order`)
	}()
	// Create the triggers in descending name order so that creation order and name order differ.
	if _, err := db.Exec(`CREATE TRIGGER trg_tbls_trigger_order_b AFTER INSERT ON tbls_trigger_order FOR EACH ROW EXECUTE PROCEDURE update_updated()`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER trg_tbls_trigger_order_a AFTER INSERT ON tbls_trigger_order FOR EACH ROW EXECUTE PROCEDURE update_updated()`); err != nil {
		t.Fatal(err)
	}

	driver := New(db)
	sc := &schema.Schema{Name: "testdb"}
	if err := driver.Analyze(sc); err != nil {
		t.Fatal(err)
	}

	tbl, err := sc.FindTableByName("public.tbls_trigger_order")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, trig := range tbl.Triggers {
		got = append(got, trig.Name)
	}
	want := []string{"trg_tbls_trigger_order_b", "trg_tbls_trigger_order_a"}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Error(diff)
	}
}

func TestParseFK(t *testing.T) {
	tests := []struct {
		in              string
		wantCols        []string
		wantParentTable string
		wantParentCols  []string
	}{
		{"FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE", []string{"user_id"}, "users", []string{"id"}},
		{"FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL (user_id)", []string{"user_id"}, "users", []string{"id"}},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			gotCols, gotParentTable, gotParentCols, err := parseFK(tt.in)
			if err != nil {
				t.Error(err)
				return
			}
			if diff := cmp.Diff(gotCols, tt.wantCols, nil); diff != "" {
				t.Error(diff)
			}
			if gotParentTable != tt.wantParentTable {
				t.Errorf("got %v want %v", gotParentTable, tt.wantParentTable)
			}
			if diff := cmp.Diff(gotParentCols, tt.wantParentCols, nil); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestAnalyzeConstraintReferencedTable(t *testing.T) {
	driver := New(db)
	if err := driver.Analyze(s); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		tableName      string
		constraintName string
		want           string
	}{
		{"user_options", "user_options_user_id_fk", "public.users"},
		{"backup.blog_options", "blog_options_blog_id_fk", "backup.blogs"},
		{"time.referencing", "referencing_bar_id", "time.bar"},
	}

	for _, tt := range tests {
		t.Run(tt.constraintName, func(t *testing.T) {
			table, err := s.FindTableByName(tt.tableName)
			if err != nil {
				t.Fatal(err)
			}
			constraint, err := table.FindConstraintByName(tt.constraintName)
			if err != nil {
				t.Fatal(err)
			}
			if constraint.ReferencedTable == nil {
				t.Fatalf("constraint %s has nil referenced table", tt.constraintName)
			}
			if got := *constraint.ReferencedTable; got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}
		})
	}
}
