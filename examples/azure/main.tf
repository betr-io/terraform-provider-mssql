terraform {
  required_version = "~> 0.13"
  required_providers {
    mssql = {
      versions = ["0.0.1"]
      source   = "betr.io/betr/mssql"
    }
  }
}

provider "mssql" {
  debug = "true"
}

data "mssql_server" "db" {
  name = "betr-ci-sql-server"
  azure_administrator {}
}

data "mssql_roles" "all" {
  server_encoded = data.mssql_server.db.encoded
  database       = "surveyjs"
}

output "roles" {
  value = data.mssql_roles.all.roles
}

resource "mssql_az_sp_login" "reportingAPI" {
  server {
    name = "betr-ci-sql-server"
    azure_administrator {}
  }
  #  client_id = "02a8acd6-b32b-4141-89df-58813ae8e2fd"
  #  client_secret = "wD5puzkg@yEx0PrK3uQ6f0IqDp1RU3J0rdem"
  database  = "surveyjs"
  username  = "reportingAPI"
  client_id = "02a8acd6-b32b-4141-89df-58813ae8e2fd"
  schema    = "dbo"
  roles      = ["db_owner"]
}

# data "mssql_roles" "other" {
#   server {
#     name = "betr-ci-sql-server"
#     azure_administrator {
#       client_id     = "02a8acd6-b32b-4141-89df-58813ae8e2fd"
#       client_secret = "wD5puzkg@yEx0PrK3uQ6f0IqDp1RU3J0rdem"
#       tenant_id     = "30ddf688-b59c-470c-b8e1-143ccbb6ce33"
#     }
#   }
#   database = "surveyjs"
# }

# output "other" {
#   value = data.mssql_roles.all.roles
# }
