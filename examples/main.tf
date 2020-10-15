terraform {
  required_version = "~> 0.13"
  required_providers {
    mssql = {
      versions = ["0.0.1"]
      source   = "betr.io/betr/mssql"
    }
  }
}

provider "mssql" {}

data "mssql_database" "db" {
  server_name   = "betr-ci-sql-server"
  database_name = "surveyjs"
  azure_administrator {}
}

data "mssql_roles" "all" {
  database_id = data.mssql_database.db.id
}

output "roles" {
  value = data.mssql_roles.all
}
