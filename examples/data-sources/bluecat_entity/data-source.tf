data "bluecat_entity" "config" {
    name = "ConfigName"
    type = "Configuration"
}

output "bluecat_config_id" {
    value = data.bluecat_entity.config.id
}
