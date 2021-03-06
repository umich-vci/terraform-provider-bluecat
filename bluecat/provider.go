package bluecat

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := &Config{
		Username:        d.Get("username").(string),
		Password:        d.Get("password").(string),
		BlueCatEndpoint: d.Get("bluecat_endpoint").(string),
		SSLVerify:       d.Get("ssl_verify").(bool),
	}

	return config, nil
}

// Provider returns a terraform resource provider
func Provider() *schema.Provider {
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
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("BLUECAT_PASSWORD", nil),
				Description: "The BlueCat Address Manager password.",
			},
			"bluecat_endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("BLUECAT_ENDPOINT", nil),
				Description: "The BlueCat Address Manager endpoint hostname",
			},
			"ssl_verify": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Verify the SSL certificate of the BlueCat Address Manager endpoint",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"bluecat_host_record":           resourceHostRecord(),
			"bluecat_ip4_address":           resourceIP4Address(),
			"bluecat_ip4_available_network": resourceIP4AvailableNetwork(),
			"bluecat_ip4_network":           resourceIP4Network(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"bluecat_entity":                  dataSourceEntity(),
			"bluecat_host_record":             dataSourceHostRecord(),
			"bluecat_ip4_address":             dataSourceIP4Address(),
			"bluecat_ip4_network-block-range": dataSourceIP4Network(),
		},
		ConfigureFunc: providerConfigure,
	}
}

var mutex = &sync.Mutex{}
