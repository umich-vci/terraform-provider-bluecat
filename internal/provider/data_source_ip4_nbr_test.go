package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIP4NBRDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccIP4NBRDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("data.bluecat_ip4_nbr.test", "id", validateObjectID),
				),
			},
		},
	})
}

const testAccIP4NBRDataSourceConfig = testAccEntityDataSourceConfig + `
variable "ip4_address" {
	type = string
}

data "bluecat_ip4_nbr" "test" {
	container_id = data.bluecat_entity.config.id
	address      = var.ip4_address
	type         = "IP4Network"
  }
`
