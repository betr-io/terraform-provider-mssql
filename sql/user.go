package sql

import (
  "context"
  "database/sql"
  "strings"
  "terraform-provider-mssql/mssql/model"
)

func (c *Connector) GetUser(ctx context.Context, database, username string) (*model.User, error) {
  cmd := `DECLARE @stmt nvarchar(max)
          SET @stmt = 'WITH CTE_Roles (principal_id, role_principal_id) AS ' +
                      '(' +
                      '  SELECT member_principal_id, role_principal_id FROM ' + QuoteName(@database) + '.[sys].[database_role_members] WHERE member_principal_id = DATABASE_PRINCIPAL_ID(' + QuoteName(@username, '''') + ')' +
                      '  UNION ALL ' +
                      '  SELECT member_principal_id, drm.role_principal_id FROM ' + QuoteName(@database) + '.[sys].[database_role_members] drm' +
                      '    INNER JOIN CTE_Roles cr ON drm.member_principal_id = cr.role_principal_id' +
                      ') ' +
                      'SELECT p.principal_id, p.name, p.authentication_type_desc, COALESCE(p.default_schema_name, ''''), COALESCE(p.default_language_name, ''''), COALESCE(sl.name, ''''), COALESCE(STRING_AGG(USER_NAME(r.role_principal_id), '',''), '''') ' +
                      'FROM ' + QuoteName(@database) + '.[sys].[database_principals] p' +
                      '  LEFT JOIN CTE_Roles r ON p.principal_id = r.principal_id ' +
                      '  LEFT JOIN [master].[sys].[sql_logins] sl ON p.sid = sl.sid ' +
                      'WHERE p.name = ' + QuoteName(@username, '''') + ' ' +
                      'GROUP BY p.principal_id, p.name, p.authentication_type_desc, p.default_schema_name, p.default_language_name, sl.name'
          EXEC (@stmt)`
  var (
    user  model.User
    roles string
  )
  err := c.
    setDatabase(&database).
    QueryRowContext(ctx, cmd,
      func(r *sql.Row) error {
        return r.Scan(&user.PrincipalID, &user.Username, &user.AuthType, &user.DefaultSchema, &user.DefaultLanguage, &user.LoginName, &roles)
      },
      sql.Named("database", database),
      sql.Named("username", username),
    )
  if err != nil {
    if err == sql.ErrNoRows {
      return nil, nil
    }
    return nil, err
  }
  if roles == "" {
    user.Roles = make([]string, 0)
  } else {
    user.Roles = strings.Split(roles, ",")
  }
  return &user, nil
}

func (c *Connector) CreateUser(ctx context.Context, database string, user *model.User) error {
  cmd := `DECLARE @stmt nvarchar(max)
          DECLARE @language nvarchar(max) = @defaultLanguage
          IF @language = '' SET @language = NULL
          IF @authType = 'INSTANCE'
            BEGIN
              SET @stmt = 'CREATE USER ' + QuoteName(@username) + ' FOR LOGIN ' + QuoteName(@loginName) + ' ' +
                          'WITH DEFAULT_SCHEMA = ' + QuoteName(@defaultSchema)
            END
          IF @authType = 'DATABASE'
            BEGIN
              SET @stmt = 'CREATE USER ' + QuoteName(@username) + ' WITH PASSWORD = ' + QuoteName(@password, '''') + ', ' +
                          'DEFAULT_SCHEMA = ' + QuoteName(@defaultSchema)
              IF NOT @@VERSION LIKE 'Microsoft SQL Azure%'
                BEGIN
                  SET @stmt = @stmt + ', DEFAULT_LANGUAGE = ' + Coalesce(QuoteName(@language), 'NONE')
                END
            END
          IF @authType = 'EXTERNAL'
            BEGIN
              IF @@VERSION LIKE 'Microsoft SQL Azure%'
                BEGIN
                  SET @stmt = 'CREATE USER ' + QuoteName(@username) + ' FROM EXTERNAL PROVIDER'
                END
              ELSE
                BEGIN
                  SET @stmt = 'CREATE USER ' + QuoteName(@username) + ' FOR LOGIN ' + QuoteName(@username) + ' FROM EXTERNAL PROVIDER ' +
                              'WITH DEFAULT_SCHEMA = ' + QuoteName(@defaultSchema) + ', ' +
                              'DEFAULT_LANGUAGE = ' + Coalesce(QuoteName(@language), 'NONE')
                END
            END
          SET @stmt = @stmt + '; ' +
                      'DECLARE role_cur CURSOR FOR SELECT name FROM ' + QuoteName(@database) + '.[sys].[database_principals] WHERE type = ''R'' AND name != ''public'' AND name IN (SELECT value FROM String_Split(' + QuoteName(@roles, '''') + ', '',''));' +
                      'DECLARE @role nvarchar(max);' +
                      'OPEN role_cur;' +
                      'FETCH NEXT FROM role_cur INTO @role;' +
                      'WHILE @@FETCH_STATUS = 0' +
                      '  BEGIN' +
                      '    DECLARE @sql nvarchar(max);' +
                      '    SET @sql = ''ALTER ROLE '' + QuoteName(@role) + '' ADD MEMBER ' + QuoteName(@username) + ''';' +
                      '    EXEC (@sql);' +
                      '    FETCH NEXT FROM role_cur INTO @role;' +
                      '  END;' +
                      'CLOSE role_cur;' +
                      'DEALLOCATE role_cur;'
          EXEC (@stmt)`
  _, err := c.GetLogin(ctx, user.LoginName)
  if err != nil {
    return err
  }
  return c.
    setDatabase(&database).
    ExecContext(ctx, cmd,
      sql.Named("database", database),
      sql.Named("username", user.Username),
      sql.Named("loginName", user.LoginName),
      sql.Named("password", user.Password),
      sql.Named("authType", user.AuthType),
      sql.Named("defaultSchema", user.DefaultSchema),
      sql.Named("defaultLanguage", user.DefaultLanguage),
      sql.Named("roles", strings.Join(user.Roles, ",")),
    )
}

