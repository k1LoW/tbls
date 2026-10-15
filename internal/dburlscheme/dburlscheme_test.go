package dburlscheme

import (
	"testing"

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
