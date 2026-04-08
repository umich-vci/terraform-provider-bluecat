package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIP4AvailableNetworkResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckEnvVars(t, "TF_VAR_config_name", "TF_VAR_ip4_network_parent_id") },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIP4AvailableNetworkResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bluecat_ip4_available_network.test", "id", "-"),
					resource.TestCheckResourceAttrSet("bluecat_ip4_available_network.test", "network_id"),
				),
			},
		},
	})
}

const testAccIP4AvailableNetworkResourceConfig = testAccIP4NetworkResourceConfig + `
resource "bluecat_ip4_available_network" "test" {
  network_id_list = [bluecat_ip4_network.test.id]
}
`
