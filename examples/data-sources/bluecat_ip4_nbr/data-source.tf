data "bluecat_ip4_nbr" "network" {
  container_id = data.bluecat_entity.config.id
  address      = "192.168.1.1"
  type         = "IP4Network"
}

output "bluecat_network_name" {
  value = data.bluecat_ip4_nbr.network.name
}
