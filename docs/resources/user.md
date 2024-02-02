# mssql_user

The `mssql_user` resource creates and manages a user on a SQL Server database.

## Example Usage

### Basic usage

```hcl
resource "mssql_user" "example" {
  server {
    host = "example-sql-server.database.windows.net"
    azure_login {
      tenant_id     = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      client_id     = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      client_secret = "terriblySecretSecret"
    }
  }
  username = "user@example.com"
  roles    = [ "db_owner" ]
}
```

### Using managed identities

```hcl
resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_user_assigned_identity" "example" {
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location

  name = "my-sql-identity"
}

resource "mssql_user" "example" {
  server {
    host = "example-sql-server.database.windows.net"
    azure_login {
    }
  }

  database  = "my-database"
  username  = azurerm_user_assigned_identity.example.name
  object_id = azurerm_user_assigned_identity.example.client_id

  roles     = ["db_datareader"]
}
```

> Note that in order to create an external user referencing an Azure AD entity (user, application), the Azure SQL Server needs to be a member of an Azure AD group assigned the Azure AD role `Directory Readers`. If it is not possible to give the Azure SQL Server this role (through the group), you can use the `object id` of the Azure AD entity instead.

## Argument Reference

The following arguments are supported:

* `server` - (Required) Server and login details for the SQL Server. The attributes supported in the `server` block is detailed below.
* `database` - (Optional) The user will be created in this database. Defaults to `master`. Changing this forces a new resource to be created.
* `username` - (Required) The name of the database user. Changing this forces a new resource to be created.
* `password` - (Optional) The password of the database user. Conflicts with the `login_name` argument. Changing this forces a new resource to be created.
* `login_name` - (Optional) The login name of the database user. This must refer to an existing SQL Server login name. Conflicts with the `password` argument. Changing this forces a new resource to be created.
* `default_schema` - (Optional) Specifies the first schema that will be searched by the server when it resolves the names of objects for this database user. Defaults to `dbo`.
* `default_language` - (Optional) Specifies the default language for the user. If no default language is specified, the default language for the user will bed the default language of the database. This argument does not apply to Azure SQL Database or if the user is not a contained database user.
* `user_type` - (Optional) For Azure SQL, only used in azure_login auth context, specify `E` for EXTERNAL_USER of `X` for EXTERNAL_GROUP
* `roles` - (Optional) List of database roles the user has. Defaults to none.

-> If only `username` is specified, an external user is created. The username must be in a format appropriate to the external user created, and will vary between SQL Server types. If `password` is specified, a user that authenticates at the database is created, and if `login_name` is specified, a user that authenticates at the server is created.

The `server` block supports the following arguments:

* `host` - (Required) The host of the SQL Server. Changing this forces a new resource to be created.
* `port` - (Optional) The port of the SQL Server. Defaults to `1433`. Changing this forces a new resource to be created.
* `login` - (Optional) SQL Server login for managing the database resources. The attributes supported in the `login` block is detailed below.
* `azure_login` - (Optional) Azure AD login for managing the database resources. The attributes supported in the `azure_login` block is detailed below.
* `azuread_default_chain_auth` - (Optional) Use a chain of strategies for authenticating when managing the database resources. This auth strategy is very similar to how the Azure CLI authenticates. For more information, see [DefaultAzureCredential](https://github.com/Azure/azure-sdk-for-go/wiki/Set-up-Your-Environment-for-Authentication#configure-defaultazurecredential). This block has no attributes.
* `azuread_managed_identity_auth` - (Optional) Use a managed identity for authenticating when managing the database resources. This is mainly useful for specifying a user-assigned managed identity. The attributes supported in the `azuread_managed_identity_auth` block is detailed below.

The `login` block supports the following arguments:

* `username` - (Required) The username of the SQL Server login. Can also be sourced from the `MSSQL_USERNAME` environment variable.
* `password` - (Required) The password of the SQL Server login. Can also be sourced from the `MSSQL_PASSWORD` environment variable.
* `object_id` - (Optional) The object id of the external username. Only used in azure_login auth context when AAD role delegation to sql server identity is not possible.

The `azure_login` block supports the following arguments:

* `tenant_id` - (Required) The tenant ID of the principal used to login to the SQL Server. Can also be sourced from the `MSSQL_TENANT_ID` environment variable.
* `client_id` - (Required) The client ID of the principal used to login to the SQL Server. Can also be sourced from the `MSSQL_CLIENT_ID` environment variable.
* `client_secret` - (Required) The client secret of the principal used to login to the SQL Server. Can also be sourced from the `MSSQL_CLIENT_SECRET` environment variable.

The `azuread_managed_identity_auth` block supports the following arguments:

* `user_id` - (Optional) Id of a user-assigned managed identity to assume. Omitting this property instructs the provider to assume a system-assigned managed identity.

-> Only one of `login`, `azure_login`, `azuread_default_chain_auth` and `azuread_managed_identity_auth` can be specified.

## Attribute Reference

The following attributes are exported:

* `principal_id` - The principal id of this database user.
* `sid` - The security identifier (SID) of this database user in String format.
* `authentication_type` - One of `DATABASE`, `INSTANCE`, or `EXTERNAL`.

## Import

Before importing `mssql_user`, you must to configure the authentication to your sql server:

1. Using Azure AD authentication, you must set the following environment variables: `MSSQL_TENANT_ID`, `MSSQL_CLIENT_ID` and `MSSQL_CLIENT_SECRET`.
2. Using SQL authentication, you must set the following environment variables: `MSSQL_USERNAME` and `MSSQL_PASSWORD`.

After that you can import the SQL Server database user using the server URL and `login name`, e.g.

```shell
terraform import mssql_user.example 'mssql://example-sql-server.database.windows.net/master/user@example.com'
```
