package mssql

import (
  "context"
  "github.com/betr-io/terraform-provider-mssql/mssql/model"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/pkg/errors"
)

type dataLoginConnector interface {
  GetLogin(ctx context.Context, name string) (*model.Login, error)
}

func dataLogin() *schema.Resource {
  return &schema.Resource{
    ReadContext: dataLoginRead,
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
    },
    Timeouts: &schema.ResourceTimeout{
      Default: defaultTimeout,
    },
  }
}

func dataLoginRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "login", "read")
  logger.Debug().Msgf("Read %s", getLoginID(data))

  loginName := data.Get(loginNameProp).(string)

  connector, err := datagetLoginConnector(meta, data)
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
    data.SetId(getLoginID(data))
  }

  return nil
}

func datagetLoginConnector(meta interface{}, data *schema.ResourceData) (dataLoginConnector, error) {
  provider := meta.(model.Provider)
  connector, err := provider.GetConnector(serverProp, data)
  if err != nil {
    return nil, err
  }
  return connector.(dataLoginConnector), nil
}
