package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIP4BlockResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccIP4BlockResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("bluecat_ip4_block.test", "id", validateObjectID),
					resource.TestCheckResourceAttr("bluecat_ip4_block.test", "name", "Test IPv4 Block"),
				),
			},
		},
	})
}

const testAccIP4BlockResourceConfig = testAccEntityDataSourceConfig + `
variable "ip4_block_parent_id" {
  type = number
}

resource "bluecat_ip4_block" "test" {
	parent_id = var.ip4_block_parent_id
	name      = "Test IPv4 Block"
	size      = 256
  }
`
