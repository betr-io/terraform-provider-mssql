package sql

import (
  "context"
  "database/sql"
  "github.com/betr-io/terraform-provider-mssql/mssql/model"
)

func (c *Connector) GetDatabaseCredential(ctx context.Context, database, credentialname string) (*model.DatabaseCredential, error) {
  var scopedcredential model.DatabaseCredential
  err := c.
    setDatabase(&database).
    QueryRowContext(ctx,
    "SELECT name, principal_id, credential_id, credential_identity FROM [sys].[database_scoped_credentials] WHERE [name] = @credentialname",
    func(r *sql.Row) error {
      return r.Scan(&scopedcredential.CredentialName, &scopedcredential.PrincipalID, &scopedcredential.CredentialID, &scopedcredential.IdentityName)
    },
    sql.Named("credentialname", credentialname),
  )
  if err != nil {
    if err == sql.ErrNoRows {
      return nil, nil
    }
    return nil, err
  }
  return &scopedcredential, nil
}

func (c *Connector) CreateDatabaseCredential(ctx context.Context, database, credentialname, identityname, secret string) error {
  cmd := `DECLARE @stmt nvarchar(max)
          SET @stmt = 'CREATE DATABASE SCOPED CREDENTIAL ' + QuoteName(@credentialname) + ' WITH IDENTITY = ' + QuoteName(@identityname, '''') + ', SECRET = ' + QuoteName(@secret, '''')
          EXEC (@stmt)`
  return c.
    setDatabase(&database).
    ExecContext(ctx, cmd,
      sql.Named("credentialname", credentialname),
      sql.Named("identityname", identityname),
      sql.Named("secret", secret),
    )
}

func (c *Connector) UpdateDatabaseCredential(ctx context.Context, database, credentialname, identityname, secret string) error {
  cmd := `DECLARE @stmt nvarchar(max)
          SET @stmt = 'ALTER DATABASE SCOPED CREDENTIAL ' + QuoteName(@credentialname) + ' WITH IDENTITY = ' + QuoteName(@identityname, '''')
          IF @secret != ''
            BEGIN
              SET @stmt = @stmt + ', SECRET = ' + QuoteName(@secret, '''')
            END
          EXEC (@stmt)`
  return c.
    setDatabase(&database).
    ExecContext(ctx, cmd,
      sql.Named("credentialname", credentialname),
      sql.Named("identityname", identityname),
      sql.Named("secret", secret),
    )
}

func (c *Connector) DeleteDatabaseCredential(ctx context.Context, database, credentialname string) error {
  cmd := `DECLARE @stmt nvarchar(max)
          SET @stmt = 'IF EXISTS (SELECT 1 FROM [sys].[database_scoped_credentials] WHERE [name] = ' + QuoteName(@credentialname, '''') + ') ' +
                      'DROP DATABASE SCOPED CREDENTIAL ' + QuoteName(@credentialname)
          EXEC (@stmt)`
  return c.
  setDatabase(&database).
    ExecContext(ctx, cmd,
      sql.Named("credentialname", credentialname),
    )
}