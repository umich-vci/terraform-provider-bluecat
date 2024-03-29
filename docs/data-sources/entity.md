---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bluecat_entity Data Source - terraform-provider-bluecat"
subcategory: ""
description: |-
  Data source to access the attributes of a BlueCat entity.
---

# bluecat_entity (Data Source)

Data source to access the attributes of a BlueCat entity.

## Example Usage

```terraform
data "bluecat_entity" "config" {
  name = "ConfigName"
  type = "Configuration"
}

output "bluecat_config_id" {
  value = data.bluecat_entity.config.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the entity to find.
- `parent_id` (Number) The object ID of the parent object that contains the entity. Configurations are stored in ID `0`.
- `type` (String) The type of the entity you want to retrieve.

### Read-Only

- `id` (String) Entity identifier
- `properties` (String) The properties of the entity as returned by the API (pipe delimited).
