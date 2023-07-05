// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strconv"
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
					resource.TestCheckResourceAttrWith("data.bluecat_entity.config", "id", func(value string) error {
						valueInt, err := strconv.Atoi(value)
						if err != nil {
							return err
						}

						if valueInt <= 0 {
							return fmt.Errorf("should be a value greater than 0")
						}
						return nil
					}),
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
