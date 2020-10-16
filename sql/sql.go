package sql

import (
  "context"
  "database/sql"
  "database/sql/driver"
  "fmt"
  "github.com/Azure/go-autorest/autorest/adal"
  "github.com/Azure/go-autorest/autorest/azure"
  mssql "github.com/denisenkom/go-mssqldb"
  "github.com/pkg/errors"
  "log"
  "net"
  "net/url"
  "strings"
  "time"
)

const DefaultPort = "1433"

type Connector struct {
  Host               string `json:"host"`
  Port               string `json:"port"`
  Database           string `json:"database"`
  Administrator      *AdministratorLogin
  AzureAdministrator *AzureAdministrator
  Timeout            time.Duration `json:"timeout,omitempty"`
}

type AdministratorLogin struct {
  Username string `json:"admin_username,omitempty"`
  Password string `json:"admin_password,omitempty"`
}

type AzureAdministrator struct {
  TenantID     string `json:"azure_tenant_id,omitempty"`
  ClientID     string `json:"azure_client_id,omitempty"`
  ClientSecret string `json:"azure_client_secret,omitempty"`
}

func (c *Connector) ID() string {
  host := c.Host
  if c.Port != DefaultPort {
    host = net.JoinHostPort(host, c.Port)
  }
  id := (&url.URL{
    Scheme: "sqlserver",
    Host:   host,
    Path:   c.Database,
  }).String()
  return id
}

func (c *Connector) PingContext(ctx context.Context) error {
  db, err := c.db()
  if err != nil {
    return err
  }

  err = db.PingContext(ctx)
  if err != nil {
    return errors.Wrap(err, "In ping")
  }

  return nil
}

// Execute an SQL statement and ignore the results
func (c *Connector) ExecContext(ctx context.Context, command string, args ...interface{}) error {
  db, err := c.db()
  if err != nil {
    return err
  }
  defer db.Close()

  _, err = db.ExecContext(ctx, command, args...)
  if err != nil {
    return err
  }

  return nil
}

func (c *Connector) QueryContext(ctx context.Context, query string, scanner func(*sql.Rows) error, args ...interface{}) error {
  db, err := c.db()
  if err != nil {
    return err
  }
  defer db.Close()

  rows, err := db.QueryContext(ctx, query, args...)
  if err != nil {
    return err
  }
  defer rows.Close()

  err = scanner(rows)
  if err != nil {
    return err
  }

  return nil
}

func (c *Connector) db() (*sql.DB, error) {
  if c == nil {
    panic("No connector")
  }
  conn, err := c.connector()
  if err != nil {
    return nil, err
  }
  if db, err := connectLoop(conn, c.Timeout); err != nil {
    return nil, err
  } else {
    return db, nil
  }
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

func connectLoop(connector driver.Connector, timeout time.Duration) (*sql.DB, error) {
  ticker := time.NewTicker(250 * time.Millisecond)
  defer ticker.Stop()

  timeoutExceeded := time.After(timeout)
  for {
    select {
    case <-timeoutExceeded:
      return nil, fmt.Errorf("db connection failed after %s timeout", timeout)

    case <-ticker.C:
      db, err := connect(connector)
      if err == nil {
        return db, nil
      }
      if strings.Contains(err.Error(), "Login error") {
        return nil, err
      }
      log.Println(errors.Wrap(err, "failed to connect to database"))
    }
  }
}

func connect(connector driver.Connector) (*sql.DB, error) {
  db := sql.OpenDB(connector)
  if err := db.Ping(); err != nil {
    db.Close()
    return nil, err
  }
  return db, nil
}
