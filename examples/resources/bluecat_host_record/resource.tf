resource "bluecat_host_record" "hostname" {
  view      = data.bluecat_entity.view.id
  name      = "hostname"
  dns_zone  = "example.com"
  addresses = ["192.168.1.100"]
}

output "bluecat_hostname_fqdn" {
  value = bluecat_host_record.hostname.absolute_name
}
