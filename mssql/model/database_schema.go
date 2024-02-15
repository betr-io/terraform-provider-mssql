
package model

// Schema represents a SQL Server schema
type DatabaseSchema struct {
  SchemaID   int
  SchemaName string
  OwnerName  string
  OwnerId    int
}
