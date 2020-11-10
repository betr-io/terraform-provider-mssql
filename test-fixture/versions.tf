terraform {
  required_version = "~> 0.13"
  required_providers {
    azuread = {
      source  = "hashicorp/azuread"
      version = "~> 1.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 2.35.0"
    }
    docker = {
      source  = "terraform-providers/docker"
      version = "~> 2.7.2"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
}
