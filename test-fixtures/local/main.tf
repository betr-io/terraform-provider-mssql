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
  ports {
    internal = 1433
    external = 1433
  }
  env = ["ACCEPT_EULA=Y", "SA_PASSWORD=${local.local_password}"]
}


#
# Writes information necessary to log in to the SQL Server to file. This file is used by the Makefile when running acceptance tests.
#
resource "local_file" "local_env" {
  filename             = "${path.root}/../../.local.env"
  directory_permission = "0755"
  file_permission      = "0600"
  sensitive_content    = <<-EOT
                         export MSSQL_USERNAME='${local.local_username}'
                         export MSSQL_PASSWORD='${local.local_password}'
                         EOT
}
