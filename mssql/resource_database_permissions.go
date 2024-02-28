package mssql

import (
  "context"
  "strings"
  "github.com/betr-io/terraform-provider-mssql/mssql/model"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/pkg/errors"
)

func resourceDatabasePermissions() *schema.Resource {
  return &schema.Resource{
    CreateContext: resourceDatabasePermissionsCreate,
    ReadContext:   resourceDatabasePermissionsRead,
    UpdateContext: resourceDatabasePermissionUpdate,
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
        Required: true,
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
  GetDatabasePermissions(ctx context.Context, database string, username string) (*model.DatabasePermissions, error)
  UpdateDatabasePermissions(ctx context.Context, dbPermission *model.DatabasePermissions) error
  DeleteDatabasePermissions(ctx context.Context, dbPermission *model.DatabasePermissions) error
}

func resourceDatabasePermissionsCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "databasepermissions", "create")
  logger.Debug().Msgf("Create %s", getDatabasePermissionsID(data))

  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)
  permissions := data.Get(permissionsProp).(*schema.Set).List()

  connector, err := getDatabasePermissionsConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  dbPermissionModel := &model.DatabasePermissions{
    DatabaseName: database,
    UserName:  username,
    Permissions:  toStringSlice(permissions),
  }
  if err = connector.CreateDatabasePermissions(ctx, dbPermissionModel); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to create database permissions [%s] on database [%s] for user [%s]", permissions, database, username))
  }

  data.SetId(getDatabasePermissionsID(data))

  logger.Info().Msgf("created database permissions [%s] on database [%s] for user [%s]", permissions, database, username)

  return resourceDatabasePermissionsRead(ctx, data, meta)
}

func resourceDatabasePermissionsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "databasepermissions", "read")
  logger.Debug().Msgf("Read %s", data.Id())

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
  }

  return nil
}

func resourceDatabasePermissionDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "databasepermissions", "delete")
  logger.Debug().Msgf("Delete %s", data.Id())

  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)
  permissions := data.Get(permissionsProp).(*schema.Set).List()

  connector, err := getDatabasePermissionsConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  dbPermissionModel := &model.DatabasePermissions{
    DatabaseName: database,
    UserName:  username,
    Permissions:  toStringSlice(permissions),
  }
  if err = connector.DeleteDatabasePermissions(ctx, dbPermissionModel); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to delete permissions for user [%s] on database [%s]", username, database))
  }

  logger.Info().Msgf("deleted permissions for user [%s] on database [%s]", username, database)

  // d.SetId("") is automatically called assuming delete returns no errors, but it is added here for explicitness.
  data.SetId("")

  return nil
}

func resourceDatabasePermissionUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "databasepermissions", "update")
  logger.Debug().Msgf("Update %s", data.Id())

  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)
  permissions := data.Get(permissionsProp).(*schema.Set).List()

  connector, err := getDatabasePermissionsConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  dbPermissionModel := &model.DatabasePermissions{
    DatabaseName: database,
    UserName:  username,
    Permissions:  toStringSlice(permissions),
  }
  if err = connector.UpdateDatabasePermissions(ctx, dbPermissionModel); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to update permissions for user [%s] on database [%s]", username, database))
  }

  data.SetId(getDatabasePermissionsID(data))

  logger.Info().Msgf("updated permissions for user [%s] on database [%s]", username, database)

  return resourceDatabasePermissionsRead(ctx, data, meta)
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
  if err = data.Set(usernameProp, parts[2]); err != nil {
    return nil, err
  }

  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)

  data.SetId(getDatabasePermissionsID(data))

  connector, err := getDatabasePermissionsConnector(meta, data)
  if err != nil {
    return nil, err
  }

  permissions, err := connector.GetDatabasePermissions(ctx, database, username)
  if err != nil {
    return nil, errors.Wrapf(err, "unable to import permissions for user [%s] on database [%s]", username, database)
  }

  if permissions == nil {
    return nil, errors.Errorf("no permissions found for user [%s] on database [%s] for import", username, database)
  }

  if err = data.Set(databaseProp, permissions.DatabaseName); err != nil {
    return nil, err
  }
  if err = data.Set(usernameProp, permissions.UserName); err != nil {
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
