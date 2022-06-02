terraform {
  required_version = "~> 0.13"
  required_providers {
    azuread = {
      source  = "hashicorp/azuread"
      version = "~> 2.22.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.8.0"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 2.16.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.2"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.2"
    }
  }
}
