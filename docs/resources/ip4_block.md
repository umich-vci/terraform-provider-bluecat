---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bluecat_ip4_block Resource - terraform-provider-bluecat"
subcategory: ""
description: |-
  Resource to create an IPv4 block.
---

# bluecat_ip4_block (Resource)

Resource to create an IPv4 block.

## Example Usage

```terraform
resource "bluecat_ip4_block" "block" {
  parent_id = data.bluecat_ip4_network-block-range.block.id
  name      = "New Block"
  size      = 256
}

output "bluecat_ip4_block_cidr" {
  value = bluecat_ip4_block.block.cidr
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `parent_id` (Number) The object ID of the parent object that will contain the new IPv4 block. If this argument is changed, then the resource will be recreated.
- `size` (Number) The size of the IPv4 block expressed as a power of 2. For example, 256 would create a /24. If this argument is changed, then the resource will be recreated.

### Optional

- `allow_duplicate_host` (Boolean) Duplicate host names check.
- `default_domains` (Set of Number) The object ids of the default DNS domains.
- `default_view` (Number) The object id of the default DNS View for the block.
- `dns_restrictions` (Set of Number) The object ids of the DNS restrictions for the block.
- `inherit_allow_duplicate_host` (Boolean) Duplicate host names check is inherited.
- `inherit_default_domains` (Boolean) Default domains are inherited.
- `inherit_default_view` (Boolean) The default DNS View is inherited.
- `inherit_dns_restrictions` (Boolean) DNS restrictions are inherited.
- `inherit_ping_before_assign` (Boolean) PingBeforeAssign option inheritance check option property.
- `is_larger_allowed` (Boolean) (Optional) Is it ok to return a block that is larger than the size specified?
- `location_code` (String) The location code of the block.
- `name` (String) The display name of the IPv4 block.
- `ping_before_assign` (Boolean) Option to ping check. The possible values are enable and disable.
- `traversal_method` (String) The traversal method used to find the range to allocate the block. Must be one of "NO_TRAVERSAL", "DEPTH_FIRST", or "BREADTH_FIRST".
- `user_defined_fields` (Map of String) A map of all user-definied fields associated with the IP4 Block.

### Read-Only

- `cidr` (String) The CIDR value of the block (if it forms a valid CIDR).
- `end` (String) The end of the block (if it does not form a valid CIDR).
- `id` (String) IPv4 Block identifier.
- `location_inherited` (Boolean) The location is inherited.
- `properties` (String) The properties of the resource as returned by the API (pipe delimited).
- `start` (String) The start of the block (if it does not form a valid CIDR).
- `type` (String) The type of the resource.