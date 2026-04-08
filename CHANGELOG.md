## 0.7.0 (Unreleased)

BREAKING CHANGES:
* provider: The `ssl_verify` attribute has been renamed to `skip_ssl_verify` and its semantics have been inverted. TLS certificate verification is now **enabled by default**. Users who previously set `ssl_verify = false` to enable verification should remove that setting. Users who need to skip verification should set `skip_ssl_verify = true`.

BUG FIXES:
* datasource/bluecat_ip4_nbr: Fixed `inherit_ping_before_assign` being written to the wrong field (`inherit_allow_duplicate_host`), causing both attributes to have incorrect values.
* resource/bluecat_ip4_available_network: Fixed infinite loop when `random = true` and all networks in the list have no free addresses.
* resource/bluecat_ip4_network, resource/bluecat_ip4_block: Fixed `dns_restrictions` being formatted as a Go slice literal instead of a comma-separated string during updates.
* datasource/bluecat_ip4_nbr: Removed spurious mutex unlock in `getIP4NetworkAddressUsage` that could cause a panic on error.
* Fixed property parsing across all resources and data sources to handle malformed input without panicking and to preserve `=` characters in property values.
* datasource/bluecat_ip4_nbr: Fixed `dnsRestrictions` parsing overwriting the diagnostics accumulator variable.
* datasource/bluecat_ip4_address: Fixed schema/model mismatch — added missing attributes (`router_port_info`, `switch_port_info`, `vlan_info`, `lease_time`, `expiry_time`, `parameter_request_list`, `vendor_class_identifier`, `location_code`, `location_inherited`) and renamed `custom_properties` to `user_defined_fields` to match the model.
* resource/bluecat_ip4_network: Fixed `dynamic_update` not being set in state after an update operation.
* provider: Fixed wrong error message when `ssl_verify` (now `skip_ssl_verify`) has an unknown value.
* provider: Fixed `clientLogout` reporting errors as "login error" instead of "logout error".
* resource/bluecat_ip4_network, resource/bluecat_ip4_block: Fixed validation error message for `dns_restrictions` incorrectly referencing `allow_duplicate_host`.
* Fixed error messages in `flattenIP4AddressProperties` incorrectly referencing `flattenIP4Network`.
* Fixed `Configure()` error messages across all resources and data sources incorrectly saying "Expected *http.Client" instead of "Expected *loginClient".

## 0.6.0 (July 14, 2025)
IMPROVEMENTS:
* Updated [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs) to 0.22.0
* Updated [terraform-plugin-framework](https://github.com/hashicorp/terraform-plugin-framework) to 1.15.0
* Updated [terraform-plugin-framework-validators](https://github.com/hashicorp/terraform-plugin-framework-validators) to 0.18.0
* Updated [terraform-plugin-go](https://github.com/hashicorp/terraform-plugin-go) to 0.28.0
* Updated [terraform-plugin-log](https://github.com/hashicorp/terraform-plugin-log) to 0.9.0
* Updated [terraform-plugin-testing](https://github.com/hashicorp/terraform-plugin-testing) to 1.13.2
* resource/bluecat_ip4_network, datasource/bluecat_ip4_network - add support for dynamicUpdate property introduced in BlueCat 9.6.0


## 0.5.0 (November 21, 2024)
FEATURES:
* **New Resource:** `bluecat_ip4_block` ([#113](https://github.com/umich-vci/terraform-provider-bluecat/pull/113))
  Thanks to @aaronmaxlevy

IMPROVEMENTS:
* Updated [terraform-plugin-testing](https://github.com/hashicorp/terraform-plugin-testing) to 1.11.0

## 0.4.5 (November 15, 2024)
BUG FIXES:
* resource/bluecat_ip4_available_network fix problem introduced with migration from sdk to framework

IMPROVEMENTS:
* Updated [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs) to 0.20.0
* Updated [terraform-plugin-framework](https://github.com/hashicorp/terraform-plugin-framework) to 1.13.0
* Updated [terraform-plugin-framework-validators](https://github.com/hashicorp/terraform-plugin-framework-validators) to 0.15.0
* Updated [terraform-plugin-go](https://github.com/hashicorp/terraform-plugin-go) to 0.25.0
* Updated [terraform-plugin-log](https://github.com/hashicorp/terraform-plugin-log) to 0.9.0
* Updated [terraform-plugin-testing](https://github.com/hashicorp/terraform-plugin-testing) to 1.10.0

## 0.4.4 (March 26, 2024)

BUG FIXES:
* Adjust plan modifiers for attributes only required for creation to allow for import.

## 0.4.3 (March 26, 2024)

BREAKING CHANGES:
* IDs must be treated as a string so that import works correctly.

## 0.4.2 (March 25, 2024)

BUG FIXES:
* Fix logic error with ping_before_assign value check on `resource/bluecat_ip4_network`

IMPROVEMENTS:
* Updated [terraform-plugin-framework](https://github.com/hashicorp/terraform-plugin-framework) to 1.7.0

## 0.4.1 (March 19, 2024)

BREAKING CHANGES:
* The schema on several resources has been changed, but now fields exposed
  track what is documented for the Legacy API.
* There were breaking changes in 0.4.0 as well, but they were not documented or as consistent.

IMPROVEMENTS:
* All fields on network, address, and host_record resource that are configurable
  via the API should work now.
* Updated dependencies

## 0.4.0 (February 29, 2024)

IMPROVEMENTS:

* Reworked code to use terraform-plugin-framework instead of terraform-plugin-sdk/v2
* Simplified how we handle logouts

## 0.3.1 (October 15, 2021)

BUG FIXES:

* bluecat_ip4_network was not returning all properties.

## 0.3.0 (October 10, 2021)

FEATURES:

* **New Datasource:** `bluecat_ip4_network`

BUG FIXES:

* Error if no entity is found with the bluecat_entity datasource.

IMPROVEMENTS:

* Updated [Terraform SDK](https://github.com/hashicorp/terraform-plugin-sdk/) to 2.8.0

* Updated [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs) to 0.5.0

* Reworked code to follow [Terraform Provider Scaffolding](https://github.com/hashicorp/terraform-provider-scaffolding)

* Use generated documentation.

* Use context aware functions.

* Now building with go 1.16

## 0.2.0 (December 9, 2020)

BREAKING CHANGES:

* resource/bluecat_ip4_address: Removed `parent_id_list` argument and `computed_parent_id` attribute.
  `parent_id` argument is now required.
  ([#4](https://github.com/umich-vci/terraform-provider-bluecat/issues/4))

FEATURES:

* **New Resource:** `bluecat_ip4_available_network` ([#4](https://github.com/umich-vci/terraform-provider-bluecat/issues/4))

IMPROVEMENTS:

* Updated [gobam](https://github.com/umich-vci/gobam) to 20201026200032-5742f663694f and added a new
  provider configuration argument `ssl_verify` to allow ignoring SSL certificate validation errors.
  ([#1](https://github.com/umich-vci/terraform-provider-bluecat/issues/1))

* Switched from Terraform SDK v1 to v2 ([#5](https://github.com/umich-vci/terraform-provider-bluecat/pull/5))

* Now building with go 1.15

## 0.1.0 (October 14, 2020)

First release of provider.
