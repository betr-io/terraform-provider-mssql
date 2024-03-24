package mssql

import (
  "fmt"
  "regexp"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/rs/zerolog"
  "github.com/betr-io/terraform-provider-mssql/mssql/model"
)

func getLoginID(data *schema.ResourceData) string {
  host := data.Get(serverProp + ".0.host").(string)
  port := data.Get(serverProp + ".0.port").(string)
  loginName := data.Get(loginNameProp).(string)
  return fmt.Sprintf("sqlserver://%s:%s/%s", host, port, loginName)
}

func getUserID(data *schema.ResourceData) string {
  host := data.Get(serverProp + ".0.host").(string)
  port := data.Get(serverProp + ".0.port").(string)
  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)
  return fmt.Sprintf("sqlserver://%s:%s/%s/%s", host, port, database, username)
}

func getDatabasePermissionsID(data *schema.ResourceData) string {
  host := data.Get(serverProp + ".0.host").(string)
  port := data.Get(serverProp + ".0.port").(string)
  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)
  return fmt.Sprintf("sqlserver://%s:%s/%s/%s/%s", host, port, database, username, "permissions")
}

func getDatabaseRoleID(data *schema.ResourceData) string {
  host := data.Get(serverProp + ".0.host").(string)
  port := data.Get(serverProp + ".0.port").(string)
  database := data.Get(databaseProp).(string)
  roleName := data.Get(roleNameProp).(string)
  return fmt.Sprintf("sqlserver://%s:%s/%s/%s", host, port, database, roleName)
}

func getDatabaseSchemaID(data *schema.ResourceData) string {
  host := data.Get(serverProp + ".0.host").(string)
  port := data.Get(serverProp + ".0.port").(string)
  database := data.Get(databaseProp).(string)
  schemaName := data.Get(schemaNameProp).(string)
  return fmt.Sprintf("sqlserver://%s:%s/%s/%s", host, port, database, schemaName)
}

func getDatabaseCredentialID(data *schema.ResourceData) string {
  host := data.Get(serverProp + ".0.host").(string)
  port := data.Get(serverProp + ".0.port").(string)
  database := data.Get(databaseProp).(string)
  credentialname := data.Get(credentialNameProp).(string)
  return fmt.Sprintf("sqlserver://%s:%s/%s/%s", host, port, database, credentialname)
}

func getDatabaseMasterkeyID(data *schema.ResourceData) string {
  host := data.Get(serverProp + ".0.host").(string)
  port := data.Get(serverProp + ".0.port").(string)
  database := data.Get(databaseProp).(string)
  return fmt.Sprintf("sqlserver://%s:%s/%s/%s", host, port, database, "masterkey")
}

func loggerFromMeta(meta interface{}, resource, function string) zerolog.Logger {
  return meta.(model.Provider).ResourceLogger(resource, function)
}

func toStringSlice(values []interface{}) []string {
  result := make([]string, len(values))
  for i, v := range values {
    result[i] = v.(string)
  }
  return result
}

func SQLIdentifier(v interface{}, k string) (warns []string, errors []error) {
  value := v.(string)
  if match, _ := regexp.Match("^[a-zA-Z_@#\\.][a-zA-Z\\.\\d@$#_-]*$", []byte(value)); !match {
    errors = append(errors, fmt.Errorf(
      "invalid SQL identifier. SQL identifier allows letters, digits, @, $, #, . or _, start with letter, _, @ or # .Got %q", value))
  }

  if 1 > len(value) {
    errors = append(errors, fmt.Errorf("%q cannot be less than 1 character: %q", k, value))
  }

  if len(value) > 128 {
    errors = append(errors, fmt.Errorf("%q cannot be longer than 128 characters: %q %d", k, value, len(value)))
  }

  return
}

func SQLIdentifierName(v interface{}, k string) (warns []string, errors []error) {
  value := v.(string)
  if (!regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(value)) && (!regexp.MustCompile("SHARED ACCESS SIGNATURE").MatchString(value)) {
    errors = append(errors, fmt.Errorf(
      "any combination of alphanumeric characters including hyphens and underscores are allowed in %q: %q", k, value))
  }

  if 1 > len(value) {
    errors = append(errors, fmt.Errorf("%q cannot be less than 1 character: %q", k, value))
  }

  if len(value) > 128 {
    errors = append(errors, fmt.Errorf("%q cannot be longer than 128 characters: %q %d", k, value, len(value)))
  }

  return
}