func (c *Connector) UpdateUser(ctx context.Context, database string, user *model.User) error {
  cmd := `DECLARE @stmt nvarchar(max)
          SET @stmt = 'ALTER USER ' + QuoteName(@username) + ' '
          DECLARE @language nvarchar(max) = @defaultLanguage
          IF @language = '' SET @language = NULL
          SET @stmt = @stmt + 'WITH DEFAULT_SCHEMA = ' + QuoteName(@defaultSchema)
          DECLARE @auth_type nvarchar(max) = (SELECT authentication_type_desc FROM [sys].[database_principals] WHERE name = @username)
          IF NOT @@VERSION LIKE 'Microsoft SQL Azure%' AND @auth_type != 'INSTANCE'
            BEGIN
              SET @stmt = @stmt + ', DEFAULT_LANGUAGE = ' + Coalesce(QuoteName(@language), 'NONE')
            END
          SET @stmt = @stmt + '; ' +
                      'DECLARE @sql nvarchar(max);' +
                      'DECLARE @role nvarchar(max);' +
                      'DECLARE del_role_cur CURSOR FOR SELECT name FROM ' + QuoteName(@database) + '.[sys].[database_principals] WHERE type = ''R'' AND name != ''public'' AND name IN (SELECT name FROM ' + QuoteName(@database) + '.[sys].[database_role_members] drm, ' + QuoteName(@database) + '.[sys].[database_principals] db WHERE drm.member_principal_id = DATABASE_PRINCIPAL_ID(' + QuoteName(@username, '''') + ') AND drm.role_principal_id = db.principal_id) AND name NOT IN(SELECT value FROM STRING_SPLIT(' + QuoteName(@roles, '''') + ', '',''));' +
                      'DECLARE add_role_cur CURSOR FOR SELECT name FROM ' + QuoteName(@database) + '.[sys].[database_principals] WHERE type = ''R'' AND name != ''public'' AND name NOT IN (SELECT name FROM ' + QuoteName(@database) + '.[sys].[database_role_members] drm, ' + QuoteName(@database) + '.[sys].[database_principals] db WHERE drm.member_principal_id = DATABASE_PRINCIPAL_ID(' + QuoteName(@username, '''') + ') AND drm.role_principal_id = db.principal_id) AND name IN(SELECT value FROM STRING_SPLIT(' + QuoteName(@roles, '''') + ', '',''));' +
                      'OPEN del_role_cur;' +
                      'FETCH NEXT FROM del_role_cur INTO @role;' +
                      'WHILE @@FETCH_STATUS = 0' +
                      '  BEGIN' +
                      '    SET @sql = ''ALTER ROLE '' + QuoteName(@role) + '' DROP MEMBER ' + QuoteName(@username) + ''';' +
                      '    EXEC (@sql);' +
                      '    FETCH NEXT FROM del_role_cur INTO @role;' +
                      '  END;' +
                      'CLOSE del_role_cur;' +
                      'DEALLOCATE del_role_cur;' +
                      'OPEN add_role_cur;' +
                      'FETCH NEXT FROM add_role_cur INTO @role;' +
                      'WHILE @@FETCH_STATUS = 0' +
                      '  BEGIN' +
                      '    SET @sql = ''ALTER ROLE '' + QuoteName(@role) + '' ADD MEMBER ' + QuoteName(@username) + ''';' +
                      '    EXEC (@sql);' +
                      '    FETCH NEXT FROM add_role_cur INTO @role;' +
                      '  END;' +
                      'CLOSE add_role_cur;' +
                      'DEALLOCATE add_role_cur;'
          EXEC (@stmt)`
  return c.
    setDatabase(&database).
    ExecContext(ctx, cmd,
      sql.Named("database", database),
      sql.Named("username", user.Username),
      sql.Named("defaultSchema", user.DefaultSchema),
      sql.Named("defaultLanguage", user.DefaultLanguage),
      sql.Named("roles", strings.Join(user.Roles, ",")),
    )
}

func (c *Connector) DeleteUser(ctx context.Context, database, username string) error {
  cmd := `DECLARE @stmt nvarchar(max)
          SET @stmt = 'IF EXISTS (SELECT 1 FROM ' + QuoteName(@database) + '.[sys].[database_principals] WHERE [name] = ' + QuoteName(@username, '''') + ') ' +
                      'DROP USER ' + QuoteName(@username)
          EXEC (@stmt)`
  return c.
    setDatabase(&database).
    ExecContext(ctx, cmd, sql.Named("database", database), sql.Named("username", username))
}

func (c *Connector) setDatabase(database *string) *Connector {
  if *database == "" {
    *database = "master"
  }
  c.Database = *database
  return c
}
