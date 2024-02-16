package sql

import (
  "context"
  "database/sql"
  "github.com/betr-io/terraform-provider-mssql/mssql/model"
)

func (c *Connector) GetDatabaseRole(ctx context.Context, database, roleName string) (*model.DatabaseRole, error) {
  cmd := `IF @@VERSION LIKE 'Microsoft SQL Azure%' AND @database = 'master'
            BEGIN
              IF @ownerName = 'dbo' OR @ownerName = ''
                BEGIN
                  SELECT dp2.principal_id, dp2.name, dp2.owning_principal_id, '' AS ownerName FROM [sys].[database_principals] dp1 INNER JOIN [sys].[database_principals] dp2 ON dp1.principal_id = dp2.owning_principal_id AND dp2.type = 'R' AND dp2.name = @roleName
                END
              ELSE
                BEGIN
                  SELECT dp2.principal_id, dp2.name, dp2.owning_principal_id, dp1.name AS ownerName FROM [sys].[database_principals] dp1 INNER JOIN [sys].[database_principals] dp2 ON dp1.principal_id = dp2.owning_principal_id AND dp2.type = 'R' AND dp2.name = @roleName
                END
            END
          ELSE
            BEGIN
              SELECT dp2.principal_id, dp2.name, dp2.owning_principal_id, dp1.name AS ownerName FROM [sys].[database_principals] dp1 INNER JOIN [sys].[database_principals] dp2 ON dp1.principal_id = dp2.owning_principal_id AND dp2.type = 'R' AND dp2.name = @roleName
            END`
  var role model.DatabaseRole
  err := c.
    setDatabase(&database).
    QueryRowContext(ctx, cmd,
      func(r *sql.Row) error {
        return r.Scan(&role.RoleID, &role.RoleName, &role.OwnerId, &role.OwnerName)
      },
      sql.Named("database", database),
      sql.Named("roleName", roleName),
      sql.Named("ownerName", role.OwnerName),
    )
  if err != nil {
    if err == sql.ErrNoRows {
      return nil, nil
    }
    return nil, err
  }
  return &role, nil
}

func (c *Connector) CreateDatabaseRole(ctx context.Context, database, roleName string, ownerName string) error {
  cmd := `DECLARE @sql nvarchar(max);
          IF @ownerName = 'dbo' OR @ownerName = ''
            BEGIN
              SET @sql = 'CREATE ROLE ' + QuoteName(@roleName)
            END
          ELSE
            BEGIN
              SET @sql = 'CREATE ROLE ' + QuoteName(@roleName) + ' AUTHORIZATION ' + QuoteName(@ownerName)
            END
          EXEC (@sql);`

  return c.
    setDatabase(&database).
    ExecContext(ctx, cmd,
      sql.Named("roleName", roleName),
      sql.Named("ownerName", ownerName),
      sql.Named("database", database),
    )
}

func (c *Connector) DeleteDatabaseRole(ctx context.Context, database, roleName string) error {
  cmd := `DECLARE @stmt nvarchar(max)
          DECLARE @sql NVARCHAR(max)
          DECLARE @user_name NVARCHAR(max) = (SELECT USER_NAME())
          DECLARE @roleNameowner NVARCHAR(max) = (SELECT dp2.name FROM [sys].[database_principals] dp1 INNER JOIN [sys].[database_principals] dp2 ON dp1.principal_id = dp2.owning_principal_id AND dp1.name = @roleName)
          SET @sql =  'IF EXISTS (SELECT 1 FROM ' + QuoteName(@database) + '.[sys].[database_principals] dp1 INNER JOIN ' + QuoteName(@database) + '.[sys].[database_principals] dp2 ON dp1.principal_id = dp2.owning_principal_id AND dp1.name = ' + QuoteName(@roleName, '''') + ') ' +
                      'ALTER AUTHORIZATION ON ROLE:: [' + @roleNameowner + '] TO [' + @user_name + ']'
          EXEC sp_executesql @sql;
          SET @stmt = 'IF EXISTS (SELECT 1 FROM ' + QuoteName(@database) + '.[sys].[database_principals] WHERE [name] = ' + QuoteName(@roleName, '''') + ') ' +
                      'DROP ROLE ' + QuoteName(@roleName)
          EXEC (@stmt)`

  return c.
    setDatabase(&database).
    ExecContext(ctx, cmd,
      sql.Named("database", database),
      sql.Named("roleName", roleName),
    )
}

func (c *Connector) UpdateDatabaseRole(ctx context.Context, database string, roleId int, roleName string, ownerName string) error {
  cmd := `DECLARE @sql NVARCHAR(max)
          IF @ownerName = 'dbo' OR @ownerName = ''
            BEGIN
              SET @ownerName = (SELECT USER_NAME())
            END
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
      sql.Named("roleName", roleName),
      sql.Named("principalId", roleId),
      sql.Named("ownerName", ownerName),
    )
}
