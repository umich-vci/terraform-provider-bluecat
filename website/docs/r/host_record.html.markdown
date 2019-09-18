---
layout: "bluecat"
page_title: "BlueCat: bluecat_host_record"
sidebar_current: "docs-bluecat-resource-host_record"
description: |-
 Create a host record for address(es).
---

# bluecat\_host\_record

Use this resource create a host record.

## Example Usage

```hcl
resource "bluecat_host_record" "hostname" {
    view = data.bluecat_entity.view.id
    name = "hostname"
    dns_zone = "example.com
    addresses = ["192.168.1.100"]
}

output "bluecat_hostname_fqdn" {
    value = bluecat_host_record.hostname.absolute_name
}
```

## Argument Reference

* `view_id` - (Required) The object ID of the View that host record should be created in.
  If changed, forces a new resource.

* `name` - (Required) The name of the host record to be created.
  Combined with `dns_zone` to make the fqdn.
  
* `dns_zone` - (Required) The DNS zone to create the host record in.
  Combined with `name` to make the fqdn.  If changed, forces a new resource.

* `addresses` - (Required) The address(es) to be associated with the host record.

* `ttl` - (Optional) The TTL for the host record.  Defaults to -1 which ignores the TTL.

* `reverse_record` - (Optional) If a reverse record should be created for addresses.
  Defaults to true.

* `comments` - (Optional) Comments to be associated with the host record.

* `custom_properties` - (Optional) A map of all custom properties associated with the host record.

## Attributes Reference

* `properties` -  The properties of the IPv4 address as returned by the API (pipe delimited).

* `type` - The type of the resource.

* `absolute_name` - The absolute name (fqdn) of the host record.
