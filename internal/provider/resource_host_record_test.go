package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccHostRecordResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccCheckEnvVars(t, "TF_VAR_config_name", "TF_VAR_ip4_network_parent_id", "TF_VAR_host_record_dns_zone", "TF_VAR_dns_view_name")
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHostRecordResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("bluecat_host_record.test", "id", validateObjectID),
					resource.TestCheckResourceAttr("bluecat_host_record.test", "name", "tfacc-host-record"),
				),
			},
		},
	})
}

const testAccDNSViewDataSourceConfig = `
variable "dns_view_name" {
  type = string
}

data "bluecat_entity" "dns_view" {
  name      = var.dns_view_name
  parent_id = data.bluecat_entity.config.id
  type      = "View"
}
`

const testAccHostRecordResourceConfig = testAccIP4AddressResourceConfig + testAccDNSViewDataSourceConfig + `
resource "bluecat_host_record" "test" {
  name      = "tfacc-host-record"
  dns_zone  = var.host_record_dns_zone
  view_id   = data.bluecat_entity.dns_view.id
  addresses = [bluecat_ip4_address.test.address]
}
`
