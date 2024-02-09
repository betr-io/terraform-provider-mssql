package sql

import (
  "context"
  "database/sql"
  "github.com/betr-io/terraform-provider-mssql/mssql/model"
)

func (c *Connector) GetRole(ctx context.Context, database, roleName string) (*model.DatabaseRole, error) {
  var role model.DatabaseRole

  err := c.
    setDatabase(&database).
    QueryRowContext(ctx,
      "SELECT dp2.principal_id, dp2.name, dp1.name AS ownerName FROM [sys].[database_principals] dp1 INNER JOIN [sys].[database_principals] dp2 ON dp1.principal_id = dp2.owning_principal_id AND dp2.type = 'R' AND dp2.name = @roleName",
      func(r *sql.Row) error {
        return r.Scan(&role.RoleID, &role.RoleName, &role.OwnerName)
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

func (c *Connector) CreateRole(ctx context.Context, database, roleName string, ownerName string) error {
  cmd := `DECLARE @sql nvarchar(max);
          SET @sql = 'CREATE ROLE ' + QuoteName(@roleName)
          IF @ownerName != ''
            BEGIN
              SET @sql = @sql + ' AUTHORIZATION ' + QuoteName(@ownerName);
            END
          EXEC (@sql);`

  return c.
    setDatabase(&database).
    ExecContext(ctx, cmd,
      sql.Named("roleName", roleName),
      sql.Named("ownerName", ownerName),
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

func (c *Connector) UpdateRole(ctx context.Context, database string, role *model.DatabaseRole) error {
  cmd := `DECLARE @sql NVARCHAR(max)
          DECLARE @old_role_name NVARCHAR(max) = (SELECT name FROM [sys].[database_principals]  WHERE [type] = 'R' AND [principal_id] = @principalId)
          DECLARE @old_owner_name NVARCHAR(max) = (SELECT dp1.name AS ownerName FROM [sys].[database_principals] dp1 INNER JOIN [sys].[database_principals] dp2 ON dp1.principal_id = dp2.owning_principal_id AND dp2.type = 'R' AND dp2.name = @roleName)
          IF (@old_role_name != @roleName) AND (@old_owner_name = @ownerName)
          BEGIN
            SET @sql = 'ALTER ROLE ' + QuoteName(@old_role_name) + ' WITH NAME = ' + QuoteName(@roleName)
          END
          IF (@old_owner_name != @ownerName) AND (@old_role_name = @roleName)
          BEGIN
            SET @sql = 'ALTER AUTHORIZATION ON ROLE:: [' + @roleName + '] TO [' + @ownerName + ']'
          END
          ELSE
          BEGIN
            SET @sql = 'ALTER ROLE ' + QuoteName(@old_role_name) + ' WITH NAME = ' + QuoteName(@roleName) + ';'
            SET @sql = @sql + 'ALTER AUTHORIZATION ON ROLE:: [' + @roleName + '] TO [' + @ownerName + '];'
          END
          EXEC (@sql)`

  return c.
    setDatabase(&database).
    ExecContext(ctx, cmd,
      sql.Named("roleName", role.RoleName),
      sql.Named("principalId", role.RoleID),
      sql.Named("ownerName", role.OwnerName),
    )
}
