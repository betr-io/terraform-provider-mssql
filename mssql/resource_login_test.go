package mssql

import (
  "fmt"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
  "os"
  "testing"
)

func TestAccLogin_Local_Basic(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckLogin(t, "basic", false, map[string]interface{}{"login_name": "login_basic", "password": "valueIsH8kd$¡"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckLoginExists("mssql_login.basic"),
          testAccCheckLoginWorks("mssql_login.basic"),
          resource.TestCheckResourceAttr("mssql_login.basic", "login_name", "login_basic"),
          resource.TestCheckResourceAttr("mssql_login.basic", "password", "valueIsH8kd$¡"),
          resource.TestCheckResourceAttr("mssql_login.basic", "default_database", "master"),
          resource.TestCheckResourceAttr("mssql_login.basic", "default_language", "us_english"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.host", "localhost"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.login.#", "1"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_login.basic", "principal_id"),
        ),
      },
    },
  })
}

func TestAccLogin_Local_Basic_SID(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckLogin(t, "basic", false, map[string]interface{}{"login_name": "login_basic", "password": "valueIsH8kd$¡", "sid": "0xB7BDEF7990D03541BAA2AD73E4FF18E8"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckLoginExists("mssql_login.basic"),
          testAccCheckLoginWorks("mssql_login.basic"),
          resource.TestCheckResourceAttr("mssql_login.basic", "login_name", "login_basic"),
          resource.TestCheckResourceAttr("mssql_login.basic", "password", "valueIsH8kd$¡"),
          resource.TestCheckResourceAttr("mssql_login.basic", "sid", "0xB7BDEF7990D03541BAA2AD73E4FF18E8"),
          resource.TestCheckResourceAttr("mssql_login.basic", "default_database", "master"),
          resource.TestCheckResourceAttr("mssql_login.basic", "default_language", "us_english"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.host", "localhost"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.login.#", "1"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_login.basic", "principal_id"),
        ),
      },
    },
  })
}

func TestAccLogin_Azure_Basic(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckLogin(t, "basic", true, map[string]interface{}{"login_name": "login_basic", "password": "valueIsH8kd$¡"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckLoginExists("mssql_login.basic"),
          resource.TestCheckResourceAttr("mssql_login.basic", "login_name", "login_basic"),
          resource.TestCheckResourceAttr("mssql_login.basic", "password", "valueIsH8kd$¡"),
          resource.TestCheckResourceAttr("mssql_login.basic", "default_database", "master"),
          resource.TestCheckResourceAttr("mssql_login.basic", "default_language", "us_english"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.#", "1"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_login.basic", "principal_id"),
        ),
      },
    },
  })
}

func TestAccLogin_Azure_Basic_SID(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckLogin(t, "basic", true, map[string]interface{}{"login_name": "login_basic", "password": "valueIsH8kd$¡", "sid": "0x01060000000000640000000000000000BAF5FC800B97EF49AC6FD89469C4987F"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckLoginExists("mssql_login.basic"),
          resource.TestCheckResourceAttr("mssql_login.basic", "login_name", "login_basic"),
          resource.TestCheckResourceAttr("mssql_login.basic", "password", "valueIsH8kd$¡"),
          resource.TestCheckResourceAttr("mssql_login.basic", "sid", "0x01060000000000640000000000000000BAF5FC800B97EF49AC6FD89469C4987F"),
          resource.TestCheckResourceAttr("mssql_login.basic", "default_database", "master"),
          resource.TestCheckResourceAttr("mssql_login.basic", "default_language", "us_english"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.#", "1"),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
          resource.TestCheckResourceAttr("mssql_login.basic", "server.0.login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_login.basic", "principal_id"),
        ),
      },
    },
  })
}

func TestAccLogin_Local_UpdateLoginName(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckLogin(t, "test_update", false, map[string]interface{}{"login_name": "login_update_pre", "password": "valueIsH8kd$¡"}),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("mssql_login.test_update", "login_name", "login_update_pre"),
          testAccCheckLoginExists("mssql_login.test_update"),
          testAccCheckLoginWorks("mssql_login.test_update"),
        ),
      },
      {
        Config: testAccCheckLogin(t, "test_update", false, map[string]interface{}{"login_name": "login_update_post", "password": "valueIsH8kd$¡"}),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("mssql_login.test_update", "login_name", "login_update_post"),
          testAccCheckLoginExists("mssql_login.test_update"),
          testAccCheckLoginWorks("mssql_login.test_update"),
        ),
      },
    }})
}

func TestAccLogin_Local_UpdatePassword(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckLogin(t, "test_update", false, map[string]interface{}{"login_name": "login_update", "password": "valueIsH8kd$¡"}),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("mssql_login.test_update", "password", "valueIsH8kd$¡"),
          testAccCheckLoginExists("mssql_login.test_update"),
          testAccCheckLoginWorks("mssql_login.test_update"),
        ),
      },
      {
        Config: testAccCheckLogin(t, "test_update", false, map[string]interface{}{"login_name": "login_update", "password": "otherIsH8kd$¡"}),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("mssql_login.test_update", "password", "otherIsH8kd$¡"),
          testAccCheckLoginExists("mssql_login.test_update"),
          testAccCheckLoginWorks("mssql_login.test_update"),
        ),
      },
    }})
}

