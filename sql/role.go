package sql

import (
  "context"
  "database/sql"
  "github.com/betr-io/terraform-provider-mssql/mssql/model"
)

func (c *Connector) GetRole(ctx context.Context, database, roleName string) (*model.Role, error) {
  var role model.Role

  err := c.
    setDatabase(&database).
    QueryRowContext(ctx,
      "SELECT principal_id, name FROM [sys].[database_principals]  WHERE [type] = 'R' AND [name] = @roleName",
      func(r *sql.Row) error {
        return r.Scan(&role.RoleID, &role.RoleName)
      },
      sql.Named("roleName", roleName),
    )
  if err != nil {
    if err == sql.ErrNoRows {
      return nil, nil
    }
    return nil, err
  }
  return &role, nil
}

func (c *Connector) CreateRole(ctx context.Context, database, roleName string) error {
  cmd := `DECLARE @sql nvarchar(max);
          SET @sql = 'CREATE ROLE ' + QuoteName(@roleName);
          EXEC (@sql);`

  return c.
    setDatabase(&database).
    ExecContext(ctx, cmd,
      sql.Named("roleName", roleName),
    )
}

func (c *Connector) DeleteRole(ctx context.Context, database, roleName string) error {
  cmd := `DECLARE @sql nvarchar(max);
          SET @sql = 'DROP ROLE ' + QuoteName(@roleName);
          EXEC (@sql);`

  return c.
    setDatabase(&database).
    ExecContext(ctx, cmd,
      sql.Named("roleName", roleName),
    )
}

func (c *Connector) UpdateRole(ctx context.Context, database string, role *model.Role) error {
  cmd := `DECLARE @sql nvarchar(max);
          DECLARE @old_role_name nvarchar(max) = (SELECT name FROM [sys].[database_principals]  WHERE [type] = 'R' AND [principal_id] = @principalId);
          SET @sql = 'ALTER ROLE ' + QuoteName(@old_role_name) + ' WITH NAME = ' + QuoteName(@roleName);
          EXEC (@sql);`
  return c.
    setDatabase(&database).
    ExecContext(ctx, cmd,
      sql.Named("roleName", role.RoleName),
      sql.Named("principalId", role.RoleID),
    )
}
