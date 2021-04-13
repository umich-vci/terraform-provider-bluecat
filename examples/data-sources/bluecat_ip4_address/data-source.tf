data "bluecat_ip4_address" "addr" {
    container_id = data.bluecat_entity.config.id
    address = "192.168.1.1"
}

output "bluecat_address_mac" {
    value = data.bluecat_ip4_address.addr.mac_address
}
