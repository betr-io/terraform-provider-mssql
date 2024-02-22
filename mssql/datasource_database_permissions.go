package mssql

import (
  "context"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/pkg/errors"
)

func dataSourceDatabasePermissions() *schema.Resource {
  return &schema.Resource{
    ReadContext:   dataSourceDatabasePermissionsRead,
    Schema: map[string]*schema.Schema{
      serverProp: {
        Type:     schema.TypeList,
        MaxItems: 1,
        Required: true,
        ForceNew: true,
        Elem: &schema.Resource{
          Schema: getServerSchema(serverProp),
        },
      },
      databaseProp: {
        Type:     schema.TypeString,
        Required: true,
        ForceNew: true,
      },
      usernameProp: {
        Type:     schema.TypeString,
        Required: true,
        ForceNew: true,
      },
      principalIdProp: {
        Type:     schema.TypeInt,
				Computed: true,
      },
      permissionsProp: {
        Type:     schema.TypeSet,
        Computed: true,
        Elem: &schema.Schema{
          Type: schema.TypeString,
        },
      },
    },
    Timeouts: &schema.ResourceTimeout{
      Default: defaultTimeout,
    },
  }
}
func dataSourceDatabasePermissionsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "databasepermissions", "read")
  logger.Debug().Msgf("Read %s", getDatabasePermissionsID(data))

  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)

  connector, err := getDatabasePermissionsConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  permissions, err := connector.GetDatabasePermissions(ctx, database, username)
  if err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to read permissions for user [%s] on database [%s]", username, database))
  }
  if permissions == nil {
    logger.Info().Msgf("No permissions found for user [%s] on database [%s]", username, database)
    data.SetId("")
  } else {
    if err = data.Set(databaseProp, permissions.DatabaseName); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(usernameProp, permissions.UserName); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(principalIdProp, permissions.PrincipalID); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(permissionsProp, permissions.Permissions); err != nil {
      return diag.FromErr(err)
    }
		data.SetId(getDatabasePermissionsID(data))
  }

  return nil
}
