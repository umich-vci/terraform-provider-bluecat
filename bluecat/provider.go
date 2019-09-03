package bluecat

import (
	"sync"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := &Config{
		Username:        d.Get("username").(string),
		Password:        d.Get("password").(string),
		BlueCatEndpoint: d.Get("bluecat_endpoint").(string),
	}

	return config, nil
}

// Provider returns a terraform resource provider
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("BLUECAT_USERNAME", nil),
				Description: "A BlueCat Address Manager username.",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("BLUECAT_PASSWORD", nil),
				Description: "The BlueCat Address Manager password.",
			},
			"bluecat_endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("BLUECAT_ENDPOINT", nil),
				Description: "The BlueCat Address Manager endpoint hostname",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"bluecat_ip4_address": resourceIP4Address(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"bluecat_entity":      dataSourceEntity(),
			"bluecat_host_record": dataSourceHostRecord(),
			"bluecat_ip4_network": dataSourceIP4Network(),
			"bluecat_ip4_address": dataSourceIP4Address(),
		},
		ConfigureFunc: providerConfigure,
	}
}

var mutex = &sync.Mutex{}
