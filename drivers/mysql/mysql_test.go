//go:build mysql

package mysql

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-cmp/cmp"
	_ "github.com/k1LoW/tbls/internal/dburlscheme"
	"github.com/k1LoW/tbls/schema"
	"github.com/xo/dburl"
)

var s *schema.Schema
var db *sql.DB

func TestMain(m *testing.M) {
	s = &schema.Schema{
		Name: "testdb",
	}
	db, _ = dburl.Open("my://root:mypass@localhost:33308/testdb")
	defer db.Close()
	exit := m.Run()
	if exit != 0 {
		os.Exit(exit)
	}
}

func TestAnalyzeView(t *testing.T) {
	driver, err := New(db)
	if err != nil {
		t.Fatal(err)
	}

	err = driver.Analyze(s)
	if err != nil {
		t.Fatal(err)
	}
	view, _ := s.FindTableByName("post_comments")
	if got := view.Def; got == "" {
		t.Errorf("got not empty string.")
	}
}

func TestExtraDef(t *testing.T) {
	driver, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	if err := driver.Analyze(s); err != nil {
		t.Fatal(err)
	}
	tbl, _ := s.FindTableByName("comments")
	{
		c, _ := tbl.FindColumnByName("id")
		got := c.ExtraDef
		if want := "auto_increment"; got != want {
			t.Errorf("got %v\nwant %v", got, want)
		}
	}
	{
		c, _ := tbl.FindColumnByName("post_id_desc")
		got := c.ExtraDef
		if want := "GENERATED ALWAYS AS (`post_id` * -(1)) VIRTUAL"; got != want {
			t.Errorf("got %v\nwant %v", got, want)
		}
	}
}

func TestInfo(t *testing.T) {
	driver, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	d, err := driver.Info()
	if err != nil {
		t.Fatal(err)
	}
	if d.Name != "mysql" {
		t.Errorf("got %v\nwant %v", d.Name, "mysql")
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
	// mysql orders triggers by name because information_schema has no stable creation-order key, so create
	// them in descending name order to detect the creation order leaking into the output.
	if _, err := db.Exec(`CREATE TRIGGER trg_tbls_trigger_order_b BEFORE INSERT ON tbls_trigger_order FOR EACH ROW SET NEW.a = NEW.a`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER trg_tbls_trigger_order_a BEFORE INSERT ON tbls_trigger_order FOR EACH ROW SET NEW.a = NEW.a`); err != nil {
		t.Fatal(err)
	}

	driver, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	sc := &schema.Schema{Name: "testdb"}
	if err := driver.Analyze(sc); err != nil {
		t.Fatal(err)
	}

	tbl, err := sc.FindTableByName("tbls_trigger_order")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, trig := range tbl.Triggers {
		got = append(got, trig.Name)
	}
	want := []string{"trg_tbls_trigger_order_a", "trg_tbls_trigger_order_b"}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Error(diff)
	}
}

func TestForeignKeyActions(t *testing.T) {
	// A run that is killed before its cleanup leaves these behind, and the CREATE statements below would
	// then fail on every later run. The table goes first because it references the external database.
	for _, q := range []string{
		`DROP TABLE IF EXISTS tbls_fk_actions`,
		`DROP DATABASE IF EXISTS tbls_fk_actions_external`,
	} {
		if _, err := db.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`CREATE DATABASE tbls_fk_actions_external`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec(`DROP DATABASE tbls_fk_actions_external`); err != nil {
			t.Error(err)
		}
	})
	if _, err := db.Exec(`CREATE TABLE tbls_fk_actions_external.users (id int PRIMARY KEY)`); err != nil {
		t.Fatal(err)
	}
	// The same table and constraint names in another schema must not affect this schema.
	if _, err := db.Exec(`
CREATE TABLE tbls_fk_actions_external.tbls_fk_actions (
  external_user_id int,
  CONSTRAINT tbls_fk_actions_external_fk FOREIGN KEY (external_user_id)
    REFERENCES tbls_fk_actions_external.users (id) ON DELETE CASCADE
)`); err != nil {
		t.Fatal(err)
	}
	// Reuse the fixture's composite key on comments(post_id, user_id).
	if _, err := db.Exec(`
CREATE TABLE tbls_fk_actions (
  child_user_id int,
  child_post_id bigint,
  external_user_id int,
  CONSTRAINT tbls_fk_actions_fk FOREIGN KEY (child_post_id, child_user_id)
    REFERENCES comments (post_id, user_id) ON UPDATE CASCADE ON DELETE SET NULL,
  CONSTRAINT tbls_fk_actions_external_fk FOREIGN KEY (external_user_id)
    REFERENCES tbls_fk_actions_external.users (id) ON UPDATE SET NULL
)`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec(`DROP TABLE tbls_fk_actions`); err != nil {
			t.Error(err)
		}
	})

	driver, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	sc := &schema.Schema{Name: "testdb"}
	if err := driver.Analyze(sc); err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		table string
		want  map[string]string
	}{
		{
			table: "tbls_fk_actions",
			want: map[string]string{
				"tbls_fk_actions_fk":          "FOREIGN KEY (child_post_id, child_user_id) REFERENCES comments (post_id, user_id) ON UPDATE CASCADE ON DELETE SET NULL",
				"tbls_fk_actions_external_fk": "FOREIGN KEY (external_user_id) REFERENCES tbls_fk_actions_external.users (id) ON UPDATE SET NULL",
			},
		},
		{
			table: "posts",
			want: map[string]string{
				"posts_user_id_fk": "FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE",
				"PRIMARY":          "PRIMARY KEY (id)",
				"user_id":          "UNIQUE KEY user_id (user_id, title)",
			},
		},
		{
			table: "comments",
			want: map[string]string{
				"comments_post_id_fk": "FOREIGN KEY (post_id) REFERENCES posts (id)",
				"comments_user_id_fk": "FOREIGN KEY (user_id) REFERENCES users (id)",
				"PRIMARY":             "PRIMARY KEY (id)",
				"post_id":             "UNIQUE KEY post_id (post_id, user_id)",
			},
		},
	} {
		t.Run(tt.table, func(t *testing.T) {
			tbl, err := sc.FindTableByName(tt.table)
			if err != nil {
				t.Fatal(err)
			}
			got := make(map[string]string, len(tbl.Constraints))
			for _, c := range tbl.Constraints {
				if _, dup := got[c.Name]; dup {
					t.Errorf("duplicate constraint %q", c.Name)
				}
				got[c.Name] = c.Def
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("constraint definitions (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCheckConstraintsOrder(t *testing.T) {
	if _, err := db.Exec(`DROP TABLE IF EXISTS tbls_check_order`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
CREATE TABLE tbls_check_order (
  a int,
  b int,
  CONSTRAINT chk_tbls_check_order_b CHECK (b > 0),
  CONSTRAINT chk_tbls_check_order_a CHECK (a > 0)
)`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.Exec(`DROP TABLE IF EXISTS tbls_check_order`)
	}()

	driver, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	sc := &schema.Schema{Name: "testdb"}
	if err := driver.Analyze(sc); err != nil {
		t.Fatal(err)
	}

	tbl, err := sc.FindTableByName("tbls_check_order")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, c := range tbl.Constraints {
		if c.Type == "CHECK" {
			got = append(got, c.Name)
		}
	}
	want := []string{"chk_tbls_check_order_a", "chk_tbls_check_order_b"}
	if len(got) != len(want) {
		t.Fatalf("got %v\nwant %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got %v\nwant %v", got, want)
			break
		}
	}
}
