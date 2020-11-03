package mssql

import (
  "fmt"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/rs/zerolog"
)

func getLoginID(data *schema.ResourceData) string {
  host := data.Get(serverProp + ".0.host").(string)
  port := data.Get(serverProp + ".0.port").(string)
  loginName := data.Get(loginNameProp).(string)
  return fmt.Sprintf("sqlserver://%s:%s/%s", host, port, loginName)
}

func loggerFromMeta(meta interface{}, resource, function string) zerolog.Logger {
  return meta.(Provider).logger.With().Str("resource", resource).Str("func", function).Logger()
}
