# bluecat\_ip4\_address Data Source

Use this data source to access the attributes of an IPv4 address.

## Example Usage

```hcl
data "bluecat_ip4_address" "addr" {
    container_id = data.bluecat_entity.config.id
    address = "192.168.1.1"
}

output "bluecat_address_mac" {
    value = data.bluecat_ip4_address.addr.mac_address
}
```

## Argument Reference

* `container_id` - (Required) The object ID of the container that has the specified `address`.  This can be a
  Configuration, IPv4 Block, IPv4 Network, or DHCP range.

* `address` - (Required) The IPv4 address to get data for.

## Attribute Reference

* `name` - The name assigned to the IPv4 address.  This is not related to DNS.

* `properties` -  The properties of the IPv4 address as returned by the API (pipe delimited).

* `type` - The type of the resource.

* `state` - The state of the IPv4 address.

* `mac_address` - The MAC address associated with the IPv4 address.

* `custom_properties` - A map of all custom properties associated with the IPv4 address.
