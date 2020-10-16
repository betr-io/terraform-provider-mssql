package sql

import (
  "context"
  "database/sql"
  "fmt"
  "github.com/google/uuid"
  gouuid "github.com/hashicorp/go-uuid"
  "strings"
)

type Login struct {
  PrincipalID int64
  Type        string
  Username    string
  SID         uuid.UUID
  Schema      string
  Roles       []string
}

func (c *Connector) GetUserLogin(ctx context.Context, username string) (*Login, error) {
  cmd := `WITH CTE_Roles (role_principal_id) AS
          (
              SELECT role_principal_id FROM sys.database_role_members WHERE member_principal_id = DATABASE_PRINCIPAL_ID(@username)
              UNION ALL
              SELECT drm.role_principal_id FROM sys.database_role_members drm
                  INNER JOIN CTE_Roles CR ON drm.member_principal_id = CR.role_principal_id
         )
         SELECT p.principal_id, p.type, p.default_schema_name, p.sid, STRING_AGG(USER_NAME(r.role_principal_id), ',') FROM sys.database_principals p, CTE_Roles r
         WHERE p.name = @username
         GROUP BY p.principal_id, p.type, p.default_schema_name, p.sid;`
  var principalID int64 = -1
  var typ, schema string
  var bytes []byte
  var roles string

  err := c.QueryContext(
    ctx,
    cmd,
    func(r *sql.Rows) error {
      for r.Next() {
        err := r.Scan(&principalID, &typ, &schema, &bytes, &roles)
        if err != nil {
          return err
        }
      }
      return nil
    },
    sql.Named("username", username),
  )

  if err != nil {
    return nil, err
  }

  sid, err := uuid.FromBytes(bytes)
  if err != nil {
    return nil, err
  }
  if principalID != -1 {
    return &Login{
      PrincipalID: principalID,
      Type:        typ,
      Username:    username,
      SID:         sid,
      Schema:      schema,
      Roles:       strings.Split(roles, ","),
    }, nil
  }

  return nil, nil
}

func (c *Connector) CreateUserLogin(ctx context.Context, database, username, password string) error {
  cmd := `DECLARE @stmt nvarchar(max)
          SET @stmt = 'CREATE LOGIN ' + QuoteName(@username) + ' ' +
                      'WITH PASSWORD = ' + QuoteName(@password, '''') + ', ' +
                      'DEFAULT_DATABASE = ' + QuoteName(@database)
          EXEC (@stmt)`
  return c.ExecContext(ctx, cmd, sql.Named("database", database), sql.Named("username", username), sql.Named("password", password))
}

func (c *Connector) UpdateUserLogin(ctx context.Context, username string, password string) error {
  cmd := `DECLARE @sql nvarchar(max)
          SET @sql = 'IF EXISTS (SELECT 1 FROM [master].[sys].[server_principals] WHERE [name] = ' + QuoteName(@username, '''') + ') ' +
                     'ALTER LOGIN ' + QuoteName(@username) + ' ' +
                     'WITH PASSWORD = ' + QuoteName(@password, '''')
          EXEC (@sql)`
  return c.ExecContext(ctx, cmd, sql.Named("username", username), sql.Named("password", password))
}

func (c *Connector) DeleteUserLogin(ctx context.Context, username string) error {
  cmd := `DECLARE @sql nvarchar(max)
          SET @sql = 'IF EXISTS (SELECT 1 FROM [master].[sys].[server_principals] WHERE [name] = ' + QuoteName(@username, '''') + ') ' +
                     'DROP LOGIN ' + QuoteName(@username)
          EXEC (@sql)`
  return c.ExecContext(ctx, cmd, sql.Named("username", username))
}

func (c *Connector) CreateAzureADLogin(ctx context.Context, username, sid, schema string, roles []interface{}) error {
  sid, err := convertToSid(sid)
  if err != nil {
    return err
  }
  dbRoles := make([]string, len(roles))
  for i, v := range roles {
    dbRoles[i] = v.(string)
  }
  cmd := `DECLARE @stmt nvarchar(max)
          SET @stmt = 'IF NOT EXISTS(SELECT 1 FROM sys.database_principals WHERE name = ' + QuoteName(@username, '''') + ') ' +
                      '  BEGIN ' +
                      '    CREATE USER ' + QuoteName(@username) + ' WITH DEFAULT_SCHEMA=' + QuoteName(@schema) + ', SID = ' + @sid + ', TYPE = E;' +
                      '    DECLARE role_cur CURSOR FOR SELECT value FROM String_Split(''' @roles ''', '','');' +
                      '    DECLARE @role nvarchar(100);' +
                      '    OPEN role_cur;' +
                      '    FETCH NEXT FROM role_cur INTO @role;' +
                      '    WHILE @@FETCH_STATUS = 0' +
                      '      BEGIN' +
                      '        IF EXISTS(SELECT 1 FROM sys.database_principals WHERE )' +
                      '          BEGIN' +
                      '            ALTER ROLE ' + QuoteName(@role) + ' ADD MEMBER ' + QuoteName(@username) + ';' +
                      '          END;' +
                      '        FETCH NEXT FROM role_cur INTO @role;' +
                      '      END;' +
                      '    CLOSE role_cur;' +
                      '    DEALLOCATE role_cur;' +
                      '    SELECT 1;' +
                      '  END' +
                      'ELSE ' +
                      '  BEGIN' +
                      '    SELECT 0;' +
                      '  END;' +
          EXEC(@stmt)`
  return c.ExecContext(ctx, cmd, sql.Named("username", username), sql.Named("schema", schema), sql.Named("sid", sid), sql.Named("roles", strings.Join(dbRoles, ",")))
}

func convertToSid(s string) (string, error) {
  u, err := gouuid.ParseUUID(s)
  if err != nil {
    return "", err
  }
  return fmt.Sprintf("0x%x", u), nil
}
