
package model

// Role represents a SQL Server role
type DatabaseRole struct {
  RoleID    int
  RoleName  string
	OwnerName string
	OwnerId   int
}
