resource "bluecat_ip4_block" "block" {
  parent_id = data.bluecat_ip4_network-block-range.block.id
  name      = "New Block"
  size      = 256
}

output "bluecat_ip4_block_cidr" {
  value = bluecat_ip4_block.block.cidr
}
