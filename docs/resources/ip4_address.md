# bluecat\_ip4\_address Resource

Use this resource to reserve an IPv4 address.

## Example Usage

```hcl
resource "bluecat_ip4_address" "addr" {
    configuration_id = data.bluecat_entity.config.id
    name = "IP Reserved for Example"
    parent_id = data.bluecat_ip4_network.example_net.id
}

output "allocated_address" {
    value = bluecat_ip4_address.addr.address
}
```

## Argument Reference

* `configuration_id` - (Required) The object ID of the Configuration that will hold the new address.

* `parent_id` - (Optional) The object ID of the Configuration, Block, or Network to find the next available
  IPv4 address in.  If changed, forces a new resource.  If not set, `parent_id_list` is required.

* `parent_id_list` - (Optional) A list of object IDs of the Configuration, Block, or Network to find the next available
  IPv4 address in.  The list will be parsed and the network with the most available addresses will be selected.
  The list is only used at object creation, so it might be beneficial to use a `lifecycle` customization to ignore
  changes to the list after resource creation.  If not set, `parent_id` is required.

* `name` - (Required) The name assigned to the IPv4 address.  This is not related to DNS.
  
* `mac_address` - (Optional) The MAC address to associate with the IPv4 address.

* `action` - (Optional) The action to take on the next available IPv4 address.  Must be one of:
  MAKE_STATIC, MAKE_RESERVED, or MAKE_DHCP_RESERVED.  Defaults to MAKE_STATIC.

* `custom_properties` - (Optional) A map of all custom properties associated with the IPv4 address.

## Attributes Reference

* `address` -  The IPv4 address that was allocated.

* `properties` -  The properties of the IPv4 address as returned by the API (pipe delimited).

* `state` - The state of the IPv4 address.

* `type` - The type of the resource.

* `computed_parent_id` - The ID network that was selected if `parent_id_list` was used to allocate the network.
  Will contain the same value as `parent_id` if that is used instead.
