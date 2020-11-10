resource "random_password" "sa" {
  keepers = {
    username = "sa"
  }
  length = 16
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
  env = ["ACCEPT_EULA=Y", "SA_PASSWORD=${random_password.sa.result}"]
}

resource "local_file" "local_env" {
  filename             = "${path.root}/../.local.env"
  directory_permission = "0755"
  file_permission      = "0600"
  sensitive_content    = <<-EOT
                         export MSSQL_USERNAME='${random_password.sa.keepers.username}'
                         export MSSQL_PASSWORD='${random_password.sa.result}'
                         EOT
}
