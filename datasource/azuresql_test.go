package datasource

import (
	"strings"
	"testing"
)

func TestParseAzureSQLDatabaseName(t *testing.T) {
	tests := []struct {
		name    string
		urlstr  string
		want    string
		wantErr string
	}{
		{
			name:   "valid Service Principal DSN",
			urlstr: "azuresql://myhost.datawarehouse.fabric.microsoft.com?database=mydb&fedauth=ActiveDirectoryServicePrincipal&user+id=client123@tenant456&password=secret",
			want:   "mydb",
		},
		{
			name:    "missing database param",
			urlstr:  "azuresql://myhost.datawarehouse.fabric.microsoft.com?fedauth=ActiveDirectoryServicePrincipal",
			wantErr: "no database name in azuresql connection string",
		},
		{
			name:    "malformed URL",
			urlstr:  "azuresql://[invalid",
			wantErr: "invalid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAzureSQLDatabaseName(tt.urlstr)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("parseAzureSQLDatabaseName() expected error but got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("parseAzureSQLDatabaseName() error = %v, want error containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("parseAzureSQLDatabaseName() unexpected error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("parseAzureSQLDatabaseName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildAzureSQLConnStr(t *testing.T) {
	tests := []struct {
		name    string
		urlstr  string
		want    string
		wantErr string
	}{
		{
			name:   "full Service Principal DSN",
			urlstr: "azuresql://myhost.datawarehouse.fabric.microsoft.com?database=mydb&fedauth=ActiveDirectoryServicePrincipal&user+id=client123@tenant456&password=mysecret",
			want:   "Server=myhost.datawarehouse.fabric.microsoft.com;Database=mydb;fedauth=ActiveDirectoryServicePrincipal;User ID=client123@tenant456;Password=mysecret;Encrypt=true;TrustServerCertificate=false",
		},
		{
			name:    "missing database",
			urlstr:  "azuresql://myhost.datawarehouse.fabric.microsoft.com?fedauth=ActiveDirectoryServicePrincipal",
			wantErr: "no database name",
		},
		{
			name:    "missing host",
			urlstr:  "azuresql://?database=mydb",
			wantErr: "no server host",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildAzureSQLConnStr(tt.urlstr)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("buildAzureSQLConnStr() expected error but got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("buildAzureSQLConnStr() error = %v, want error containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("buildAzureSQLConnStr() unexpected error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("buildAzureSQLConnStr() = %v, want %v", got, tt.want)
			}
		})
	}
}
