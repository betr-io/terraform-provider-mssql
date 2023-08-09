package mssql

import "github.com/betr-io/terraform-provider-mssql/sql"

var (
	_ AadLoginConnector = &sql.Connector{}
)
