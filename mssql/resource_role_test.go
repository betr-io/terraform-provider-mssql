package mssql

import (
  "fmt"
  "os"
  "testing"

  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestRole_Create(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckRoleDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckRole(t, "test_create", "test-role-name", map[string]interface{}{}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_role.test_create"),
          resource.TestCheckResourceAttr("mssql_role.test_create", "database", "master"),
          resource.TestCheckResourceAttr("mssql_role.test_create", "role_name", "test-role-name"),
          resource.TestCheckResourceAttr("mssql_role.test_create", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("mssql_role.test_create", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttrSet("mssql_role.test_create", "principal_id"),
          resource.TestCheckNoResourceAttr("mssql_role.test_create", "password"),
        ),
      },
    },
  })
}

func TestRole_Update(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckRoleDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckRole(t, "test_update", "test-role-pre", map[string]interface{}{}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_role.test_update"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "database", "master"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "role_name", "test-role-pre"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttrSet("mssql_role.test_update", "principal_id"),
          resource.TestCheckNoResourceAttr("mssql_role.test_update", "password"),
        ),
      },
      {
        Config: testAccCheckRole(t, "test_update", "test-role-post", map[string]interface{}{}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_role.test_update"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "database", "master"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "role_name", "test-role-post"),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("mssql_role.test_update", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttrSet("mssql_role.test_update", "principal_id"),
          resource.TestCheckNoResourceAttr("mssql_role.test_update", "password"),
        ),
      },
    },
  })
}

func testAccCheckRole(t *testing.T, name string, roleName string, data map[string]interface{}) string {
  text := `resource "mssql_role" "{{ .name }}" {
             server {
               host = "localhost"
         login {}
             }
       {{ with .database }}database = "{{ . }}"{{ end }}
             role_name = "{{ .role_name }}"
           }`

  data["name"] = name

  data["role_name"] = roleName
  data["host"] = "localhost"

  res, err := templateToString(roleName, text, data)
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
