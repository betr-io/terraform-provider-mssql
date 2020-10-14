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
