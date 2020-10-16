package mssql

const (
  serverProp        = "server"
  databaseProp      = "database"
  principalIdProp   = "principal_id"
  usernameProp      = "username"
  passwordProp      = "password"
  clientIdProp      = "client_id"
  schemaProp        = "schema"
  schemaPropDefault = "dbo"
  rolesProp         = "roles"
)

var (
  rolesPropDefault = []string{"db_owner"}
)
