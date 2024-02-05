package mssql

import (
  "fmt"
  "os"
  "testing"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRole_Local_Basic_Create(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckRoleDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckRole(t, "test_create", "login", map[string]interface{}{"role_name": "test-role-name"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_role.test_create"),
          resource.TestCheckResourceAttr("mssql_role.test_create", "database", "master"),
          resource.TestCheckResourceAttr("mssql_role.test_create", "role_name", "test-role-name"),
          resource.TestCheckResourceAttr("mssql_role.test_create", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_role.test_create", "server.0.host", "localhost"),
          resource.TestCheckResourceAttr("mssql_role.test_create", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_role.test_create", "server.0.login.#", "1"),
          resource.TestCheckResourceAttr("mssql_role.test_create", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("mssql_role.test_create", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttr("mssql_role.test_create", "server.0.azure_login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_role.test_create", "principal_id"),
          resource.TestCheckNoResourceAttr("mssql_role.test_create", "password"),
        ),
      },
    },
  })
}

func TestAccRole_Azure_Basic_Create(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckRoleDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckRole(t, "test_create", "azure", map[string]interface{}{"role_name": "test-role-name"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_role.test_create"),
          resource.TestCheckResourceAttr("mssql_role.test_create", "database", "master"),
          resource.TestCheckResourceAttr("mssql_role.test_create", "role_name", "test-role-name"),
          resource.TestCheckResourceAttr("mssql_role.test_create", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_role.test_create", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
          resource.TestCheckResourceAttr("mssql_role.test_create", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_role.test_create", "server.0.azure_login.#", "1"),
          resource.TestCheckResourceAttr("mssql_role.test_create", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
          resource.TestCheckResourceAttr("mssql_role.test_create", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
          resource.TestCheckResourceAttr("mssql_role.test_create", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
          resource.TestCheckResourceAttr("mssql_role.test_create", "server.0.login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_role.test_create", "principal_id"),
          resource.TestCheckNoResourceAttr("mssql_role.test_create", "password"),
        ),
      },
    },
  })
}

func TestAccRole_Local_Basic_Update(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckRoleDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckRole(t, "test_update", "login", map[string]interface{}{"role_name": "test-role-pre"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_role.test_update"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "database", "master"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "role_name", "test-role-pre"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.host", "localhost"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.login.#", "1"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.azure_login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_role.test_update", "principal_id"),
          resource.TestCheckNoResourceAttr("mssql_role.test_update", "password"),
        ),
      },
      {
        Config: testAccCheckRole(t, "test_update", "login", map[string]interface{}{"role_name": "test-role-post"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_role.test_update"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "database", "master"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "role_name", "test-role-post"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.host", "localhost"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.login.#", "1"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.azure_login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_role.test_update", "principal_id"),
          resource.TestCheckNoResourceAttr("mssql_role.test_update", "password"),
        ),
      },
    },
  })
}

func TestAccRole_Azure_Basic_Update(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckRoleDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckRole(t, "test_update", "azure", map[string]interface{}{"role_name": "test-role-pre"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_role.test_update"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "database", "master"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "role_name", "test-role-pre"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.azure_login.#", "1"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_role.test_update", "principal_id"),
          resource.TestCheckNoResourceAttr("mssql_role.test_update", "password"),
        ),
      },
      {
        Config: testAccCheckRole(t, "test_update", "azure", map[string]interface{}{"role_name": "test-role-post"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_role.test_update"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "database", "master"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "role_name", "test-role-post"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.azure_login.#", "1"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_role.test_update", "principal_id"),
          resource.TestCheckNoResourceAttr("mssql_role.test_update", "password"),
        ),
      },
    },
  })
}

func testAccCheckRole(t *testing.T, name string, login string, data map[string]interface{}) string {
  text := `resource "mssql_role" "{{ .name }}" {
             server {
              host = "{{ .host }}"
              {{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
             }
             {{ with .database }}database = "{{ . }}"{{ end }}
             role_name = "{{ .role_name }}"
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

func testAccCheckRoleDestroy(state *terraform.State) error {
  for _, rs := range state.RootModule().Resources {
    if rs.Type != "mssql_role" {
      continue
    }

    connector, err := getTestConnector(rs.Primary.Attributes)
    if err != nil {
      return err
    }

    database := rs.Primary.Attributes["database"]
    roleName := rs.Primary.Attributes["username"]
    role, err := connector.GetRole(database, roleName)
    if role != nil {
      return fmt.Errorf("role still exists")
    }
    if err != nil {
      return fmt.Errorf("expected no error, got %s", err)
    }
  }
  return nil
}

func testAccCheckRoleExists(resource string) resource.TestCheckFunc {
  return func(state *terraform.State) error {
    rs, ok := state.RootModule().Resources[resource]
    if !ok {
      return fmt.Errorf("not found: %s", resource)
    }
    if rs.Type != "mssql_role" {
      return fmt.Errorf("expected resource of type %s, got %s", "mssql_role", rs.Type)
    }
    if rs.Primary.ID == "" {
      return fmt.Errorf("no record ID is set")
    }
    connector, err := getTestConnector(rs.Primary.Attributes)
    if err != nil {
      return err
    }
    database := rs.Primary.Attributes["database"]
    roleName := rs.Primary.Attributes["role_name"]
    role, err := connector.GetRole(database, roleName)
    if err != nil {
      return fmt.Errorf("error: %s", err)
    }
    if role.RoleName != roleName {
      return fmt.Errorf("expected to be role %s, got %s", roleName, role.RoleName)
    }
    return nil
  }
}
