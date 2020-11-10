package mssql

import (
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
  "testing"
)

func TestAccLogin_importBasic(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(t, state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckMssqlLoginImporterBasic(),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckMssqlLoginExists("mssql_login.test_import"),
        ),
      },
      {
        ResourceName:            "mssql_login.test_import",
        ImportState:             true,
        ImportStateVerify:       true,
        ImportStateVerifyIgnore: []string{"password"},
      },
    },
  })
}

func testAccCheckMssqlLoginImporterBasic() string {
  return `resource "mssql_login" "test_import" {
            server {
              host = "localhost"
              login {}
            }
            login_name = "testlogin"
            password   = "valueIsH8kd$¡"
         }`
}
