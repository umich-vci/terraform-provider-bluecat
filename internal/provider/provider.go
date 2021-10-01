package provider

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/umich-vci/gobam"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		desc := s.Description
		if s.Default != nil {
			desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
		}
		return strings.TrimSpace(desc)
	}
}

//New instance of the BlueCat Terraform Provider
func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"bluecat_endpoint": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("BLUECAT_ENDPOINT", nil),
					Description: "The BlueCat Address Manager endpoint hostname. Can also use the environment variable `BLUECAT_ENDPOINT`",
				},
				"password": {
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("BLUECAT_PASSWORD", nil),
					Description: "The BlueCat Address Manager password. Can also use the environment variable `BLUECAT_PASSWORD`",
				},
				"username": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("BLUECAT_USERNAME", nil),
					Description: "A BlueCat Address Manager username. Can also use the environment variable `BLUECAT_USERNAME`",
				},
				"ssl_verify": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     true,
					Description: "Verify the SSL certificate of the BlueCat Address Manager endpoint?",
				},
			},

			DataSourcesMap: map[string]*schema.Resource{
				"bluecat_entity":                  dataSourceEntity(),
				"bluecat_host_record":             dataSourceHostRecord(),
				"bluecat_ip4_address":             dataSourceIP4Address(),
				"bluecat_ip4_network":             dataSourceIP4Network(),
				"bluecat_ip4_network-block-range": dataSourceIP4NBR(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"bluecat_host_record":           resourceHostRecord(),
				"bluecat_ip4_address":           resourceIP4Address(),
				"bluecat_ip4_available_network": resourceIP4AvailableNetwork(),
				"bluecat_ip4_network":           resourceIP4Network(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

type apiClient struct {
	Client   gobam.ProteusAPI
	Username string
	Password string
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {

		username := d.Get("username").(string)
		password := d.Get("password").(string)
		endpoint := d.Get("bluecat_endpoint").(string)
		sslVerify := d.Get("ssl_verify").(bool)

		client := gobam.NewClient(endpoint, sslVerify)

		return &apiClient{Client: client, Username: username, Password: password}, nil
	}
}

var mutex = &sync.Mutex{}

// Config holds the provider configuration
type Config struct {
	Username        string
	Password        string
	BlueCatEndpoint string
	SSLVerify       bool
}
