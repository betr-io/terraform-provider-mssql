package mssql

import (
  "context"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
  return &schema.Provider{
    ResourcesMap: map[string]*schema.Resource{},
    DataSourcesMap: map[string]*schema.Resource{
      "mssql_database": dataSourceDatabase(),
      "mssql_roles":    dataSourceRoles(),
    },
    ConfigureContextFunc: providerConfigure,
  }
}

func providerConfigure(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
  // Warnings or errors can be collected in a slice type
  var diags diag.Diagnostics

  return make(map[string]*Connector), diags
}
