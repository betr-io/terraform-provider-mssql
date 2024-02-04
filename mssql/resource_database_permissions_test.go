package mssql

import (
  "fmt"
  "os"
  "strconv"
  "testing"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatabasePermissions_Local_Basic(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabasePermissionsDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckDatabasePermissions(t, "database", "login", map[string]interface{}{"database":"master", "username": "db_user_perm", "permissions": "[\"REFERENCES\", \"UPDATE\"]", "login_name": "db_login_perm", "login_password": "valueIsH8kd$¡", "roles": "[\"db_owner\"]"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckDatabasePermissionsExist("mssql_database_permissions.database"),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "database", "master"),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "permissions.#", "2"),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "permissions.0", "REFERENCES"),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "permissions.1", "UPDATE"),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "server.0.host", "localhost"),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "server.0.login.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "server.0.azure_login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_database_permissions.database", "principal_id"),
        ),
      },
    },
  })
}

func TestAccDatabasePermissions_Azure_Basic(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabasePermissionsDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckDatabasePermissions(t, "database", "login", map[string]interface{}{"database":"testdb", "username": "db_user_perm", "permissions": "[\"EXECUTE\"]", "login_name": "db_login_perm", "login_password": "valueIsH8kd$¡", "roles": "[\"db_owner\"]"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckDatabasePermissionsExist("mssql_database_permissions.database"),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "database", "testdb"),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "permissions.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "permissions.0", "EXECUTE"),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "server.0.azure_login.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
          resource.TestCheckResourceAttr("mssql_database_permissions.database", "server.0.login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_database_permissions.database", "principal_id"),
        ),
      },
    },
  })
}

func testAccCheckDatabasePermissions(t *testing.T, name string, login string, data map[string]interface{}) string {
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
            principal_id = mssql_user.{{ .name }}.principal_id
            permissions  = {{ .permissions }}
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

func testAccCheckDatabasePermissionsDestroy(state *terraform.State) error {
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
    principalId := rs.Primary.Attributes["principal_id"]

    var pId int
    if pId, err := strconv.Atoi(principalId); err != nil {
      return fmt.Errorf("pId=%d, type: %T\n", pId, pId)
    }

    permissions, err := connector.GetDatabasePermissions(database, pId)
    if permissions != nil {
      return fmt.Errorf("permissions still exist")
    }
    if err != nil {
      return fmt.Errorf("expected no error, got %s", err)
    }
  }
  return nil
}

func testAccCheckDatabasePermissionsExist(resource string, checks ...Check) resource.TestCheckFunc {
  return func(state *terraform.State) error {
    rs, ok := state.RootModule().Resources[resource]
    if !ok {
      return fmt.Errorf("not found: %s", resource)
    }
    if rs.Type != "mssql_database_permissions" {
      return fmt.Errorf("expected resource of type %s, got %s", "mssql_database_permissions", rs.Type)
    }
    if rs.Primary.ID == "" {
      return fmt.Errorf("no record ID is set")
    }
    connector, err := getTestConnector(rs.Primary.Attributes)
    if err != nil {
      return err
    }

    database := rs.Primary.Attributes["database"]
    principalId := rs.Primary.Attributes["principal_id"]

    var pId int
    if pId, err := strconv.Atoi(principalId); err != nil {
      return fmt.Errorf("pId=%d, type: %T\n", pId, pId)
    }

    permissions, err := connector.GetDatabasePermissions(database, pId)
    if permissions == nil {
      return fmt.Errorf("permissions do not exist")
    }
    if err != nil {
      return fmt.Errorf("expected no error, got %s", err)
    }

    var actual interface{}
    for _, check := range checks {
      switch check.name {
      case "principal_id":
        actual = permissions.PrincipalID
      case "permission":
        actual = permissions.Permissions
      case "database":
        actual = permissions.DatabaseName
      default:
        return fmt.Errorf("unknown property %s", check.name)
      }
      if (check.op == "" || check.op == "==") && check.expected != actual {
        return fmt.Errorf("expected %s == %s, got %s", check.name, check.expected, actual)
      }
      if check.op == "!=" && check.expected == actual {
        return fmt.Errorf("expected %s != %s, got %s", check.name, check.expected, actual)
      }
    }
    return nil
  }
}
