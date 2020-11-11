package mssql

import (
  "fmt"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
  "os"
  "testing"
)

func TestAccLogin_Basic(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(t, state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckLogin(t, "basic", map[string]string{"login_name": "login_basic", "password": "valueIsH8kd$¡"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckLoginExists("mssql_login.basic"),
          testAccCheckLoginWorks("mssql_login.basic"),
          resource.TestCheckResourceAttr("mssql_login.basic", "login_name", "login_basic"),
          resource.TestCheckResourceAttr("mssql_login.basic", "password", "valueIsH8kd$¡"),
          resource.TestCheckResourceAttr("mssql_login.basic", "default_database", "master"),
          resource.TestCheckResourceAttr("mssql_login.basic", "default_language", "us_english"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.host", "localhost"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.login.#", "1"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_login.basic", "principal_id"),
        ),
      },
    },
  })
}

func TestAccLogin_Update(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(t, state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckLogin(t, "test_update", map[string]string{"login_name": "login_update_pre", "password": "valueIsH8kd$¡", "database": "tempdb", "language": "russian"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckLoginExists("mssql_login.test_update"),
          testAccCheckLoginWorks("mssql_login.test_update"),
          resource.TestCheckResourceAttr("mssql_login.test_update", "login_name", "login_update_pre"),
          resource.TestCheckResourceAttr("mssql_login.test_update", "default_database", "tempdb"),
          resource.TestCheckResourceAttr("mssql_login.test_update", "default_language", "russian"),
        ),
      },
      {
        Config: testAccCheckLogin(t, "test_update", map[string]string{"login_name": "login_update_pre", "password": "valueIsH8kd$¡"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckLoginExists("mssql_login.test_update"),
          testAccCheckLoginWorks("mssql_login.test_update"),
          resource.TestCheckResourceAttr("mssql_login.test_update", "login_name", "login_update_pre"),
          resource.TestCheckResourceAttr("mssql_login.test_update", "default_database", "master"),
          resource.TestCheckResourceAttr("mssql_login.test_update", "default_language", "us_english"),
        ),
      },
      {
        Config: testAccCheckLogin(t, "test_update", map[string]string{"login_name": "login_update_pre", "password": "otherIsH8kd$¡", "database": "tempdb", "language": "russian"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckLoginExists("mssql_login.test_update"),
          testAccCheckLoginWorks("mssql_login.test_update"),
          resource.TestCheckResourceAttr("mssql_login.test_update", "login_name", "login_update_pre"),
          resource.TestCheckResourceAttr("mssql_login.test_update", "password", "otherIsH8kd$¡"),
          resource.TestCheckResourceAttr("mssql_login.test_update", "default_database", "tempdb"),
          resource.TestCheckResourceAttr("mssql_login.test_update", "default_language", "russian"),
        ),
      },
      {
        Config: testAccCheckLogin(t, "test_update", map[string]string{"login_name": "login_update_post", "password": "otherIsH8kd$¡"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckLoginExists("mssql_login.test_update"),
          testAccCheckLoginWorks("mssql_login.test_update"),
          resource.TestCheckResourceAttr("mssql_login.test_update", "login_name", "login_update_post"),
          resource.TestCheckResourceAttr("mssql_login.test_update", "default_database", "master"),
          resource.TestCheckResourceAttr("mssql_login.test_update", "default_language", "us_english"),
        ),
      },
    },
  })
}

func testAccCheckLogin(t *testing.T, name string, data map[string]string) string {
  text := `resource "mssql_login" "{{ .name }}" {
             server {
               host = "localhost"
               login {}
             }
             login_name = "{{ .login_name }}"
             password   = "{{ .password }}"
             {{ with .database }}default_database = "{{ . }}"{{ end }}
             {{ with .language }}default_language = "{{ . }}"{{ end }}
           }`
  data["name"] = name
  res, err := templateToString(name, text, data)
  if err != nil {
    t.Fatalf("%s", err)
  }
  return res
}

func testAccCheckLoginDestroy(t *testing.T, state *terraform.State) error {
  for _, rs := range state.RootModule().Resources {
    if rs.Type != "mssql_login" {
      continue
    }

    connector, err := getTestConnector(rs.Primary.Attributes)
    if err != nil {
      return err
    }

    loginName := rs.Primary.Attributes["login_name"]
    login, err := connector.GetLogin(loginName)
    if login != nil {
      return fmt.Errorf("login still exists")
    }
    if err != nil {
      return fmt.Errorf("expected no error, got %s", err)
    }
  }
  return nil
}

func testAccCheckLoginExists(resource string) resource.TestCheckFunc {
  return func(state *terraform.State) error {
    rs, ok := state.RootModule().Resources[resource]
    if !ok {
      return fmt.Errorf("not found: %s", resource)
    }
    if rs.Type != "mssql_login" {
      return fmt.Errorf("expected resource of type %s, got %s", "mssql_login", rs.Type)
    }
    if rs.Primary.ID == "" {
      return fmt.Errorf("no record ID is set")
    }
    connector, err := getTestConnector(rs.Primary.Attributes)
    if err != nil {
      return err
    }

    loginName := rs.Primary.Attributes["login_name"]
    login, err := connector.GetLogin(loginName)
    if login == nil {
      return fmt.Errorf("login does not exist")
    }
    if err != nil {
      return fmt.Errorf("expected no error, got %s", err)
    }
    return nil
  }
}

func testAccCheckLoginWorks(resource string) resource.TestCheckFunc {
  return func(state *terraform.State) error {
    rs, ok := state.RootModule().Resources[resource]
    if !ok {
      return fmt.Errorf("not found: %s", resource)
    }
    if rs.Type != "mssql_login" {
      return fmt.Errorf("expected resource of type %s, got %s", "mssql_login", rs.Type)
    }
    if rs.Primary.ID == "" {
      return fmt.Errorf("no record ID is set")
    }
    connector, err := getTestLoginConnector(rs.Primary.Attributes)
    if err != nil {
      return err
    }
    systemUser, err := connector.GetSystemUser()
    if systemUser != rs.Primary.Attributes[loginNameProp] {
      return fmt.Errorf("expected to log in as %s, got %s", rs.Primary.Attributes[loginNameProp], systemUser)
    }
    return nil
  }
}
