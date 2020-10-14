<a href="https://terraform.io">
    <img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" alt="Terraform logo" title="Terraform" align="right" height="50" />
</a>

# Terraform Provider for BlueCat Address Manager

A Terraform provider for BlueCat Address Manager.  Curently the provider is able to work with
IPv4 Addresses, IPv4 Networks, and host records.

Currently, the provider does not have working tests so it should probably be considered beta.

## Building/Installing

Running `GO111MODULE=on go get -u github.com/umich-vci/terraform-provider-bluecat` should download
the code and result in a binary at `$GOPATH/bin/terraform-provider-bluecat`. You can then move the
binary to `~/.terraform.d/plugins` to use it with Terraform.

This has been tested with Terraform 0.12.x and BlueCat Address Manager 9.1.0.

## License

This project is licensed under the Mozilla Public License Version 2.0.
