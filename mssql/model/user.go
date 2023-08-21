package model

type User struct {
	PrincipalID     int64
	Username        string
	ObjectId        string
	LoginName       string
	Password        string
	SIDStr          string
	AuthType        string
	DefaultSchema   string
	DefaultLanguage string
	Roles           []string
}
