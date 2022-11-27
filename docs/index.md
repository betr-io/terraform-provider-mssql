# Microsoft SQL Server Provider

The SQL Server provider exposes resources used to manage the configuration of resources in a Microsoft SQL Server and an Azure SQL Database. It might also work for other Microsoft SQL Server products like Azure Managed SQL Server, but it has not been tested against these resources.

## Example Usage

```hcl
terraform {
  required_providers {
    mssql = {
      source = "betr-io/mssql"
      version = "0.3.0"
    }
  }
}

provider "mssql" {
  debug = "false"
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
