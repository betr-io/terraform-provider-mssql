package mssql

import (
  "context"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/pkg/errors"
)

func dataSourceDatabaseRole() *schema.Resource {
  return &schema.Resource{
    ReadContext: dataSourceDatabaseRoleRead,
    Schema: map[string]*schema.Schema{
      serverProp: {
        Type:     schema.TypeList,
        MaxItems: 1,
        Required: true,
        Elem: &schema.Resource{
          Schema: getServerSchema(serverProp),
        },
      },
      databaseProp: {
        Type:     schema.TypeString,
        Optional: true,
        ForceNew: true,
        Default:  "master",
      },
      roleNameProp: {
        Type:        schema.TypeString,
        Required: true,
        ForceNew: true,
      },
      ownerNameProp: {
        Type:     schema.TypeString,
        Computed: true,
      },
      ownerIdProp: {
        Type:     schema.TypeInt,
        Computed: true,
      },
      principalIdProp: {
        Type:     schema.TypeInt,
        Computed: true,
      },
    },
    Timeouts: &schema.ResourceTimeout{
      Default: defaultTimeout,
    },
  }
}
func dataSourceDatabaseRoleRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "role", "read")
  logger.Debug().Msgf("Read %s", getDatabaseRoleID(data))

  database := data.Get(databaseProp).(string)
  roleName := data.Get(roleNameProp).(string)

  connector, err := getDatabaseRoleConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  role, err := connector.GetDatabaseRole(ctx, database, roleName)
  if err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to get role [%s].[%s]", database, roleName))
  }

  if role == nil {
    logger.Info().Msgf("role [%s].[%s] does not exist", database, roleName)
    data.SetId("")
  } else {
    if err = data.Set(principalIdProp, role.RoleID); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(roleNameProp, role.RoleName); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(ownerNameProp, role.OwnerName); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(ownerIdProp, role.OwnerId); err != nil {
      return diag.FromErr(err)
    }
    data.SetId(getDatabaseRoleID(data))
  }

  logger.Info().Msgf("read role [%s].[%s]", database, roleName)

  return nil
}
