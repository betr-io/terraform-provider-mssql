terraform {
  required_version = "~> 1.3.6"
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 2.23.1"
    }
    mssql = {
      source  = "betr-io/mssql"
      version = "~> 0.2.6"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.4.3"
    }
    time = {
      source  = "hashicorp/time"
      version = "0.9.0"
    }
  }
}

provider "docker" {}

provider "mssql" {
  debug = true
}

provider "random" {}

#
# Creates a SQL Server running in a docker container on the local machine.
#
locals {
  local_username = "sa"
  local_password = "!!up3R!!3cR37"
}

resource "docker_image" "mssql" {
  name         = "mcr.microsoft.com/mssql/server"
  keep_locally = true
}

resource "docker_container" "mssql" {
  name  = "mssql"
  image = docker_image.mssql.latest
  env   = ["ACCEPT_EULA=Y", "SA_PASSWORD=${local.local_password}"]
  ports {
    internal = 1433
    external = 1433
  }
}

resource "time_sleep" "wait_5_seconds" {
  depends_on = [docker_container.mssql]

  create_duration = "5s"
}

#
# Creates a login and user in the SQL Server
#
resource "random_password" "example" {
  keepers = {
    login_name = "testlogin"
    username   = "testuser"
  }
  length  = 32
  special = true
}

resource "mssql_login" "example" {
  server {
    host = docker_container.mssql.ip_address
    login {
      username = local.local_username
      password = local.local_password
    }
  }
  login_name = random_password.example.keepers.login_name
  password   = random_password.example.result

  depends_on = [time_sleep.wait_5_seconds]
}

resource "mssql_user" "example" {
  server {
    host = docker_container.mssql.ip_address
    login {
      username = local.local_username
      password = local.local_password
    }
  }
  username   = random_password.example.keepers.username
  login_name = mssql_login.example.login_name
}

resource "mssql_role" "example" {
  server {
    host = docker_container.mssql.ip_address
    login {
      username = local.local_username
      password = local.local_password
    }
  }
  database = "master"
  role_name = "testrole"
}

output "login" {
  sensitive = true
  value = {
    login_name = mssql_login.example.login_name,
    password   = mssql_login.example.password
  }
}
