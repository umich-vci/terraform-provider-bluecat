// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccHostRecordDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccHostRecordDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("data.bluecat_host_record.test", "id", validateObjectID),
				),
			},
		},
	})
}

const testAccHostRecordDataSourceConfig = `
variable "absolute_name" {
	type = string
}

data "bluecat_host_record" "test" {
	absolute_name = var.absolute_name
}
`
