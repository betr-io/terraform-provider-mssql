package mssql

import (
  "fmt"
  "os"
  "testing"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatabaseMasterkey_Local_Basic(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseMasterkeyDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckDatabaseMasterkey(t, "local_test_masterkey", "login", map[string]interface{}{"database": "master", "password": "V3ryS3cretP@asswd"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckDatabaseMasterkeyExists("mssql_database_masterkey.local_test_masterkey"),
          resource.TestCheckResourceAttr("mssql_database_masterkey.local_test_masterkey", "database", "master"),
          resource.TestCheckResourceAttr("mssql_database_masterkey.local_test_masterkey", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_masterkey.local_test_masterkey", "server.0.host", "localhost"),
          resource.TestCheckResourceAttr("mssql_database_masterkey.local_test_masterkey", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_database_masterkey.local_test_masterkey", "server.0.login.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_masterkey.local_test_masterkey", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("mssql_database_masterkey.local_test_masterkey", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttr("mssql_database_masterkey.local_test_masterkey", "server.0.azure_login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_database_masterkey.local_test_masterkey", "principal_id"),
          resource.TestCheckResourceAttrSet("mssql_database_masterkey.local_test_masterkey", "key_guid"),
        ),
      },
    },
  })
}

func TestAccDatabaseMasterkey_Local_Basic_update(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseMasterkeyDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckDatabaseMasterkey(t, "update", "login", map[string]interface{}{"database": "master", "password": "V3ryS3cretP@asswd"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckDatabaseMasterkeyExists("mssql_database_masterkey.update"),
          resource.TestCheckResourceAttr("mssql_database_masterkey.update", "database", "master"),
          resource.TestCheckResourceAttr("mssql_database_masterkey.update", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_masterkey.update", "server.0.host", "localhost"),
          resource.TestCheckResourceAttr("mssql_database_masterkey.update", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_database_masterkey.update", "server.0.login.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_masterkey.update", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("mssql_database_masterkey.update", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttr("mssql_database_masterkey.update", "server.0.azure_login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_database_masterkey.update", "principal_id"),
          resource.TestCheckResourceAttrSet("mssql_database_masterkey.update", "key_guid"),
        ),
      },
      {
        Config: testAccCheckDatabaseMasterkey(t, "update", "login", map[string]interface{}{"database": "master", "password": "V3ryS3cretP@asswdUpdated123"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckDatabaseMasterkeyExists("mssql_database_masterkey.update"),
          resource.TestCheckResourceAttr("mssql_database_masterkey.update", "database", "master"),
          resource.TestCheckResourceAttr("mssql_database_masterkey.update", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_masterkey.update", "server.0.host", "localhost"),
          resource.TestCheckResourceAttr("mssql_database_masterkey.update", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_database_masterkey.update", "server.0.login.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_masterkey.update", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("mssql_database_masterkey.update", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttr("mssql_database_masterkey.update", "server.0.azure_login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_database_masterkey.update", "principal_id"),
          resource.TestCheckResourceAttrSet("mssql_database_masterkey.update", "key_guid"),
        ),
      },
    },
  })
}

func testAccCheckDatabaseMasterkey(t *testing.T, name string, login string, data map[string]interface{}) string {
  text := `resource "mssql_database_masterkey" "{{ .name }}" {
             server {
               host = "{{ .host }}"
               {{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
             }
             database = "{{ .database }}"
             password = "{{ .password }}"
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

func testAccCheckDatabaseMasterkeyDestroy(state *terraform.State) error {
  for _, rs := range state.RootModule().Resources {
    if rs.Type != "mssql_database_masterkey" {
      continue
    }

    connector, err := getTestConnector(rs.Primary.Attributes)
    if err != nil {
      return err
    }

    database := rs.Primary.Attributes["database"]
    masterkey, err := connector.GetDatabaseMasterkey(database)
    if masterkey != nil {
      return fmt.Errorf("database master key still exists")
    }
    if err != nil {
      return fmt.Errorf("expected no error, got %s", err)
    }
  }
  return nil
}

func testAccCheckDatabaseMasterkeyExists(resource string, checks ...Check) resource.TestCheckFunc {
  return func(state *terraform.State) error {
    rs, ok := state.RootModule().Resources[resource]
    if !ok {
      return fmt.Errorf("not found: %s", resource)
    }
    if rs.Type != "mssql_database_masterkey" {
      return fmt.Errorf("expected resource of type %s, got %s", "mssql_database_masterkey", rs.Type)
    }
    if rs.Primary.ID == "" {
      return fmt.Errorf("no record ID is set")
    }
    connector, err := getTestConnector(rs.Primary.Attributes)
    if err != nil {
      return err
    }
    database := rs.Primary.Attributes["database"]
    keyguid := rs.Primary.Attributes["key_guid"]
    masterkey, err := connector.GetDatabaseMasterkey(database)
    if err != nil {
      return fmt.Errorf("error: %s", err)
    }
    if masterkey.KeyGuid != keyguid {
      return fmt.Errorf("expected to be key_guid %s, got %s", keyguid, masterkey.KeyGuid)
    }

    var actual interface{}
    for _, check := range checks {
      switch check.name {
      case "key_guid":
        actual = masterkey.KeyGuid
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
