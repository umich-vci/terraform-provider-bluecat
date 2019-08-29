package bluecat

import (
	"github.com/umich-vci/golang-bluecat"
)

// Config holds the provider configuration
type Config struct {
	Username        string
	Password        string
	BlueCatEndpoint string
}

// Client returns a new client for accessing BlueCat Address Manager
func (c *Config) Client() (bam.ProteusAPI, error) {
	client, err := bam.Client(c.Username, c.Password, c.BlueCatEndpoint)

	return client, err
}
