package sql

import (
  "context"
  "database/sql"
  "github.com/betr-io/terraform-provider-mssql/mssql/model"
)

func (c *Connector) GetDatabaseMasterkey(ctx context.Context, database string) (*model.DatabaseMasterkey, error) {
  var masterkey model.DatabaseMasterkey
  err := c.
    setDatabase(&database).
    QueryRowContext(ctx,
    "SELECT name, principal_id, symmetric_key_id, key_length, key_algorithm, algorithm_desc, CONVERT(VARCHAR(85), [key_guid], 1) FROM [sys].[symmetric_keys] WHERE name = '##MS_DatabaseMasterKey##'",
    func(r *sql.Row) error {
      return r.Scan(&masterkey.KeyName, &masterkey.PrincipalID, &masterkey.SymmetricKeyID, &masterkey.KeyLength, &masterkey.KeyAlgorithm, &masterkey.AlgorithmDesc, &masterkey.KeyGuid)
    },
  )
  if err != nil {
    if err == sql.ErrNoRows {
      return nil, nil
    }
    return nil, err
  }
  return &masterkey, nil
}

func (c *Connector) CreateDatabaseMasterkey(ctx context.Context, database, password string) error {
  cmd := `DECLARE @stmt nvarchar(max)
          SET @stmt = 'CREATE MASTER KEY ENCRYPTION BY PASSWORD = ' + QuoteName(@password, '''')
          EXEC (@stmt)`
  return c.
    setDatabase(&database).
    ExecContext(ctx, cmd,
      sql.Named("password", password),
    )
}

func (c *Connector) UpdateDatabaseMasterkey(ctx context.Context, database, password string) error {
  cmd := `DECLARE @stmt nvarchar(max)
          SET @stmt = 'ALTER MASTER KEY REGENERATE WITH ENCRYPTION BY PASSWORD = ' + QuoteName(@password, '''')
          EXEC (@stmt)`
  return c.
    setDatabase(&database).
    ExecContext(ctx, cmd,
      sql.Named("password", password),
    )
}

func (c *Connector) DeleteDatabaseMasterkey(ctx context.Context, database string) error {
  cmd := `DECLARE @stmt nvarchar(max)
          SET @stmt = 'IF EXISTS (SELECT 1 FROM [sys].[symmetric_keys] WHERE name = ' + QuoteName('##MS_DatabaseMasterKey##', '''') + ') ' +
                      'DROP MASTER KEY'
          EXEC (@stmt)`
  return c.
  setDatabase(&database).
    ExecContext(ctx, cmd,
    )
}
