package model

type DatabasePermissions struct {
	DatabaseName string
	PrincipalID  int
	Permissions  []string
}
