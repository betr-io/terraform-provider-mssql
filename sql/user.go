package sql

import (
  "context"
  "database/sql"
  "terraform-provider-mssql/mssql/model"
)

func (c *Connector) GetUser(ctx context.Context, database, username string) (*model.User, error) {
  cmd := `DECLARE @stmt nvarchar(max)
          SET @stmt = 'SELECT principal_id, name, authentication_type_desc, COALESCE(default_schema_name, ''''), COALESCE(default_language_name, '''') ' +
                      'FROM ' + QuoteName(@database) + '.[sys].[database_principals] WHERE [name] = ' + QuoteName(@username, '''')
          EXEC (@stmt)`
  var user model.User
  err := c.
    setDatabase(&database).
    QueryRowContext(ctx, cmd,
      func(r *sql.Row) error {
        return r.Scan(&user.PrincipalID, &user.Username, &user.AuthType, &user.DefaultSchema, &user.DefaultLanguage)
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
          EXEC (@stmt)`
  return c.
    setDatabase(&database).
    ExecContext(ctx, cmd,
      sql.Named("username", user.Username),
      sql.Named("loginName", user.LoginName),
      sql.Named("password", user.Password),
      sql.Named("authType", user.AuthType),
      sql.Named("defaultSchema", user.DefaultSchema),
      sql.Named("defaultLanguage", user.DefaultLanguage))
}

func (c *Connector) UpdateUser(ctx context.Context, database string, user *model.User) error {
  cmd := `DECLARE @stmt nvarchar(max)
          SET @stmt = 'ALTER USER ' + QuoteName(@username) + ' '
          DECLARE @language nvarchar(max) = @defaultLanguage
          IF @language = '' SET @language = NULL
          SET @stmt = @stmt + 'WITH DEFAULT_SCHEMA = ' + QuoteName(@defaultSchema)
          IF NOT @@VERSION LIKE 'Microsoft SQL Azure%'
            BEGIN
              SET @stmt = @stmt + ', DEFAULT_LANGUAGE = ' + Coalesce(QuoteName(@language), 'NONE')
            END
          EXEC (@stmt)`
  return c.
    setDatabase(&database).
    ExecContext(ctx, cmd,
      sql.Named("username", user.Username),
      sql.Named("defaultSchema", user.DefaultSchema),
      sql.Named("defaultLanguage", user.DefaultLanguage))
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
