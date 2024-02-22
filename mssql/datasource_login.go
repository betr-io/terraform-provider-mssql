package mssql

import (
  "context"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/pkg/errors"
)

func dataSourceLogin() *schema.Resource {
  return &schema.Resource{
    ReadContext: dataSourceLoginRead,
    Schema: map[string]*schema.Schema{
      serverProp: {
        Type:     schema.TypeList,
        MaxItems: 1,
        Required: true,
        Elem: &schema.Resource{
          Schema: getServerSchema(serverProp),
        },
      },
      loginNameProp: {
        Type:     schema.TypeString,
        Required: true,
        ForceNew: true,
      },
      sidStrProp: {
        Type:     schema.TypeString,
        Computed: true,
      },
      principalIdProp: {
        Type:     schema.TypeInt,
        Computed: true,
      },
      defaultDatabaseProp: {
        Type:     schema.TypeString,
        Computed: true,
      },
      defaultLanguageProp: {
        Type:     schema.TypeString,
        Computed: true,
      },
    },
    Timeouts: &schema.ResourceTimeout{
      Default: defaultTimeout,
    },
  }
}

func dataSourceLoginRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "login", "read")
  logger.Debug().Msgf("Read %s", getLoginID(data))

  loginName := data.Get(loginNameProp).(string)

  connector, err := getLoginConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  login, err := connector.GetLogin(ctx, loginName)
  if err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to read login [%s]", loginName))
  }
  if login == nil {
    logger.Info().Msgf("No login found for [%s]", loginName)
    data.SetId("")
  } else {
    if err = data.Set(principalIdProp, login.PrincipalID); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(sidStrProp, login.SIDStr); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(defaultDatabaseProp, login.DefaultDatabase); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(defaultLanguageProp, login.DefaultLanguage); err != nil {
      return diag.FromErr(err)
    }
    data.SetId(getLoginID(data))
  }

  return nil
}
