package mssql

import (
  "context"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/pkg/errors"
)

func dataSourceDatabaseSchema() *schema.Resource {
  return &schema.Resource{
    ReadContext:   dataSourceDatabaseSchemaRead,
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
      schemaNameProp: {
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
      schemaIdProp: {
        Type:     schema.TypeInt,
        Computed: true,
      },
    },
    Timeouts: &schema.ResourceTimeout{
      Default: defaultTimeout,
    },
  }
}

func dataSourceDatabaseSchemaRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "schema", "read")
  logger.Debug().Msgf("Read %s", getDatabaseSchemaID(data))

  database := data.Get(databaseProp).(string)
  schemaName := data.Get(schemaNameProp).(string)

  connector, err := getDatabaseSchemaConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  sqlschema, err := connector.GetDatabaseSchema(ctx, database, schemaName)
  if err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to get schema [%s].[%s]", database, schemaName))
  }

  if sqlschema == nil {
    logger.Info().Msgf("schema [%s].[%s] does not exist", database, schemaName)
    data.SetId("")
  } else {
    if err = data.Set(schemaIdProp, sqlschema.SchemaID); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(schemaNameProp, sqlschema.SchemaName); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(ownerNameProp, sqlschema.OwnerName); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(ownerIdProp, sqlschema.OwnerId); err != nil {
      return diag.FromErr(err)
    }
    data.SetId(getDatabaseSchemaID(data))
  }

  return nil
}
