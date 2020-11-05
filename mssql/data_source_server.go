package mssql

import (
  "context"
  "encoding/json"
  "errors"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "net"
  "net/url"
  "os"
  "strings"
)

type ServerConnector interface {
  ID() string
}

func getServerConnector(meta interface{}, prefix string, data *schema.ResourceData) (ServerConnector, error) {
  provider := meta.(Provider)
  connector, err := provider.GetConnector(prefix, data)
  if err != nil {
    return nil, err
  }
  return connector.(ServerConnector), nil
}

func getPrefix(prefix string) string {
  if len(prefix) > 0 {
    return prefix + ".0."
  }
  return prefix
}

func dataSourceServer() *schema.Resource {
  return &schema.Resource{
    ReadContext: dataSourceServerRead,
    Schema: getServerSchema("", true, &map[string]*schema.Schema{
      "encoded": {
        Type:      schema.TypeString,
        Computed:  true,
        Sensitive: true,
      },
    }),
    Timeouts: &schema.ResourceTimeout{
      Read: defaultReadTimeout,
    },
  }
}

func dataSourceServerRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  connector, err := getServerConnector(meta, "", data)
  if err != nil {
    return diag.FromErr(err)
  }

  encoded, err := json.Marshal(connector)
  if err != nil {
    return diag.FromErr(err)
  }
  data.Set("encoded", string(encoded))

  logger := meta.(Provider).DataSourceLogger("server", "read")
  logger.Info().Msgf("Created connector for %s", connector.ID())

  data.SetId(connector.ID())

  return nil
}

func getServerSchema(prefix string, allowSqlLogin bool, extensions *map[string]*schema.Schema) map[string]*schema.Schema {
  prefix = getPrefix(prefix)
  s := map[string]*schema.Schema{
    "host": {
      Type:     schema.TypeString,
      Required: true,
      ForceNew: true,
      DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
        return strings.ToLower(old) == strings.ToLower(new)
      },
    },
    "port": {
      Type:     schema.TypeString,
      Optional: true,
      ForceNew: true,
      Default:  1433,
    },
    "azure_login": {
      Type:     schema.TypeList,
      MaxItems: 1,
      Required: true,
      Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
          "tenant_id": {
            Type:        schema.TypeString,
            Required:    true,
            DefaultFunc: schema.EnvDefaultFunc("MSSQL_TENANT_ID", nil),
          },
          "client_id": {
            Type:        schema.TypeString,
            Required:    true,
            DefaultFunc: schema.EnvDefaultFunc("MSSQL_CLIENT_ID", nil),
          },
          "client_secret": {
            Type:        schema.TypeString,
            Required:    true,
            Sensitive:   true,
            DefaultFunc: schema.EnvDefaultFunc("MSSQL_CLIENT_SECRET", nil),
          },
        },
      },
    },
  }
  if allowSqlLogin {
    s["azure_login"].Optional = true
    s["azure_login"].Required = false
    s["azure_login"].ExactlyOneOf = []string{prefix + "login", prefix + "azure_login"}
    s["login"] = &schema.Schema{
      Type:         schema.TypeList,
      MaxItems:     1,
      Optional:     true,
      ExactlyOneOf: []string{prefix + "login", prefix + "azure_login"},
      Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
          "username": {
            Type:        schema.TypeString,
            Required:    true,
            DefaultFunc: schema.EnvDefaultFunc("MSSQL_USERNAME", nil),
          },
          "password": {
            Type:        schema.TypeString,
            Required:    true,
            Sensitive:   true,
            DefaultFunc: schema.EnvDefaultFunc("MSSQL_PASSWORD", nil),
          },
        },
      },
    }
  }
  if extensions != nil {
    for k, v := range *extensions {
      s[k] = v
    }
  }

  return s
}

const DefaultPort = "1433"
func serverFromId(id string, allowLogin bool) ([]map[string]interface{}, *url.URL, error) {
  u, err := url.Parse(id)
  if err != nil {
    return nil, nil, err
  }

  if u.Scheme != "sqlserver" && u.Scheme != "mssql" {
    return nil, nil, errors.New("invalid schema in ID")
  }

  host := u.Host
  port := DefaultPort

  if strings.IndexRune(host, ':') != -1 {
    var err error
    if host, port, err = net.SplitHostPort(u.Host); err != nil {
      return nil, nil, err
    }
  }

  values := u.Query()

  login, loginInValues := getLogin(values)
  azureLogin, azureInValues := getAzureLogin(values)
  if login == nil && azureLogin == nil {
    return nil, nil, errors.New("neither login nor azure login specified")
  }
  if loginInValues && azureInValues {
    return nil, nil, errors.New("both login and azure login specified in resource")
  }
  if login != nil && azureLogin != nil {
    // prefer azure login
    login = nil
  }

  if allowLogin {
    return []map[string]interface{}{{
      "host":        host,
      "port":        port,
      "login":       login,
      "azure_login": azureLogin,
    }}, u, nil
  }

  if azureLogin == nil {
    return nil, nil, errors.New("this resource requires azure login")
  }
  return []map[string]interface{}{{
    "host":        host,
    "port":        port,
    "azure_login": azureLogin,
  }}, u, nil
}

func getLogin(values url.Values) ([]map[string]interface{}, bool) {
  var inValues bool

  username := values.Get("username")
  if username == "" {
    username = os.Getenv("MSSQL_USERNAME")
  } else {
    inValues = true
  }

  password := values.Get("password")
  if password == "" {
    password = os.Getenv("MSSQL_PASSWORD")
  } else {
    inValues = true
  }

  if username == "" || password == "" {
    return nil, false
  }

  return []map[string]interface{}{{
    "username": username,
    "password": password,
  }}, inValues
}

func getAzureLogin(values url.Values) ([]map[string]interface{}, bool) {
  var inValues bool

  tenantId := values.Get("tenant_id")
  if tenantId == "" {
    tenantId = os.Getenv("MSSQL_TENANT_ID")
  } else {
    inValues = true
  }

  clientId := values.Get("client_id")
  if clientId == "" {
    clientId = os.Getenv("MSSQL_CLIENT_ID")
  } else {
    inValues = true
  }

  clientSecret := values.Get("client_secret")
  if clientSecret == "" {
    clientSecret = os.Getenv("MSSQL_CLIENT_SECRET")
  } else {
    inValues = true
  }

  if tenantId == "" || clientId == "" || clientSecret == "" {
    return nil, false
  }

  return []map[string]interface{}{{
    "tenant_id":     tenantId,
    "client_id":     clientId,
    "client_secret": clientSecret,
  }}, inValues
}
