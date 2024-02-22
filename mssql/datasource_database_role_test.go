package mssql

import (
  "fmt"
  "os"
  "testing"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataDatabaseRole_Local_Basic(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckDataRoleDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckDataRole(t, "data_local_test", "login", map[string]interface{}{"role_name": "data_test_role"}),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("data.mssql_database_role.data_local_test", "id", "sqlserver://localhost:1433/master/data_test_role"),
          resource.TestCheckResourceAttr("data.mssql_database_role.data_local_test", "database", "master"),
          resource.TestCheckResourceAttr("data.mssql_database_role.data_local_test", "role_name", "data_test_role"),
          resource.TestCheckResourceAttr("data.mssql_database_role.data_local_test", "server.#", "1"),
          resource.TestCheckResourceAttr("data.mssql_database_role.data_local_test", "server.0.host", "localhost"),
          resource.TestCheckResourceAttr("data.mssql_database_role.data_local_test", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("data.mssql_database_role.data_local_test", "server.0.login.#", "1"),
          resource.TestCheckResourceAttr("data.mssql_database_role.data_local_test", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("data.mssql_database_role.data_local_test", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttr("data.mssql_database_role.data_local_test", "server.0.azure_login.#", "0"),
          resource.TestCheckResourceAttrSet("data.mssql_database_role.data_local_test", "principal_id"),
        ),
      },
    },
  })
}

func TestAccDataDatabaseRole_Azure_Basic(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckDataRoleDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckDataRole(t, "data_azure_test", "azure", map[string]interface{}{"database": "testdb", "role_name": "data_test_role"}),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("data.mssql_database_role.data_azure_test", "id", "sqlserver://"+os.Getenv("TF_ACC_SQL_SERVER")+":1433/testdb/data_test_role"),
          resource.TestCheckResourceAttr("data.mssql_database_role.data_azure_test", "database", "testdb"),
          resource.TestCheckResourceAttr("data.mssql_database_role.data_azure_test", "role_name", "data_test_role"),
          resource.TestCheckResourceAttr("data.mssql_database_role.data_azure_test", "server.#", "1"),
          resource.TestCheckResourceAttr("data.mssql_database_role.data_azure_test", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
          resource.TestCheckResourceAttr("data.mssql_database_role.data_azure_test", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("data.mssql_database_role.data_azure_test", "server.0.azure_login.#", "1"),
          resource.TestCheckResourceAttr("data.mssql_database_role.data_azure_test", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
          resource.TestCheckResourceAttr("data.mssql_database_role.data_azure_test", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
          resource.TestCheckResourceAttr("data.mssql_database_role.data_azure_test", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
          resource.TestCheckResourceAttr("data.mssql_database_role.data_azure_test", "server.0.login.#", "0"),
          resource.TestCheckResourceAttrSet("data.mssql_database_role.data_azure_test", "principal_id"),
        ),
      },
    },
  })
}

func testAccCheckDataRole(t *testing.T, name string, login string, data map[string]interface{}) string {
  text := `resource "mssql_database_role" "{{ .name }}" {
             server {
               host = "{{ .host }}"
               {{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
             }
             {{ with .database }}database = "{{ . }}"{{ end }}
             role_name = "{{ .role_name }}"
           }
           data "mssql_database_role" "{{ .name }}" {
             server {
               host = "{{ .host }}"
               {{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
             }
             {{ with .database }}database = "{{ . }}"{{ end }}
             role_name = "{{ .role_name }}"
             depends_on = [mssql_database_role.{{ .name }}]
           }`

  data["name"] = name
  data["login"] = login
  if login == "fedauth" || login == "msi" || login == "azure" {
    data["host"] = os.Getenv("TF_ACC_SQL_SERVER")
  } else if login == "login" {
    data["host"] = "localhost"
  } else {
    t.Fatalf("login expected to be one of 'login', 'azure', 'msi', 'fedauth', got %s", login)
  }
  res, err := templateToString(name, text, data)
  if err != nil {
    t.Fatalf("%s", err)
  }
  return res
}

func testAccCheckDataRoleDestroy(state *terraform.State) error {
  for _, rs := range state.RootModule().Resources {
    if rs.Type != "mssql_database_role" {
      continue
    }

    connector, err := getTestConnector(rs.Primary.Attributes)
    if err != nil {
      return err
    }

    database := rs.Primary.Attributes["database"]
    roleName := rs.Primary.Attributes["role_name"]
    role, err := connector.GetDatabaseRole(database, roleName)
    if role != nil {
      return fmt.Errorf("role still exists")
    }
    if err != nil {
      return fmt.Errorf("expected no error, got %s", err)
    }
  }
  return nil
}