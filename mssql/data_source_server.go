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
  "terraform-provider-mssql/sql"
)

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
  connector, diags := createConnector("", data)
  if len(diags) > 0 {
    return diags
  }

  encoded, err := json.Marshal(connector)
  if err != nil {
    return append(diags, diag.FromErr(err)[0])
  }
  data.Set("encoded", string(encoded))

  logger := meta.(Provider).logger
  logger.Info().Msgf("Created connector for %s", connector.ID())

  data.SetId(connector.ID())

  return nil
}

func getServerSchema(prefix string, allowAdministratorLogin bool, extensions *map[string]*schema.Schema) map[string]*schema.Schema {
  prefix = getPrefix(prefix)
  s := map[string]*schema.Schema{
    "name": {
      Type:     schema.TypeString,
      Required: true,
      ForceNew: true,
      DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
        return strings.ToLower(old) == strings.ToLower(new)
      },
    },
    "fqdn": {
      Type:     schema.TypeString,
      Optional: true,
      ForceNew: true,
      DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
        old = strings.ToLower(old)
        if new == "" {
          name := strings.ToLower(d.Get(prefix + "name").(string))
          return old == name || old == name+".database.windows.net"
        }
        return strings.ToLower(old) == strings.ToLower(new)
      },
    },
    "port": {
      Type:     schema.TypeString,
      Optional: true,
      ForceNew: true,
      Default:  1433,
    },
    "azure_administrator": {
      Type:     schema.TypeList,
      MaxItems: 1,
      Required: true,
      Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
          "tenant_id": {
            Type:        schema.TypeString,
            Required:    true,
            DefaultFunc: schema.EnvDefaultFunc("MSSQL_ADMIN_TENANT_ID", nil),
          },
          "client_id": {
            Type:        schema.TypeString,
            Required:    true,
            DefaultFunc: schema.EnvDefaultFunc("MSSQL_ADMIN_CLIENT_ID", nil),
          },
          "client_secret": {
            Type:        schema.TypeString,
            Required:    true,
            Sensitive:   true,
            DefaultFunc: schema.EnvDefaultFunc("MSSQL_ADMIN_CLIENT_SECRET", nil),
          },
        },
      },
    },
  }
  if allowAdministratorLogin {
    s["azure_administrator"].Optional = true
    s["azure_administrator"].Required = false
    s["azure_administrator"].ExactlyOneOf = []string{prefix + "administrator_login", prefix + "azure_administrator"}
    s["administrator_login"] = &schema.Schema{
      Type:         schema.TypeList,
      MaxItems:     1,
      Optional:     true,
      ExactlyOneOf: []string{prefix + "administrator_login", prefix + "azure_administrator"},
      Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
          "username": {
            Type:        schema.TypeString,
            Required:    true,
            DefaultFunc: schema.EnvDefaultFunc("MSSQL_ADMIN_USERNAME", nil),
          },
          "password": {
            Type:        schema.TypeString,
            Required:    true,
            Sensitive:   true,
            DefaultFunc: schema.EnvDefaultFunc("MSSQL_ADMIN_PASSWORD", nil),
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

func serverFromId(id string) ([]map[string]interface{}, error) {
  u, err := url.Parse(id)
  if err != nil {
    return nil, err
  }

  if u.Scheme != "sqlserver" && u.Scheme != "mssql" {
    return nil, errors.New("invalid schema in ID")
  }

  host := u.Host
  port := sql.DefaultPort

  if strings.IndexRune(host, ':') != -1 {
    var err error
    if host, port, err = net.SplitHostPort(u.Host); err != nil {
      return nil, err
    }
  }

  name := host
  if strings.HasSuffix(host, ".database.windows.net") {
    name = host[:len(host)-len(".database.windows.net")]
  }

  values := u.Query()

  administratorLogin, adminInValues := getAdministratorLogin(values)
  azureAdministrator, azureInValues := getAzureAdministrator(values)
  if administratorLogin == nil && azureAdministrator == nil {
    return nil, errors.New("neither administrator login nor azure administrator specified in ID")
  }
  if adminInValues && azureInValues {
    return nil, errors.New("both administrator login and azure administrator specified in ID")
  }
  if administratorLogin != nil {
    azureAdministrator = nil
  }

  return []map[string]interface{}{{
    "name":                name,
    "fqdn":                host,
    "port":                port,
    "administrator_login": administratorLogin,
    "azure_administrator": azureAdministrator,
  }}, nil
}

func getAdministratorLogin(values url.Values) ([]map[string]interface{}, bool) {
  var inValues bool

  username := values.Get("admin_username")
  if username == "" {
    username = os.Getenv("MSSQL_ADMIN_USERNAME")
  } else {
    inValues = true
  }

  password := values.Get("admin_password")
  if password == "" {
    password = os.Getenv("MSSQL_ADMIN_PASSWORD")
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

func getAzureAdministrator(values url.Values) ([]map[string]interface{}, bool) {
  var inValues bool

  tenantId := values.Get("admin_tenant_id")
  if tenantId == "" {
    tenantId = os.Getenv("MSSQL_ADMIN_TENANT_ID")
  } else {
    inValues = true
  }

  clientId := values.Get("admin_client_id")
  if clientId == "" {
    clientId = os.Getenv("MSSQL_ADMIN_CLIENT_ID")
  } else {
    inValues = true
  }

  clientSecret := values.Get("admin_client_secret")
  if clientSecret == "" {
    clientSecret = os.Getenv("MSSQL_ADMIN_CLIENT_SECRET")
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

func getConnector(prefix string, data *schema.ResourceData) (*sql.Connector, diag.Diagnostics) {
  if encoded, isOk := data.GetOk(prefix + "_encoded"); isOk {
    c := &sql.Connector{}
    err := json.Unmarshal([]byte(encoded.(string)), c)
    if err != nil {
      return nil, diag.FromErr(err)
    }
    return c, nil
  }

  return createConnector(prefix, data)
}

func createConnector(prefix string, data *schema.ResourceData) (*sql.Connector, diag.Diagnostics) {
  prefix = getPrefix(prefix)

  // Warnings or errors can be collected in a slice type
  var diags diag.Diagnostics

  connector := &sql.Connector{
    Host:    data.Get(prefix + "name").(string),
    Port:    data.Get(prefix + "port").(string),
    Timeout: data.Timeout(schema.TimeoutRead),
  }

  if admin, ok := data.GetOk(prefix + "administrator_login.0"); ok {
    admin := admin.(map[string]interface{})
    connector.Administrator = &sql.AdministratorLogin{
      Username: admin["username"].(string),
      Password: admin["password"].(string),
    }
  }

  if admin, ok := data.GetOk(prefix + "azure_administrator.0"); ok {
    admin := admin.(map[string]interface{})

    connector.AzureAdministrator = &sql.AzureAdministrator{
      TenantID:     admin["tenant_id"].(string),
      ClientID:     admin["client_id"].(string),
      ClientSecret: admin["client_secret"].(string),
    }
  }

  if fqdn, ok := data.GetOk(prefix + "fqdn"); ok {
    connector.Host = fqdn.(string)
  } else if connector.AzureAdministrator != nil && !strings.HasSuffix(connector.Host, ".database.windows.net") {
    connector.Host = connector.Host + ".database.windows.net"
    if prefix == "" {
      if err := data.Set(prefix+"fqdn", connector.Host); err != nil {
        diags = append(diags, diag.FromErr(err)[0])
      }
    } else {
      servers := data.Get("server").([]interface{})
      server := servers[0].(map[string]interface{})
      server["fqdn"] = connector.Host
    }
  }

  return connector, diags
}

func GetConnector(prefix string, data *schema.ResourceData) *sql.Connector {
  prefix = getPrefix(prefix)

  connector := &sql.Connector{
    Host:    data.Get(prefix + "fqdn").(string),
    Port:    data.Get(prefix + "port").(string),
    Timeout: data.Timeout(schema.TimeoutRead),
  }

  if admin, ok := data.GetOk(prefix + "administrator_login.0"); ok {
    admin := admin.(map[string]interface{})
    connector.Administrator = &sql.AdministratorLogin{
      Username: admin["username"].(string),
      Password: admin["password"].(string),
    }
  }

  if admin, ok := data.GetOk(prefix + "azure_administrator.0"); ok {
    admin := admin.(map[string]interface{})
    connector.AzureAdministrator = &sql.AzureAdministrator{
      TenantID:     admin["tenant_id"].(string),
      ClientID:     admin["client_id"].(string),
      ClientSecret: admin["client_secret"].(string),
    }
  }

  return connector
}

func getPrefix(prefix string) string {
  if len(prefix) > 0 {
    return prefix + ".0."
  }
  return prefix
}
