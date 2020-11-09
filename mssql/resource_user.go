package mssql

import (
  "context"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/pkg/errors"
  "strings"
  "terraform-provider-mssql/mssql/model"
)

const authenticationTypeProp = "authentication_type"
const defaultSchemaProp = "default_schema"

func resourceUser() *schema.Resource {
  return &schema.Resource{
    CreateContext: resourceUserCreate,
    ReadContext:   resourceUserRead,
    UpdateContext: resourceUserUpdate,
    DeleteContext: resourceUserDelete,
    Importer: &schema.ResourceImporter{
      StateContext: resourceUserImport,
    },
    Schema: map[string]*schema.Schema{
      serverProp: {
        Type:         schema.TypeList,
        MaxItems:     1,
        Optional:     true,
        ExactlyOneOf: []string{serverProp, serverEncodedProp},
        Elem: &schema.Resource{
          Schema: getServerSchema(serverProp, true, nil),
        },
      },
      serverEncodedProp: {
        Type:         schema.TypeString,
        Optional:     true,
        Sensitive:    true,
        ExactlyOneOf: []string{serverProp, serverEncodedProp},
      },
      databaseProp: {
        Type:     schema.TypeString,
        Optional: true,
        ForceNew: true,
        Default:  "master",
      },
      usernameProp: {
        Type:     schema.TypeString,
        Required: true,
        ForceNew: true,
      },
      loginNameProp: {
        Type:     schema.TypeString,
        Optional: true,
        ForceNew: true,
      },
      passwordProp: {
        Type:     schema.TypeString,
        Optional: true,
        ForceNew: true,
      },
      authenticationTypeProp: {
        Type:     schema.TypeString,
        Computed: true,
      },
      principalIdProp: {
        Type:     schema.TypeInt,
        Computed: true,
      },
      defaultSchemaProp: {
        Type:     schema.TypeString,
        Optional: true,
        Default:  "dbo",
      },
      defaultLanguageProp: {
        Type:     schema.TypeString,
        Optional: true,
      },
    },
  }
}

type UserConnector interface {
  CreateUser(ctx context.Context, database string, user *model.User) error
  GetUser(ctx context.Context, database, username string) (*model.User, error)
  UpdateUser(ctx context.Context, database string, user *model.User) error
  DeleteUser(ctx context.Context, database, username string) error
}

func resourceUserCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "user", "create")
  logger.Debug().Msgf("Create %s", getUserID(data))

  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)
  loginName := data.Get(loginNameProp).(string)
  password := data.Get(passwordProp).(string)
  defaultSchema := data.Get(defaultSchemaProp).(string)
  defaultLanguage := data.Get(defaultLanguageProp).(string)

  if loginName != "" && password != "" {
    return diag.Errorf(loginNameProp + " and " + passwordProp + " cannot both be set")
  }
  var authType string
  if loginName != "" {
    authType = "INSTANCE"
  } else if password != "" {
    authType = "DATABASE"
  } else {
    authType = "EXTERNAL"
  }
  if defaultSchema == "" {
    return diag.Errorf(defaultSchemaProp + " cannot be empty")
  }

  connector, err := getUserConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  user := &model.User{
    Username:        username,
    LoginName:       loginName,
    Password:        password,
    AuthType:        authType,
    DefaultSchema:   defaultSchema,
    DefaultLanguage: defaultLanguage,
  }
  if err = connector.CreateUser(ctx, database, user); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to create user [%s].[%s]", database, username))
  }

  data.SetId(getUserID(data))

  logger.Info().Msgf("created user [%s].[%s]", database, username)

  return resourceUserRead(ctx, data, meta)
}

func resourceUserRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "user", "read")
  logger.Debug().Msgf("Read %s", data.Id())

  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)

  connector, err := getUserConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  user, err := connector.GetUser(ctx, database, username)
  if err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to read user [%s].[%s]", database, username))
  }
  if user == nil {
    logger.Info().Msgf("No user found for [%s].[%s]", database, username)
    data.SetId("")
  } else {
    if err = data.Set(authenticationTypeProp, user.AuthType); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(principalIdProp, user.PrincipalID); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(defaultSchemaProp, user.DefaultSchema); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(defaultLanguageProp, user.DefaultLanguage); err != nil {
      return diag.FromErr(err)
    }
  }

  return nil
}

func resourceUserUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "user", "update")
  logger.Debug().Msgf("Update %s", data.Id())

  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)
  defaultSchema := data.Get(defaultSchemaProp).(string)
  defaultLanguage := data.Get(defaultLanguageProp).(string)

  connector, err := getUserConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  user := &model.User{
    Username:        username,
    DefaultSchema:   defaultSchema,
    DefaultLanguage: defaultLanguage,
  }
  if err = connector.UpdateUser(ctx, database, user); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to update user [%s].[%s]", database, username))
  }

  data.SetId(getUserID(data))

  logger.Info().Msgf("updated user [%s].[%s]", database, username)

  return resourceUserRead(ctx, data, meta)
}

func resourceUserDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "user", "delete")
  logger.Debug().Msgf("Delete %s", data.Id())

  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)

  connector, err := getUserConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  if err = connector.DeleteUser(ctx, database, username); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to delete user [%s].[%s]", database, username))
  }

  logger.Info().Msgf("deleted user [%s].[%s]", database, username)

  // d.SetId("") is automatically called assuming delete returns no errors, but it is added here for explicitness.
  data.SetId("")

  return nil
}

func resourceUserImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
  logger := loggerFromMeta(meta, "user", "import")
  logger.Debug().Msgf("Import %s", data.Id())

  server, u, err := serverFromId(data.Id(), true)
  if err != nil {
    return nil, err
  }
  if err = data.Set(serverProp, server); err != nil {
    return nil, err
  }

  parts := strings.Split(u.Path, "/")
  if len(parts) != 3 {
    return nil, errors.New("invalid ID")
  }
  if err = data.Set(databaseProp, parts[1]); err != nil {
    return nil, err
  }
  if err = data.Set(usernameProp, parts[2]); err != nil {
    return nil, err
  }

  data.SetId(getUserID(data))

  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)

  connector, err := getUserConnector(meta, data)
  if err != nil {
    return nil, err
  }

  login, err := connector.GetUser(ctx, database, username)
  if err != nil {
    return nil, errors.Wrapf(err, "unable to read user [%s].[%s] for import", database, username)
  }

  if login == nil {
    return nil, errors.Errorf("no user [%s].[%s] found for import", database, username)
  }

  if err = data.Set(authenticationTypeProp, login.AuthType); err != nil {
    return nil, err
  }
  if err = data.Set(principalIdProp, login.PrincipalID); err != nil {
    return nil, err
  }
  if err = data.Set(defaultSchemaProp, login.DefaultSchema); err != nil {
    return nil, err
  }
  if err = data.Set(defaultLanguageProp, login.DefaultLanguage); err != nil {
    return nil, err
  }

  return []*schema.ResourceData{data}, nil
}

func getUserConnector(meta interface{}, data *schema.ResourceData) (UserConnector, error) {
  provider := meta.(model.Provider)
  connector, err := provider.GetConnector(serverProp, data)
  if err != nil {
    return nil, err
  }
  return connector.(UserConnector), nil
}
