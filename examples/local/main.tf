terraform {
  required_version = "~> 0.13"
  required_providers {
    docker = {
      source  = "terraform-providers/docker"
      version = "~> 2.7.2"
    }
    mssql = {
      versions = ["0.0.1"]
      source   = "betr.io/betr/mssql"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 2.3"
    }
  }
}

provider "docker" {}

provider "mssql" {
  debug = true
}

provider "random" {}

resource "docker_image" "mssql" {
  name         = "mcr.microsoft.com/mssql/server"
  keep_locally = true
}

resource "docker_container" "mssql" {
  name  = "mssql"
  image = docker_image.mssql.latest
  ports {
    internal = 1433
    external = 1433
  }
  env = ["ACCEPT_EULA=Y", "SA_PASSWORD=$$up3R$$3cR37"]
}

resource "random_string" "username" {
  length  = 16
  special = false
}

resource "random_password" "password" {
  length  = 32
  special = true
}

data "mssql_server" "db" {
  name = "db"
  fqdn = "localhost"
  administrator_login {
    username = "sa"
    password = "$$up3R$$3cR37"
  }

  depends_on = [
    docker_container.mssql
  ]
}

resource "mssql_user_login" "new" {
  server_encoded = data.mssql_server.db.encoded
  username       = random_string.username.result
  password       = random_password.password.result

  depends_on = [
    docker_container.mssql
  ]
}

resource "mssql_user_login" "other" {
  server_encoded = data.mssql_server.db.encoded
  username       = "magne"
  password       = "1gaMnessumsaR"

  depends_on = [
    docker_container.mssql
  ]
}

data "mssql_roles" "all" {
  server_encoded = data.mssql_server.db.encoded
}

data "mssql_roles" "other" {
  server {
    name = "db"
    fqdn = "localhost"
    administrator_login {
      username = "sa"
      password = "$$up3R$$3cR37"
    }
  }

  depends_on = [
    docker_container.mssql
  ]
}

output "roles" {
  value = data.mssql_roles.all.roles
}
