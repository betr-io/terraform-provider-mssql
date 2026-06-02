# Microsoft SQL Server Provider

The SQL Server provider exposes resources used to manage the configuration of resources in a Microsoft SQL Server and an Azure SQL Database. It might also work for other Microsoft SQL Server products like Azure Managed SQL Server, but it has not been tested against these resources.

## Example Usage

```hcl
terraform {
  required_providers {
    mssql = {
      source = "betr-io/mssql"
      version = "0.1.0"
    }
  }
}

provider "mssql" {
  debug                    = false
  max_open_connections     = 0
  max_idle_connections     = 2
  connection_max_lifetime  = 0
  connection_max_idle_time = 0
}

resource "mssql_login" "example" {
  server {
    host = "localhost"
    login {
      username = "sa"
      password = "MySuperSecr3t!"
    }
  }
  login_name = "testlogin"
  password   = "NotSoS3cret?"
}

resource "mssql_user" "example" {
  server {
    host = "localhost"
    login {
      username = "sa"
      password = "MySuperSecr3t!"
    }
  }
  username   = "testuser"
  login_name = mssql_login.example.login_name
}
```

## Argument Reference

The following arguments are supported:

* `debug` - (Optional) Either `false` or `true`. Defaults to `false`. If `true`, the provider will write a debug log to `terraform-provider-mssql.log`.
* `max_open_connections` - (Optional) Maximum number of open connections to the database. If set to 0, there is no limit on the number of open connections. Defaults to `0`.
* `max_idle_connections` - (Optional) Maximum number of idle connections in the pool. If set to 0, no idle connections are retained. Defaults to `2`.
* `connection_max_lifetime` - (Optional) Maximum lifetime of a connection in seconds. If set to 0, connections are not closed due to a connection's age. Defaults to `0`.
* `connection_max_idle_time` - (Optional) Maximum time in seconds a connection may be idle before being closed. If set to 0, connections are not closed due to a connection's idle time. Defaults to `0`.
