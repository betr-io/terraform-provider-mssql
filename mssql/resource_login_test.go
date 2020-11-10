package mssql

import (
  "context"
  "fmt"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
  "terraform-provider-mssql/sql"
  "testing"
  "time"
)

func TestAccLogin_Basic(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(t, state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckLoginBasic(),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckMssqlLoginExists("mssql_login.test"),
          resource.TestCheckResourceAttr("mssql_login.test", "login_name", "testlogin"),
          resource.TestCheckResourceAttr("mssql_login.test", "password", "valueIsH8kd$¡"),
          resource.TestCheckResourceAttr("mssql_login.test", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_login.test", "server.0.host", "localhost"),
          resource.TestCheckResourceAttr("mssql_login.test", "server.0.login.#", "1"),
          resource.TestCheckResourceAttr("mssql_login.test", "server.0.login.0.username", "sa"),
        ),
      },
    },
  })
}

func TestAccItem_Update(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(t, state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckLoginUpdatePre(),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckMssqlLoginExists("mssql_login.test_update"),
          resource.TestCheckResourceAttr("mssql_login.test_update", "login_name", "testlogin_update_pre"),
        ),
      },
      {
        Config: testAccCheckLoginUpdatePost(),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckMssqlLoginExists("mssql_login.test_update"),
          resource.TestCheckResourceAttr("mssql_login.test_update", "login_name", "testlogin_update_post"),
        ),
      },
    },
  })
}

func testAccCheckLoginBasic() string {
  return `resource "mssql_login" "test" {
            server {
              host = "localhost"
              login {}
            }
            login_name = "testlogin"
            password   = "valueIsH8kd$¡"
          }`
}

func testAccCheckLoginUpdatePre() string {
  return `resource "mssql_login" "test_update" {
            server {
              host = "localhost"
              login {}
            }
            login_name = "testlogin_update_pre"
            password   = "valueIsH8kd$¡"
          }`
}

func testAccCheckLoginUpdatePost() string {
  return `resource "mssql_login" "test_update" {
            server {
              host = "localhost"
              login {}
            }
            login_name = "testlogin_update_post"
            password   = "valueIsH8kd$¡"
          }`
}

func testAccCheckLoginDestroy(t *testing.T, state *terraform.State) error {
  for _, rs := range state.RootModule().Resources {
    if rs.Type != "mssql_login" {
      continue
    }

    connector, err := getTestLoginConnector("server", rs.Primary.Attributes)
    if err != nil {
      return err
    }

    loginName := rs.Primary.Attributes["login_name"]
    login, err := connector.GetLogin(context.Background(), loginName)
    if login != nil {
      return fmt.Errorf("login still exists")
    }
    if err != nil {
      return fmt.Errorf("expected no error, got %s", err)
    }
  }
  return nil
}

func testAccCheckMssqlLoginExists(resource string) resource.TestCheckFunc {
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
    connector, err := getTestLoginConnector("server", rs.Primary.Attributes)
    if err != nil {
      return err
    }

    loginName := rs.Primary.Attributes["login_name"]
    login, err := connector.GetLogin(context.Background(), loginName)
    if login == nil {
      return fmt.Errorf("login does not exist")
    }
    if err != nil {
      return fmt.Errorf("expected no error, got %s", err)
    }
    return nil
  }
}

func getTestLoginConnector(prefix string, a map[string]string) (LoginConnector, error) {
  if len(prefix) > 0 {
    prefix = prefix + ".0."
  }

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

  return connector, nil
}
