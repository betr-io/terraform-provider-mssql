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
        Config: testAccCheckRole(t, "local_test_create", "login", map[string]interface{}{"role_name": "test_role_create"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_database_role.local_test_create"),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_create", "database", "master"),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_create", "role_name", "test_role_create"),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_create", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_create", "server.0.host", "localhost"),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_create", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_create", "server.0.login.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_create", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_create", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_create", "server.0.azure_login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_database_role.local_test_create", "principal_id"),
          resource.TestCheckNoResourceAttr("mssql_database_role.local_test_create", "password"),
        ),
      },
    },
  })
}

func TestAccRole_Local_Basic_Create_with_Authorization(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckRoleDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckRole(t, "test_create_auth", "login", map[string]interface{}{"database": "master", "role_name": "test_role_auth", "owner_name": "db_user_role", "username": "db_user_role", "login_name": "db_login_role", "login_password": "valueIsH8kd$¡", "roles": "[\"db_owner\"]"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_database_role.test_create_auth"),
          resource.TestCheckResourceAttr("mssql_database_role.test_create_auth", "database", "master"),
          resource.TestCheckResourceAttr("mssql_database_role.test_create_auth", "role_name", "test_role_auth"),
          resource.TestCheckResourceAttr("mssql_database_role.test_create_auth", "owner_name", "db_user_role"),
          resource.TestCheckResourceAttr("mssql_database_role.test_create_auth", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_role.test_create_auth", "server.0.host", "localhost"),
          resource.TestCheckResourceAttr("mssql_database_role.test_create_auth", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_database_role.test_create_auth", "server.0.login.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_role.test_create_auth", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("mssql_database_role.test_create_auth", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttr("mssql_database_role.test_create_auth", "server.0.azure_login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_database_role.test_create_auth", "principal_id"),
          resource.TestCheckNoResourceAttr("mssql_database_role.test_create_auth", "password"),
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
        Config: testAccCheckRole(t, "azure_test_create", "azure", map[string]interface{}{"role_name": "test_role_create"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_database_role.test_create"),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_create", "database", "master"),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_create", "role_name", "test_role_create"),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_create", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_create", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_create", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_create", "server.0.azure_login.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_create", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_create", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_create", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_create", "server.0.login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_database_role.azure_test_create", "principal_id"),
          resource.TestCheckNoResourceAttr("mssql_database_role.azure_test_create", "password"),
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
        Config: testAccCheckRole(t, "local_test_update", "login", map[string]interface{}{"role_name": "test_role_pre"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_database_role.local_test_update", Check{"role_name", "==", "test_role_pre"}),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_update", "database", "master"),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_update", "role_name", "test_role_pre"),
        ),
      },
      {
        Config: testAccCheckRole(t, "local_test_update", "login", map[string]interface{}{"role_name": "test_role_post"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_database_role.local_test_update", Check{"role_name", "==", "test_role_post"}),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_update", "database", "master"),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_update", "role_name", "test_role_post"),
        ),
      },
    },
  })
}

func TestAccRole_Local_Basic_Update_with_Authorization(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckRoleDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckRole(t, "local_test_update_auth", "login", map[string]interface{}{"database": "master", "role_name": "test_role_owner", "owner_name": "db_user_owner_pre", "username": "db_user_owner_pre", "login_name": "db_login_owner1", "login_password": "valueIsH8kd$¡", "roles": "[\"db_owner\"]"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_database_role.local_test_update_auth", Check{"owner_name", "==", "db_user_owner_pre"}),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_update_auth", "role_name", "test_role_owner"),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_update_auth", "owner_name", "db_user_owner_pre"),
        ),
      },
      {
        Config: testAccCheckRole(t, "local_test_update_auth", "login", map[string]interface{}{"database": "master", "role_name": "test_role_owner", "owner_name": "db_user_owner_post", "username": "db_user_owner_post", "login_name": "db_login_owner2", "login_password": "valueIsH8kd$¡", "roles": "[\"db_owner\"]"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_database_role.local_test_update_auth", Check{"owner_name", "==", "db_user_owner_post"}),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_update_auth", "role_name", "test_role_owner"),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_update_auth", "owner_name", "db_user_owner_post"),
        ),
      },
    },
  })
}

