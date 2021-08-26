resource "bluecat_ip4_address" "addr" {
  configuration_id = data.bluecat_entity.config.id
  name             = "IP Reserved for Example"
  parent_id        = data.bluecat_ip4_network.example_net.id
}

output "allocated_address" {
  value = bluecat_ip4_address.addr.address
}
