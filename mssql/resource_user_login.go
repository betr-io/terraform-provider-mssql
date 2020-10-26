package mssql

import (
  "context"
  "fmt"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/pkg/errors"
  "net"
  "net/url"
  "os"
  "strings"
  "terraform-provider-mssql/sql"
)

const serverEncodedProp = "server_encoded"

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
    },
  }
}

func resourceUserLoginImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
  // logger := meta.(Provider).logger

  id := data.Id()
  u, err := url.Parse(id)
  if err != nil {
    return nil, err
  }

  host := u.Host
  port := sql.DefaultPort

  if strings.IndexRune(host, ':') != -1 {
    var err error
    if host, port, err = net.SplitHostPort(u.Host); err != nil {
      return nil, err
    }
  }

  parts := strings.SplitN(u.Path, "/", 3)
  databaseName := parts[1]
  username := parts[2]

  var found bool
  values := u.Query()
  adminUsername := values.Get("admin_username")
  if adminUsername == "" {
    if adminUsername, found = os.LookupEnv("MSSQL_USERNAME"); !found {
      return nil, errors.New("Missing admin_username query value or MSSQL_USERNAME environment variable")
    }
  }
  adminPassword := values.Get("admin_password")
  if adminPassword == "" {
    if adminPassword, found = os.LookupEnv("MSSQL_PASSWORD"); !found {
      return nil, errors.New("Missing MSSQL_PASSWORD environment variable")
    }
  }

  administratorLogin := make([]map[string]interface{}, 1)
  administratorLogin[0] = map[string]interface{}{
    "username": adminUsername,
    "password": adminPassword,
  }
  server := make([]map[string]interface{}, 1)
  server[0] = map[string]interface{}{
    "host":                host,
    "port":                port,
    "administrator_login": administratorLogin,
  }
  if err = data.Set("server", server); err != nil {
    return nil, err
  }
  if err = data.Set(databaseProp, parts[0]); err != nil {
    return nil, err
  }
  if err = data.Set(usernameProp, parts[1]); err != nil {
    return nil, err
  }

  data.SetId(fmt.Sprintf("sqlserver://%s:%d/%s/%s", host, port, databaseName, username))

  return []*schema.ResourceData{data}, nil
}

func resourceUserLoginCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  connector, diags := getConnector("server", data)
  if diags != nil {
    return diags
  }

  connector.Database = data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)
  password := data.Get(passwordProp).(string)

  if err := connector.CreateUserLogin(ctx, connector.Database, username, password); err != nil {
    return diag.FromErr(err)
  }

  data.SetId(connector.ID() + "/" + username)

  return resourceUserLoginRead(ctx, data, meta)
}

func resourceUserLoginRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := meta.(Provider).logger
  logger.Info().Msgf("Read user login %v %+v", data.Id(), data)

  connector, diags := getConnector("server", data)
  if diags != nil {
    return diags
  }
  connector.Database = data.Get(databaseProp).(string)

  logger.Info().Msgf("Connector %+v", connector)

  u, err := url.Parse(data.Id())
  if err != nil {
    return diag.FromErr(err)
  }
  username := strings.SplitN(u.Path, "/", 3)[2]

  login, err := connector.GetUserLogin(ctx, username)
  if err != nil {
    logger.Err(err)
    return diag.FromErr(errors.Wrap(err, "UserLoginRead"))
  }
  if login == nil {
    logger.Info().Msgf("No login found for user [%s].[%s]'", connector.Database, username)
    data.SetId("")
  }

  logger.Info().Msgf("data %#v\nlogin %+v", *data, login)

  data.Set(usernameProp, login.Username)
  data.Set(serverProp, connector)

  return nil
}

func resourceUserLoginUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  connector, diags := getConnector("server", data)
  if diags != nil {
    return diags
  }

  username := data.Get(usernameProp).(string)
  password := data.Get(passwordProp).(string)

  if err := connector.UpdateUserLogin(ctx, username, password); err != nil {
    return diag.FromErr(err)
  }

  return resourceUserLoginRead(ctx, data, meta)
}

func resourceUserLoginDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  connector, diags := getConnector("server", data)
  if diags != nil {
    return diags
  }
  connector.Database = data.Get(databaseProp).(string)

  username := strings.Split(connector.ID(), "/")[1]

  if err := connector.DeleteUserLogin(ctx, username); err != nil {
    return diag.FromErr(err)
  }

  return nil
}
