package model

type DatabasePermissions struct {
  DatabaseName string
	UserName     string
  PrincipalID  int
  Permissions  []string
}
