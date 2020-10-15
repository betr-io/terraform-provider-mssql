package mssql

import (
  "context"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "io"
  "strconv"
  "time"
)

func dataSourceRoles() *schema.Resource {
  return &schema.Resource{
    ReadContext: dataSourceRolesRead,
    Schema: map[string]*schema.Schema{
      "database_id": {
        Type:     schema.TypeString,
        Required: true,
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
  // Warnings or errors can be collected in a slice type
  var diags diag.Diagnostics

  id := data.Get("database_id").(string)
  c := meta.(map[string]*Connector)[id]

  rows, err := c.QueryContext(ctx, "SELECT uid, name FROM sys.sysusers WHERE issqlrole = 1")
  if err != nil {
    return diag.FromErr(err)
  }
  defer checkClose(rows, &diags)

  roles := make([]map[string]interface{}, 0)
  for rows.Next() {
    var id int
    var name string
    err := rows.Scan(&id, &name)
    if err != nil {
      return diag.FromErr(err)
    }

    roles = append(roles, map[string]interface{}{"id": id, "name": name})
  }

  if err := data.Set("roles", roles); err != nil {
    return diag.FromErr(err)
  }

  // always run
  data.SetId(strconv.FormatInt(time.Now().Unix(), 10))

  return diags
}

func checkClose(c io.Closer, diags *diag.Diagnostics) {
  if err := c.Close(); err != nil {
    *diags = append(*diags, diag.FromErr(err)[0])
  }
}
