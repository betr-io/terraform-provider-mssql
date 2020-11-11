package mssql

import (
  "bytes"
  "context"
  sql2 "database/sql"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "os"
  "terraform-provider-mssql/mssql/model"
  "terraform-provider-mssql/sql"
  "testing"
  "text/template"
  "time"
)

var testAccProvider *schema.Provider
var testAccProviders map[string]func() (*schema.Provider, error)

func init() {
  testAccProvider = Provider(sql.GetFactory())
  testAccProviders = map[string]func() (*schema.Provider, error){
    "mssql": func() (*schema.Provider, error) {
      return testAccProvider, nil
    },
  }
}

func TestProvider(t *testing.T) {
  if err := testAccProvider.InternalValidate(); err != nil {
    t.Fatalf("err: %s", err)
  }
}

func testAccPreCheck(t *testing.T) {
  if v := os.Getenv("MSSQL_USERNAME"); v == "" {
    t.Fatal("MSSQL_USERNAME must be set for acceptance tests")
  }
  if v := os.Getenv("MSSQL_PASSWORD"); v == "" {
    t.Fatal("MSSQL_PASSWORD must be set for acceptance tests")
  }
}

type TestConnector interface {
  GetLogin(name string) (*model.Login, error)
  GetSystemUser() (string, error)
}

type testConnector struct {
  c interface{}
}

func getTestConnector(a map[string]string) (TestConnector, error) {
  prefix := serverProp + ".0."

  connector := &sql.Connector{
    Host:    a[prefix+"host"],
    Port:    a[prefix+"port"],
    Timeout: 60 * time.Second,
  }

  if username, ok := a[prefix+"login.0.username"]; ok {
    connector.Login = &sql.LoginUser{
      Username: username,
      Password: a[prefix+"login.0.password"],
    }
  }

  if tenantId, ok := a[prefix+"azure_login.0.tenant_id"]; ok {
    connector.AzureLogin = &sql.AzureLogin{
      TenantID:     tenantId,
      ClientID:     a[prefix+"azure_login.0.client_id"],
      ClientSecret: a[prefix+"azure_login.0.client_secret"],
    }
  }

  return testConnector{c: connector}, nil
}

func getTestLoginConnector(a map[string]string) (TestConnector, error) {
  prefix := serverProp + ".0."
  connector := &sql.Connector{
    Host:    a[prefix+"host"],
    Port:    a[prefix+"port"],
    Timeout: 60 * time.Second,
  }
  if password, ok := a[passwordProp]; ok {
    connector.Login = &sql.LoginUser{
      Username: a[loginNameProp],
      Password: password,
    }
  }

  return testConnector{c: connector}, nil
}

func (t testConnector) GetLogin(name string) (*model.Login, error) {
  return t.c.(LoginConnector).GetLogin(context.Background(), name)
}

func (t testConnector) GetSystemUser() (string, error) {
  var systemUser string
  err := t.c.(*sql.Connector).QueryRowContext(context.Background(), "SELECT SYSTEM_USER;", func(row *sql2.Row) error {
    return row.Scan(&systemUser)
  })
  return systemUser, err
}

func templateToString(name, text string, data interface{}) (string, error) {
  t, err := template.New(name).Parse(text)
  if err != nil {
    return "", err
  }
  var doc bytes.Buffer
  if err = t.Execute(&doc, data); err != nil {
    return "", err
  }
  return doc.String(), nil
}
