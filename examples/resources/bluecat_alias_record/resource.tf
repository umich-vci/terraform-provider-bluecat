resource "bluecat_alias_record" "alias" {
  view_id            = data.bluecat_entity.view.id
  name               = "alias"
  dns_zone           = "example.com"
  linked_record_name = "hostname.example.com"
}

output "bluecat_alias_fqdn" {
  value = bluecat_alias_record.alias.absolute_name
}
