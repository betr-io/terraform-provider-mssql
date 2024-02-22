package mssql

import (
  "fmt"
  "os"
  "testing"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatabaseSchema_Local_Basic_Create(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckSchemaDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckSchema(t, "local_test_create", "login", map[string]interface{}{"schema_name": "test_schema_create"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckSchemaExists("mssql_database_schema.local_test_create"),
          resource.TestCheckResourceAttr("mssql_database_schema.local_test_create", "database", "master"),
          resource.TestCheckResourceAttr("mssql_database_schema.local_test_create", "schema_name", "test_schema_create"),
          resource.TestCheckResourceAttr("mssql_database_schema.local_test_create", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_schema.local_test_create", "server.0.host", "localhost"),
          resource.TestCheckResourceAttr("mssql_database_schema.local_test_create", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_database_schema.local_test_create", "server.0.login.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_schema.local_test_create", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("mssql_database_schema.local_test_create", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttr("mssql_database_schema.local_test_create", "server.0.azure_login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_database_schema.local_test_create", "schema_id"),
          resource.TestCheckNoResourceAttr("mssql_database_schema.local_test_create", "password"),
        ),
      },
    },
  })
}

func TestAccDatabaseSchema_Local_Basic_Create_owner(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckSchemaDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckSchema(t, "test_create_auth", "login", map[string]interface{}{"database": "master", "schema_name": "test_schema_auth", "owner_name": "db_user_schema", "username": "db_user_schema", "login_name": "db_login_schema", "login_password": "valueIsH8kd$¡", "roles": "[\"db_owner\"]"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckSchemaExists("mssql_database_schema.test_create_auth"),
          resource.TestCheckResourceAttr("mssql_database_schema.test_create_auth", "database", "master"),
          resource.TestCheckResourceAttr("mssql_database_schema.test_create_auth", "schema_name", "test_schema_auth"),
          resource.TestCheckResourceAttr("mssql_database_schema.test_create_auth", "owner_name", "db_user_schema"),
          resource.TestCheckResourceAttr("mssql_database_schema.test_create_auth", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_schema.test_create_auth", "server.0.host", "localhost"),
          resource.TestCheckResourceAttr("mssql_database_schema.test_create_auth", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_database_schema.test_create_auth", "server.0.login.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_schema.test_create_auth", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("mssql_database_schema.test_create_auth", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttr("mssql_database_schema.test_create_auth", "server.0.azure_login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_database_schema.test_create_auth", "schema_id"),
          resource.TestCheckNoResourceAttr("mssql_database_schema.test_create_auth", "password"),
        ),
      },
    },
  })
}

func TestAccDatabaseSchema_Azure_Basic_Create(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckSchemaDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckSchema(t, "azure_test_create", "azure", map[string]interface{}{"database":"testdb", "schema_name": "test_schema_create"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckSchemaExists("mssql_database_schema.azure_test_create"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create", "database", "testdb"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create", "schema_name", "test_schema_create"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create", "server.0.azure_login.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create", "server.0.login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_database_schema.azure_test_create", "schema_id"),
          resource.TestCheckNoResourceAttr("mssql_database_schema.azure_test_create", "password"),
        ),
      },
    },
  })
}

func TestAccDatabaseSchema_Azure_Basic_Create_owner(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckSchemaDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckSchema(t, "azure_test_create_auth", "azure", map[string]interface{}{"database": "testdb", "schema_name": "test_schema_auth", "owner_name": "db_user_schema", "username": "db_user_schema", "password": "valueIsH8kd$¡", "roles": "[\"db_owner\"]"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckSchemaExists("mssql_database_schema.azure_test_create_auth"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create_auth", "database", "testdb"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create_auth", "schema_name", "test_schema_auth"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create_auth", "owner_name", "db_user_schema"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create_auth", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create_auth", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create_auth", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create_auth", "server.0.azure_login.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create_auth", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create_auth", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create_auth", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_create_auth", "server.0.login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_database_schema.azure_test_create_auth", "schema_id"),
          resource.TestCheckNoResourceAttr("mssql_database_schema.azure_test_create_auth", "password"),
        ),
      },
    },
  })
}

func TestAccDatabaseSchema_Local_Basic_Update_owner(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckSchemaDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckSchema(t, "local_test_update_auth", "login", map[string]interface{}{"database": "master", "schema_name": "test_schema_owner", "owner_name": "db_user_owner_schema_pre", "username": "db_user_owner_schema_pre", "login_name": "db_login_owner_schema_pre", "login_password": "valueIsH8kd$¡", "roles": "[\"db_owner\"]"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckSchemaExists("mssql_database_schema.local_test_update_auth", Check{"owner_name", "==", "db_user_owner_schema_pre"}),
          resource.TestCheckResourceAttr("mssql_database_schema.local_test_update_auth", "schema_name", "test_schema_owner"),
          resource.TestCheckResourceAttr("mssql_database_schema.local_test_update_auth", "owner_name", "db_user_owner_schema_pre"),
        ),
      },
      {
        Config: testAccCheckSchema(t, "local_test_update_auth", "login", map[string]interface{}{"database": "master", "schema_name": "test_schema_owner", "owner_name": "db_user_owner_schema_post", "username": "db_user_owner_schema_post", "login_name": "db_login_owner_schema_post", "login_password": "valueIsH8kd$¡", "roles": "[\"db_owner\"]"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckSchemaExists("mssql_database_schema.local_test_update_auth", Check{"owner_name", "==", "db_user_owner_schema_post"}),
          resource.TestCheckResourceAttr("mssql_database_schema.local_test_update_auth", "schema_name", "test_schema_owner"),
          resource.TestCheckResourceAttr("mssql_database_schema.local_test_update_auth", "owner_name", "db_user_owner_schema_post"),
        ),
      },
    },
  })
}

