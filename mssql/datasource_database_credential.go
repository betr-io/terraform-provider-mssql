package mssql

import (
  "context"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/pkg/errors"
)
func datasourceDatabaseCredential() *schema.Resource {
  return &schema.Resource{
    ReadContext:   datasourceDatabaseCredentialRead,
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
      credentialNameProp: {
        Type:     schema.TypeString,
        Required: true,
        ForceNew: true,
      },
      identitynameProp: {
        Type:     schema.TypeString,
        Computed: true,
      },
      principalIdProp: {
        Type:     schema.TypeInt,
        Computed: true,
      },
      credentialIdProp: {
        Type:     schema.TypeInt,
        Computed: true,
      },
    },
    Timeouts: &schema.ResourceTimeout{
      Default: defaultTimeout,
    },
  }
}

func datasourceDatabaseCredentialRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "databasecredential", "read")
  logger.Debug().Msgf("Read %s", data.Id())

  database := data.Get(databaseProp).(string)
  credentialname := data.Get(credentialNameProp).(string)

  connector, err := getDatabaseCredentialConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  scopedcredential, err := connector.GetDatabaseCredential(ctx, database, credentialname)
  if err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to read database scoped credential [%s] on database [%s]", credentialname, database))
  }
  if scopedcredential == nil {
    logger.Info().Msgf("No database scoped credential [%s] found on database [%s]", credentialname, database)
    data.SetId("")
  } else {
    if err = data.Set(credentialNameProp, scopedcredential.CredentialName); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(identitynameProp, scopedcredential.IdentityName); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(principalIdProp, scopedcredential.PrincipalID); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(credentialIdProp, scopedcredential.CredentialID); err != nil {
      return diag.FromErr(err)
    }
    data.SetId(getDatabaseCredentialID(data))
  }

  return nil
}
