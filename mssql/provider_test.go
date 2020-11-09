package mssql

import (
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "os"
  "terraform-provider-mssql/sql"
  "testing"
)

var testAccProviders map[string]func() (*schema.Provider, error)
var testAccProvider *schema.Provider

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
