package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIP4NBRDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckEnvVars(t, "TF_VAR_config_name", "TF_VAR_ip4_address") },
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

func TestAccIP4NBRDataSourceCIDR(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccCheckEnvVars(t, "TF_VAR_config_name", "TF_VAR_ip4_address")
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIP4NBRDataSourceCIDRConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("data.bluecat_ip4_nbr.test_cidr", "id", validateObjectID),
					// both lookups should resolve to the same network
					resource.TestCheckResourceAttrPair("data.bluecat_ip4_nbr.test_cidr", "id", "data.bluecat_ip4_nbr.by_address", "id"),
				),
			},
		},
	})
}

func TestAccIP4NBRDataSourceCIDRRangeError(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccIP4NBRDataSourceCIDRRangeConfig,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("Invalid type for CIDR lookup"),
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

const testAccIP4NBRDataSourceCIDRConfig = testAccEntityDataSourceConfig + `
variable "ip4_address" {
	type = string
}

# Look up the network by address to discover its CIDR
data "bluecat_ip4_nbr" "by_address" {
	container_id = data.bluecat_entity.config.id
	address      = var.ip4_address
	type         = "IP4Network"
}

# Look up the parent block by address to use as container_id for the CIDR lookup
data "bluecat_ip4_nbr" "parent_block" {
	container_id = data.bluecat_entity.config.id
	address      = var.ip4_address
	type         = "IP4Block"
}

# Look up the same network by CIDR using the direct parent block
data "bluecat_ip4_nbr" "test_cidr" {
	container_id = data.bluecat_ip4_nbr.parent_block.id
	cidr         = data.bluecat_ip4_nbr.by_address.cidr
	type         = "IP4Network"
}
`

const testAccIP4NBRDataSourceCIDRRangeConfig = `
data "bluecat_ip4_nbr" "test_range_error" {
	container_id = 1
	cidr         = "10.0.0.0/24"
	type         = "DHCP4Range"
}
`
