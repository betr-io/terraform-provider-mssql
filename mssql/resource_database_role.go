package mssql

import (
  "context"
  "strings"
  "github.com/betr-io/terraform-provider-mssql/mssql/model"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/pkg/errors"
)

const ownerIdProp   = "owning_principal_id"
const ownerNameProp = "owner_name"
const defaultOwnerNameDefault = "dbo"

func resourceDatabaseRole() *schema.Resource {
  return &schema.Resource{
    CreateContext: resourceDatabaseRoleCreate,
    ReadContext:   resourceDatabaseRoleRead,
    UpdateContext: resourceDatabaseRoleUpdate,
    DeleteContext: resourceDatabaseRoleDelete,
    Importer: &schema.ResourceImporter{
      StateContext: resourceDatabaseRoleImport,
    },
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
        Required:    true,
        Description: "The name of the role",
      },
      ownerNameProp: {
        Type:     schema.TypeString,
        Optional: true,
        Default:  defaultOwnerNameDefault,
        DiffSuppressFunc: func(k, old, new string, data *schema.ResourceData) bool {
          return (old == "" && new == defaultOwnerNameDefault) || (old == defaultOwnerNameDefault && new == "")
        },
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

type DatabaseRoleConnector interface {
  CreateDatabaseRole(ctx context.Context, database string, roleName string, ownerName string) error
  GetDatabaseRole(ctx context.Context, database, roleName string) (*model.DatabaseRole, error)
  UpdateDatabaseRole(ctx context.Context, database string, roleId int, roleName string, ownerName string) error
  DeleteDatabaseRole(ctx context.Context, database, roleName string) error
}

func resourceDatabaseRoleRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
  }

  logger.Info().Msgf("read role [%s].[%s]", database, roleName)

  return nil
}

func resourceDatabaseRoleCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "role", "create")
  logger.Debug().Msgf("Create %s", getDatabaseRoleID(data))

  database := data.Get(databaseProp).(string)
  roleName := data.Get(roleNameProp).(string)
  ownerName := data.Get(ownerNameProp).(string)

  connector, err := getDatabaseRoleConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  if err = connector.CreateDatabaseRole(ctx, database, roleName, ownerName); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to create role [%s].[%s]", database, roleName))
  }

  data.SetId(getDatabaseRoleID(data))

  logger.Info().Msgf("created role [%s].[%s]", database, roleName)

  return resourceDatabaseRoleRead(ctx, data, meta)
}

func resourceDatabaseRoleDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "role", "delete")
  logger.Debug().Msgf("Delete %s", getDatabaseRoleID(data))

  database := data.Get(databaseProp).(string)
  roleName := data.Get(roleNameProp).(string)

  connector, err := getDatabaseRoleConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  if err = connector.DeleteDatabaseRole(ctx, database, roleName); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to delete role [%s].[%s]", database, roleName))
  }

  // d.SetId("") is automatically called assuming delete returns no errors, but it is added here for explicitness.
  data.SetId("")

  logger.Info().Msgf("deleted role [%s].[%s]", database, roleName)

  return nil
}

func resourceDatabaseRoleUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "role", "update")
  logger.Debug().Msgf("Update %s", getDatabaseRoleID(data))

  database := data.Get(databaseProp).(string)
  roleId := data.Get(principalIdProp).(int)
  roleName := data.Get(roleNameProp).(string)
  ownerName := data.Get(ownerNameProp).(string)

  connector, err := getDatabaseRoleConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  if err = connector.UpdateDatabaseRole(ctx, database, roleId, roleName, ownerName); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to update role [%s].[%s]", database, roleName))
  }

  data.SetId(getDatabaseRoleID(data))

  logger.Info().Msgf("updated role [%s].[%s]", database, roleName)

  return resourceDatabaseRoleRead(ctx, data, meta)
}

func resourceDatabaseRoleImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
  logger := loggerFromMeta(meta, "role", "import")
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
  if err = data.Set(roleNameProp, parts[2]); err != nil {
    return nil, err
  }

  data.SetId(getDatabaseRoleID(data))

  database := data.Get(databaseProp).(string)
  role_name := data.Get(roleNameProp).(string)

  connector, err := getDatabaseRoleConnector(meta, data)
  if err != nil {
    return nil, err
  }

  role, err := connector.GetDatabaseRole(ctx, database, role_name)
  if err != nil {
    return nil, errors.Wrapf(err, "unable to get role [%s].[%s]", database, role_name)
  }

  if role == nil {
    return nil, errors.Errorf("role [%s].[%s] does not exist", database, role_name)
  }

  if err = data.Set(principalIdProp, role.RoleID); err != nil {
    return nil, err
  }
  if err = data.Set(ownerNameProp, role.OwnerName); err != nil {
    return nil, err
  }

  return []*schema.ResourceData{data}, nil
}

func getDatabaseRoleConnector(meta interface{}, data *schema.ResourceData) (DatabaseRoleConnector, error) {
  provider := meta.(model.Provider)
  connector, err := provider.GetConnector(serverProp, data)
  if err != nil {
    return nil, err
  }
  return connector.(DatabaseRoleConnector), nil
}
