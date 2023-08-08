package sql

import (
	"context"
	"database/sql"
	"github.com/betr-io/terraform-provider-mssql/mssql/model"
)

func (c *Connector) GetAadLogin(ctx context.Context, name string) (*model.AadLogin, error) {
	var login model.AadLogin
	err := c.QueryRowContext(ctx,
		"SELECT name, default_database_name, default_language_name FROM [master].[sys].[server_principals] WHERE [type] NOT IN ('G', 'R') and [name] = @name",
		func(r *sql.Row) error {
			return r.Scan(&login.LoginName, &login.DefaultDatabase, &login.DefaultLanguage)
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

func (c *Connector) CreateAadLogin(ctx context.Context, name, defaultDatabase, defaultLanguage string) error {
	cmd := `DECLARE @sql nvarchar(max)
          SET @sql = 'CREATE LOGIN ' + QuoteName(@name) + ' FROM EXTERNAL PROVIDER '
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
			sql.Named("defaultDatabase", defaultDatabase),
			sql.Named("defaultLanguage", defaultLanguage))
}

func (c *Connector) DeleteAadLogin(ctx context.Context, name string) error {
	if err := c.killSessionsForLogin(ctx, name); err != nil {
		return err
	}
	cmd := `DECLARE @sql nvarchar(max)
          SET @sql = 'IF EXISTS (SELECT 1 FROM [master].[sys].[server_principals] WHERE [name] = ' + QuoteName(@name, '''') + ') ' +
                     'DROP LOGIN ' + QuoteName(@name)
          EXEC (@sql)`
	return c.ExecContext(ctx, cmd, sql.Named("name", name))
}
