package main

import (
  "context"
  "database/sql"
  "database/sql/driver"
  "fmt"
  "github.com/Azure/go-autorest/autorest/adal"
  "github.com/Azure/go-autorest/autorest/azure"
  mssql "github.com/denisenkom/go-mssqldb"
  "net/url"
  "os"
)

func main() {
  args := os.Args[1:]

  loginLocal("localhost", "", os.Getenv("MSSQL_USERNAME"), os.Getenv("MSSQL_PASSWORD"))
  loginLocal("localhost", "master", os.Getenv("MSSQL_USERNAME"), os.Getenv("MSSQL_PASSWORD"))

  loginLocal(os.Getenv("TF_ACC_SQL_SERVER"), "", os.Getenv("TF_ACC_AZURE_MSSQL_USERNAME"), os.Getenv("TF_ACC_AZURE_MSSQL_PASSWORD"))
  loginLocal(os.Getenv("TF_ACC_SQL_SERVER"), "master", os.Getenv("TF_ACC_AZURE_MSSQL_USERNAME"), os.Getenv("TF_ACC_AZURE_MSSQL_PASSWORD"))
  loginLocal(os.Getenv("TF_ACC_SQL_SERVER"), "testdb", os.Getenv("TF_ACC_AZURE_MSSQL_USERNAME"), os.Getenv("TF_ACC_AZURE_MSSQL_PASSWORD"))

  for _, v := range args {
    loginLocal(os.Getenv("TF_ACC_SQL_SERVER"), "", v, os.Getenv("valueIsH8kd$¡"))
    loginLocal(os.Getenv("TF_ACC_SQL_SERVER"), "master", v, os.Getenv("valueIsH8kd$¡"))
    loginLocal(os.Getenv("TF_ACC_SQL_SERVER"), "testdb", v, os.Getenv("valueIsH8kd$¡"))
  }

  loginAzure(os.Getenv("TF_ACC_SQL_SERVER"), "", os.Getenv("MSSQL_TENANT_ID"), os.Getenv("MSSQL_CLIENT_ID"), os.Getenv("MSSQL_CLIENT_SECRET"))
  loginAzure(os.Getenv("TF_ACC_SQL_SERVER"), "master", os.Getenv("MSSQL_TENANT_ID"), os.Getenv("MSSQL_CLIENT_ID"), os.Getenv("MSSQL_CLIENT_SECRET"))
  loginAzure(os.Getenv("TF_ACC_SQL_SERVER"), "testdb", os.Getenv("MSSQL_TENANT_ID"), os.Getenv("MSSQL_CLIENT_ID"), os.Getenv("MSSQL_CLIENT_SECRET"))

  loginAzure(os.Getenv("TF_ACC_SQL_SERVER"), "", os.Getenv("MSSQL_TENANT_ID"), os.Getenv("TF_ACC_AZURE_USER_CLIENT_ID"), os.Getenv("TF_ACC_AZURE_USER_CLIENT_SECRET"))
  loginAzure(os.Getenv("TF_ACC_SQL_SERVER"), "master", os.Getenv("MSSQL_TENANT_ID"), os.Getenv("TF_ACC_AZURE_USER_CLIENT_ID"), os.Getenv("TF_ACC_AZURE_USER_CLIENT_SECRET"))
  loginAzure(os.Getenv("TF_ACC_SQL_SERVER"), "testdb", os.Getenv("MSSQL_TENANT_ID"), os.Getenv("TF_ACC_AZURE_USER_CLIENT_ID"), os.Getenv("TF_ACC_AZURE_USER_CLIENT_SECRET"))
}

func loginLocal(host, database, username, password string) {
  fmt.Printf("Logging in local at %s/%s (%s, ...) -> ", host, database, username)

  connector, err := connector(host, database, username, password)
  if err != nil {
    fmt.Printf("%s\n", err)
    return
  }

  login(connector)
}

func loginAzure(host, database, tenantId, clientId, clientSecret string) {
  fmt.Printf("Logging in azure at %s/%s (%s, ...) -> ", host, database, clientId)

  connector, err := connectorAzure(host, database, tenantId, clientId, clientSecret)
  if err != nil {
    fmt.Printf("%s\n", err)
    return
  }

  login(connector)
}

func login(c driver.Connector) {
  var res int
  db := sql.OpenDB(c)
  err := db.QueryRowContext(context.Background(), "SELECT 1").Scan(&res)
  if err != nil {
    fmt.Printf("%s\n", err)
    return
  }
  fmt.Printf("OK\n")
}

func connector(host, database, username, password string) (driver.Connector, error) {
  query := url.Values{}
  if database != "" {
    query.Set("database", database)
  }
  connectionString := (&url.URL{
    Scheme:   "sqlserver",
    User:     url.UserPassword(username, password),
    Host:     host,
    RawQuery: query.Encode(),
  }).String()
  return mssql.NewConnector(connectionString)
}

func connectorAzure(host, database, tenantId, clientId, clientSecret string) (driver.Connector, error) {
  query := url.Values{}
  if database != "" {
    query.Set("database", database)
  }
  connectionString := (&url.URL{
    Scheme:   "sqlserver",
    Host:     host,
    RawQuery: query.Encode(),
  }).String()
  return mssql.NewAccessTokenConnector(connectionString, func() (string, error) { return tokenProvider(tenantId, clientId, clientSecret) })
}

func tokenProvider(tenantId, clientId, clientSecret string) (string, error) {
  const resourceID = "https://database.windows.net/"

  oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, tenantId)
  if err != nil {
    return "", err
  }

  spt, err := adal.NewServicePrincipalToken(*oauthConfig, clientId, clientSecret, resourceID)
  if err != nil {
    return "", err
  }

  err = spt.EnsureFresh()
  if err != nil {
    return "", err
  }

  return spt.OAuthToken(), nil
}
