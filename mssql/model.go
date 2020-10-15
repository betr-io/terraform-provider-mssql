package mssql

import (
  "context"
  "database/sql"
  "database/sql/driver"
  "github.com/Azure/go-autorest/autorest/adal"
  "github.com/Azure/go-autorest/autorest/azure"
  mssql "github.com/denisenkom/go-mssqldb"
  "net/url"
)

type Connector struct {
  Host          string
  Port          int
  Database      string
  Administrator *struct {
    Username string
    Password string
  }
  AzureAdministrator *struct {
    TenantID     string
    ClientID     string
    ClientSecret string
  }
  _db *sql.DB
}

func (c *Connector) ID() string {
  return c.Host + "/" + c.Database
}

func (c *Connector) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
  if db, err := c.db(); err != nil {
    return nil, err
  } else {
    return db.QueryContext(ctx, query, args...)
  }
}

func (c *Connector) db() (*sql.DB, error) {
  if c._db == nil {
    conn, err := c.connector()
    if err != nil {
      return nil, err
    }
    c._db = sql.OpenDB(conn)
  }
  return c._db, nil
}

func (c *Connector) connector() (driver.Connector, error) {
  query := url.Values{
    "database": {c.Database},
  }
  connectionString := (&url.URL{
    Scheme:   "sqlserver",
    User:     c.userPassword(),
    Host:     c.Host,
    RawQuery: query.Encode(),
  }).String()
  if c.Administrator != nil {
    return mssql.NewConnector(connectionString)
  }
  return mssql.NewAccessTokenConnector(connectionString, func() (string, error) { return c.tokenProvider() })
}

func (c *Connector) userPassword() *url.Userinfo {
  if c.Administrator != nil {
    return url.UserPassword(c.Administrator.Username, c.Administrator.Password)
  }
  return nil
}

func (c *Connector) tokenProvider() (string, error) {
  const resourceID = "https://database.windows.net/"

  admin := c.AzureAdministrator
  oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, admin.TenantID)
  if err != nil {
    return "", err
  }

  spt, err := adal.NewServicePrincipalToken(*oauthConfig, admin.ClientID, admin.ClientSecret, resourceID)
  if err != nil {
    return "", err
  }

  err = spt.EnsureFresh()
  if err != nil {
    return "", err
  }

  return spt.OAuthToken(), nil
}
