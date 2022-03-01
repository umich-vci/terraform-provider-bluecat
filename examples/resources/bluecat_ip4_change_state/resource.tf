resource "bluecat_ip4_change_state" "ip4_action" {
  address_id  = data.bluecat_ip4_address.addr.id
  action      = "MAKE_DHCP_RESERVED"
  mac_address = data.bluecat_ip4_address.addr.mac_address
}

output "ip4_final_state" {
  value = bluecat_ip4_change_state.ip4_action.state
}