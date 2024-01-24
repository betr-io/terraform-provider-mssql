terraform {
  required_version = "~> 1.5"
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 3.0"
    }
    mssql = {
      source  = "betr-io/mssql"
      version = "~> 0.2"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.10"
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
  image = docker_image.mssql.image_id
  env   = ["ACCEPT_EULA=Y", "SA_PASSWORD=${local.local_password}"]
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
    host = docker_container.mssql.network_data[0].ip_address
    login {
      username = local.local_username
      password = local.local_password
    }
  }
  login_name = random_password.example.keepers.login_name
  password   = random_password.example.result

  depends_on = [time_sleep.wait_5_seconds]
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
  sid        = "0xB7BDEF7990D03541BAA2AD73E4FF18E8"

  depends_on = [time_sleep.wait_5_seconds]
}

resource "mssql_user" "example" {
  server {
    host = docker_container.mssql.network_data[0].ip_address
    login {
      username = local.local_username
      password = local.local_password
    }
  }
  username   = random_password.example.keepers.username
  login_name = mssql_login.example.login_name
}

output "login" {
  value = {
    login_name = mssql_login.example.login_name,
    password   = mssql_login.example.password
  }
  sensitive = true
}

data "mssql_login" "example" {
  server {
    host = docker_container.mssql.ip_address
    login {
      username = local.local_username
      password = local.local_password
    }
  }
  login_name = mssql_login.example.login_name

  depends_on = [mssql_login.example]
}

output "datalogin" {
  value = {
    principal_id = data.mssql_login.example.principal_id
    sid          = data.mssql_login.example.sid
  }
}