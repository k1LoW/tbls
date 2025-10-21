package datasource

import (
	"strings"
	"testing"
)

func TestValidateDatabricksAuth(t *testing.T) {
	tests := []struct {
		name         string
		token        string
		clientID     string
		clientSecret string
		wantErr      string
	}{
		{
			name:    "valid PAT authentication",
			token:   "dapi1234567890",
			wantErr: "",
		},
		{
			name:         "valid OAuth authentication",
			clientID:     "client123",
			clientSecret: "secret456",
			wantErr:      "",
		},
		{
			name:    "no authentication provided",
			wantErr: "authentication required: provide either 'token' for PAT authentication or both 'client_id' and 'client_secret' for OAuth authentication",
		},
		{
			name:         "conflicting authentication methods",
			token:        "dapi1234567890",
			clientID:     "client123",
			clientSecret: "secret456",
			wantErr:      "conflicting authentication methods: provide either 'token' for PAT authentication OR 'client_id'/'client_secret' for OAuth authentication, not both",
		},
		{
			name:     "incomplete OAuth - only clientID",
			clientID: "client123",
			wantErr:  "authentication required: provide either 'token' for PAT authentication or both 'client_id' and 'client_secret' for OAuth authentication",
		},
		{
			name:         "incomplete OAuth - only clientSecret",
			clientSecret: "secret456",
			wantErr:      "authentication required: provide either 'token' for PAT authentication or both 'client_id' and 'client_secret' for OAuth authentication",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDatabricksAuth(tt.token, tt.clientID, tt.clientSecret)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("validateDatabricksAuth() expected error but got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("validateDatabricksAuth() error = %v, want error containing %v", err, tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("validateDatabricksAuth() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestBuildDatabricksDSN(t *testing.T) {
	tests := []struct {
		name         string
		host         string
		path         string
		catalog      string
		schema       string
		token        string
		clientID     string
		clientSecret string
		want         string
	}{
		{
			name:    "token auth with schema",
			host:    "host:443",
			path:    "/sql/1.0/warehouses/abc123",
			catalog: "my_catalog",
			schema:  "my_schema",
			token:   "dapi1234567890",
			want:    "token:dapi1234567890@host:443/sql/1.0/warehouses/abc123?catalog=my_catalog&schema=my_schema",
		},
		{
			name:    "token auth without schema",
			host:    "host:443",
			path:    "/sql/1.0/warehouses/abc123",
			catalog: "my_catalog",
			schema:  "",
			token:   "dapi1234567890",
			want:    "token:dapi1234567890@host:443/sql/1.0/warehouses/abc123?catalog=my_catalog",
		},
		{
			name:         "oauth auth with schema",
			host:         "host:443",
			path:         "/sql/1.0/warehouses/abc123",
			catalog:      "my_catalog",
			schema:       "my_schema",
			token:        "",
			clientID:     "client123",
			clientSecret: "secret456",
			want:         "host:443/sql/1.0/warehouses/abc123?authType=OauthM2M&clientID=client123&clientSecret=secret456&catalog=my_catalog&schema=my_schema",
		},
		{
			name:         "oauth auth without schema",
			host:         "host:443",
			path:         "/sql/1.0/warehouses/abc123",
			catalog:      "my_catalog",
			schema:       "",
			token:        "",
			clientID:     "client123",
			clientSecret: "secret456",
			want:         "host:443/sql/1.0/warehouses/abc123?authType=OauthM2M&clientID=client123&clientSecret=secret456&catalog=my_catalog",
		},
		{
			name:    "token auth with different catalog",
			host:    "dbc-123.cloud.databricks.com:443",
			path:    "/sql/1.0/warehouses/xyz",
			catalog: "production",
			schema:  "analytics",
			token:   "token_abc",
			want:    "token:token_abc@dbc-123.cloud.databricks.com:443/sql/1.0/warehouses/xyz?catalog=production&schema=analytics",
		},
		{
			name:         "oauth auth with different catalog",
			host:         "dbc-123.cloud.databricks.com:443",
			path:         "/sql/1.0/warehouses/xyz",
			catalog:      "production",
			schema:       "analytics",
			token:        "",
			clientID:     "oauth_client",
			clientSecret: "oauth_secret",
			want:         "dbc-123.cloud.databricks.com:443/sql/1.0/warehouses/xyz?authType=OauthM2M&clientID=oauth_client&clientSecret=oauth_secret&catalog=production&schema=analytics",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildDatabricksDSN(tt.host, tt.path, tt.catalog, tt.schema, tt.token, tt.clientID, tt.clientSecret)
			if got != tt.want {
				t.Errorf("buildDatabricksDSN() = %v, want %v", got, tt.want)
			}
		})
	}
}
