package datasource

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"github.com/k1LoW/tbls/drivers/databricks"
	"github.com/k1LoW/tbls/schema"
)

// AnalyzeDatabricks analyze `databricks://`
func AnalyzeDatabricks(urlstr string) (_ *schema.Schema, err error) {
	s := &schema.Schema{}

	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}

	catalog := u.Query().Get("catalog")
	if catalog == "" {
		return nil, errors.New("no catalog name in the connection string")
	}
	schemaName := u.Query().Get("schema")
	if schemaName == "" {
		return nil, errors.New("no schema name in the connection string")
	}

	// Extract authentication parameters
	token := u.Query().Get("token")
	clientID := u.Query().Get("client_id")
	clientSecret := u.Query().Get("client_secret")

	// Validate authentication parameter combinations
	hasToken := token != ""
	hasOAuth := clientID != "" && clientSecret != ""

	if !hasToken && !hasOAuth {
		return nil, errors.New("authentication required: provide either 'token' for PAT authentication or both 'client_id' and 'client_secret' for OAuth authentication")
	}

	if hasToken && hasOAuth {
		return nil, errors.New("conflicting authentication methods: provide either 'token' for PAT authentication OR 'client_id'/'client_secret' for OAuth authentication, not both")
	}

	if clientID != "" && clientSecret == "" {
		return nil, errors.New("incomplete OAuth credentials: 'client_secret' is required when 'client_id' is provided")
	}

	if clientSecret != "" && clientID == "" {
		return nil, errors.New("incomplete OAuth credentials: 'client_id' is required when 'client_secret' is provided")
	}

	s.Name = fmt.Sprintf("%s.%s", catalog, schemaName)

	// Build databricks driver DSN based on authentication method
	var databricksDSN string
	if hasToken {
		// PAT token authentication: token:TOKEN@host:port/path?catalog=CATALOG&schema=SCHEMA
		databricksDSN = fmt.Sprintf("token:%s@%s%s?catalog=%s&schema=%s", token, u.Host, u.Path, catalog, schemaName)
	} else {
		// OAuth client credentials authentication: host:port/path?authType=OauthM2M&clientID=ID&clientSecret=SECRET&catalog=CATALOG&schema=SCHEMA
		databricksDSN = fmt.Sprintf("%s%s?authType=OauthM2M&clientID=%s&clientSecret=%s&catalog=%s&schema=%s", 
			u.Host, u.Path, clientID, clientSecret, catalog, schemaName)
	}

	db, err := sql.Open("databricks", databricksDSN)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	driver := databricks.New(db)
	if err := driver.Analyze(s); err != nil {
		return nil, err
	}

	return s, nil
}
