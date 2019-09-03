---
layout: "bluecat"
page_title: "Provider: BlueCat"
sidebar_current: "docs-bluecat-index"
description: |-
  The BlueCat provider is used to interact with an instance of BlueCat Address Manager.
---

# BlueCat Provider

 The BlueCat provider is used to interact with BlueCat Address Manager.

## Example Usage

```hcl
// Configure the BlueCat Provider
provider "bluecat" {
    username = "username"
    password = "password123"
    bluecat_endpoint = bam.example.com
}

// Get information about a BAM Configuration
data "bluecat_entity" "config" {
    name = "Your Config"
    type = "Configuration"
}
```

## Configuration Reference

The following keys can be used to configure the provider.

* `username` - (Optional) This is the username to use to access BlueCat Address Manager.
  This must be provided in the config or in the environment variable `BLUECAT_USERNAME`.

* `password` - (Optional) This is the password to use to access BlueCat Address Manager.
  This must be provided in the config or in the environment variable `BLUECAT_PASSWORD`.

* `bluecat_endpoint` - (Optional) This is the hostname or IP address of the BlueCat Address Manager.
  This must be provided in the config or in the environment variable `BLUECAT_ENDPOINT`.
