package model

type Login struct {
	PrincipalID     int64
	LoginName       string
	SIDStr          string
	DefaultDatabase string
	DefaultLanguage string
}
