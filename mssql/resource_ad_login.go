package mssql

import (
	"context"
	"github.com/betr-io/terraform-provider-mssql/mssql/model"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"strings"
)

const AadLoginNameProp = "aad_login_name"

type AadLoginConnector interface {
	CreateAadLogin(ctx context.Context, name, defaultDatabase, defaultLanguage string) error
	GetAadLogin(ctx context.Context, name string) (*model.AadLogin, error)
	DeleteAadLogin(ctx context.Context, name string) error
}

func resourceAadLogin() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAadLoginCreate,
		ReadContext:   resourceAadLoginRead,
		DeleteContext: resourceAadLoginDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAadLoginImport,
		},
		Schema: map[string]*schema.Schema{
			serverProp: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: getServerSchema(serverProp),
				},
			},
			AadLoginNameProp: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			defaultDatabaseProp: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  defaultDatabaseDefault,
				DiffSuppressFunc: func(k, old, new string, data *schema.ResourceData) bool {
					return (old == "" && new == defaultDatabaseDefault) || (old == defaultDatabaseDefault && new == "")
				},
			},
			defaultLanguageProp: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return (old == "" && new == "us_english") || (old == "us_english" && new == "")
				},
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Default: defaultTimeout,
		},
	}
}

func resourceAadLoginCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "AadLogin", "create")
	logger.Debug().Msgf("Create %s", getAadLoginID(data))

	AadLoginName := data.Get(AadLoginNameProp).(string)
	defaultDatabase := data.Get(defaultDatabaseProp).(string)
	defaultLanguage := data.Get(defaultLanguageProp).(string)

	connector, err := GetSqlConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.CreateAadLogin(ctx, AadLoginName, defaultDatabase, defaultLanguage); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to create AadLogin [%s]", AadLoginName))
	}

	data.SetId(getAadLoginID(data))

	logger.Info().Msgf("created AadLogin [%s]", AadLoginName)

	return resourceAadLoginRead(ctx, data, meta)
}

func resourceAadLoginRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "AadLogin", "read")
	logger.Debug().Msgf("Read %s", getAadLoginID(data))

	AadLoginName := data.Get(AadLoginNameProp).(string)

	connector, err := GetSqlConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	AadLogin, err := connector.GetAadLogin(ctx, AadLoginName)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to read AadLogin [%s]", AadLoginName))
	}
	if AadLogin == nil {
		logger.Info().Msgf("No AadLogin found for [%s]", AadLoginName)
		data.SetId("")
	} else {
		if err = data.Set(defaultDatabaseProp, AadLogin.DefaultDatabase); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(defaultLanguageProp, AadLogin.DefaultLanguage); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceAadLoginDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "AadLogin", "delete")
	logger.Debug().Msgf("Delete %s", data.Id())

	AadLoginName := data.Get(AadLoginNameProp).(string)

	connector, err := GetSqlConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.DeleteAadLogin(ctx, AadLoginName); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to delete AadLogin [%s]", AadLoginName))
	}

	logger.Info().Msgf("deleted AadLogin [%s]", AadLoginName)

	// d.SetId("") is automatically called assuming delete returns no errors, but it is added here for explicitness.
	data.SetId("")

	return nil
}

func resourceAadLoginImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	logger := loggerFromMeta(meta, "AadLogin", "import")
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
	if err = data.Set(AadLoginNameProp, parts[1]); err != nil {
		return nil, err
	}

	data.SetId(getAadLoginID(data))

	AadLoginName := data.Get(AadLoginNameProp).(string)

	connector, err := GetSqlConnector(meta, data)
	if err != nil {
		return nil, err
	}

	AadLogin, err := connector.GetAadLogin(ctx, AadLoginName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read AadLogin [%s] for import", AadLoginName)
	}

	if AadLogin == nil {
		return nil, errors.Errorf("no AadLogin [%s] found for import", AadLoginName)
	}

	if err = data.Set(defaultDatabaseProp, AadLogin.DefaultDatabase); err != nil {
		return nil, err
	}
	if err = data.Set(defaultLanguageProp, AadLogin.DefaultLanguage); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{data}, nil
}

func GetSqlConnector(meta interface{}, data *schema.ResourceData) (AadLoginConnector, error) {
	provider := meta.(model.Provider)
	connector, err := provider.GetConnector(serverProp, data)
	if err != nil {
		return nil, err
	}
	return connector.(AadLoginConnector), nil
}
