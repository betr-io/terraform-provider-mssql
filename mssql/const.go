package mssql

const (
  serverProp               = "server"
  serverEncodedProp        = "server_encoded"
  databaseProp             = "database"
  principalIdProp          = "principal_id"
  usernameProp             = "username"
  passwordProp             = "password"
  clientIdProp             = "client_id"
  authenticationTypeProp   = "authentication_type"
  defaultSchemaProp        = "default_schema"
  defaultSchemaPropDefault = "dbo"
  rolesProp                = "roles"
)

var (
  rolesPropDefault = []string{"db_owner"}
)
