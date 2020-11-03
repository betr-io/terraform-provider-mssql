package mssql

import (
  "context"
  sql2 "database/sql"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/pkg/errors"
  "strconv"
  "time"
)

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
  logger := meta.(Provider).logger.With().Str("datasource", "roles").Str("func", "read").Logger()
  logger.Debug().Msgf("read %s", data.Id())

  connector, err := GetConnector("server", data)
  if err != nil {
    return diag.FromErr(err)
  }
  connector.Database = data.Get("database").(string)

  roles := make([]map[string]interface{}, 0)
  err = connector.QueryContext(ctx, "SELECT uid, name FROM [sys].[sysusers] WHERE [issqlrole] = 1", func(r *sql2.Rows) error {
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

  logger.Info().Msgf("Token = %s", connector.Token)

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
