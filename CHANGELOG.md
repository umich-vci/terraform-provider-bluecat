## 0.3.0 (October 10, 2021)

FEATURES:

* **New Datasoruce:** `bluecat_ip4_network`

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
