package mssql

import (
  "context"
  "database/sql"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/pkg/errors"
  "strconv"
  "terraform-provider-mssql/mssql/model"
  "time"
)

type RawConnector interface {
  GetDatabase() string
  SetDatabase(database string)
  QueryContext(ctx context.Context, query string, scanner func(*sql.Rows) error, args ...interface{}) error
}

func dataSourceRoles() *schema.Resource {
  return &schema.Resource{
    ReadContext: dataSourceRolesRead,
    Schema: map[string]*schema.Schema{
      "server": {
        Type:         schema.TypeList,
        MaxItems:     1,
        Optional:     true,
        ExactlyOneOf: []string{"server", "server_encoded"},
        Elem: &schema.Resource{
          Schema: getServerSchema("server", true, nil),
        },
      },
      "server_encoded": {
        Type:         schema.TypeString,
        Optional:     true,
        Sensitive:    true,
        ExactlyOneOf: []string{"server", "server_encoded"},
      },
      "database": {
        Type:     schema.TypeString,
        Optional: true,
        Default:  "master",
      },
      "roles": {
        Type:     schema.TypeList,
        Computed: true,
        Elem: &schema.Resource{
          Schema: map[string]*schema.Schema{
            "id": {
              Type:     schema.TypeInt,
              Computed: true,
            },
            "name": {
              Type:     schema.TypeString,
              Computed: true,
            },
          },
        },
      },
    },
  }
}

func dataSourceRolesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := meta.(model.Provider).DataSourceLogger("roles", "read")
  logger.Debug().Msgf("read %s", data.Id())

  connector, err := getRawConnector(meta, "server", data)
  if err != nil {
    return diag.FromErr(err)
  }
  connector.SetDatabase(data.Get("database").(string))

  roles := make([]map[string]interface{}, 0)
  err = connector.QueryContext(ctx, "SELECT uid, name FROM [sys].[sysusers] WHERE [issqlrole] = 1", func(r *sql.Rows) error {
    for r.Next() {
      var id int64
      var name string
      err := r.Scan(&id, &name)
      if err != nil {
        return err
      }
      roles = append(roles, map[string]interface{}{"id": id, "name": name})
    }
    return nil
  })

  if err != nil {
    return diag.FromErr(errors.Wrap(err, "RolesRead"))
  }

  if err := data.Set("roles", roles); err != nil {
    return diag.FromErr(err)
  }

  // always run
  data.SetId(strconv.FormatInt(time.Now().Unix(), 10))

  return nil
}

func getRawConnector(meta interface{}, prefix string, data *schema.ResourceData) (RawConnector, error) {
  provider := meta.(model.Provider)
  connector, err := provider.GetConnector(prefix, data)
  if err != nil {
    return nil, err
  }
  return connector.(RawConnector), nil
}
