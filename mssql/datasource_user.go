package mssql

import (
  "context"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/pkg/errors"
)

func dataSourceUser() *schema.Resource {
  return &schema.Resource{
    ReadContext: dataSourceUserRead,
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
      usernameProp: {
        Type:     schema.TypeString,
        Required: true,
        ForceNew: true,
      },
      objectIdProp: {
        Type:     schema.TypeString,
        Computed: true,
      },
      loginNameProp: {
        Type:     schema.TypeString,
        Computed: true,
      },
      sidStrProp: {
        Type:     schema.TypeString,
        Computed: true,
      },
      authenticationTypeProp: {
        Type:     schema.TypeString,
        Computed: true,
      },
      principalIdProp: {
        Type:     schema.TypeInt,
        Computed: true,
      },
      defaultSchemaProp: {
        Type:     schema.TypeString,
        Computed: true,
      },
      defaultLanguageProp: {
        Type:     schema.TypeString,
        Computed: true,
      },
      rolesProp: {
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

func dataSourceUserRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "user", "read")
  logger.Debug().Msgf("Read %s", getUserID(data))

  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)

  connector, err := getUserConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  user, err := connector.GetUser(ctx, database, username)
  if err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to read user [%s].[%s]", database, username))
  }
  if user == nil {
    logger.Info().Msgf("No user found for [%s].[%s]", database, username)
    data.SetId("")
  } else {
    if err = data.Set(loginNameProp, user.LoginName); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(sidStrProp, user.SIDStr); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(authenticationTypeProp, user.AuthType); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(principalIdProp, user.PrincipalID); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(defaultSchemaProp, user.DefaultSchema); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(defaultLanguageProp, user.DefaultLanguage); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(rolesProp, user.Roles); err != nil {
      return diag.FromErr(err)
    }
    data.SetId(getUserID(data))
  }

  return nil
}
