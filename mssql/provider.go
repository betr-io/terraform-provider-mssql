package mssql

import (
  "context"
  "fmt"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/rs/zerolog"
  "github.com/rs/zerolog/log"
  "io"
  "os"
  "time"
)

type ConnectorFactory interface {
  GetConnector(prefix string, data *schema.ResourceData) (interface{}, error)
}

type Provider interface {
  GetConnector(prefix string, data *schema.ResourceData) (interface{}, error)
  ResourceLogger(resource, function string) zerolog.Logger
  DataSourceLogger(datasource, function string) zerolog.Logger
}

type mssqlProvider struct {
  factory ConnectorFactory
  logger  *zerolog.Logger
}

var defaultReadTimeout = schema.DefaultTimeout(30 * time.Second)

const providerLogFile = "terraform-provider-mssql.log"

// NewProvider -
func NewProvider(factory ConnectorFactory) *schema.Provider {
  return &schema.Provider{
    Schema: map[string]*schema.Schema{
      "debug": {
        Type:        schema.TypeBool,
        Description: fmt.Sprintf("Enable provider debug logging (logs to file %s)", providerLogFile),
        Optional:    true,
        Default:     false,
      },
    },
    ResourcesMap: map[string]*schema.Resource{
      "mssql_login":       resourceLogin(),
      "mssql_user":        resourceUser(),
      "mssql_user_login":  resourceUserLogin(),
      "mssql_az_sp_login": resourceAzSpLogin(),
    },
    DataSourcesMap: map[string]*schema.Resource{
      "mssql_server": dataSourceServer(),
      "mssql_roles":  dataSourceRoles(),
    },
    ConfigureContextFunc: func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
      return providerConfigure(ctx, data, factory)
    },
  }
}

func providerConfigure(ctx context.Context, data *schema.ResourceData, factory ConnectorFactory) (Provider, diag.Diagnostics) {
  isDebug := data.Get("debug").(bool)
  logger := newLogger(isDebug)

  logger.Info().Msg("Created provider")

  return mssqlProvider{factory: factory, logger: logger}, nil
}

func (p mssqlProvider) GetConnector(prefix string, data *schema.ResourceData) (interface{}, error) {
  return p.factory.GetConnector(prefix, data)
}

func (p mssqlProvider) ResourceLogger(resource, function string) zerolog.Logger {
  return p.logger.With().Str("resource", resource).Str("func", function).Logger()
}

func (p mssqlProvider) DataSourceLogger(datasource, function string) zerolog.Logger {
  return p.logger.With().Str("datasource", datasource).Str("func", function).Logger()
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
