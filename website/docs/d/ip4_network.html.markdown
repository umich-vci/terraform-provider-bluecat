---
layout: "bluecat"
page_title: "BlueCat: bluecat_ip4_network"
sidebar_current: "docs-bluecat-datasource-ip4-network"
description: |-
 Gets information about an existing IPv4 network.
---

# bluecat\_ip4\_network

Use this data source to access the attributes of an IPv4 network.  If the API returns more than one
IPv4 network that matches the specified hint, an error will be returned.

## Example Usage

```hcl
data "bluecat_ip4_network" "network" {
    container_id = data.bluecat_entity.config.id
    hint = "192.168.1.0/24"
    hint_type = "cidr"
}

output "bluecat_network_name" {
    value = data.bluecat_ip4_network.network.name
}
```

## Argument Reference

* `container_id` - (Required) The object ID of the container that has the specified IPv4 network.

* `start` - (Optional) The start index of the search results the API should return.  Defaults to 0.
  You most likely want to leave this alone.

* `result_count` - (Optional) The number of results the API should return.  Defaults to 10.
  This must be between 1 and 10.  You most likely want to leave this alone.

* `hint` - (Required) Name or CIDR address of the network to find.

* `hint_type` - (Required) Must be either "name" or "cidr"

## Attributes Reference

* `name` - The name assigned to the IPv4 network.

* `properties` -  The properties of the IPv4 network as returned by the API (pipe delimited).

* `type` - The type of the resource.

* `cidr` - The CIDR address of the IPv4 network.

* `allow_duplicate_host` - Duplicate host names check.

* `inherit_allow_duplicate_host` -  Duplicate host names check is inherited.

* `ping_before_assign` - The network pings an address before assignment.

* `inherit_ping_before_assign` - The network pings an address before assignment is inherited.
