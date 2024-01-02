package sql

import (
  "context"
  "database/sql"
  "github.com/betr-io/terraform-provider-mssql/mssql/model"
)

func (c *Connector) GetLogin(ctx context.Context, name string) (*model.Login, error) {
  var login model.Login
  err := c.QueryRowContext(ctx,
    "SELECT principal_id, name, CONVERT(VARCHAR(1000), [sid], 1), default_database_name, default_language_name FROM [master].[sys].[sql_logins] WHERE [name] = @name",
    func(r *sql.Row) error {
      return r.Scan(&login.PrincipalID, &login.LoginName, &login.SIDStr, &login.DefaultDatabase, &login.DefaultLanguage)
    },
    sql.Named("name", name),
  )
  if err != nil {
    if err == sql.ErrNoRows {
      return nil, nil
    }
    return nil, err
  }
  return &login, nil
}

func (c *Connector) CreateLogin(ctx context.Context, name, password, sid, defaultDatabase, defaultLanguage string) error {
  cmd := `DECLARE @sql nvarchar(max)
          SET @sql = 'CREATE LOGIN ' + QuoteName(@name) + ' ' +
                     'WITH PASSWORD = ' + QuoteName(@password, '''')
          IF NOT @sid = ''
            BEGIN
              SET @sql = @sql + ', SID = ' + CONVERT(VARCHAR(1000), @sid, 1)
            END
          IF @@VERSION NOT LIKE 'Microsoft SQL Azure%'
            BEGIN
              IF @defaultDatabase = '' SET @defaultDatabase = 'master'
              IF NOT @defaultDatabase = 'master'
                BEGIN
                  SET @sql = @sql + ', DEFAULT_DATABASE = ' + QuoteName(@defaultDatabase)
                END
              DECLARE @serverLanguage nvarchar(max) = (SELECT lang.name FROM [sys].[configurations] c INNER JOIN [sys].[syslanguages] lang ON c.[value] = lang.langid WHERE c.name = 'default language')
              IF NOT @defaultLanguage IN ('', @serverLanguage)
                BEGIN
                  SET @sql = @sql + ', DEFAULT_LANGUAGE = ' + QuoteName(@defaultLanguage)
                END
            END
          EXEC (@sql)`
  database := "master"
  return c.
    setDatabase(&database).
    ExecContext(ctx, cmd,
    sql.Named("name", name),
    sql.Named("password", password),
    sql.Named("sid", sid),
    sql.Named("defaultDatabase", defaultDatabase),
    sql.Named("defaultLanguage", defaultLanguage))
}

func (c *Connector) UpdateLogin(ctx context.Context, name, password, defaultDatabase, defaultLanguage string) error {
  cmd := `DECLARE @sql nvarchar(max)
          SET @sql = 'ALTER LOGIN ' + QuoteName(@name) + ' ' +
                     'WITH PASSWORD = ' + QuoteName(@password, '''')
          IF @@VERSION NOT LIKE 'Microsoft SQL Azure%'
            BEGIN
              IF @defaultDatabase = '' SET @defaultDatabase = 'master'
              IF NOT @defaultDatabase IN (SELECT default_database_name FROM [master].[sys].[sql_logins] WHERE [name] = @name)
                BEGIN
                  SET @sql = @sql + ', DEFAULT_DATABASE = ' + QuoteName(@defaultDatabase)
                END
                DECLARE @language nvarchar(max) = @defaultLanguage
              IF @language = '' SET @language = (SELECT lang.name FROM [sys].[configurations] c INNER JOIN [sys].[syslanguages] lang ON c.[value] = lang.langid WHERE c.name = 'default language')
              IF @language != (SELECT default_language_name FROM [master].[sys].[sql_logins] WHERE [name] = @name)
                BEGIN
                  SET @sql = @sql + ', DEFAULT_LANGUAGE = ' + QuoteName(@language)
                END
              END
          EXEC (@sql)`
  return c.ExecContext(ctx, cmd,
    sql.Named("name", name),
    sql.Named("password", password),
    sql.Named("defaultDatabase", defaultDatabase),
    sql.Named("defaultLanguage", defaultLanguage))
}

func (c *Connector) DeleteLogin(ctx context.Context, name string) error {
  if err := c.killSessionsForLogin(ctx, name); err != nil {
    return err
  }
  cmd := `DECLARE @sql nvarchar(max)
          SET @sql = 'IF EXISTS (SELECT 1 FROM [master].[sys].[sql_logins] WHERE [name] = ' + QuoteName(@name, '''') + ') ' +
                     'DROP LOGIN ' + QuoteName(@name)
          EXEC (@sql)`
  return c.ExecContext(ctx, cmd, sql.Named("name", name))
}

func (c *Connector) killSessionsForLogin(ctx context.Context, name string) error {
  cmd := `-- adapted from https://stackoverflow.com/a/5178097/38055
          DECLARE sessionsToKill CURSOR FAST_FORWARD FOR
            SELECT session_id
            FROM sys.dm_exec_sessions
            WHERE login_name = @name
          OPEN sessionsToKill
          DECLARE @sessionId INT
          DECLARE @statement NVARCHAR(200)
          FETCH NEXT FROM sessionsToKill INTO @sessionId
          WHILE @@FETCH_STATUS = 0
          BEGIN
            PRINT 'Killing session ' + CAST(@sessionId AS NVARCHAR(20)) + ' for login ' + @name
            SET @statement = 'KILL ' + CAST(@sessionId AS NVARCHAR(20))
            EXEC sp_executesql @statement
            FETCH NEXT FROM sessionsToKill INTO @sessionId
          END
          CLOSE sessionsToKill
          DEALLOCATE sessionsToKill`
  return c.ExecContext(ctx, cmd, sql.Named("name", name))
}
