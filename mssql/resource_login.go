package mssql

import (
  "context"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/pkg/errors"
  "strings"
  "github.com/betr-io/terraform-provider-mssql/mssql/model"
)

const loginNameProp = "login_name"
const defaultDatabaseProp = "default_database"
const defaultDatabaseDefault = "master"
const defaultLanguageProp = "default_language"

type LoginConnector interface {
  CreateLogin(ctx context.Context, name, password, defaultDatabase, defaultLanguage string) error
  GetLogin(ctx context.Context, name string) (*model.Login, error)
  UpdateLogin(ctx context.Context, name, password, defaultDatabase, defaultLanguage string) error
  DeleteLogin(ctx context.Context, name string) error
}

func resourceLogin() *schema.Resource {
  return &schema.Resource{
    CreateContext: resourceLoginCreate,
    ReadContext:   resourceLoginRead,
    UpdateContext: resourceLoginUpdate,
    DeleteContext: resourceLoginDelete,
    Importer: &schema.ResourceImporter{
      StateContext: resourceLoginImport,
    },
    Schema: map[string]*schema.Schema{
      serverProp: {
        Type:         schema.TypeList,
        MaxItems:     1,
        Required:     true,
        Elem: &schema.Resource{
          Schema: getServerSchema(serverProp),
        },
      },
      loginNameProp: {
        Type:     schema.TypeString,
        Required: true,
        ForceNew: true,
      },
      passwordProp: {
        Type:      schema.TypeString,
        Required:  true,
        Sensitive: true,
      },
      defaultDatabaseProp: {
        Type:     schema.TypeString,
        Optional: true,
        Default:  defaultDatabaseDefault,
        DiffSuppressFunc: func(k, old, new string, data *schema.ResourceData) bool {
          return (old == "" && new == defaultDatabaseDefault) || (old == defaultDatabaseDefault && new == "")
        },
      },
      defaultLanguageProp: {
        Type:     schema.TypeString,
        Optional: true,
        DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
          return (old == "" && new == "us_english") || (old == "us_english" && new == "")
        },
      },
      principalIdProp: {
        Type:     schema.TypeInt,
        Computed: true,
      },
    },
    Timeouts: &schema.ResourceTimeout{
      Default: defaultTimeout,
    },
  }
}

func resourceLoginCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "login", "create")
  logger.Debug().Msgf("Create %s", getLoginID(data))

  loginName := data.Get(loginNameProp).(string)
  password := data.Get(passwordProp).(string)
  defaultDatabase := data.Get(defaultDatabaseProp).(string)
  defaultLanguage := data.Get(defaultLanguageProp).(string)

  connector, err := getLoginConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  if err = connector.CreateLogin(ctx, loginName, password, defaultDatabase, defaultLanguage); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to create login [%s]", loginName))
  }

  data.SetId(getLoginID(data))

  logger.Info().Msgf("created login [%s]", loginName)

  return resourceLoginRead(ctx, data, meta)
}

func resourceLoginRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "login", "read")
  logger.Debug().Msgf("Read %s", getLoginID(data))

  loginName := data.Get(loginNameProp).(string)

  connector, err := getLoginConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  login, err := connector.GetLogin(ctx, loginName)
  if err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to read login [%s]", loginName))
  }
  if login == nil {
    logger.Info().Msgf("No login found for [%s]", loginName)
    data.SetId("")
  } else {
    if err = data.Set(principalIdProp, login.PrincipalID); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(defaultDatabaseProp, login.DefaultDatabase); err != nil {
      return diag.FromErr(err)
    }
    if err = data.Set(defaultLanguageProp, login.DefaultLanguage); err != nil {
      return diag.FromErr(err)
    }
  }

  return nil
}

func resourceLoginUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "login", "update")
  logger.Debug().Msgf("Update %s", data.Id())

  loginName := data.Get(loginNameProp).(string)
  password := data.Get(passwordProp).(string)
  defaultDatabase := data.Get(defaultDatabaseProp).(string)
  defaultLanguage := data.Get(defaultLanguageProp).(string)

  connector, err := getLoginConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  if err = connector.UpdateLogin(ctx, loginName, password, defaultDatabase, defaultLanguage); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to update login [%s]", loginName))
  }

  logger.Info().Msgf("updated login [%s]", loginName)

  return resourceLoginRead(ctx, data, meta)
}

func resourceLoginDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  logger := loggerFromMeta(meta, "login", "delete")
  logger.Debug().Msgf("Delete %s", data.Id())

  loginName := data.Get(loginNameProp).(string)

  connector, err := getLoginConnector(meta, data)
  if err != nil {
    return diag.FromErr(err)
  }

  if err = connector.DeleteLogin(ctx, loginName); err != nil {
    return diag.FromErr(errors.Wrapf(err, "unable to delete login [%s]", loginName))
  }

  logger.Info().Msgf("deleted login [%s]", loginName)

  // d.SetId("") is automatically called assuming delete returns no errors, but it is added here for explicitness.
  data.SetId("")

  return nil
}

func resourceLoginImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
  logger := loggerFromMeta(meta, "login", "import")
  logger.Debug().Msgf("Import %s", data.Id())

  server, u, err := serverFromId(data.Id())
  if err != nil {
    return nil, err
  }
  if err = data.Set(serverProp, server); err != nil {
    return nil, err
  }

  parts := strings.Split(u.Path, "/")
  if len(parts) != 2 {
    return nil, errors.New("invalid ID")
  }
  if err = data.Set(loginNameProp, parts[1]); err != nil {
    return nil, err
  }

  data.SetId(getLoginID(data))

  loginName := data.Get(loginNameProp).(string)

  connector, err := getLoginConnector(meta, data)
  if err != nil {
    return nil, err
  }

  login, err := connector.GetLogin(ctx, loginName)
  if err != nil {
    return nil, errors.Wrapf(err, "unable to read login [%s] for import", loginName)
  }

  if login == nil {
    return nil, errors.Errorf("no login [%s] found for import", loginName)
  }

  if err = data.Set(principalIdProp, login.PrincipalID); err != nil {
    return nil, err
  }
  if err = data.Set(defaultDatabaseProp, login.DefaultDatabase); err != nil {
    return nil, err
  }
  if err = data.Set(defaultLanguageProp, login.DefaultLanguage); err != nil {
    return nil, err
  }

  return []*schema.ResourceData{data}, nil
}

func getLoginConnector(meta interface{}, data *schema.ResourceData) (LoginConnector, error) {
  provider := meta.(model.Provider)
  connector, err := provider.GetConnector(serverProp, data)
  if err != nil {
    return nil, err
  }
  return connector.(LoginConnector), nil
}
