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
# Creates an Azure SQL Database running in a temporary resource group on Azure.
#

# Random names and secrets
resource "random_string" "random" {
  length  = 16
  upper   = false
  special = false
}

locals {
  prefix = "${var.prefix}-${substr(random_string.random.result, 0, 4)}"
}

# An Azure AD group assigned the role 'Directory Readers'. The Azure SQL Server needs to be assigned to this group to enable external logins.
data "azuread_group" "sql_servers" {
  display_name = var.sql_servers_group
}

# An Azure AD service principal used as Azure Administrator for the Azure SQL Server resource
resource "azuread_application" "sa" {
  display_name = "${local.prefix}-sa"
  web {
    homepage_url = "https://test.example.com"
  }
}

resource "azuread_service_principal" "sa" {
  application_id = azuread_application.sa.application_id
}

resource "azuread_service_principal_password" "sa" {
  service_principal_id = azuread_service_principal.sa.id
}

# An Azure AD service principal used to test creating an external login to the Azure SQL server resource
resource "azuread_application" "user" {
  display_name = "${local.prefix}-user"
  web {
    homepage_url = "https://test.example.com"
  }
}

resource "azuread_service_principal" "user" {
  application_id = azuread_application.user.application_id
}

resource "azuread_service_principal_password" "user" {
  service_principal_id = azuread_service_principal.user.id
}

# Temporary resource group
resource "azurerm_resource_group" "rg" {
  name     = "${lower(var.prefix)}-${random_string.random.result}"
  location = var.location
}

# An Azure SQL Server
resource "azurerm_mssql_server" "sql_server" {
  name                = "${lower(local.prefix)}-sql-server"
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location

  version                      = "12.0"
  administrator_login          = "SuperAdministrator"
  administrator_login_password = azuread_service_principal_password.sa.value

  azuread_administrator {
    tenant_id      = var.tenant_id
    object_id      = azuread_service_principal.sa.application_id
    login_username = azuread_service_principal.sa.display_name
  }

  identity {
    type = "SystemAssigned"
  }
}

resource "azuread_group_member" "sql" {
  group_object_id  = data.azuread_group.sql_servers.id
  member_object_id = azurerm_mssql_server.sql_server.identity[0].principal_id
}

resource "azurerm_mssql_firewall_rule" "sql_server_fw_rule" {
  count            = length(var.local_ip_addresses)
  name             = "AllowIP ${count.index}"
  server_id        = azurerm_mssql_server.sql_server.id
  start_ip_address = var.local_ip_addresses[count.index]
  end_ip_address   = var.local_ip_addresses[count.index]
}

# The Azure SQL Database used in tests
resource "azurerm_mssql_database" "db" {
  name      = "testdb"
  server_id = azurerm_mssql_server.sql_server.id
  sku_name  = "Basic"
}


#
# Writes information necessary to log in to the SQL Server to file. This file is used by the Makefile when running acceptance tests.
#
resource "local_sensitive_file" "local_env" {
  filename             = "${path.root}/../../.local.env"
  directory_permission = "0755"
  file_permission      = "0600"
  content              = <<-EOT
                         export TF_ACC=1
                         export MSSQL_USERNAME='${local.local_username}'
                         export MSSQL_PASSWORD='${local.local_password}'
                         export MSSQL_TENANT_ID='${var.tenant_id}'
                         export MSSQL_CLIENT_ID='${azuread_service_principal.sa.application_id}'
                         export MSSQL_CLIENT_SECRET='${azuread_service_principal_password.sa.value}'
                         export TF_ACC_SQL_SERVER='${azurerm_mssql_server.sql_server.fully_qualified_domain_name}'
                         export TF_ACC_AZURE_MSSQL_USERNAME='${azurerm_mssql_server.sql_server.administrator_login}'
                         export TF_ACC_AZURE_MSSQL_PASSWORD='${azurerm_mssql_server.sql_server.administrator_login_password}'
                         export TF_ACC_AZURE_USER_CLIENT_ID='${azuread_service_principal.user.application_id}'
                         export TF_ACC_AZURE_USER_CLIENT_USER='${azuread_service_principal.user.display_name}'
                         export TF_ACC_AZURE_USER_CLIENT_SECRET='${azuread_service_principal_password.user.value}'
                         # Configuration for fedauth which uses env vars via DefaultAzureCredential
                         export AZURE_TENANT_ID='${var.tenant_id}'
                         export AZURE_CLIENT_ID='${azuread_service_principal.sa.application_id}'
                         export AZURE_CLIENT_SECRET='${azuread_service_principal_password.sa.value}'
                         EOT
}
