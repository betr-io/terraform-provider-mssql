package mssql

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccLogin_Local_BasicImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLogin(t, "test_import", false, map[string]interface{}{"login_name": "login_import", "password": "valueIsH8kd$¡"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoginExists("mssql_login.test_import"),
				),
			},
			{
				ResourceName:            "mssql_login.test_import",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
				ImportStateIdFunc:       testAccImportStateId("mssql_login.test_import", false),
			},
		},
	})
}
