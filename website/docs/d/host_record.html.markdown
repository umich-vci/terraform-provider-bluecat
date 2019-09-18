---
layout: "bluecat"
page_title: "BlueCat: bluecat_host_record"
sidebar_current: "docs-bluecat-datasource-host-record"
description: |-
 Gets information about an existing host record.
---

# bluecat\_host\_record

Use this data source to access the attributes of a host record.  If the API returns more than one
host record that matches, an error will be returned.

## Example Usage

```hcl
data "bluecat_host_record" "host" {
    absolute_name = "host.example.com"
}

output "bluecat_host_addresses" {
    value = data.bluecat_host_record.host.addresses
}
```

## Argument Reference

* `absolute_name` - (Required) The absolute name/fqdn of the host record.

* `start` - (Optional) The start index of the search results the API should return.  Defaults to 0.
  You most likely want to leave this alone.

* `result_count` - (Optional) The number of results the API should return.  Defaults to 10.
  This must be between 1 and 10.  You most likely want to leave this alone.

## Attributes Reference

* `name` - The short name of the host record.

* `properties` -  The properties of the host record as returned by the API (pipe delimited).

* `type` - The type of the resource.

* `parent_id` - The ID of the parent of the host record.

* `parent_type` - The type of the parent of the host record.

* `reverse_record` -  A boolean that represents if the host record should set reverse records.

* `addresses` - A set of all addresses associated with the host record.

* `address_ids` - A set of all address ids associated with the host record.

* `custom_properties` - A map of all custom properties associated with the host record.

* `ttl` - The TTL of the host record.
