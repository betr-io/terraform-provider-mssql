package model

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

type ConnectorFactory interface {
	GetConnector(prefix string, data *schema.ResourceData) (interface{}, error)
}
