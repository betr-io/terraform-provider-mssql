package mssql

import (
  "fmt"
  "os"
  "testing"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataDatabaseCredential_Azure_Basic(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckDataCredentialDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckDataCredential(t, "data_azure_test", "azure", map[string]interface{}{"database": "testdb", "credential_name": "test_scoped_data_cred", "identity_name": "test_identity_data_name", "secret": "V3ryS3cretP@asswd", "password": "V3ryS3cretP@asswd!Key"}),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("data.mssql_database_credential.data_azure_test", "id", "sqlserver://"+os.Getenv("TF_ACC_SQL_SERVER")+":1433/testdb/test_scoped_data_cred"),
          resource.TestCheckResourceAttr("data.mssql_database_credential.data_azure_test", "database", "testdb"),
          resource.TestCheckResourceAttr("data.mssql_database_credential.data_azure_test", "credential_name", "test_scoped_data_cred"),
          resource.TestCheckResourceAttr("data.mssql_database_credential.data_azure_test", "server.#", "1"),
          resource.TestCheckResourceAttr("data.mssql_database_credential.data_azure_test", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
          resource.TestCheckResourceAttr("data.mssql_database_credential.data_azure_test", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("data.mssql_database_credential.data_azure_test", "server.0.azure_login.#", "1"),
          resource.TestCheckResourceAttr("data.mssql_database_credential.data_azure_test", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
          resource.TestCheckResourceAttr("data.mssql_database_credential.data_azure_test", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
          resource.TestCheckResourceAttr("data.mssql_database_credential.data_azure_test", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
          resource.TestCheckResourceAttr("data.mssql_database_credential.data_azure_test", "server.0.login.#", "0"),
          resource.TestCheckResourceAttrSet("data.mssql_database_credential.data_azure_test", "principal_id"),
          resource.TestCheckResourceAttrSet("data.mssql_database_credential.data_azure_test", "credential_id"),
        ),
      },
    },
  })
}

func testAccCheckDataCredential(t *testing.T, name string, login string, data map[string]interface{}) string {
  text := `resource "mssql_database_masterkey" "{{ .name }}" {
             server {
               host = "{{ .host }}"
               {{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
             }
             database = "{{ .database }}"
             password = "{{ .password }}"
           }
           resource "mssql_database_credential" "{{ .name }}" {
             server {
               host = "{{ .host }}"
               {{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
             }
             database = "{{ .database }}"
             credential_name = "{{ .credential_name }}"
             identity_name = "{{ .identity_name }}"
             {{ with .secret }}secret = "{{ . }}"{{ end }}
             depends_on = [mssql_database_masterkey.{{ .name }}]
            }
           data "mssql_database_credential" "{{ .name }}" {
             server {
               host = "{{ .host }}"
               {{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
             }
             database = "{{ .database }}"
             credential_name = "{{ .credential_name }}"
             depends_on = [mssql_database_credential.{{ .name }}]
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

func testAccCheckDataCredentialDestroy(state *terraform.State) error {
  for _, rs := range state.RootModule().Resources {
    if rs.Type != "mssql_database_credential" {
      continue
    }
    if rs.Type != "mssql_database_masterkey" {
      continue
    }
    connector, err := getTestConnector(rs.Primary.Attributes)
    if err != nil {
      return err
    }

    database := rs.Primary.Attributes["database"]
    credentialName := rs.Primary.Attributes["credential_name"]
    scopedcredential, err := connector.GetDatabaseCredential(database, credentialName)
    if scopedcredential != nil {
      return fmt.Errorf("scoped credential still exists")
    }
    if err != nil {
      return fmt.Errorf("expected no error, got %s", err)
    }
  }
  return nil
}