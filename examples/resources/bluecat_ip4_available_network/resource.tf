resource "bluecat_ip4_available_network" "network" {
  network_id_list = [1234, 5678, 9101]
}

output "network_id" {
  value = bluecat_ip4_available_network.network.network_id
}
