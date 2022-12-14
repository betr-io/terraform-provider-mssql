terraform {
  required_version = "~> 1.3.6"
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 2.23.1"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.8.0"
    }
    azuread = {
      source  = "hashicorp/azuread"
      version = "~> 2.22.0"
    }
    mssql = {
      source  = "betr-io/mssql"
      version = "~> 0.2.6"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.2"
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
