package bluecat

import "github.com/umich-vci/gobam"

// Config holds the provider configuration
type Config struct {
	Username        string
	Password        string
	BlueCatEndpoint string
	SSLVerify       bool
}

// Client returns a new client for accessing BlueCat Address Manager
func (c *Config) Client() (gobam.ProteusAPI, error) {
	client, err := gobam.Client(c.Username, c.Password, c.BlueCatEndpoint, c.SSLVerify)

	return client, err
}
