package model

type User struct {
  PrincipalID     int64
  Username        string
  LoginName       string
  Password        string
  AuthType        string
  DefaultSchema   string
  DefaultLanguage string
  Roles           []string
}
