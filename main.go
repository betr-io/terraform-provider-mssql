package main

import (
	"context"
	"log"

	"github.com/betr-io/terraform-provider-mssql/mssql"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
)

// These will be set by goreleaser to appropriate values for the compiled binary
var (
	version string = "dev"
	commit  string = "none"
)

func main() {
	ctx := context.Background()
	upgradedSdkProvider, err := tf5to6server.UpgradeServer(
		ctx,
		mssql.New(version, commit)().GRPCProvider,
	)
	if err != nil {
		log.Fatal(err)
	}

	providers := []func() tfprotov6.ProviderServer{
		func() tfprotov6.ProviderServer { return upgradedSdkProvider },
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		log.Fatal(err)
	}

	err = tf6server.Serve(
		"registry.terraform.io/betr-io/mssql",
		muxServer.ProviderServer,
	)
	if err != nil {
		log.Fatal(err)
	}
}