func TestAccRole_Local_Basic_Update_Role_and_Authorization(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckRoleDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckRole(t, "local_test_update_role_auth", "login", map[string]interface{}{"database": "master", "role_name": "test_role_owner_pre", "owner_name": "db_user_owner_pre", "username": "db_user_owner_pre", "login_name": "db_login_owner1", "login_password": "valueIsH8kd$¡", "roles": "[\"db_owner\"]"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_database_role.local_test_update_role_auth"),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_update_role_auth", "role_name", "test_role_owner_pre"),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_update_role_auth", "owner_name", "db_user_owner_pre"),
        ),
      },
      {
        Config: testAccCheckRole(t, "local_test_update_role_auth", "login", map[string]interface{}{"database": "master", "role_name": "test_role_owner_post", "owner_name": "db_user_owner_post", "username": "db_user_owner_post", "login_name": "db_login_owner2", "login_password": "valueIsH8kd$¡", "roles": "[\"db_owner\"]"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_database_role.local_test_update_role_auth"),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_update_role_auth", "role_name", "test_role_owner_post"),
          resource.TestCheckResourceAttr("mssql_database_role.local_test_update_role_auth", "owner_name", "db_user_owner_post"),
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
        Config: testAccCheckRole(t, "azure_test_update", "azure", map[string]interface{}{"role_name": "test_role_pre"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_database_role.test_update"),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_update", "database", "master"),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_update", "role_name", "test_role_pre"),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_update", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_update", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_update", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_update", "server.0.azure_login.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_update", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_update", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_update", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_update", "server.0.login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_database_role.azure_test_update", "principal_id"),
          resource.TestCheckNoResourceAttr("mssql_database_role.azure_test_update", "password"),
        ),
      },
      {
        Config: testAccCheckRole(t, "azure_test_update", "azure", map[string]interface{}{"role_name": "test_role_post"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_database_role.test_update"),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_update", "database", "master"),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_update", "role_name", "test_role_post"),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_update", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_update", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_update", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_update", "server.0.azure_login.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_update", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_update", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_update", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
          resource.TestCheckResourceAttr("mssql_database_role.azure_test_update", "server.0.login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_database_role.azure_test_update", "principal_id"),
          resource.TestCheckNoResourceAttr("mssql_database_role.azure_test_update", "password"),
        ),
      },
    },
  })
}

func testAccCheckRole(t *testing.T, name string, login string, data map[string]interface{}) string {
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
           {{ if .username }}
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
           {{ end }}
           resource "mssql_database_role" "{{ .name }}" {
             server {
               host = "{{ .host }}"
               {{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
             }
             {{ with .database }}database = "{{ . }}"{{ end }}
             role_name = "{{ .role_name }}"
             {{ with .owner_name }}owner_name = "{{ . }}"{{ end }}
             {{ if .username }}
             depends_on = [mssql_user.{{ .name }}]
             {{ end }}
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
    if rs.Type != "mssql_database_role" {
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

func testAccCheckRoleExists(resource string, checks ...Check) resource.TestCheckFunc {
  return func(state *terraform.State) error {
    rs, ok := state.RootModule().Resources[resource]
    if !ok {
      return fmt.Errorf("not found: %s", resource)
    }
    if rs.Type != "mssql_database_role" {
      return fmt.Errorf("expected resource of type %s, got %s", "mssql_database_role", rs.Type)
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

    var actual interface{}
    for _, check := range checks {
      switch check.name {
      case "role_name":
        actual = role.RoleName
      case "owner_name":
        actual = role.OwnerName
      default:
        return fmt.Errorf("unknown property %s", check.name)
      }
      if (check.op == "" || check.op == "==") && !equal(check.expected, actual) {
        return fmt.Errorf("expected %s == %s, got %s", check.name, check.expected, actual)
      }
      if check.op == "!=" && equal(check.expected, actual) {
        return fmt.Errorf("expected %s != %s, got %s", check.name, check.expected, actual)
      }
    }
    return nil
  }
}
