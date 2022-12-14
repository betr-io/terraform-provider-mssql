package main

import (
  "flag"

  "github.com/betr-io/terraform-provider-mssql/mssql"
  "github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

// These will be set by goreleaser to appropriate values for the compiled binary
var (
  version string = "dev"
  commit  string = "none"
)

func main() {
  var debug bool

  flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
  flag.Parse()

  opts := &plugin.ServeOpts{
    Debug:        debug,
    ProviderAddr: "registry.terraform.io/betr-io/mssql",
    ProviderFunc: mssql.New(version, commit),
  }

  plugin.Serve(opts)
}
