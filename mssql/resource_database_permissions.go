package mssql

import (
	"context"

	"github.com/betr-io/terraform-provider-mssql/mssql/model"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func resourceDatabasePermissions() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabasePermissionsCreate,
		ReadContext:   resourceDatabasePermissionsRead,
		// UpdateContext: resourceDatabasePermissionUpdate,
		DeleteContext: resourceDatabasePermissionDelete,
		// Importer: &schema.ResourceImporter{
		// 	StateContext: resourceUserImport,
		// },
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

	// TODO: @favoretti handle updates gracefully instead of recreating all permissions
	// UpdateDatabasePermissions(ctx context.Context, dbPermission *model.DatabasePermissions) error

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

// func resourceDatabasePermissionUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	logger := loggerFromMeta(meta, "database_permissions", "update")
// 	logger.Debug().Msgf("Update %s", data.Id())
//
// 	database := data.Get(databaseProp).(string)
// 	principalId := data.Get(principalIdProp).(int)
// 	permissions := data.Get(permissionsProp).(*schema.Set).List()
//
// 	connector, err := getDatabasePermissionsConnector(meta, data)
// 	if err != nil {
// 		return diag.FromErr(err)
// 	}
//
// 	dbPermissionModel := &model.DatabasePermissions{
// 		DatabaseName: database,
// 		PrincipalID:  principalId,
// 		Permissions:  toStringSlice(permissions),
// 	}
// 	if err = connector.UpdateDatabasePermissions(ctx, dbPermissionModel); err != nil {
// 		return diag.FromErr(errors.Wrapf(err, "unable to update permissions for principal [%d] on database [%s]", principalId, database))
// 	}
//
// 	data.SetId(getDatabasePermissionsID(data))
//
// 	logger.Info().Msgf("updated permissions for principal [%d] on database [%s]", principalId, database)
//
// 	return resourceDatabasePermissionsRead(ctx, data, meta)
// }

func resourceDatabasePermissionDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "user", "delete")
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

// func resourceUserImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
// 	logger := loggerFromMeta(meta, "user", "import")
// 	logger.Debug().Msgf("Import %s", data.Id())
//
// 	server, u, err := serverFromId(data.Id())
// 	if err != nil {
// 		return nil, err
// 	}
// 	if err = data.Set(serverProp, server); err != nil {
// 		return nil, err
// 	}
//
// 	parts := strings.Split(u.Path, "/")
// 	if len(parts) != 3 {
// 		return nil, errors.New("invalid ID")
// 	}
// 	if err = data.Set(databaseProp, parts[1]); err != nil {
// 		return nil, err
// 	}
// 	if err = data.Set(usernameProp, parts[2]); err != nil {
// 		return nil, err
// 	}
//
// 	data.SetId(getUserID(data))
//
// 	database := data.Get(databaseProp).(string)
// 	username := data.Get(usernameProp).(string)
//
// 	connector, err := getUserConnector(meta, data)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	login, err := connector.GetUser(ctx, database, username)
// 	if err != nil {
// 		return nil, errors.Wrapf(err, "unable to read user [%s].[%s] for import", database, username)
// 	}
//
// 	if login == nil {
// 		return nil, errors.Errorf("no user [%s].[%s] found for import", database, username)
// 	}
//
// 	if err = data.Set(authenticationTypeProp, login.AuthType); err != nil {
// 		return nil, err
// 	}
// 	if err = data.Set(principalIdProp, login.PrincipalID); err != nil {
// 		return nil, err
// 	}
// 	if err = data.Set(defaultSchemaProp, login.DefaultSchema); err != nil {
// 		return nil, err
// 	}
// 	if err = data.Set(defaultLanguageProp, login.DefaultLanguage); err != nil {
// 		return nil, err
// 	}
// 	if err = data.Set(rolesProp, login.Roles); err != nil {
// 		return nil, err
// 	}
//
// 	return []*schema.ResourceData{data}, nil
// }

func getDatabasePermissionsConnector(meta interface{}, data *schema.ResourceData) (DatabasePermissionsConnector, error) {
	provider := meta.(model.Provider)
	connector, err := provider.GetConnector(serverProp, data)
	if err != nil {
		return nil, err
	}
	return connector.(DatabasePermissionsConnector), nil
}
