package mssql

import (
  "fmt"
  "os"
  "testing"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataDatabasePermissions_Local_Basic(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckDataDatabasePermissionsDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckDataDataBasepermissions(t, "database", "login", map[string]interface{}{"database": "master", "username": "db_user_perm", "permissions": "[\"REFERENCES\", \"UPDATE\"]", "login_name": "db_login_perm", "login_password": "valueIsH8kd$¡", "roles": "[\"db_owner\"]"}),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("data.mssql_database_permissions.database", "id", "sqlserver://localhost:1433/master/db_user_perm/permissions"), //guess user principal-ID = 7
          resource.TestCheckResourceAttr("data.mssql_database_permissions.database", "database", "master"),
          resource.TestCheckResourceAttr("data.mssql_database_permissions.database", "permissions.#", "2"),
          resource.TestCheckResourceAttr("data.mssql_database_permissions.database", "permissions.0", "REFERENCES"),
          resource.TestCheckResourceAttr("data.mssql_database_permissions.database", "permissions.1", "UPDATE"),
          resource.TestCheckResourceAttr("data.mssql_database_permissions.database", "server.#", "1"),
          resource.TestCheckResourceAttr("data.mssql_database_permissions.database", "server.0.host", "localhost"),
          resource.TestCheckResourceAttr("data.mssql_database_permissions.database", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("data.mssql_database_permissions.database", "server.0.login.#", "1"),
          resource.TestCheckResourceAttr("data.mssql_database_permissions.database", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("data.mssql_database_permissions.database", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttr("data.mssql_database_permissions.database", "server.0.azure_login.#", "0"),
          resource.TestCheckResourceAttrSet("data.mssql_database_permissions.database", "principal_id"),
        ),
      },
    },
  })
}

func testAccCheckDataDataBasepermissions(t *testing.T, name string, login string, data map[string]interface{}) string {
  text := `{{ if .login_name }}
           resource "mssql_login" "{{ .name }}" {
             server {
               host = "{{ .host }}"
               {{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
             }
             login_name = "{{ .login_name }}"
             password   = "{{ .login_password }}"
           }
           {{ end }}
           resource "mssql_user" "{{ .name }}" {
             server {
               host = "{{ .host }}"
               {{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
             }
             {{ with .database }}database = "{{ . }}"{{ end }}
             {{ with .username }}username = "{{ . }}"{{ end }}
             {{ with .password }}password = "{{ . }}"{{ end }}
             {{ with .login_name }}login_name = "{{ . }}"{{ end }}
             {{ with .default_schema }}default_schema = "{{ . }}"{{ end }}
             {{ with .default_language }}default_language = "{{ . }}"{{ end }}
             {{ with .roles }}roles = {{ . }}{{ end }}
           }
           resource "mssql_database_permissions" "{{ .name }}" {
             server {
               host = "{{ .host }}"
               {{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
             }
             database     = "{{ .database }}"
             username = mssql_user.{{ .name }}.username
             permissions  = {{ .permissions }}
            }
           data "mssql_database_permissions" "{{ .name }}" {
             server {
               host = "{{ .host }}"
               {{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
             }
             database     = "{{ .database }}"
             username = mssql_user.{{ .name }}.username
             depends_on = [mssql_database_permissions.{{ .name }}]
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

func testAccCheckDataDatabasePermissionsDestroy(state *terraform.State) error {
  for _, rs := range state.RootModule().Resources {
    if rs.Type != "mssql_database_permissions" {
      continue
    }
    if rs.Type != "mssql_user" {
      continue
    }
    if rs.Type != "mssql_login" {
      continue
    }
    connector, err := getTestConnector(rs.Primary.Attributes)
    if err != nil {
      return err
    }

    database := rs.Primary.Attributes["database"]
    username := rs.Primary.Attributes["username"]

    permissions, err := connector.GetDatabasePermissions(database, username)
    if permissions != nil {
      return fmt.Errorf("permissions still exist")
    }
    if err != nil {
      return fmt.Errorf("expected no error, got %s", err)
    }
  }
  return nil
}
