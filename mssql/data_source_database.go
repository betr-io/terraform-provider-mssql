package mssql

import (
  "context"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "strconv"
  "strings"
)

func dataSourceDatabase() *schema.Resource {
  return &schema.Resource{
    ReadContext: dataSourceDatabaseRead,
    Schema: map[string]*schema.Schema{
      "server_name": {
        Type:     schema.TypeString,
        Required: true,
      },
      "server_fqdn": {
        Type:     schema.TypeString,
        Optional: true,
      },
      "server_port": {
        Type:     schema.TypeString,
        Optional: true,
        Default:  1433,
      },
      "database_name": {
        Type:     schema.TypeString,
        Required: true,
      },
      "administrator_login": {
        Type:         schema.TypeList,
        MaxItems:     1,
        Optional:     true,
        ExactlyOneOf: []string{"administrator_login", "azure_administrator"},
        Elem: &schema.Resource{
          Schema: map[string]*schema.Schema{
            "username": {
              Type:     schema.TypeString,
              Required: true,
            },
            "password": {
              Type:      schema.TypeString,
              Required:  true,
              Sensitive: true,
            },
          },
        },
      },
      "azure_administrator": {
        Type:         schema.TypeList,
        MaxItems:     1,
        Optional:     true,
        ExactlyOneOf: []string{"administrator_login", "azure_administrator"},
        Elem: &schema.Resource{
          Schema: map[string]*schema.Schema{
            "tenant_id": {
              Type:        schema.TypeString,
              Required:    true,
              DefaultFunc: schema.EnvDefaultFunc("ARM_TENANT_ID", nil),
            },
            "client_id": {
              Type:        schema.TypeString,
              Required:    true,
              DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_ID", nil),
            },
            "client_secret": {
              Type:        schema.TypeString,
              Required:    true,
              Sensitive:   true,
              DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_SECRET", nil),
            },
          },
        },
      },
    },
  }
}

func dataSourceDatabaseRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  //connectors := meta.(map[string]*Connector)

  // Warnings or errors can be collected in a slice type
  var diags diag.Diagnostics

  port, err := strconv.Atoi(data.Get("server_port").(string))
  if err != nil {
    return diag.FromErr(err)
  }

  connector := Connector{
    Host:     data.Get("server_name").(string),
    Port:     port,
    Database: data.Get("database_name").(string),
  }

  if admin, ok := data.GetOk("administrator_login.0"); ok {
    admin := admin.(map[string]interface{})
    connector.Administrator = &struct {
      Username string
      Password string
    }{
      Username: admin["username"].(string),
      Password: admin["password"].(string),
    }
  }

  if admin, ok := data.GetOk("azure_administrator.0"); ok {
    admin := admin.(map[string]interface{})

    connector.AzureAdministrator = &struct {
      TenantID     string
      ClientID     string
      ClientSecret string
    }{
      TenantID:     admin["tenant_id"].(string),
      ClientID:     admin["client_id"].(string),
      ClientSecret: admin["client_secret"].(string),
    }
  }

  if fqdn, fqdnOk := data.GetOk("server_fqdn"); fqdnOk {
    connector.Host = fqdn.(string)
  } else if connector.AzureAdministrator != nil && !strings.HasSuffix(connector.Host, ".database.windows.net") {
    connector.Host = connector.Host + ".database.windows.net"
    data.Set("server_fqdn", connector.Host)
  }

  connectors := meta.(map[string]*Connector)
  connectors[connector.ID()] = &connector

  data.SetId(connector.ID())

  return diags
}
