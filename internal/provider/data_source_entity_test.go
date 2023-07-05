// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEntityDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccEntityDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("data.bluecat_entity.config", "id", validateObjectID),
				),
			},
		},
	})
}

const testAccEntityDataSourceConfig = `
variable "config_name" {
	type = string
}

data "bluecat_entity" "config" {
	name      = var.config_name
	parent_id = 0
	type      = "Configuration"
}
`
