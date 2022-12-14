terraform {
  required_version = "~> 1.3.6"
  required_providers {
    azuread = {
      source  = "hashicorp/azuread"
      version = "~> 1.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 2.34.0"
    }
    mssql = {
      source  = "betr-io/mssql"
      version = "0.2.6"
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

provider "azuread" {}

provider "azurerm" {
  features {}
}

provider "mssql" {
  debug = "true"
}

provider "random" {}

variable "prefix" {
  description = "A prefix used when naming Azure resources"
  type        = string
}

variable "sql_servers_group" {
  description = "The name of an Azure AD group assigned the role 'Directory Reader'. The Azure SQL Server will be added to this group to enable external logins."
  type        = string
  default     = "SQL Servers"
}

variable "location" {
  description = "The location of the Azure resources."
  type        = string
  default     = "East US"
}

variable "tenant_id" {
  description = "The tenant id of the Azure AD tenant"
  type        = string
}

variable "local_ip_addresses" {
  description = "The external IP addresses of the machines running the acceptance tests. This is necessary to allow access to the Azure SQL Server resource."
  type        = list(string)
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
  count               = length(var.local_ip_addresses)
  name                = "AllowIP ${count.index}"
  resource_group_name = azurerm_mssql_server.sql_server.resource_group_name
  server_name         = azurerm_mssql_server.sql_server.name
  start_ip_address    = var.local_ip_addresses[count.index]
  end_ip_address      = var.local_ip_addresses[count.index]
}

# The Azure SQL Database used in tests
resource "azurerm_mssql_database" "db" {
  name      = "testdb"
  server_id = azurerm_mssql_server.sql_server.id
  sku_name  = "Basic"
}

resource "time_sleep" "wait_15_seconds" {
  depends_on = [azurerm_mssql_database.db]

  create_duration = "15s"
}


#
# Creates a login and user in the SQL Server
#
resource "random_password" "server" {
  keepers = {
    login_name = "testlogin"
    username   = "testuser"
  }
  length  = 32
  special = true
}

resource "mssql_login" "server" {
  server {
    host = azurerm_mssql_server.sql_server.fully_qualified_domain_name
    login {
      username = azurerm_mssql_server.sql_server.administrator_login
      password = azurerm_mssql_server.sql_server.administrator_login_password
    }
  }
  login_name = random_password.server.keepers.login_name
  password   = random_password.server.result

  depends_on = [time_sleep.wait_15_seconds]
}

resource "mssql_user" "server" {
  server {
    host = azurerm_mssql_server.sql_server.fully_qualified_domain_name
    login {
      username = azurerm_mssql_server.sql_server.administrator_login
      password = azurerm_mssql_server.sql_server.administrator_login_password
    }
  }
  database   = azurerm_mssql_database.db.name
  username   = random_password.server.keepers.username
  login_name = mssql_login.server.login_name
}

resource "mssql_role" "server" {
  server {
    host = azurerm_mssql_server.sql_server.fully_qualified_domain_name
    login {
      username = azurerm_mssql_server.sql_server.administrator_login
      password = azurerm_mssql_server.sql_server.administrator_login_password
    }
  }
  database = azurerm_mssql_database.db.name
  role_name = "testrole"
}

output "instance" {
  value = {
    login_name = mssql_login.server.login_name,
    password   = mssql_login.server.password
  }
}


#
# Creates a user with login in the SQL Server database
#

resource "random_password" "database" {
  keepers = {
    username = "testuser2"
  }
  length  = 32
  special = true
}

resource "mssql_user" "database" {
  server {
    host = azurerm_mssql_server.sql_server.fully_qualified_domain_name
    login {
      username = azurerm_mssql_server.sql_server.administrator_login
      password = azurerm_mssql_server.sql_server.administrator_login_password
    }
  }
  database = azurerm_mssql_database.db.name
  username = random_password.database.keepers.username
  password = random_password.database.result
}

output "database" {
  value = {
    username = mssql_user.database.username,
    password = mssql_user.database.password
  }
}


#
# Creates a login and user from Azure AD in the SQL Server
#

resource "mssql_user" "external" {
  server {
    host = azurerm_mssql_server.sql_server.fully_qualified_domain_name
    azure_login {
      tenant_id     = var.tenant_id
      client_id     = azuread_service_principal.sa.application_id
      client_secret = azuread_service_principal_password.sa.value
    }
  }
  database = azurerm_mssql_database.db.name
  username = azuread_service_principal.user.display_name
}

output "external" {
  value = {
    tenant_id     = var.tenant_id
    client_id     = azuread_service_principal.user.application_id
    client_secret = azuread_service_principal_password.user.value
  }
}
