package mssql

import (
  "context"
  "strings"
  "github.com/betr-io/terraform-provider-mssql/mssql/model"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/pkg/errors"
)

func resourceRole() *schema.Resource {
  return &schema.Resource{
    CreateContext: resourceRoleCreate,
    ReadContext:   resourceRoleRead,
    UpdateContext: resourceRoleUpdate,
    DeleteContext: resourceRoleDelete,
    Importer: &schema.ResourceImporter{
      StateContext: resourceRoleImport,
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

type RoleConnector interface {
  CreateRole(ctx context.Context, database string, roleName string) error
  GetRole(ctx context.Context, database, roleName string) (*model.Role, error)
  UpdateRole(ctx context.Context, database string, role *model.Role) error
  DeleteRole(ctx context.Context, database, roleName string) error
}

func resourceRoleRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "role", "read")
  logger.Debug().Msgf("Read %s", getRoleID(data))

  database := data.Get(databaseProp).(string)
  roleName := data.Get(roleNameProp).(string)

  connector, err := getRoleConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  role, err := connector.GetRole(ctx, database, roleName)
  if err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to get role [%s].[%s]", database, roleName))
  }

  if role == nil {
    logger.Info().Msgf("role [%s].[%s] does not exist", database, roleName)
    data.SetId("")
    return nil
  } else {
    if err = data.Set(principalIdProp, role.RoleID); err != nil {
      return diag.FromErr(err)
    }
  }

  logger.Info().Msgf("read role [%s].[%s]", database, roleName)

  return nil
}

func resourceRoleCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "user", "create")
  logger.Debug().Msgf("Create role %s", getRoleID(data))

  database := data.Get(databaseProp).(string)
  roleName := data.Get(roleNameProp).(string)

  connector, err := getRoleConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  if err = connector.CreateRole(ctx, database, roleName); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to create role [%s].[%s]", database, roleName))
  }

  data.SetId(getRoleID(data))

  logger.Info().Msgf("created role [%s].[%s]", database, roleName)

  return resourceRoleRead(ctx, data, meta)
}

func resourceRoleDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "role", "delete")
  logger.Debug().Msgf("Delete %s", getRoleID(data))

  database := data.Get(databaseProp).(string)
  roleName := data.Get(roleNameProp).(string)

  connector, err := getRoleConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  if err = connector.DeleteRole(ctx, database, roleName); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to delete role [%s].[%s]", database, roleName))
  }

  // d.SetId("") is automatically called assuming delete returns no errors, but it is added here for explicitness.
  data.SetId("")

  logger.Info().Msgf("deleted role [%s].[%s]", database, roleName)

  return nil
}

func resourceRoleUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "role", "update")
  logger.Debug().Msgf("Update %s", getRoleID(data))

  database := data.Get(databaseProp).(string)
  roleId := data.Get(principalIdProp).(int)
  roleName := data.Get(roleNameProp).(string)

  connector, err := getRoleConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  role := &model.Role{
    RoleID:   int64(roleId),
    RoleName: roleName,
  }

  if err = connector.UpdateRole(ctx, database, role); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to update role [%s].[%s]", database, roleName))
  }

  data.SetId(getRoleID(data))

  logger.Info().Msgf("updated role [%s].[%s]", database, roleName)

  return resourceRoleRead(ctx, data, meta)
}

func resourceRoleImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

  data.SetId(getRoleID(data))

  database := data.Get(databaseProp).(string)
  role_name := data.Get(roleNameProp).(string)

  connector, err := getRoleConnector(meta, data)
  if err != nil {
    return nil, err
  }

  role, err := connector.GetRole(ctx, database, role_name)
  if err != nil {
    return nil, errors.Wrapf(err, "unable to get role [%s].[%s]", database, role_name)
  }

  if role == nil {
    return nil, errors.Errorf("role [%s].[%s] does not exist", database, role_name)
  }

  if err = data.Set(principalIdProp, role.RoleID); err != nil {
    return nil, err
  }

  return []*schema.ResourceData{data}, nil
}

func getRoleConnector(meta interface{}, data *schema.ResourceData) (RoleConnector, error) {
  provider := meta.(model.Provider)
  connector, err := provider.GetConnector(serverProp, data)
  if err != nil {
    return nil, err
  }
  return connector.(RoleConnector), nil
}
