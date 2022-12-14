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

// Fixes the error:
// Error: Value for undeclared variable
// A variable named "operating_system" was assigned on the command line, but the root module does not declare a variable of that name. To use this value, add a "variable" block to the configuration. 
variable "operating_system" {
  description = "dummy variable"
  type        = string
  default     = "dummyvariable"
}
