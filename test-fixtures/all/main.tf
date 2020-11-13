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

resource "random_password" "sa" {
  length  = 32
  special = true
  keepers = {
    name = "${local.prefix}-sa"
  }
}

resource "random_password" "user" {
  length  = 32
  special = true
  keepers = {
    name = "${local.prefix}-user"
  }
}

# An Azure AD group assigned the role 'Directory Readers'. The Azure SQL Server needs to be assigned to this group to enable external logins.
data "azuread_group" "sql_servers" {
  name = var.sql_servers_group
}

# An Azure AD service principal used as Azure Administrator for the Azure SQL Server resource
resource "azuread_application" "sa" {
  name     = random_password.sa.keepers.name
  homepage = "https://test.example.com"
}

resource "azuread_service_principal" "sa" {
  application_id = azuread_application.sa.application_id
}

resource "azuread_service_principal_password" "sa" {
  service_principal_id = azuread_service_principal.sa.id
  value                = random_password.sa.result
  end_date_relative    = "360h"
}

# An Azure AD service principal used to test creating an external login to the Azure SQL server resource
resource "azuread_application" "user" {
  name     = random_password.user.keepers.name
  homepage = "https://test.example.com"
}

resource "azuread_service_principal" "user" {
  application_id = azuread_application.user.application_id
}

resource "azuread_service_principal_password" "user" {
  service_principal_id = azuread_service_principal.user.id
  value                = random_password.user.result
  end_date_relative    = "360h"
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
  administrator_login_password = random_password.sa.result

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

resource "azurerm_sql_firewall_rule" "sql_server_fw_rule" {
  name                = "AllowIP"
  resource_group_name = azurerm_mssql_server.sql_server.resource_group_name
  server_name         = azurerm_mssql_server.sql_server.name
  start_ip_address    = var.local_ip_address
  end_ip_address      = var.local_ip_address
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
resource "local_file" "local_env" {
  filename             = "${path.root}/../../.local.env"
  directory_permission = "0755"
  file_permission      = "0600"
  sensitive_content    = <<-EOT
                         export MSSQL_USERNAME='${local.local_username}'
                         export MSSQL_PASSWORD='${local.local_password}'
                         export MSSQL_TENANT_ID='${var.tenant_id}'
                         export MSSQL_CLIENT_ID='${azuread_service_principal.sa.application_id}'
                         export MSSQL_CLIENT_SECRET='${azuread_service_principal_password.sa.value}'
                         export TF_ACC_SQL_SERVER='${azurerm_mssql_server.sql_server.fully_qualified_domain_name}'
                         export TF_ACC_AZURE_MSSQL_USERNAME='${azurerm_mssql_server.sql_server.administrator_login}'
                         export TF_ACC_AZURE_MSSQL_PASSWORD='${azurerm_mssql_server.sql_server.administrator_login_password}'
                         EOT
}
