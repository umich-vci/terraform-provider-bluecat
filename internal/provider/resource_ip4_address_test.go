package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIP4AddressResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckEnvVars(t, "TF_VAR_config_name", "TF_VAR_ip4_network_parent_id", "TF_VAR_host_record_dns_zone") },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIP4AddressResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("bluecat_ip4_address.test", "id", validateObjectID),
					resource.TestCheckResourceAttrSet("bluecat_ip4_address.test", "name"),
				),
			},
		},
	})
}

const testAccIP4AddressResourceConfig = testAccIP4NetworkResourceConfig + `
variable "host_record_dns_zone" {
  type = string
}

resource "bluecat_ip4_address" "test" {
  configuration_id = data.bluecat_entity.config.id
  parent_id        = bluecat_ip4_network.test.id
  name             = "tfacc-host-record.${var.host_record_dns_zone}"
}
`
