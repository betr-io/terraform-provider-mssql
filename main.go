package main

import (
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
  "terraform-provider-mssql/sql"

  "terraform-provider-mssql/mssql"
)

// These will be set by goreleaser to appropriate values for the compiled binary
var (
  version string = "dev"
  commit  string = "none"
)

func main() {
  plugin.Serve(&plugin.ServeOpts{
    ProviderFunc: func() *schema.Provider {
      return mssql.Provider(sql.GetFactory())
    },
  })
}
