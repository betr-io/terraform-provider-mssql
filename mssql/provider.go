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

type Provider struct {
  logger *zerolog.Logger
}

var defaultReadTimeout = schema.DefaultTimeout(30 * time.Second)

const providerLogFile = "terraform-provider-mssql.log"

// NewProvider -
func NewProvider() *schema.Provider {
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
      "mssql_login":      resourceLogin(),
      "mssql_user_login": resourceUserLogin(),
      "mssql_az_sp_login": resourceAzSpLogin(),
    },
    DataSourcesMap: map[string]*schema.Resource{
      "mssql_server": dataSourceServer(),
      "mssql_roles":  dataSourceRoles(),
    },
    ConfigureContextFunc: providerConfigure,
  }
}

func providerConfigure(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
  isDebug := data.Get("debug").(bool)
  logger := newLogger(isDebug)

  logger.Info().Msg("Created provider")

  return Provider{logger: logger}, nil
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
