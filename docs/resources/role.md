# mssql_role

The `mssql_role` resource creates and manages a role on a SQL Server database.

## Example Usage

### Basic usage

```hcl
resource "mssql_role" "example" {
  server {
    host = "example-sql-server.database.windows.net"
    azure_login {
      tenant_id     = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      client_id     = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      client_secret = "terriblySecretSecret"
    }
  }
  database      = "master"
  role_name     = "example-role-name"
}
```

## Argument Reference

The following arguments are supported:

* `server` - (Required) Server and login details for the SQL Server. The attributes supported in the `server` block is detailed below.
* `role_name` - (Required) The name of the role. Changing this resource property modifies the existing resource.
* `database` - (Optional) The user will be created in this database. Defaults to `master`. Changing this forces a new resource to be created.

The `server` block supports the following arguments:

* `host` - (Required) The host of the SQL Server. Changing this forces a new resource to be created.
* `port` - (Optional) The port of the SQL Server. Defaults to `1433`. Changing this forces a new resource to be created.
* `login` - (Optional) SQL Server login for managing the database resources. The attributes supported in the `login` block is detailed below.

The `login` block supports the following arguments:

* `username` - (Required) The username of the SQL Server login. Can also be sourced from the `MSSQL_USERNAME` environment variable.
* `password` - (Required) The password of the SQL Server login. Can also be sourced from the `MSSQL_PASSWORD` environment variable.
* `object_id` - (Optional) The object id of the external username. Only used in azure_login auth context when AAD role delegation to sql server identity is not possible.

## Import

Before importing `mssql_role`, you must to configure the authentication to your sql server:

1. Using Azure AD authentication, you must set the following environment variables: `MSSQL_TENANT_ID`, `MSSQL_CLIENT_ID` and `MSSQL_CLIENT_SECRET`.
2. Using SQL authentication, you must set the following environment variables: `MSSQL_USERNAME` and `MSSQL_PASSWORD`.

After that you can import the SQL Server database role using the server URL and `role name`, e.g.

```shell
terraform import mssql_role.example 'mssql://example-sql-server.database.windows.net/master/testrole'
```