func TestAccLogin_Local_UpdateDefaultDatabase(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckLogin(t, "test_update", false, map[string]interface{}{"login_name": "login_update", "password": "valueIsH8kd$¡"}),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("mssql_login.test_update", "default_database", "master"),
          testAccCheckLoginExists("mssql_login.test_update", Check{"default_database", "==", "master"}),
          testAccCheckLoginWorks("mssql_login.test_update"),
        ),
      },
      {
        Config: testAccCheckLogin(t, "test_update", false, map[string]interface{}{"login_name": "login_update", "password": "valueIsH8kd$¡", "default_database": "tempdb"}),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("mssql_login.test_update", "default_database", "tempdb"),
          testAccCheckLoginExists("mssql_login.test_update", Check{"default_database", "==", "tempdb"}),
          testAccCheckLoginWorks("mssql_login.test_update"),
        ),
      },
    }})
}

func TestAccLogin_Local_UpdateDefaultLanguage(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    IsUnitTest:        runLocalAccTests,
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckLogin(t, "test_update", false, map[string]interface{}{"login_name": "login_update", "password": "valueIsH8kd$¡"}),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("mssql_login.test_update", "default_language", "us_english"),
          testAccCheckLoginExists("mssql_login.test_update"),
          testAccCheckLoginWorks("mssql_login.test_update"),
        ),
      },
      {
        Config: testAccCheckLogin(t, "test_update", false, map[string]interface{}{"login_name": "login_update", "password": "valueIsH8kd$¡", "default_language": "russian"}),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("mssql_login.test_update", "default_language", "russian"),
          testAccCheckLoginExists("mssql_login.test_update", Check{"default_language", "==", "russian"}),
          testAccCheckLoginWorks("mssql_login.test_update"),
        ),
      },
    }})
}

func TestAccLogin_Azure_UpdateLoginName(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckLogin(t, "test_update", true, map[string]interface{}{"login_name": "login_update_pre", "password": "valueIsH8kd$¡"}),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("mssql_login.test_update", "login_name", "login_update_pre"),
          testAccCheckLoginExists("mssql_login.test_update"),
        ),
      },
      {
        Config: testAccCheckLogin(t, "test_update", true, map[string]interface{}{"login_name": "login_update_post", "password": "valueIsH8kd$¡"}),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("mssql_login.test_update", "login_name", "login_update_post"),
          testAccCheckLoginExists("mssql_login.test_update"),
        ),
      },
    }})
}

func TestAccLogin_Azure_UpdatePassword(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckLogin(t, "test_update", true, map[string]interface{}{"login_name": "login_update", "password": "valueIsH8kd$¡"}),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("mssql_login.test_update", "password", "valueIsH8kd$¡"),
          testAccCheckLoginExists("mssql_login.test_update"),
        ),
      },
      {
        Config: testAccCheckLogin(t, "test_update", true, map[string]interface{}{"login_name": "login_update", "password": "otherIsH8kd$¡"}),
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr("mssql_login.test_update", "password", "otherIsH8kd$¡"),
          testAccCheckLoginExists("mssql_login.test_update"),
        ),
      },
    }})
}

func testAccCheckLogin(t *testing.T, name string, azure bool, data map[string]interface{}) string {
  text := `resource "mssql_login" "{{ .name }}" {
             server {
               host = "{{ .host }}"
               {{ if .azure }}azure_login {}{{ else }}login {}{{ end }}
             }
             login_name = "{{ .login_name }}"
             password   = "{{ .password }}"
             {{ with .sid }}sid = "{{ . }}"{{ end }}
             {{ with .default_database }}default_database = "{{ . }}"{{ end }}
             {{ with .default_language }}default_language = "{{ . }}"{{ end }}
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

func testAccCheckLoginDestroy(state *terraform.State) error {
  for _, rs := range state.RootModule().Resources {
    if rs.Type != "mssql_login" {
      continue
    }

    connector, err := getTestConnector(rs.Primary.Attributes)
    if err != nil {
      return err
    }

    loginName := rs.Primary.Attributes["login_name"]
    login, err := connector.GetLogin(loginName)
    if login != nil {
      return fmt.Errorf("login still exists")
    }
    if err != nil {
      return fmt.Errorf("expected no error, got %s", err)
    }
  }
  return nil
}

func testAccCheckLoginExists(resource string, checks ...Check) resource.TestCheckFunc {
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

    loginName := rs.Primary.Attributes["login_name"]
    login, err := connector.GetLogin(loginName)
    if login == nil {
      return fmt.Errorf("login does not exist")
    }
    if err != nil {
      return fmt.Errorf("expected no error, got %s", err)
    }

    var actual interface{}
    for _, check := range checks {
      switch check.name {
      case "default_database":
        actual = login.DefaultDatabase
      case "default_language":
        actual = login.DefaultLanguage
      default:
        return fmt.Errorf("unknown property %s", check.name)
      }
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

func testAccCheckLoginWorks(resource string) resource.TestCheckFunc {
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
    connector, err := getTestLoginConnector(rs.Primary.Attributes)
    if err != nil {
      return err
    }
    systemUser, err := connector.GetSystemUser()
    if err != nil {
      return err
    }
    if systemUser != rs.Primary.Attributes[loginNameProp] {
      return fmt.Errorf("expected to log in as [%s], got [%s]", rs.Primary.Attributes[loginNameProp], systemUser)
    }
    return nil
  }
}
