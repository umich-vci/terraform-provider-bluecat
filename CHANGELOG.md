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
