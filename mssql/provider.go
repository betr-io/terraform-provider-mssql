package mssql

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/betr-io/terraform-provider-mssql/mssql/model"
	"github.com/betr-io/terraform-provider-mssql/sql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type mssqlProvider struct {
	factory model.ConnectorFactory
	logger  *zerolog.Logger
}

const (
	providerLogFile = "terraform-provider-mssql.log"
)

var (
	defaultTimeout = schema.DefaultTimeout(30 * time.Second)
)

func New(version, commit string) func() *schema.Provider {
	return func() *schema.Provider {
		return Provider(sql.GetFactory())
	}
}

func Provider(factory model.ConnectorFactory) *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"debug": {
				Type:        schema.TypeBool,
				Description: fmt.Sprintf("Enable provider debug logging (logs to file %s)", providerLogFile),
				Optional:    true,
				Default:     false,
			},
			"max_open_connections": {
				Type:        schema.TypeInt,
				Description: "Maximum number of open connections to the database. Limits concurrent sessions to prevent exhausting server session limits. Set to 0 for unlimited.",
				Optional:    true,
				Default:     0,
			},
			"max_idle_connections": {
				Type:        schema.TypeInt,
				Description: "Maximum number of idle connections kept in the pool, ready for reuse. Avoids the latency of establishing new connections under load.",
				Optional:    true,
				Default:     2,
			},
			"connection_max_lifetime": {
				Type:        schema.TypeInt,
				Description: "Maximum lifetime of a connection in seconds. Expired connections are closed gracefully after use. Set to 0 for unlimited.",
				Optional:    true,
				Default:     0,
			},
			"connection_max_idle_time": {
				Type:        schema.TypeInt,
				Description: "Maximum time in seconds a connection may be idle before being closed. Helps free resources from unused connections. Set to 0 to disable.",
				Optional:    true,
				Default:     0,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"mssql_login": resourceLogin(),
			"mssql_user":  resourceUser(),
		},
		DataSourcesMap: map[string]*schema.Resource{},
		ConfigureContextFunc: func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
			return providerConfigure(ctx, data, factory)
		},
	}
}

func providerConfigure(ctx context.Context, data *schema.ResourceData, factory model.ConnectorFactory) (model.Provider, diag.Diagnostics) {
	isDebug := data.Get("debug").(bool)
	logger := newLogger(isDebug)

	// Read pool configuration from provider block
	maxOpenConns := data.Get("max_open_connections").(int)
	maxIdleConns := data.Get("max_idle_connections").(int)
	connMaxLifetime := data.Get("connection_max_lifetime").(int)
	connMaxIdleTime := data.Get("connection_max_idle_time").(int)

	poolConfig := sql.PoolConfig{
		MaxOpenConnections:    maxOpenConns,
		MaxIdleConnections:    maxIdleConns,
		ConnectionMaxLifetime: time.Duration(connMaxLifetime) * time.Second,
		ConnectionMaxIdleTime: time.Duration(connMaxIdleTime) * time.Second,
	}

	// Apply pool configuration to the existing factory
	sql.ConfigureFactory(factory, poolConfig)

	logger.Info().
		Int("max_open_connections", maxOpenConns).
		Int("max_idle_connections", maxIdleConns).
		Int("connection_max_lifetime_seconds", connMaxLifetime).
		Int("connection_max_idle_time_seconds", connMaxIdleTime).
		Msg("Created provider")

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
