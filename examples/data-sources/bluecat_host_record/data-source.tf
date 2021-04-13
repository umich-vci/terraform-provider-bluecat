data "bluecat_host_record" "host" {
    absolute_name = "host.example.com"
}

output "bluecat_host_addresses" {
    value = data.bluecat_host_record.host.addresses
}
