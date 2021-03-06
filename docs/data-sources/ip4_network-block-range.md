# bluecat\_ip4\_network-block-range Data Source

Use this data source to access the attributes of an IPv4 network, IPv4 Block, or DHCPv4 Range.

## Example Usage

```hcl
data "bluecat_ip4_network-block-range" "network" {
    container_id = data.bluecat_entity.config.id
    address = "192.168.1.1"
    type = "IP4Network"
}

output "bluecat_network_name" {
    value = data.bluecat_ip4_network-block-range.network.name
}
```

## Argument Reference

* `container_id` - (Required) The object ID of a container that contains the specified IPv4
  network, block, or range.

* `address` - (Required) IP address to find the IPv4 network, IPv4 Block, or DHCPv4 Range of.

* `type` - (Required) Must be "IP4Block", "IP4Network", "DHCP4Range", or "".
  "" will find the most specific container.

## Attributes Reference

The atributes returned will vary based on the object returned.

* `name` - The name assigned the resource.

* `properties` - The properties of the resource as returned by the API (pipe delimited).

* `cidr` - The CIDR address of the IPv4 network.

* `allow_duplicate_host` - Duplicate host names check.

* `inherit_allow_duplicate_host` - Duplicate host names check is inherited.

* `ping_before_assign` - The network pings an address before assignment.

* `inherit_ping_before_assign` - The network pings an address before assignment is inherited.

* `gateway` - The gateway of the IPv4 network.

* `inherit_default_domains` - Default domains are inherited.

* `default_view` - The object id of the default DNS View for the network.

* `inherit_default_view` - The default DNS Viewis inherited.

* `inherit_dns_restrictions` - DNS restrictions are inherited.

* `addresses_in_use` - The number of addresses allocated/in use on the network.

* `addresses_free` - The number of addresses unallocated/free on the network.

* `custom_properties` - A map of all custom properties associated with the IPv4 network.
