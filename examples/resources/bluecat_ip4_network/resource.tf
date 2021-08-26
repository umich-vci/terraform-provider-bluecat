resource "bluecat_ip4_network" "network" {
  parent_id = data.bluecat_ip4_network-block-range.block.id
  name      = "New Network"
  size      = 256
}

output "bluecat_ip4_network_cidr" {
  value = bluecat_ip4_network.network.cidr
}