func TestAccDatabaseSchema_Local_Basic_Update_remove_owner(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckSchemaDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckSchema(t, "local_test_update_rm_auth", "login", map[string]interface{}{"database": "master", "schema_name": "test_schema_owner_rm", "owner_name": "db_user_owner_rm", "username": "db_user_owner_rm", "login_name": "db_login_owner_rm", "login_password": "valueIsH8kd$¡"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckSchemaExists("mssql_database_schema.local_test_update_rm_auth"),
          resource.TestCheckResourceAttr("mssql_database_schema.local_test_update_rm_auth", "schema_name", "test_schema_owner_rm"),
          resource.TestCheckResourceAttr("mssql_database_schema.local_test_update_rm_auth", "owner_name", "db_user_owner_rm"),
        ),
      },
      {
        Config: testAccCheckSchema(t, "local_test_update_rm_auth", "login", map[string]interface{}{"database": "master", "schema_name": "test_schema_owner_rm", "owner_name": "", "username": "db_user_owner_rm", "login_name": "db_login_owner_rm", "login_password": "valueIsH8kd$¡"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckSchemaExists("mssql_database_schema.local_test_update_rm_auth"),
          resource.TestCheckResourceAttr("mssql_database_schema.local_test_update_rm_auth", "schema_name", "test_schema_owner_rm"),
          resource.TestCheckResourceAttr("mssql_database_schema.local_test_update_rm_auth", "owner_name", "dbo"),
        ),
      },
    },
  })
}

func TestAccDatabaseSchema_Azure_Basic_Update_owner(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckSchemaDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckSchema(t, "azure_test_update_owner", "azure", map[string]interface{}{"database": "testdb", "schema_name": "test_schema_auth", "owner_name": "db_user_schema", "username": "db_user_schema", "password": "valueIsH8kd$¡", "roles": "[\"db_owner\"]"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckSchemaExists("mssql_database_schema.azure_test_update_owner"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "database", "testdb"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "schema_name", "test_schema_auth"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "owner_name", "db_user_schema"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "server.0.azure_login.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "server.0.login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_database_schema.azure_test_update_owner", "schema_id"),
          resource.TestCheckNoResourceAttr("mssql_database_schema.azure_test_update_owner", "password"),
        ),
      },
      {
        Config: testAccCheckSchema(t, "azure_test_update_owner", "azure", map[string]interface{}{"database": "testdb", "schema_name": "test_schema_auth", "username": "db_user_schema", "password": "valueIsH8kd$¡", "roles": "[\"db_owner\"]"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckSchemaExists("mssql_database_schema.azure_test_update_owner"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "database", "testdb"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "schema_name", "test_schema_auth"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "owner_name", "dbo"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "server.0.azure_login.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
          resource.TestCheckResourceAttr("mssql_database_schema.azure_test_update_owner", "server.0.login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_database_schema.azure_test_update_owner", "schema_id"),
          resource.TestCheckNoResourceAttr("mssql_database_schema.azure_test_update_owner", "password"),
        ),
      },
    },
  })
}

func testAccCheckSchema(t *testing.T, name string, login string, data map[string]interface{}) string {
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
           resource "mssql_database_schema" "{{ .name }}" {
             server {
               host = "{{ .host }}"
               {{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
             }
             {{ with .database }}database = "{{ . }}"{{ end }}
             schema_name = "{{ .schema_name }}"
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

func testAccCheckSchemaDestroy(state *terraform.State) error {
  for _, rs := range state.RootModule().Resources {
    if rs.Type != "mssql_database_schema" {
      continue
    }

    connector, err := getTestConnector(rs.Primary.Attributes)
    if err != nil {
      return err
    }

    database := rs.Primary.Attributes["database"]
    schemaName := rs.Primary.Attributes["schema_name"]
    sqlschema, err := connector.GetDatabaseSchema(database, schemaName)
    if sqlschema != nil {
      return fmt.Errorf("schema still exists")
    }
    if err != nil {
      return fmt.Errorf("expected no error, got %s", err)
    }
  }
  return nil
}

func testAccCheckSchemaExists(resource string, checks ...Check) resource.TestCheckFunc {
  return func(state *terraform.State) error {
    rs, ok := state.RootModule().Resources[resource]
    if !ok {
      return fmt.Errorf("not found: %s", resource)
    }
    if rs.Type != "mssql_database_schema" {
      return fmt.Errorf("expected resource of type %s, got %s", "mssql_database_schema", rs.Type)
    }
    if rs.Primary.ID == "" {
      return fmt.Errorf("no record ID is set")
    }
    connector, err := getTestConnector(rs.Primary.Attributes)
    if err != nil {
      return err
    }
    database := rs.Primary.Attributes["database"]
    schemaName := rs.Primary.Attributes["schema_name"]
    sqlschema, err := connector.GetDatabaseSchema(database, schemaName)
    if err != nil {
      return fmt.Errorf("error: %s", err)
    }
    if sqlschema.SchemaName != schemaName {
      return fmt.Errorf("expected to be schema %s, got %s", schemaName, sqlschema.SchemaName)
    }

    var actual interface{}
    for _, check := range checks {
      switch check.name {
      case "schema_name":
        actual = sqlschema.SchemaName
      case "owner_name":
        actual = sqlschema.OwnerName
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
