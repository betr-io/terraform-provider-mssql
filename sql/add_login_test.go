package sql

import "github.com/betr-io/terraform-provider-mssql/mssql"

var (
	_ mssql.AadLoginConnector = &Connector{}
)
