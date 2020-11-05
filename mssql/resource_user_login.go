package mssql

import (
  "context"
  "fmt"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/pkg/errors"
  "strings"
)

const serverEncodedProp = "server_encoded"

type UserLoginConnector interface {
  GetDatabase() string
  SetDatabase(database string)
  CreateUserLogin(ctx context.Context, database, username, password, schema string, roles []interface{}) error
  GetUserLogin(ctx context.Context, username string) (*UserLogin, error)
  UpdateUserLogin(ctx context.Context, username string, password string) error
  DeleteUserLogin(ctx context.Context, username string) error
}

func resourceUserLogin() *schema.Resource {
  return &schema.Resource{
    CreateContext: resourceUserLoginCreate,
    ReadContext:   resourceUserLoginRead,
    UpdateContext: resourceUserLoginUpdate,
    DeleteContext: resourceUserLoginDelete,
    Importer: &schema.ResourceImporter{
      StateContext: resourceUserLoginImport,
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
        Default:  "master",
      },
      usernameProp: {
        Type:     schema.TypeString,
        Required: true,
        ForceNew: true,
      },
      passwordProp: {
        Type:      schema.TypeString,
        Required:  true,
        Sensitive: true,
      },
      principalIdProp: {
        Type:     schema.TypeInt,
        Computed: true,
      },
      schemaProp: {
        Type:     schema.TypeString,
        Optional: true,
        Default:  schemaPropDefault,
      },
      rolesProp: {
        Type:     schema.TypeList,
        Optional: true,
        DefaultFunc: func() (interface{}, error) {
          return rolesPropDefault, nil
        },
        Elem: &schema.Schema{
          Type: schema.TypeString,
        },
      },
    },
  }
}

func resourceUserLoginCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "user_login", "create")
  logger.Debug().Msgf("Create %s", resourceAzSpLoginGetID(data))

  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)
  password := data.Get(passwordProp).(string)
  defSchema := data.Get(schemaProp).(string)
  roles := data.Get(rolesProp).([]interface{})

  connector, err := getUserLoginConnector(meta, serverProp, data)
  if err != nil {
    return diag.FromErr(err)
  }
  connector.SetDatabase(database)

  if err = connector.CreateUserLogin(ctx, database, username, password, defSchema, roles); err != nil {
    return diag.FromErr(errors.Wrap(err, fmt.Sprintf("unable to create login [%s].[%s]", database, username)))
  }

  data.SetId(resourceUserLoginGetID(data))

  logger.Info().Msgf("created login [%s].[%s]", database, username)

  return resourceUserLoginRead(ctx, data, meta)
}

func resourceUserLoginRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "user_login", "read")
  logger.Debug().Msgf("Read %s", resourceUserLoginGetID(data))

  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)

  connector, err := getUserLoginConnector(meta, serverProp, data)
  if err != nil {
    return diag.FromErr(err)
  }
  connector.SetDatabase(database)

  login, err := connector.GetUserLogin(ctx, username)
  if err != nil {
    return diag.FromErr(errors.Wrap(err, fmt.Sprintf("unable to read login [%s].[%s]", database, username)))
  }
  if login == nil {
    logger.Info().Msgf("No login found for user [%s].[%s]'", connector.GetDatabase(), username)
    data.SetId("")
  } else {
    if err = data.Set(principalIdProp, login.PrincipalID); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(schemaProp, login.Schema); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(rolesProp, login.Roles); err != nil {
      return diag.FromErr(err)
    }
  }

  return nil
}

func resourceUserLoginUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "user_login", "update")
  logger.Debug().Msgf("Update %s", data.Id())

  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)
  password := data.Get(passwordProp).(string)

  connector, err := getUserLoginConnector(meta, serverProp, data)
  if err != nil {
    return diag.FromErr(err)
  }
  connector.SetDatabase(database)

  if err := connector.UpdateUserLogin(ctx, username, password); err != nil {
    return diag.FromErr(err)
  }

  return resourceUserLoginRead(ctx, data, meta)
}

func resourceUserLoginDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "user_login", "delete")
  logger.Debug().Msgf("Delete %s", data.Id())

  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)

  connector, err := getUserLoginConnector(meta, serverProp, data)
  if err != nil {
    return diag.FromErr(err)
  }
  connector.SetDatabase(database)

  if err := connector.DeleteUserLogin(ctx, username); err != nil {
    return diag.FromErr(err)
  }

  // d.SetId("") is automatically called assuming delete returns no errors, but it is added here for explicitness.
  data.SetId("")

  return nil
}

func resourceUserLoginImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
  logger := loggerFromMeta(meta, "user_login", "import")
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

  data.SetId(resourceUserLoginGetID(data))

  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)

  connector, err := getUserLoginConnector(meta, serverProp, data)
  if err != nil {
    return nil, err
  }
  connector.SetDatabase(database)

  login, err := connector.GetUserLogin(ctx, username)
  if err != nil {
    return nil, errors.Wrap(err, fmt.Sprintf("unable to read login [%s].[%s] for import", database, username))
  }

  if login == nil {
    return nil, errors.Errorf("no login found for user [%s].[%s] for import", connector.GetDatabase(), username)
  }

  if err = data.Set(principalIdProp, login.PrincipalID); err != nil {
    return nil, err
  }
  if err = data.Set(schemaProp, login.Schema); err != nil {
    return nil, err
  }
  if err = data.Set(rolesProp, login.Roles); err != nil {
    return nil, err
  }

  return []*schema.ResourceData{data}, nil
}

func resourceUserLoginGetID(data *schema.ResourceData) string {
  host := data.Get(serverProp + ".0.host").(string)
  port := data.Get(serverProp + ".0.port").(string)
  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)
  return fmt.Sprintf("sqlserver://%s:%s/%s/%s", host, port, database, username)
}

func getUserLoginConnector(meta interface{}, prefix string, data *schema.ResourceData) (UserLoginConnector, error) {
  provider := meta.(Provider)
  connector, err := provider.GetConnector(prefix, data)
  if err != nil {
    return nil, err
  }
  return connector.(UserLoginConnector), nil
}
