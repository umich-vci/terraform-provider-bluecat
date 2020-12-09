# bluecat\_ip4\_available\_network Resource

Use this resource to select an IPv4 network from a list of networks
based on availability of IP addresses.

## Example Usage

```hcl
resource "bluecat_ip4_available_network" "network" {
    network_id_list = [ 1234, 5678, 9101 ]
}

output "network_id" {
    value = bluecat_ip4_address.network.network_id
}
```

## Argument Reference

* `network_id_list` - (Required) A list of Network IDs to search for a free IP address.
  By default, the address with the most free addresses will be returned. See the `random`
  argument for another selection method. The resource will be recreated if the network_id_list
  is changed. You may want to use a `lifecycle` customization to ignore changes to the list
  after resource creation so that a new network is not selected if the list is changed.

* `keepers` - (Optional) An arbitrary map of values. If this argument is changed, then the
  resource will be recreated.

* `random` - (Optional) By default, the network with the most free IP addresses is returned.
  By setting this to `true` a random network from the list will be returned instead.
  The network will be validated to have at least 1 free IP address. Defaults to `false`.
  
* `seed` - (Optional) A seed for the `random` argument's generator. Can be used to try to get
  more predictable results from the random selection. The results will not be fixed however.

## Attributes Reference

* `network_id` -  The network ID of the network selected by the resource.
