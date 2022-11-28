package provider

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	providerLogFile = "terraform-provider-mssql.log"
)

func New(version, commit string) provider.ProviderWithMetadata {
	return &mssqlProvider{
		version: version,
		commit:  commit,
	}
}

type (
	mssqlProvider struct {
		version string
		commit  string
		logger  *zerolog.Logger
	}

	mssqlProviderModel struct {
		Debug types.Bool `tfsdk:"debug"`
	}
)

func (p *mssqlProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mssql"
	resp.Version = p.version
}

func (p *mssqlProvider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"debug": {
				Type:        types.BoolType,
				Optional:    true,
				Description: fmt.Sprintf("Enable provider debug logging (logs to file %s)", providerLogFile),
			},
		},
	}, nil
}

func (p *mssqlProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config mssqlProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Debug.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("debug"),
			"Unknown MSSQL Provider debug value",
			"The provider cannot create a debug logger as there is an unknown configuration value for debug attribute. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MSSQL_PROVIDER_DEBUG environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with terraform configuration value if set.

	var debug bool
	debugStr, ok := os.LookupEnv("MSSQL_PROVIDER_DEBUG")
	if ok {
		if val, err := strconv.ParseBool(debugStr); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("debug"),
				"Invalid provider debug flag",
				"Invalid provider debug flag",
			)
		} else {
			debug = val
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Debug.IsNull() {
		debug = config.Debug.ValueBool()
	}

	p.logger = newLogger(debug)

	p.logger.Info().Msgf("Created provider %s (%s)", p.version, p.commit)
}

func (p *mssqlProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
	// return []func() datasource.DataSource{
	// 	func() datasource.DataSource {
	// 		return dataSourceExample{},
	// 	},
	// }
}

func (p *mssqlProvider) Resources(ctx context.Context) []func() resource.Resource {
	return nil
	// return []func() resource.Resource{
	// 	func() resource.Resource {
	// 		return resourceExample{},
	// 	},
	// }
}

func newLogger(isDebug bool) *zerolog.Logger {
	var writer io.Writer = nil
	logLevel := zerolog.Disabled
	if isDebug {
		f, err := os.OpenFile(providerLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Err(err).Msg("error opening file")
		}
		writer = f
		logLevel = zerolog.DebugLevel
	}
	logger := zerolog.New(writer).Level(logLevel).With().Timestamp().Logger()
	return &logger
}
