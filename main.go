package main

import (
	"context"

	"github.com/betr-io/terraform-provider-mssql/mssql"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
)

// These will be set by goreleaser to appropriate values for the compiled binary
var (
	version string = "dev"
	commit  string = "none"
)

func main() {
	upgradedSdkProvider, err := tf5to6server.UpgradeServer(
		context.Background(),
		mssql.New(version, commit)().GRPCProvider,
	)
	if err != nil {
		panic(err)
	}

	err = tf6server.Serve(
		"registry.terraform.io/betr-io/mssql",
		func() tfprotov6.ProviderServer { return upgradedSdkProvider },
	)
	if err != nil {
		panic(err)
	}
}
