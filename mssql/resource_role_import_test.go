package mssql

import (
  "testing"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRole_Local_BasicImport(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckRoleDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckRole(t, "test_import", "login", map[string]interface{}{"role_name": "test-role-name"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_role.test_import"),
        ),
      },
      {
        ResourceName:      "mssql_role.test_import",
        ImportState:       true,
        ImportStateVerify: true,
        ImportStateIdFunc: testAccImportStateId("mssql_role.test_import", false),
      },
    },
  })
}
