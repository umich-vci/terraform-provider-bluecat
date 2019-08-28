package bluecat

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/tiaguinho/gosoap"
)

// Config holds the provider configuration
type Config struct {
	Username        string
	Password        string
	BlueCatEndpoint string
}

// NewConfig returns a new Config from a supplied ResourceData.
func NewConfig(d *schema.ResourceData) (*Config, error) {
	c := &Config{
		Username:        d.Get("username").(string),
		Password:        d.Get("password").(string),
		BlueCatEndpoint: d.Get("bluecat_endpoint").(string),
	}

	return c, nil
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	c, err := NewConfig(d)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Client returns a new client for accessing BlueCat.
func (c *Config) Client() (*gosoap.Client, error) {
	wdsl := "https://" + c.BlueCatEndpoint + "/Services/API?wsdl"
	client, err := gosoap.SoapClient(wdsl)

	if err != nil {
		log.Fatalf("SoapClient error: %s", err)
	}

	params := gosoap.Params{
		"username": c.Username,
		"password": c.Password,
	}

	_, err = client.Call("login", params)
	if err != nil {
		log.Fatalf("SoapClient login error: %s", err)
	}

	return client, nil
}
