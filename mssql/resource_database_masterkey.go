package mssql

import (
  "context"
  "github.com/betr-io/terraform-provider-mssql/mssql/model"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/pkg/errors"
)

const keynameProp = "key_name"
const keyguidProp   = "key_guid"
const symmetrickeyidProp = "symmetric_key_id"
const keylengthProp = "key_length"
const keyalgorithmProp = "key_algorithm"
const algorithmdescProp = "algorithm_desc"

func resourceDatabaseMasterkey() *schema.Resource {
  return &schema.Resource{
    CreateContext: resourceDatabaseMasterkeyCreate,
    ReadContext:   resourceDatabaseMasterkeyRead,
    UpdateContext: resourceDatabaseMasterkeyUpdate,
    DeleteContext: resourceDatabaseMasterkeyDelete,
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
      passwordProp: {
        Type:     schema.TypeString,
        Required: true,
        Sensitive: true,
      },
      keynameProp: {
        Type:     schema.TypeString,
        Computed: true,
      },
      keyguidProp: {
        Type:     schema.TypeString,
        Computed: true,
      },
      principalIdProp: {
        Type:     schema.TypeInt,
        Computed: true,
      },
      symmetrickeyidProp: {
        Type:     schema.TypeInt,
        Computed: true,
      },
      keylengthProp: {
        Type:     schema.TypeInt,
        Computed: true,
      },
      keyalgorithmProp: {
        Type:     schema.TypeString,
        Computed: true,
      },
      algorithmdescProp: {
        Type:     schema.TypeString,
        Computed: true,
      },
    },
    Timeouts: &schema.ResourceTimeout{
      Default: defaultTimeout,
    },
  }
}

type DatabaseMasterkeyConnector interface {
  CreateDatabaseMasterkey(ctx context.Context, database, password string) error
  GetDatabaseMasterkey(ctx context.Context, database string) (*model.DatabaseMasterkey, error)
  UpdateDatabaseMasterkey(ctx context.Context, database, password string) error
  DeleteDatabaseMasterkey(ctx context.Context, database string) error
}

func resourceDatabaseMasterkeyCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "databasemasterkey", "create")
  logger.Debug().Msgf("Create %s", getDatabaseMasterkeyID(data))

  database := data.Get(databaseProp).(string)
  password := data.Get(passwordProp).(string)

  connector, err := getDatabaseMasterkeyConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  if err = connector.CreateDatabaseMasterkey(ctx, database, password); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to create database master key on database [%s]", database))
  }

  data.SetId(getDatabaseMasterkeyID(data))

  logger.Info().Msgf("created database master key on database [%s]", database)

  return resourceDatabaseMasterkeyRead(ctx, data, meta)
}

func resourceDatabaseMasterkeyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "databasemasterkey", "read")
  logger.Debug().Msgf("Read %s", data.Id())

  database := data.Get(databaseProp).(string)

  connector, err := getDatabaseMasterkeyConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  masterkey, err := connector.GetDatabaseMasterkey(ctx, database)
  if err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to read database master key on database [%s]", database))
  }

  if masterkey == nil {
    logger.Info().Msgf("No database master key found on database [%s]", database)
    data.SetId("")
  } else {
    if err = data.Set(keynameProp, masterkey.KeyName); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(principalIdProp, masterkey.PrincipalID); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(symmetrickeyidProp, masterkey.SymmetricKeyID); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(keylengthProp, masterkey.KeyLength); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(keyalgorithmProp, masterkey.KeyAlgorithm); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(algorithmdescProp, masterkey.AlgorithmDesc); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(keyguidProp, masterkey.KeyGuid); err != nil {
      return diag.FromErr(err)
    }
  }

  return nil
}

func resourceDatabaseMasterkeyUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "databasemasterkey", "update")
  logger.Debug().Msgf("Update %s", getDatabaseMasterkeyID(data))

  database := data.Get(databaseProp).(string)
  password := data.Get(passwordProp).(string)

  connector, err := getDatabaseMasterkeyConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  if err = connector.UpdateDatabaseMasterkey(ctx, database, password); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to update database key on database [%s]", database))
  }

  data.SetId(getDatabaseMasterkeyID(data))

  logger.Info().Msgf("updated database master key on database [%s]", database)

  return resourceDatabaseMasterkeyRead(ctx, data, meta)
}

func resourceDatabaseMasterkeyDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "databasemasterkey", "delete")
  logger.Debug().Msgf("Delete %s", data.Id())

  database := data.Get(databaseProp).(string)

  connector, err := getDatabaseMasterkeyConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  if err = connector.DeleteDatabaseMasterkey(ctx, database); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to delete database master key on database [%s]", database))
  }

  logger.Info().Msgf("deleted database master key on database [%s]", database)

  // d.SetId("") is automatically called assuming delete returns no errors, but it is added here for explicitness.
  data.SetId("")

  return nil
}

func getDatabaseMasterkeyConnector(meta interface{}, data *schema.ResourceData) (DatabaseMasterkeyConnector, error) {
  provider := meta.(model.Provider)
  connector, err := provider.GetConnector(serverProp, data)
  if err != nil {
    return nil, err
  }
  return connector.(DatabaseMasterkeyConnector), nil
}
