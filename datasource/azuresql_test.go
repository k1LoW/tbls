package datasource

import (
	"net/url"
	"strings"
	"testing"

	"github.com/k1LoW/tbls/config"
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
		wantUser    string
		wantPass    string
		tls         *config.TLS
		wantCert    string
		wantTrust   string
	}{
		{
			name:    "no database param or path returns error",
			urlstr:  "azuresql://myhost.example.com?fedauth=ActiveDirectoryServicePrincipal",
			wantErr: "no database name in azuresql connection string",
		},
		{
			name:    "no database param on bare host returns error",
			urlstr:  "azuresql://myhost.example.com",
			wantErr: "no database name in azuresql connection string",
		},
		{
			name:    "malformed URL returns error",
			urlstr:  "azuresql://[invalid",
			wantErr: "invalid",
		},
		{
			name:        "database from path with no credentials defaults to ActiveDirectoryDefault",
			urlstr:      "azuresql://myhost.example.com/mydb",
			wantScheme:  "sqlserver",
			wantDB:      "mydb",
			wantFedauth: "ActiveDirectoryDefault",
			wantEncrypt: "true",
		},
		{
			name:        "database from query with no credentials defaults to ActiveDirectoryDefault",
			urlstr:      "azuresql://myhost.example.com?database=mydb",
			wantScheme:  "sqlserver",
			wantDB:      "mydb",
			wantFedauth: "ActiveDirectoryDefault",
			wantEncrypt: "true",
		},
		{
			name:        "userinfo credentials with path database defaults to ActiveDirectoryServicePrincipal",
			urlstr:      "azuresql://cid@tid:sec@myhost.example.com/mydb",
			wantScheme:  "sqlserver",
			wantDB:      "mydb",
			wantFedauth: "ActiveDirectoryServicePrincipal",
			wantEncrypt: "true",
			wantUser:    "cid@tid",
			wantPass:    "sec",
		},
		{
			name:        "userinfo without tenant defaults to ActiveDirectoryServicePrincipal",
			urlstr:      "azuresql://cid:sec@myhost.example.com/mydb",
			wantScheme:  "sqlserver",
			wantDB:      "mydb",
			wantFedauth: "ActiveDirectoryServicePrincipal",
			wantEncrypt: "true",
			wantUser:    "cid",
			wantPass:    "sec",
		},
		{
			name:        "user id query param defaults to ActiveDirectoryServicePrincipal",
			urlstr:      "azuresql://myhost.example.com?database=mydb&user+id=cid@tid&password=sec",
			wantScheme:  "sqlserver",
			wantDB:      "mydb",
			wantFedauth: "ActiveDirectoryServicePrincipal",
			wantEncrypt: "true",
		},
		{
			name:        "fedauth present is not overwritten",
			urlstr:      "azuresql://myhost.example.com/mydb?fedauth=ActiveDirectoryPassword",
			wantScheme:  "sqlserver",
			wantDB:      "mydb",
			wantFedauth: "ActiveDirectoryPassword",
			wantEncrypt: "true",
		},
		{
			name:        "scheme is swapped to sqlserver",
			urlstr:      "azuresql://cid@tid:sec@myhost.example.com?database=mydb&fedauth=ActiveDirectoryServicePrincipal",
			wantScheme:  "sqlserver",
			wantDB:      "mydb",
			wantFedauth: "ActiveDirectoryServicePrincipal",
			wantEncrypt: "true",
			wantUser:    "cid@tid",
			wantPass:    "sec",
		},
		{
			// Password containing ';' in userinfo must be preserved
			name:        "password with semicolon in userinfo",
			urlstr:      "azuresql://cid@tid:sec%3Bret@myhost.example.com/mydb",
			wantScheme:  "sqlserver",
			wantDB:      "mydb",
			wantFedauth: "ActiveDirectoryServicePrincipal",
			wantEncrypt: "true",
			wantUser:    "cid@tid",
			wantPass:    "sec;ret",
		},
		{
			// Password containing ';' must be percent-encoded in query string,
			// not interpolated raw into an ADO key=value string.
			name:        "password with semicolon in query is percent-encoded, not injected",
			urlstr:      "azuresql://myhost.example.com?database=mydb&password=sec%3Bret",
			wantScheme:  "sqlserver",
			wantDB:      "mydb",
			wantFedauth: "ActiveDirectoryDefault",
			wantEncrypt: "true",
		},
		{
			// dsn.tls must reach the connection, not be silently dropped.
			name:        "dsn.tls.ca maps to certificate and disables trustservercertificate",
			urlstr:      "azuresql://cid@tid:sec@myhost.example.com/mydb",
			tls:         &config.TLS{CA: "/etc/ssl/azure.pem"},
			wantScheme:  "sqlserver",
			wantDB:      "mydb",
			wantFedauth: "ActiveDirectoryServicePrincipal",
			wantEncrypt: "true",
			wantCert:    "/etc/ssl/azure.pem",
			wantTrust:   "false",
		},
		{
			name:        "dsn.tls keeps an explicit encrypt=strict",
			urlstr:      "azuresql://cid@tid:sec@myhost.example.com/mydb?encrypt=strict",
			tls:         &config.TLS{CA: "/etc/ssl/azure.pem"},
			wantScheme:  "sqlserver",
			wantDB:      "mydb",
			wantFedauth: "ActiveDirectoryServicePrincipal",
			wantEncrypt: "strict",
			wantCert:    "/etc/ssl/azure.pem",
			wantTrust:   "false",
		},
		{
			name:    "dsn.tls conflicting with encrypt=disable returns error",
			urlstr:  "azuresql://cid@tid:sec@myhost.example.com/mydb?encrypt=disable",
			tls:     &config.TLS{CA: "/etc/ssl/azure.pem"},
			wantErr: "dsn.tls conflicts with encrypt=disable",
		},
		{
			name:    "dsn.tls client certificate is rejected",
			urlstr:  "azuresql://cid@tid:sec@myhost.example.com/mydb",
			tls:     &config.TLS{Cert: "/etc/ssl/c.pem", Key: "/etc/ssl/c.key"},
			wantErr: "not supported for driver 'sqlserver'",
		},
		{
			name:    "dsn.tls.cert without dsn.tls.key returns error",
			urlstr:  "azuresql://cid@tid:sec@myhost.example.com/mydb",
			tls:     &config.TLS{Cert: "/etc/ssl/c.pem"},
			wantErr: "must be set together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, gotDB, err := prepareAzureSQLURL(tt.urlstr, tt.tls)
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
			if u.Path != "" {
				t.Errorf("path = %q, want empty (to avoid msdsn instance name interpretation)", u.Path)
			}
			if q.Get("fedauth") != tt.wantFedauth {
				t.Errorf("fedauth = %q, want %q", q.Get("fedauth"), tt.wantFedauth)
			}
			if q.Get("encrypt") != tt.wantEncrypt {
				t.Errorf("encrypt = %q, want %q", q.Get("encrypt"), tt.wantEncrypt)
			}
			if tt.wantUser != "" && u.User.Username() != tt.wantUser {
				t.Errorf("user = %q, want %q", u.User.Username(), tt.wantUser)
			}
			if tt.wantPass != "" {
				pass, _ := u.User.Password()
				if pass != tt.wantPass {
					t.Errorf("password = %q, want %q", pass, tt.wantPass)
				}
			}

			if tt.wantCert != "" && q.Get("certificate") != tt.wantCert {
				t.Errorf("certificate = %q, want %q", q.Get("certificate"), tt.wantCert)
			}
			if tt.wantTrust != "" && q.Get("trustservercertificate") != tt.wantTrust {
				t.Errorf("trustservercertificate = %q, want %q", q.Get("trustservercertificate"), tt.wantTrust)
			}

			// msdsn splits ADO key=value connection strings on ';', so a ';' that
			// reaches the query unencoded could inject extra options. RawQuery is
			// the encoded form, so it must never contain one.
			if strings.Contains(u.RawQuery, ";") {
				t.Errorf("raw query contains an unencoded ';': %q", u.RawQuery)
			}
		})
	}
}
