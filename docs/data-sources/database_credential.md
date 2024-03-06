# mssql_database_credential (Data Source)

The `mssql_database_credential` obtains information about user permissions on a SQL Server.

## Example Usage

```hcl
data "mssql_database_credential" "example" {
  server {
    host = "example-sql-server.database.windows.net"
    azure_login {
      tenant_id     = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      client_id     = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      client_secret = "xxxxxxxxxxxxxxxxxxxxxx"
    }
  }
  database  = "example"
  credential_name = "example-credential-name"
}
```

## Argument Reference

The following arguments are supported:

* `server` - (Required) Server and login details for the SQL Server. The attributes supported in the `server` block is detailed below.
* `database` - (Required) The database.
* `credential_name` - (Required) The database scoped credential name.

The `server` block supports the following arguments:

* `host` - (Required) The host of the SQL Server. Changing this forces a new resource to be created.
* `port` - (Optional) The port of the SQL Server. Defaults to `1433`. Changing this forces a new resource to be created.
* `login` - (Optional) SQL Server login for managing the database resources. The attributes supported in the `login` block is detailed below.

The `login` block supports the following arguments:

* `username` - (Required) The username of the SQL Server login. Can also be sourced from the `MSSQL_USERNAME` environment variable.
* `password` - (Required) The password of the SQL Server login. Can also be sourced from the `MSSQL_PASSWORD` environment variable.
* `object_id` - (Optional) The object id of the external username. Only used in azure_login auth context when AAD role delegation to sql server identity is not possible.

## Attribute Reference

The following attributes are exported:

* `principal_id` - The principal id of this database scoped credential.
* `credential_id` - The id of this database scoped credential.
* `credential_name` - The name of the database scoped credential.
* `identity_name` - The name of the account.
