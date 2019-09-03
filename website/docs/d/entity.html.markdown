---
layout: "bluecat"
page_title: "BlueCat: bluecat_entity"
sidebar_current: "docs-bluecat-datasource-entity"
description: |-
 Gets information about an a BlueCat entity.
---

# bluecat\_entity

Use this data source to access the attributes of a BlueCat entity.

## Example Usage

```hcl
data "bluecat_entity" "config" {
    name = "ConfigName"
    type = "Configuration"
}

output "bluecat_config_id" {
    value = data.bluecat_entity.config.id
}
```

## Argument Reference

* `name` - (Required) The name of the entity to find.

* `parent_id` - (Optional) The object ID of the parent object that contains the entity.
  Defaults to 0 which is where Configurations are stored.

* `Type` - (Required) The type of the entity you want to retrieve.

## Attributes Reference

* `properties` -  The properties of the entity as returned by the API (pipe delimited).
