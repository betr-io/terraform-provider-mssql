terraform {
  required_version = "~> 0.13"
  required_providers {
    azuread = {
      source  = "hashicorp/azuread"
      version = "~> 1.6.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 2.73.0"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 2.15.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.1"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.1"
    }
  }
}
