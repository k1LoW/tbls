package datasource

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"github.com/databricks/databricks-sdk-go"
	"github.com/databricks/databricks-sdk-go/service/catalog"
	driversDatabricks "github.com/k1LoW/tbls/drivers/databricks"
	"github.com/k1LoW/tbls/schema"
)

func validateDatabricksAuth(token, clientID, clientSecret string) error {
	hasToken := token != ""
	hasOAuth := clientID != "" && clientSecret != ""

	if !hasToken && !hasOAuth {
		return errors.New("authentication required: provide either 'token' for PAT authentication or both 'client_id' and 'client_secret' for OAuth authentication")
	}

	if hasToken && hasOAuth {
		return errors.New("conflicting authentication methods: provide either 'token' for PAT authentication OR 'client_id'/'client_secret' for OAuth authentication, not both")
	}

	if (clientID == "") != (clientSecret == "") {
		return errors.New("incomplete OAuth credentials: both 'client_id' and 'client_secret' are required for OAuth authentication")
	}

	return nil
}

func AnalyzeDatabricks(urlstr string) (_ *schema.Schema, err error) {
	s := &schema.Schema{}

	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}

	catalogName := u.Query().Get("catalog")
	if catalogName == "" {
		return nil, errors.New("no catalog name in the connection string")
	}
	schemaName := u.Query().Get("schema")

	token := u.Query().Get("token")
	clientID := u.Query().Get("client_id")
	clientSecret := u.Query().Get("client_secret")

	if err := validateDatabricksAuth(token, clientID, clientSecret); err != nil {
		return nil, err
	}

	if schemaName != "" {
		s.Name = fmt.Sprintf("%s.%s", catalogName, schemaName)
	} else {
		s.Name = catalogName
	}

	workspaceURL := fmt.Sprintf("https://%s", u.Host)

	cfg := &databricks.Config{
		Host: workspaceURL,
	}
	if token != "" {
		cfg.Token = token
	} else {
		cfg.ClientID = clientID
		cfg.ClientSecret = clientSecret
	}

	w, err := databricks.NewWorkspaceClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Databricks workspace client: %w", err)
	}

	databricksDSN := buildDatabricksDSN(u.Host, u.Path, catalogName, schemaName, token, clientID, clientSecret)

	db, err := sql.Open("databricks", databricksDSN)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	apiClient := &sdkTablesAPIClient{
		client:      w,
		catalogName: catalogName,
	}
	explicitSchema := schemaName != ""
	driver := driversDatabricks.New(db, apiClient, explicitSchema)

	if err := driver.Analyze(s); err != nil {
		return nil, err
	}

	return s, nil
}

func buildDatabricksDSN(host, path, catalog, schema, token, clientID, clientSecret string) string {
	baseParams := fmt.Sprintf("catalog=%s", catalog)
	if schema != "" {
		baseParams = fmt.Sprintf("%s&schema=%s", baseParams, schema)
	}

	if token != "" {
		return fmt.Sprintf("token:%s@%s%s?%s", token, host, path, baseParams)
	}
	return fmt.Sprintf("%s%s?authType=OauthM2M&clientID=%s&clientSecret=%s&%s",
		host, path, clientID, clientSecret, baseParams)
}

type sdkTablesAPIClient struct {
	client      *databricks.WorkspaceClient
	catalogName string
}

func (c *sdkTablesAPIClient) GetTable(ctx context.Context, catalogName, schemaName, tableName string) (*driversDatabricks.TableInfo, error) {
	fullName := fmt.Sprintf("%s.%s.%s", catalogName, schemaName, tableName)

	table, err := c.client.Tables.Get(ctx, catalog.GetTableRequest{
		FullName: fullName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get table info: %w", err)
	}

	columns := make([]driversDatabricks.ColumnInfo, len(table.Columns))
	for i, col := range table.Columns {
		columns[i] = driversDatabricks.ColumnInfo{
			Name:     col.Name,
			TypeName: string(col.TypeName),
			TypeText: col.TypeText,
			TypeJSON: col.TypeJson,
			Position: int(col.Position),
			Nullable: col.Nullable,
		}
	}

	return &driversDatabricks.TableInfo{
		FullName: table.FullName,
		Columns:  columns,
	}, nil
}
