package model

type DatabaseMasterkey struct {
  DatabaseName   string
  Password       string
  KeyName        string
  KeyGuid        string
  SymmetricKeyID int
  KeyLength      int
  KeyAlgorithm   string
  AlgorithmDesc  string
  PrincipalID    int
}
