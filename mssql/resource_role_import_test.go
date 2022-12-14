package mssql

import (
  "testing"

  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func Test_resourceRoleImport(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckRoleDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckRole(t, "test_import", "test-role-name", map[string]interface{}{}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckRoleExists("mssql_role.test_import"),
        ),
      },
      {
        ResourceName:      "mssql_role.test_import",
        ImportState:       true,
        ImportStateVerify: true,
        ImportStateIdFunc: testResourceTryGetStateId("mssql_role.test_import"),
      },
    },
  })
}
