package mssql

import (
  "context"
  "fmt"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/pkg/errors"
  "net/url"
  "strings"
  "time"
)

var defaultAzureTimeout = schema.DefaultTimeout(30 * time.Second)

func resourceAzSpLogin() *schema.Resource {
  return &schema.Resource{
    Description:   "Manipulate a service principal login on Azure SQL Database",
    CreateContext: resourceAzSpCreate,
    ReadContext:   resourceAzSpRead,
    UpdateContext: resourceAzSpUpdate,
    DeleteContext: resourceAzSpDelete,
    Importer: &schema.ResourceImporter{
      StateContext: resourceAzSpLoginImport,
    },
    Timeouts: &schema.ResourceTimeout{
      Default: defaultAzureTimeout,
    },
    Schema: map[string]*schema.Schema{
      serverProp: {
        Type:     schema.TypeList,
        MaxItems: 1,
        Required: true,
        Elem: &schema.Resource{
          Schema: getServerSchema(serverProp, false, nil),
        },
      },
      databaseProp: {
        Type:     schema.TypeString,
        Required: true,
      },
      usernameProp: {
        Type:     schema.TypeString,
        Required: true,
      },
      clientIdProp: {
        Type:     schema.TypeString,
        Required: true,
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

func resourceAzSpCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := meta.(Provider).logger.With().Str("resource", "az_sp_login").Str("func", "create").Logger()
  logger.Debug().Msgf("Create %s", data.Id())

  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)
  clientId := data.Get(clientIdProp).(string)
  defSchema := data.Get(schemaProp).(string)
  roles := data.Get(rolesProp).([]interface{})

  connector := GetConnector(serverProp, data)
  connector.Database = database

  err := connector.CreateAzureADLogin(ctx, username, clientId, defSchema, roles)
  if err != nil {
   return diag.FromErr(errors.Wrap(err, fmt.Sprintf("unable to create az login [%s].[%s]", database, username)))
  }

  data.SetId(resourceAzSpLoginGetID(data))

  logger.Info().Msgf("created az login [%s].[%s] for client_id %s", database, username, clientId)

  return resourceAzSpRead(ctx, data, meta)
}

func resourceAzSpRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := meta.(Provider).logger.With().Str("resource", "az_sp_login").Str("func", "read").Logger()
  logger.Debug().Msgf("Read %s", data.Id())

  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)

  connector := GetConnector(serverProp, data)
  connector.Database = database

  login, err := connector.GetUserLogin(ctx, username)
  if err != nil {
    return diag.FromErr(errors.Wrap(err, fmt.Sprintf("unable to read login [%s].[%s]", database, username)))
  }
  if login == nil {
    logger.Info().Msgf("No login found for user [%s].[%s]", database, username)
    data.SetId("")
  }

  if err = data.Set(clientIdProp, login.SID.String()); err != nil {
    return diag.FromErr(err)
  }
  if err = data.Set(principalIdProp, login.PrincipalID); err != nil {
    return diag.FromErr(err)
  }
  if err = data.Set(schemaProp, login.Schema); err != nil {
    return diag.FromErr(err)
  }
  if err = data.Set(rolesProp, login.Roles); err != nil {
    return diag.FromErr(err)
  }

  return nil
}

func resourceAzSpUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := meta.(Provider).logger.With().Str("resource", "az_sp_login").Str("func", "update").Logger()
  logger.Debug().Msgf("Update %s", data.Id())

  return nil
}

func resourceAzSpDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := meta.(Provider).logger.With().Str("resource", "az_sp_login").Str("func", "delete").Logger()
  logger.Debug().Msgf("Delete %s", data.Id())

  return nil
}

func resourceAzSpLoginImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
  logger := meta.(Provider).logger.With().Str("resource", "az_sp_login").Str("func", "import").Logger()
  logger.Debug().Msgf("Import %s", data.Id())

  server, err := serverFromId(data.Id(), false)
  if err != nil {
    return nil, err
  }
  if err = data.Set(serverProp, server); err != nil {
    return nil, err
  }

  u, err := url.Parse(data.Id())
  if err != nil {
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

  data.SetId(resourceAzSpLoginGetID(data))

  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)

  connector := GetConnector(serverProp, data)
  connector.Database = database

  login, err := connector.GetUserLogin(ctx, username)
  if err != nil {
    return nil, errors.Wrap(err, fmt.Sprintf("unable to read login [%s].[%s] for import", database, username))
  }

  if login == nil {
    return nil, errors.Errorf("no login found for user [%s].[%s] for import", connector.Database, username)
  }

  if err = data.Set(clientIdProp, login.SID.String()); err != nil {
    return nil, err
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

func resourceAzSpLoginGetID(data *schema.ResourceData) string {
  host := data.Get(serverProp + ".0.host").(string)
  port := data.Get(serverProp + ".0.port").(string)
  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)
  return fmt.Sprintf("sqlserver://%s:%s/%s/%s", host, port, database, username)
}
