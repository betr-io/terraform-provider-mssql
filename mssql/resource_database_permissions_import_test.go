package mssql

import (
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
  "testing"
)

func TestAccDatabasePermissions_Local_BasicImport(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabasePermissionsDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckDatabasePermissions(t, "test_import", "login", map[string]interface{}{"username": "db_user_import", "database":"master", "permissions": "[\"REFERENCES\"]", "login_name": "db_login_import", "login_password": "valueIsH8kd$¡"}),
        Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabasePermissionsExist("mssql_database_permissions.test_import"),
        ),
      },
      {
        ResourceName:      "mssql_database_permissions.test_import",
        ImportState:       true,
        ImportStateVerify: true,
        ImportStateIdFunc: testAccImportStateId("mssql_database_permissions.test_import", false),
      },
    },
  })
}
