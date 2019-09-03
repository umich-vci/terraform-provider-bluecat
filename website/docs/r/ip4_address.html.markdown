---
layout: "bluecat"
page_title: "BlueCat: bluecat_ip4_address"
sidebar_current: "docs-bluecat-resource-ip4-address"
description: |-
 Requests an IPv4 address from a network.
---

# bluecat\_ip4\_address

Use this resource to reserve an IPv4 address.

## Example Usage

```hcl
resource "bluecat_ip4_address" "addr" {
    container_id = data.bluecat_entity.config.id
    address = "192.168.1.1"
}

output "bluecat_address_notes" {
    value = data.bluecat_ip4_address.addr.notes
}
```

## Argument Reference

* `configuration_id` - (Required) The object ID of the Configuration that has the specified `address`.

* `parent_id` - (Required) The object ID of the Configuration, Block, or Network to find the next available
  IPv4 address in.

* `name` - (Required) The name assigned to the IPv4 address.  This is not related to DNS.
  
* `mac_address` - (Optional) The MAC address to associate with the IPv4 address.

* `action` - (Optional) The action to take on the next available IPv4 address.  Must be one of:
  MAKE_STATIC, MAKE_RESERVED, or MAKE_DHCP_RESERVED.  Defaults to MAKE_STATIC.

* `assigned_date` - (Optional) The date the IPv4 address was assigned.

* `requested_by` - (Optional) The requestor of the IPv4 address.

* `notes` -  (Optional) Notes about the IPv4 address.

## Attributes Reference

* `address` -  The IPv4 address that was allocated.

* `properties` -  The properties of the IPv4 address as returned by the API (pipe delimited).

* `state` - The state of the IPv4 address.

* `type` - The type of the resource.
