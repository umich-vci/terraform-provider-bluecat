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
resource "bluecat_ip4_block" "test" {
	parent_id = data.bluecat_entity.config.id
	name      = "Test IPv4 Block"
	size      = 256
  }
`
