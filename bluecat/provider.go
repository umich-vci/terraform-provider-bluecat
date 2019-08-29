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
			"bluecat_static_ipv4_address": resourceStaticIPv4Address(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"bluecat_entity_by_name":      dataSourceEntityByName(),
			"bluecat_ip4_network_by_name": dataSourceIP4NetworkByName(),
		},
		ConfigureFunc: providerConfigure,
	}
}

var mutex = &sync.Mutex{}
