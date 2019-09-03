---
layout: "bluecat"
page_title: "BlueCat: bluecat_ip4_address"
sidebar_current: "docs-bluecat-datasource-ip4-address"
description: |-
 Gets information about an existing IPv4 address.
---

# bluecat\_ip4\_address

Use this data source to access the attributes of an IPv4 address.

## Example Usage

```hcl
data "bluecat_ip4_address" "addr" {
    container_id = data.bluecat_entity.config.id
    address = "192.168.1.1"
}

output "bluecat_address_notes" {
    value = data.bluecat_ip4_address.addr.notes
}
```

## Argument Reference

* `container_id` - (Required) The object ID of the container that has the specified `address`.  This can be a
  Configuration, IPv4 Block, IPv4 Network, or DHCP range.

* `address` - (Required) The IPv4 address to get data for.

## Attributes Reference

* `name` - The name assigned to the IPv4 address.  This is not related to DNS.

* `properties` -  The properties of the IPv4 address as returned by the API (pipe delimited).

* `type` - The type of the resource.

* `assigned_date` - The date the IPv4 address was assigned.

* `requested_by` - The requestor of the IPv4 address.

* `notes` -  Notes about the IPv4 address.

* `state` - The state of the IPv4 address.

* `mac_address` - The MAC address associated with the IPv4 address.
