// Configure the BlueCat Provider
provider "bluecat" {
  username         = "username"
  password         = "password123"
  bluecat_endpoint = "bam.example.com"
}

// Get information about a BAM Configuration
data "bluecat_entity" "config" {
  name = "Your Config"
  type = "Configuration"
}
