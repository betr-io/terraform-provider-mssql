package mssql

import (
  "context"
  "strings"
  "github.com/betr-io/terraform-provider-mssql/mssql/model"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/pkg/errors"
)

const credentialNameProp = "credential_name"
const identitynameProp   = "identity_name"
const secretProp         = "secret"
const credentialIdProp   = "credential_id"

func resourceDatabaseCredential() *schema.Resource {
  return &schema.Resource{
    CreateContext: resourceDatabaseCredentialCreate,
    ReadContext:   resourceDatabaseCredentialRead,
    UpdateContext: resourceDatabaseCredentialUpdate,
    DeleteContext: resourceDatabaseCredentialDelete,
    Importer: &schema.ResourceImporter{
      StateContext: resourceDatabaseCredentialImport,
    },
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
        ValidateFunc: SQLIdentifierName,
      },
      identitynameProp: {
        Type:     schema.TypeString,
        Required: true,
				ValidateFunc: SQLIdentifierName,
      },
      secretProp: {
        Type:     schema.TypeString,
        Optional: true,
        Sensitive: true,
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

type DatabaseCredentialConnector interface {
  CreateDatabaseCredential(ctx context.Context, database, credentialname, identityname, secret string) error
  GetDatabaseCredential(ctx context.Context, database, credentialname string) (*model.DatabaseCredential, error)
  UpdateDatabaseCredential(ctx context.Context, database, credentialname, identityname, secret string) error
  DeleteDatabaseCredential(ctx context.Context, database, credentialname string) error
}

func resourceDatabaseCredentialCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "databasecredential", "create")
  logger.Debug().Msgf("Create %s", getDatabaseCredentialID(data))

  database := data.Get(databaseProp).(string)
  credentialname := data.Get(credentialNameProp).(string)
  identityname := data.Get(identitynameProp).(string)
  secret := data.Get(secretProp).(string)

  connector, err := getDatabaseCredentialConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  if err = connector.CreateDatabaseCredential(ctx, database, credentialname, identityname, secret); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to create database scoped credential [%s] on database [%s]", credentialname, database))
  }

  data.SetId(getDatabaseCredentialID(data))

  logger.Info().Msgf("created database scoped credential [%s] on database [%s]", credentialname, database)

  return resourceDatabaseCredentialRead(ctx, data, meta)
}

func resourceDatabaseCredentialRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
  }

  return nil
}

func resourceDatabaseCredentialUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "databasecredential", "update")
  logger.Debug().Msgf("Update %s", getDatabaseCredentialID(data))

  database := data.Get(databaseProp).(string)
  credentialname := data.Get(credentialNameProp).(string)
  identityname := data.Get(identitynameProp).(string)
  secret := data.Get(secretProp).(string)

  connector, err := getDatabaseCredentialConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  if err = connector.UpdateDatabaseCredential(ctx, database, credentialname, identityname, secret); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to update database scoped credential [%s] on database [%s]", credentialname, database))
  }

  data.SetId(getDatabaseCredentialID(data))

  logger.Info().Msgf("updated database scoped credential [%s] on database [%s]", credentialname, database)

  return resourceDatabaseCredentialRead(ctx, data, meta)
}

func resourceDatabaseCredentialDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "databasecredential", "delete")
  logger.Debug().Msgf("Delete %s", data.Id())

  database := data.Get(databaseProp).(string)
  credentialname := data.Get(credentialNameProp).(string)

  connector, err := getDatabaseCredentialConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  if err = connector.DeleteDatabaseCredential(ctx, database, credentialname); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to delete database scoped credential [%s] on database [%s]", credentialname, database))
  }

  logger.Info().Msgf("deleted database scoped credential [%s] on database [%s]", credentialname, database)

  // d.SetId("") is automatically called assuming delete returns no errors, but it is added here for explicitness.
  data.SetId("")

  return nil
}

func resourceDatabaseCredentialImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
  logger := loggerFromMeta(meta, "databasecredential", "import")
  logger.Debug().Msgf("Import %s", data.Id())

  server, u, err := serverFromId(data.Id())
  if err != nil {
    return nil, err
  }
  if err := data.Set(serverProp, server); err != nil {
    return nil, err
  }

  parts := strings.Split(u.Path, "/")
  if len(parts) != 3 {
    return nil, errors.New("invalid ID")
  }
  if err = data.Set(databaseProp, parts[1]); err != nil {
    return nil, err
  }
  if err = data.Set(credentialNameProp, parts[2]); err != nil {
    return nil, err
  }

  data.SetId(getDatabaseCredentialID(data))

  database := data.Get(databaseProp).(string)
  credentialname := data.Get(credentialNameProp).(string)

  connector, err := getDatabaseCredentialConnector(meta, data)
  if err != nil {
    return nil, err
  }

  scopedcredential, err := connector.GetDatabaseCredential(ctx, database, credentialname)
  if err != nil {
    return nil, errors.Wrapf(err, "unable to get database scoped credential [%s] on database [%s]",credentialname, database)
  }

  if scopedcredential == nil {
    return nil, errors.Errorf("database scoped credential [%s] on database [%s] does not exist",credentialname, database)
  }
  if err = data.Set(credentialNameProp, scopedcredential.CredentialName); err != nil {
    return nil, err
  }
  if err = data.Set(identitynameProp, scopedcredential.IdentityName); err != nil {
    return nil, err
  }
  if err = data.Set(principalIdProp, scopedcredential.PrincipalID); err != nil {
    return nil, err
  }
  if err = data.Set(credentialIdProp, scopedcredential.CredentialID); err != nil {
    return nil, err
  }

  return []*schema.ResourceData{data}, nil
}

func getDatabaseCredentialConnector(meta interface{}, data *schema.ResourceData) (DatabaseCredentialConnector, error) {
  provider := meta.(model.Provider)
  connector, err := provider.GetConnector(serverProp, data)
  if err != nil {
    return nil, err
  }
  return connector.(DatabaseCredentialConnector), nil
}
