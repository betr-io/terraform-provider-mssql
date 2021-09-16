# mssql_login

The `mssql_login` resource creates and manages a login on a SQL Server.

## Example Usage

```hcl
resource "mssql_login" "example" {
  server {
    host = "example-sql-server.database.windows.net"
    azure_login {}
  }
  login_name = "testlogin"
}
```

## Argument Reference

The following arguments are supported:

* `server` - (Required) Server and login details for the SQL Server. The attributes supported in the `server` block is detailed below.
* `login_name` - (Required) The name of the server login. Changing this forces a new resource to be created.
* `password` - (Required) The password of the server login.
* `default_database` - (Optional) The default database of this server login. Defaults to `master`. This argument does not apply to Azure SQL Database.
* `default_language` - (Optional) The default language of this server login. Defaults to `us_english`. This argument does not apply to Azure SQL Database.

The `server` block supports the following arguments:

* `host` - (Required) The host of the SQL Server. Changing this forces a new resource to be created.
* `port` - (Optional) The port of the SQL Server. Defaults to `1433`. Changing this forces a new resource to be created.
* `login` - (Optional) SQL Server login for managing the database resources. The attributes supported in the `login` block is detailed below.
* `azure_login` - (Optional) Azure AD login for managing the database resources. The attributes supported in the `azure_login` block is detailed below.

The `login` block supports the following arguments:

* `username` - (Required) The username of the SQL Server login. Can also be sourced from the `MSSQL_USERNAME` environment variable.
* `password` - (Required) The password of the SQL Server login. Can also be sourced from the `MSSQL_PASSWORD` environment variable.

The `azure_login` block supports the following arguments:

* `tenant_id` - (Required) The tanant ID of the principal used to login to the SQL Server. Can also be sourced from the `MSSQL_TENANT_ID` environment variable.
* `client_id` - (Required) The client ID of the principal used to login to the SQL Server. Can also be sourced from the `MSSQL_CLIENT_ID` environment variable.
* `client_secret` - (Required) The client secret of the principal used to login to the SQL Server. Can also be sourced from the `MSSQL_CLIENT_SECRET` environment variable.

-> Only one of `login` or `azure_login` can be specified. If neither is specified and both are sourced from environment variables, `azure_login` will be preferred.

## Attribute Reference

The following attributes are exported:

* `principal_id` - The principal id of this server login.

## Import

Before importing `mssql_login`, you must to configure the authentication to your sql server:

1. Using Azure AD authentication, you must set the following environment variables: `MSSQL_TENANT_ID`, `MSSQL_CLIENT_ID` and `MSSQL_CLIENT_SECRET`.
2. Using SQL authentication, you must set the following environment variables: `MSSQL_USERNAME` and `MSSQL_PASSWORD`.

After that you can import the SQL Server login using the server URL and `login name`, e.g.

```shell
terraform import mssql_login.example 'mssql://example-sql-server.database.windows.net/testlogin'
```
