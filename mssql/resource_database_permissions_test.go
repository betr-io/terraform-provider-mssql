package mssql

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatabasePermissions_Azure_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabasePermissionsDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabasePermissions(t, "test", true, map[string]interface{}{"database": "test", "principal_id": "2"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabasePermissionsExist("mssql_database_permissions.test"),
					resource.TestCheckResourceAttr("mssql_database_permissions.test", "database", "test"),
					resource.TestCheckResourceAttr("mssql_database_permissions.test", "principal_id", "2"),
					// resource.TestCheckResourceAttrSet("mssql_login.basic", "principal_id"),
				),
			},
		},
	})
}

func TestAccDatabasePermissions_Local_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabasePermissionsDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabasePermissions(t, "test", false, map[string]interface{}{"database": "test", "principal_id": "2"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabasePermissionsExist("mssql_database_permissions.test"),
					resource.TestCheckResourceAttr("mssql_database_permissions.test", "database", "test"),
					resource.TestCheckResourceAttr("mssql_database_permissions.test", "principal_id", "2"),
					// resource.TestCheckResourceAttrSet("mssql_login.basic", "principal_id"),
				),
			},
		},
	})
}

func testAccCheckDatabasePermissions(t *testing.T, name string, azure bool, data map[string]interface{}) string {
	text := `resource "mssql_database_permissions" "{{ .name }}" {
             server {
               host = "{{ .host }}"
               {{ if .azure }}azure_login {}{{ else }}login {}{{ end }}
             }
             database       = "{{ .database }}"
             principal_id   = "{{ .principal_id }}"
             permissions = [ "EXECUTE", "UPDATE" ]
           }`
	data["name"] = name
	data["azure"] = azure
	if azure {
		data["host"] = os.Getenv("TF_ACC_SQL_SERVER")
	} else {
		data["host"] = "localhost"
	}
	res, err := templateToString(name, text, data)
	if err != nil {
		t.Fatalf("%s", err)
	}
	return res
}

func testAccCheckDatabasePermissionsDestroy(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "mssql_database_permissions" {
			continue
		}

		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		database := rs.Primary.Attributes["database"]
		principalId := rs.Primary.Attributes["principal_id"]

		var pId int

		if pId, err := strconv.Atoi(principalId); err == nil {
			return fmt.Errorf("pId=%d, type: %T\n", pId, pId)
		}
		permissions, err := connector.GetDatabasePermissions(database, pId)
		if permissions != nil {
			return fmt.Errorf("permissions still exist")
		}
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}
	}
	return nil
}

func testAccCheckDatabasePermissionsExist(resource string, checks ...Check) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Type != "mssql_login" {
			return fmt.Errorf("expected resource of type %s, got %s", "mssql_login", rs.Type)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}
		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		database := rs.Primary.Attributes["database"]
		principalId := rs.Primary.Attributes["principal_id"]
		var pId int

		if pId, err := strconv.Atoi(principalId); err == nil {
			return fmt.Errorf("pId=%d, type: %T\n", pId, pId)
		}
		permissions, err := connector.GetDatabasePermissions(database, pId)
		if permissions == nil {
			return fmt.Errorf("permissions do not exist")
		}
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}

		var actual interface{}
		for _, check := range checks {
			if (check.op == "" || check.op == "==") && check.expected != actual {
				return fmt.Errorf("expected %s == %s, got %s", check.name, check.expected, actual)
			}
			if check.op == "!=" && check.expected == actual {
				return fmt.Errorf("expected %s != %s, got %s", check.name, check.expected, actual)
			}
		}
		return nil
	}
}
