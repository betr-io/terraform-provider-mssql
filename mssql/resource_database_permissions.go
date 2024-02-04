package mssql

import (
  "context"
  "strings"
  "strconv"
  "github.com/betr-io/terraform-provider-mssql/mssql/model"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/pkg/errors"
)

func resourceDatabasePermissions() *schema.Resource {
  return &schema.Resource{
    CreateContext: resourceDatabasePermissionsCreate,
    ReadContext:   resourceDatabasePermissionsRead,
    DeleteContext: resourceDatabasePermissionDelete,
    Importer: &schema.ResourceImporter{
      StateContext: resourceDatabasePermissionImport,
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
      principalIdProp: {
        Type:     schema.TypeInt,
        Required: true,
        ForceNew: true,
      },
      permissionsProp: {
        Type:     schema.TypeSet,
        Required: true,
        ForceNew: true,
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

type DatabasePermissionsConnector interface {
  CreateDatabasePermissions(ctx context.Context, dbPermission *model.DatabasePermissions) error
  GetDatabasePermissions(ctx context.Context, database string, principalId int) (*model.DatabasePermissions, error)
  DeleteDatabasePermissions(ctx context.Context, database string, principalId int) error
}

func resourceDatabasePermissionsCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "databasepermissions", "create")
  logger.Debug().Msgf("Create %s", getDatabasePermissionsID(data))

  database := data.Get(databaseProp).(string)
  principalId := data.Get(principalIdProp).(int)
  permissions := data.Get(permissionsProp).(*schema.Set).List()

  connector, err := getDatabasePermissionsConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  dbPermissionModel := &model.DatabasePermissions{
    DatabaseName: database,
    PrincipalID:  principalId,
    Permissions:  toStringSlice(permissions),
  }
  if err = connector.CreateDatabasePermissions(ctx, dbPermissionModel); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to create database permissions [%s] on database [%s] for principal [%d]", permissions, database, principalId))
  }

  data.SetId(getDatabasePermissionsID(data))

  logger.Info().Msgf("created database permissions [%s] on database [%s] for principal [%d]", permissions, database, principalId)

  return resourceDatabasePermissionsRead(ctx, data, meta)
}

func resourceDatabasePermissionsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "databasepermissions", "read")
  logger.Debug().Msgf("Read %s", data.Id())

  database := data.Get(databaseProp).(string)
  principalId := data.Get(principalIdProp).(int)

  connector, err := getDatabasePermissionsConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  permissions, err := connector.GetDatabasePermissions(ctx, database, principalId)
  if err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to read permissions for principal [%d] on database [%s]", principalId, database))
  }
  if permissions == nil {
    logger.Info().Msgf("No permissions found for principal [%d] on database [%s]", principalId, database)
    data.SetId("")
  } else {
    if err = data.Set(databaseProp, permissions.DatabaseName); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(principalIdProp, permissions.PrincipalID); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(permissionsProp, permissions.Permissions); err != nil {
      return diag.FromErr(err)
    }
  }

  return nil
}

func resourceDatabasePermissionDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "databasepermissions", "delete")
  logger.Debug().Msgf("Delete %s", data.Id())

  database := data.Get(databaseProp).(string)
  principalId := data.Get(principalIdProp).(int)

  connector, err := getDatabasePermissionsConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  if err = connector.DeleteDatabasePermissions(ctx, database, principalId); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to delete permissions for principal [%d] on database [%s]", principalId, database))
  }

  logger.Info().Msgf("deleted permissions for principal [%d] on database [%s]", principalId, database)

  // d.SetId("") is automatically called assuming delete returns no errors, but it is added here for explicitness.
  data.SetId("")

  return nil
}

func resourceDatabasePermissionImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
  logger := loggerFromMeta(meta, "databasepermissions", "import")
  logger.Debug().Msgf("Import %s", data.Id())

  server, u, err := serverFromId(data.Id())
  if err != nil {
    return nil, err
  }
  if err = data.Set(serverProp, server); err != nil {
    return nil, err
  }

  parts := strings.Split(u.Path, "/")
  if len(parts) != 4 {
    return nil, errors.New("invalid ID")
  }

  if err = data.Set(databaseProp, parts[1]); err != nil {
    return nil, err
  }

  pId, err := strconv.ParseInt(parts[2], 10, 0)
  if err != nil {
    return nil, err
  }
  if err = data.Set(principalIdProp, pId); err != nil {
    return nil, err
  }

  database := data.Get(databaseProp).(string)
  principalId := data.Get(principalIdProp).(int)

  data.SetId(getDatabasePermissionsID(data))

  connector, err := getDatabasePermissionsConnector(meta, data)
  if err != nil {
    return nil, err
  }

  permissions, err := connector.GetDatabasePermissions(ctx, database, principalId)
  if err != nil {
    return nil, errors.Wrapf(err, "unable to import permissions for principalId [%d] on database [%s]", principalId, database)
  }

  if permissions == nil {
    return nil, errors.Errorf("no permissions found for principalId [%d] on database [%s] for import", principalId, database)
  }

  if err = data.Set(databaseProp, permissions.DatabaseName); err != nil {
    return nil, err
  }
  if err = data.Set(principalIdProp, permissions.PrincipalID); err != nil {
    return nil, err
  }
  if err = data.Set(permissionsProp, permissions.Permissions); err != nil {
    return nil, err
  }

  return []*schema.ResourceData{data}, nil
}

func getDatabasePermissionsConnector(meta interface{}, data *schema.ResourceData) (DatabasePermissionsConnector, error) {
  provider := meta.(model.Provider)
  connector, err := provider.GetConnector(serverProp, data)
  if err != nil {
    return nil, err
  }
  return connector.(DatabasePermissionsConnector), nil
}
