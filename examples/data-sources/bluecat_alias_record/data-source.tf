data "bluecat_host_record" "host" {
  absolute_name = "alias.example.com"
}

output "bluecat_host_addresses" {
  value = data.bluecat_host_record.host.linked_record_name
}
