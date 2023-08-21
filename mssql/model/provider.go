package model

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rs/zerolog"
)

type Provider interface {
	GetConnector(prefix string, data *schema.ResourceData) (interface{}, error)
	ResourceLogger(resource, function string) zerolog.Logger
	DataSourceLogger(datasource, function string) zerolog.Logger
}
