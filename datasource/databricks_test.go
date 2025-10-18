package datasource

import (
	"testing"
)

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
