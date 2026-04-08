package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIP4NetworkResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckEnvVars(t, "TF_VAR_config_name", "TF_VAR_ip4_network_parent_id") },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIP4NetworkResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("bluecat_ip4_network.test", "id", validateObjectID),
					resource.TestCheckResourceAttr("bluecat_ip4_network.test", "name", "Terraform Acceptance Test - Network"),
				),
			},
		},
	})
}

const testAccIP4NetworkResourceConfig = testAccEntityDataSourceConfig + `
variable "ip4_network_parent_id" {
  type = number
}

resource "bluecat_ip4_network" "test" {
  parent_id = var.ip4_network_parent_id
  name      = "Terraform Acceptance Test - Network"
  size      = 256
}
`
