package mssql

import (
  "fmt"
  "os"
  "testing"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataUser_Local_Basic(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccDataUserDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccDataUser(t, "basic", "login", map[string]interface{}{"username": "instance", "login_name": "user_instance", "login_password": "valueIsH8kd$¡", "roles": "[\"db_owner\"]"}),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("data.mssql_user.basic", "id", "sqlserver://localhost:1433/master/instance"),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "database", "master"),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "username", "instance"),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "login_name", "user_instance"),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "roles.#", "1"),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "roles.0", "db_owner"),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "server.#", "1"),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "server.0.host", "localhost"),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "server.0.login.#", "1"),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "server.0.azure_login.#", "0"),
          resource.TestCheckResourceAttrSet("data.mssql_user.basic", "principal_id"),
          resource.TestCheckResourceAttrSet("data.mssql_user.basic", "authentication_type"),
          resource.TestCheckResourceAttrSet("data.mssql_user.basic", "sid"),
        ),
      },
    },
  })
}

func TestAccDataUser_Azure_Basic(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccDataLoginDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccDataUser(t, "basic", "azure", map[string]interface{}{"database": "testdb", "username": "instance", "login_name": "user_instance", "login_password": "valueIsH8kd$¡", "roles": "[\"db_owner\"]"}),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("data.mssql_user.basic", "id", "sqlserver://"+os.Getenv("TF_ACC_SQL_SERVER")+":1433/testdb/instance"),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "database", "testdb"),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "username", "instance"),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "server.#", "1"),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "server.0.azure_login.#", "1"),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
          resource.TestCheckResourceAttr("data.mssql_user.basic", "server.0.login.#", "0"),
          resource.TestCheckResourceAttrSet("data.mssql_user.basic", "principal_id"),
          resource.TestCheckResourceAttrSet("data.mssql_user.basic", "login_name"),
          resource.TestCheckResourceAttrSet("data.mssql_user.basic", "authentication_type"),
          resource.TestCheckResourceAttrSet("data.mssql_user.basic", "sid"),
        ),
      },
    },
  })
}

func testAccDataUser(t *testing.T, name string, login string, data map[string]interface{}) string {
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
             username = "{{ .username }}"
             {{ with .password }}password = "{{ . }}"{{ end }}
             {{ with .login_name }}login_name = "{{ . }}"{{ end }}
             {{ with .default_schema }}default_schema = "{{ . }}"{{ end }}
             {{ with .default_language }}default_language = "{{ . }}"{{ end }}
             {{ with .roles }}roles = {{ . }}{{ end }}
           }
           data "mssql_user" "{{ .name }}" {
             server {
               host = "{{ .host }}"
               {{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
             }
             {{ with .database }}database = "{{ . }}"{{ end }}
             username = "{{ .username }}"
             depends_on = [mssql_user.{{ .name }}]
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

func testAccDataUserDestroy(state *terraform.State) error {
  for _, rs := range state.RootModule().Resources {
    if rs.Type != "mssql_user" {
      continue
    }

    connector, err := getTestConnector(rs.Primary.Attributes)
    if err != nil {
      return err
    }

    database := rs.Primary.Attributes["database"]
    username := rs.Primary.Attributes["username"]
    user, err := connector.GetUser(database, username)
    if user != nil {
      return fmt.Errorf("user still exists")
    }
    if err != nil {
      return fmt.Errorf("expected no error, got %s", err)
    }
  }
  return nil
}
