terraform {
  required_version = "~> 1.1.5"
  required_providers {
    azuread = {
      source  = "hashicorp/azuread"
      version = "~> 1.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 2.99.0"
    }
    mssql = {
      source  = "betr.io/betr/mssql"
      version = "0.3.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0.0"
    }
    time = {
      source  = "hashicorp/time"
      version = "0.6.0"
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

data "azuread_client_config" "current" {}

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
  version             = "12.0"

  azuread_administrator {
    tenant_id                   = data.azuread_client_config.current.tenant_id
    object_id                   = data.azuread_client_config.current.client_id
    login_username              = "superuser"
    azuread_authentication_only = true
  }

  identity {
    type = "SystemAssigned"
  }
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
# Creates a login and user from Azure AD in the SQL Server
#

resource "mssql_user" "external" {
  server {
    host = azurerm_mssql_server.sql_server.fully_qualified_domain_name
    azuread_default_chain_auth {}
  }
  database = azurerm_mssql_database.db.name
  username = "someone@foobar.onmicrosoft.com"
}
