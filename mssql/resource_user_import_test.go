package mssql

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccUser_Local_BasicImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckUserDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckUser(t, "test_import", "login", map[string]interface{}{"username": "user_import", "login_name": "user_import", "login_password": "valueIsH8kd$¡"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("mssql_user.test_import"),
				),
			},
			{
				ResourceName:      "mssql_user.test_import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccImportStateId("mssql_user.test_import", false),
			},
		},
	})
}
