package mssql

import (
  "fmt"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
  "os"
  "testing"
)

func TestAccUser_Local_Basic(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckUserDestroy(t, state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckUser(t, "basic", map[string]string{"username": "basic", "login_name": "user_basic", "login_password": "valueIsH8kd$¡", "roles": "[\"db_owner\"]"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckUserExists("mssql_user.basic"),
          testAccCheckUserWorks("mssql_user.basic", "user_basic", "valueIsH8kd$¡"),
          resource.TestCheckResourceAttr("mssql_user.basic", "database", "master"),
          resource.TestCheckResourceAttr("mssql_user.basic", "username", "basic"),
          resource.TestCheckResourceAttr("mssql_user.basic", "login_name", "user_basic"),
          resource.TestCheckResourceAttr("mssql_user.basic", "authentication_type", "INSTANCE"),
          resource.TestCheckResourceAttr("mssql_user.basic", "default_schema", "dbo"),
          resource.TestCheckResourceAttr("mssql_user.basic", "default_language", ""),
          resource.TestCheckResourceAttr("mssql_user.basic", "roles.#", "1"),
          resource.TestCheckResourceAttr("mssql_user.basic", "roles.0", "db_owner"),
          resource.TestCheckResourceAttr("mssql_user.basic", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_user.basic", "server.0.host", "localhost"),
          resource.TestCheckResourceAttr("mssql_user.basic", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_user.basic", "server.0.login.#", "1"),
          resource.TestCheckResourceAttr("mssql_user.basic", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("mssql_user.basic", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttr("mssql_user.basic", "server.0.azure_login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_user.basic", "principal_id"),
          resource.TestCheckNoResourceAttr("mssql_user.basic", "password"),
        ),
      },
    },
  })
}

func TestAccUser_Local_Update_DefaultSchema(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(t, state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckUser(t, "update", map[string]string{"username": "test_update", "login_name": "user_update", "login_password": "valueIsH8kd$¡"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckUserExists("mssql_user.update"),
          testAccCheckUserWorks("mssql_user.update", "user_update", "valueIsH8kd$¡"),
          resource.TestCheckResourceAttr("mssql_user.update", "default_schema", "dbo"),
        ),
      },
      {
        Config: testAccCheckUser(t, "update", map[string]string{"username": "test_update", "login_name": "user_update", "login_password": "valueIsH8kd$¡", "schema": "sys"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckUserExists("mssql_user.update"),
          testAccCheckUserWorks("mssql_user.update", "user_update", "valueIsH8kd$¡"),
          resource.TestCheckResourceAttr("mssql_user.update", "default_schema", "sys"),
        ),
      },
    },
  })
}

func testAccCheckUser(t *testing.T, name string, data map[string]string) string {
  text := `{{ if .login_name }}
           resource "mssql_login" "{{ .name }}" {
             server {
               host = "localhost"
               login {}
             }
             login_name = "{{ .login_name }}"
             password   = "{{ .login_password }}"
             {{ with .database }}default_database = "{{ . }}"{{ end }}
             {{ with .language }}default_language = "{{ . }}"{{ end }}
           }
           {{ end }}
           resource "mssql_user" "{{ .name }}" {
             server {
               host = "localhost"
               login {}
             }
             username = "{{ .username }}"
             {{ with .database }}database = "{{ . }}"{{ end }}
             {{ with .login_name }}login_name = "{{ . }}"{{ end }}
             {{ with .password }}password = "{{ . }}"{{ end }}
             {{ with .schema }}default_schema = "{{ . }}"{{ end }}
             {{ with .language }}default_language = "{{ . }}"{{ end }}
             {{ with .roles }}roles = {{ . }}{{ end }}
           }`
  data["name"] = name
  res, err := templateToString(name, text, data)
  if err != nil {
    t.Fatalf("%s", err)
  }
  return res
}

func testAccCheckUserDestroy(t *testing.T, state *terraform.State) error {
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
    login, err := connector.GetUser(database, username)
    if login != nil {
      return fmt.Errorf("user still exists")
    }
    if err != nil {
      return fmt.Errorf("expected no error, got %s", err)
    }
  }
  return nil
}

func testAccCheckUserExists(resource string) resource.TestCheckFunc {
  return func(state *terraform.State) error {
    rs, ok := state.RootModule().Resources[resource]
    if !ok {
      return fmt.Errorf("not found: %s", resource)
    }
    if rs.Type != "mssql_user" {
      return fmt.Errorf("expected resource of type %s, got %s", "mssql_user", rs.Type)
    }
    if rs.Primary.ID == "" {
      return fmt.Errorf("no record ID is set")
    }
    connector, err := getTestConnector(rs.Primary.Attributes)
    if err != nil {
      return err
    }

    database := rs.Primary.Attributes["database"]
    username := rs.Primary.Attributes["username"]
    login, err := connector.GetUser(database, username)
    if login == nil {
      return fmt.Errorf("login does not exist")
    }
    if err != nil {
      return fmt.Errorf("expected no error, got %s", err)
    }
    return nil
  }
}

func testAccCheckUserWorks(resource string, username, password string) resource.TestCheckFunc {
  return func(state *terraform.State) error {
    rs, ok := state.RootModule().Resources[resource]
    if !ok {
      return fmt.Errorf("not found: %s", resource)
    }
    if rs.Type != "mssql_user" {
      return fmt.Errorf("expected resource of type %s, got %s", "mssql_user", rs.Type)
    }
    if rs.Primary.ID == "" {
      return fmt.Errorf("no record ID is set")
    }
    connector, err := getTestUserConnector(rs.Primary.Attributes, username, password)
    if err != nil {
      return err
    }
    current, system, err := connector.GetCurrentUser(rs.Primary.Attributes[databaseProp])
    if err != nil {
      return fmt.Errorf("error: %s", err)
    }
    if current != rs.Primary.Attributes[usernameProp] {
      return fmt.Errorf("expected to be user %s, got %s (%s)", rs.Primary.Attributes[usernameProp], current, system)
    }
    return nil
  }
}
