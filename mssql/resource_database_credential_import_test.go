package mssql

import (
  "os"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
  "testing"
)

func TestAccDatabaseCredential_Azure_BasicImport(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseCredemtialDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckDatabaseCredential(t, "test_import", "azure", map[string]interface{}{"database": "testdb", "credential_name": "test_scoped_cred_import", "identity_name": "test_identity_name_import", "secret": "V3ryS3cretP@asswd", "password": "V3ryS3cretP@asswd!Key"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckDatabaseCredemtialExists("mssql_database_credential.test_import"),
          resource.TestCheckResourceAttr("mssql_database_credential.test_import", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_credential.test_import", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
          resource.TestCheckResourceAttr("mssql_database_credential.test_import", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_database_credential.test_import", "server.0.azure_login.#", "1"),
          resource.TestCheckResourceAttr("mssql_database_credential.test_import", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
          resource.TestCheckResourceAttr("mssql_database_credential.test_import", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
          resource.TestCheckResourceAttr("mssql_database_credential.test_import", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
          resource.TestCheckResourceAttr("mssql_database_credential.test_import", "server.0.login.#", "0"),
        ),
      },
      {
        ResourceName:      "mssql_database_credential.test_import",
        ImportState:       true,
        ImportStateVerify: true,
        ImportStateVerifyIgnore: []string{"secret"},
        ImportStateIdFunc: testAccImportStateId("mssql_database_credential.test_import", true),
      },
    },
  })
}
