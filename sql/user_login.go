package sql

import (
  "context"
  "database/sql"
  "errors"
  "fmt"
  "github.com/google/uuid"
  "strings"
  "terraform-provider-mssql/mssql"
)

func (c *Connector) GetUserLogin(ctx context.Context, username string) (*mssql.UserLogin, error) {
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

  var sid uuid.UUID
  if len(bytes) > 0 {
    bytes[0], bytes[3] = bytes[3], bytes[0]
    bytes[1], bytes[2] = bytes[2], bytes[1]
    bytes[4], bytes[5] = bytes[5], bytes[4]
    bytes[6], bytes[7] = bytes[7], bytes[6]
    sid, err = uuid.FromBytes(bytes)
    if err != nil {
      return nil, err
    }
  }
  if principalID != -1 {
    return &mssql.UserLogin{
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

func (c *Connector) GetDatabase() string {
  return c.Database
}

func (c *Connector) SetDatabase(database string) {
  c.Database = database
}

func (c *Connector) CreateUserLogin(ctx context.Context, database, username, password, schema string, roles []interface{}) error {
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
          SET @sql = 'IF EXISTS (SELECT 1 FROM sys.server_principals WHERE name = ' + QuoteName(@username, '''') + ') ' +
                     'DROP USER ' + QuoteName(@username);
          EXEC (@sql)`
  return c.ExecContext(ctx, cmd, sql.Named("username", username))
}

func (c *Connector) CreateAzureADLogin(ctx context.Context, username, schema string, roles []interface{}) error {
  dbRoles := make([]string, len(roles))
  for i, v := range roles {
    dbRoles[i] = v.(string)
  }
  cmd := `DECLARE @stmt nvarchar(max)
          SET @stmt = 'IF NOT EXISTS(SELECT 1 FROM sys.database_principals WHERE name = ' + QuoteName(@username, '''') + ') ' +
                      '  BEGIN ' +
                      '    CREATE USER ' + QuoteName(@username) + ' FROM EXTERNAL PROVIDER WITH DEFAULT_SCHEMA=' + QuoteName(@schema) + ';' +
                      '    DECLARE role_cur CURSOR FOR SELECT name FROM sys.database_principals WHERE type = ''R'' AND name IN (SELECT value FROM String_Split(''' + @roles + ''', '',''));' +
                      '    DECLARE @role nvarchar(100);' +
                      '    OPEN role_cur;' +
                      '    FETCH NEXT FROM role_cur INTO @role;' +
                      '    WHILE @@FETCH_STATUS = 0' +
                      '      BEGIN' +
                      '        DECLARE @sql nvarchar(max); ' +
                      '        SET @sql = ''ALTER ROLE '' + QuoteName(@role) + '' ADD MEMBER ' + QuoteName(@username) + ''';' +
                      '        EXEC (@sql);' +
                      '        FETCH NEXT FROM role_cur INTO @role;' +
                      '      END;' +
                      '    CLOSE role_cur;' +
                      '    DEALLOCATE role_cur;' +
                      '    SELECT 1;' +
                      '  END;' +
                      'ELSE ' +
                      '  BEGIN' +
                      '    SELECT 0;' +
                      '  END;'
          EXEC(@stmt)`
  var result int
  err := c.QueryContext(ctx, cmd, func(r *sql.Rows) error {
    if r.Next() {
      err := r.Scan(&result)
      if err != nil {
        return err
      }
    }
    return nil
  }, sql.Named("username", username), sql.Named("schema", schema), sql.Named("roles", strings.Join(dbRoles, ",")))
  if err != nil {
    return err
  }
  if result == 0 {
    return errors.New(fmt.Sprintf("User [%s] already exists.", username))
  }
  return nil
}

func (c *Connector) DeleteAzUserLogin(ctx context.Context, username string) error {
  cmd := `DECLARE @sql nvarchar(max)
          SET @sql = 'IF EXISTS (SELECT 1 FROM sys.server_principals WHERE name = ' + QuoteName(@username, '''') + ') ' +
                     'DROP LOGIN ' + QuoteName(@username);
          EXEC (@sql)`
  return c.ExecContext(ctx, cmd, sql.Named("username", username))
}


//func (c Connector) UpdateLogin(username string, password string) error {
//  cmd := `DECLARE @sql nvarchar(max)
//					SET @sql = 'IF EXISTS (SELECT 1 FROM [master].[sys].[server_principals] WHERE [name] = ' + QuoteName(@username, '''') + ') ' +
//										 'ALTER LOGIN ' + QuoteName(@username) + ' ' +
//										 'WITH PASSWORD = ' + QuoteName(@password, '''')
//					EXEC (@sql)`
//
//  return c.Execute(cmd, sql.Named("username", username), sql.Named("password", password))
//}
//
//func (c Connector) killSessionsForLogin(username string) error {
//  cmd := ` -- adapted from https://stackoverflow.com/a/5178097/38055
//	DECLARE sessionsToKill CURSOR FAST_FORWARD FOR
//			SELECT session_id
//			FROM sys.dm_exec_sessions
//			WHERE login_name = @username
//	OPEN sessionsToKill
//	DECLARE @sessionId INT
//	DECLARE @statement NVARCHAR(200)
//	FETCH NEXT FROM sessionsToKill INTO @sessionId
//	WHILE @@FETCH_STATUS = 0
//	BEGIN
//			PRINT 'Killing session ' + CAST(@sessionId AS NVARCHAR(20)) + ' for login ' + @username
//			SET @statement = 'KILL ' + CAST(@sessionId AS NVARCHAR(20))
//			EXEC sp_executesql @statement
//			FETCH NEXT FROM sessionsToKill INTO @sessionId
//	END
//	CLOSE sessionsToKill
//	DEALLOCATE sessionsToKill`
//
//  return c.Execute(cmd, sql.Named("username", username))
//}
