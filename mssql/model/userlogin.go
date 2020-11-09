package model

import "github.com/google/uuid"

type UserLogin struct {
  PrincipalID int64
  Type        string
  Username    string
  SID         uuid.UUID
  Schema      string
  Roles       []string
}
