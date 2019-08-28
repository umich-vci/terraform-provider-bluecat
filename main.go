package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/umich-vci/terraform-provider-bluecat/bluecat"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: bluecat.Provider})
}
