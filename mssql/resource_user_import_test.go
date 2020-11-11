package mssql

import (
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
  "testing"
)

func TestAccUser_importBasic(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckUserDestroy(t, state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckUser(t, "test_import", map[string]string{"username": "user_import", "login_name": "user_import", "login_password": "valueIsH8kd$¡"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckUserExists("mssql_user.test_import"),
        ),
      },
      {
        ResourceName:            "mssql_user.test_import",
        ImportState:             true,
        ImportStateVerify:       true,
      },
    },
  })
}
