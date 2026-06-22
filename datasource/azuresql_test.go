package datasource

import (
	"net/url"
	"strings"
	"testing"
)

func TestPrepareAzureSQLURL(t *testing.T) {
	tests := []struct {
		name        string
		urlstr      string
		wantErr     string
		wantScheme  string
		wantFedauth string
		wantEncrypt string
		wantDB      string
	}{
		{
			name:        "no database param returns error",
			urlstr:      "azuresql://myhost.example.com?fedauth=ActiveDirectoryServicePrincipal",
			wantErr:     "no database name in azuresql connection string",
		},
		{
			name:        "malformed URL returns error",
			urlstr:      "azuresql://[invalid",
			wantErr:     "invalid",
		},
		{
			name:        "fedauth absent gets default",
			urlstr:      "azuresql://myhost.example.com?database=mydb",
			wantScheme:  "sqlserver",
			wantDB:      "mydb",
			wantFedauth: "ActiveDirectoryServicePrincipal",
			wantEncrypt: "true",
		},
		{
			name:        "fedauth present is not overwritten",
			urlstr:      "azuresql://myhost.example.com?database=mydb&fedauth=ActiveDirectoryPassword",
			wantScheme:  "sqlserver",
			wantDB:      "mydb",
			wantFedauth: "ActiveDirectoryPassword",
			wantEncrypt: "true",
		},
		{
			name:        "scheme is swapped to sqlserver",
			urlstr:      "azuresql://myhost.example.com?database=mydb&fedauth=ActiveDirectoryServicePrincipal",
			wantScheme:  "sqlserver",
			wantDB:      "mydb",
			wantFedauth: "ActiveDirectoryServicePrincipal",
			wantEncrypt: "true",
		},
		{
			// Password containing ';' must be percent-encoded in the output URL,
			// not interpolated raw into an ADO key=value string.
			name:        "password with semicolon is percent-encoded, not injected",
			urlstr:      "azuresql://myhost.example.com?database=mydb&password=sec%3Bret",
			wantScheme:  "sqlserver",
			wantDB:      "mydb",
			wantFedauth: "ActiveDirectoryServicePrincipal",
			wantEncrypt: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, gotDB, err := prepareAzureSQLURL(tt.urlstr)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotDB != tt.wantDB {
				t.Errorf("dbName = %q, want %q", gotDB, tt.wantDB)
			}

			u, err := url.Parse(gotURL)
			if err != nil {
				t.Fatalf("output URL not parseable: %v", err)
			}
			q := u.Query()

			if u.Scheme != tt.wantScheme {
				t.Errorf("scheme = %q, want %q", u.Scheme, tt.wantScheme)
			}
			if q.Get("fedauth") != tt.wantFedauth {
				t.Errorf("fedauth = %q, want %q", q.Get("fedauth"), tt.wantFedauth)
			}
			if q.Get("encrypt") != tt.wantEncrypt {
				t.Errorf("encrypt = %q, want %q", q.Get("encrypt"), tt.wantEncrypt)
			}

			// Verify password with ';' is not raw in the URL string
			if strings.Contains(tt.urlstr, "%3B") {
				if strings.Contains(gotURL, ";") {
					// Strip the scheme+host part to avoid false positive on sqlserver://host
					rawQuery := u.RawQuery
					if strings.Contains(rawQuery, "=sec;ret") {
						t.Error("password ';' leaked unencoded into query string — ADO injection possible")
					}
				}
			}
		})
	}
}
