package bluecat

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceStaticIPv4Address() *schema.Resource {
	return &schema.Resource{
		Create: resourceStaticIPv4AddressCreate,
		Read:   resourceStaticIPv4AddressRead,
		Update: resourceStaticIPv4AddressUpdate,
		Delete: resourceStaticIPv4AddressDelete,

		Schema: map[string]*schema.Schema{
			"configuration_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"ipv4_address": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
			},
			"mac_address": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Default:  "",
			},
			"host_info": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Default:  "",
			},
			"parent_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
			},
			"properties": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Default:  "",
			},
		},
	}
}

func resourceStaticIPv4AddressCreate(d *schema.ResourceData, m interface{}) error {
	return resourceStaticIPv4AddressRead(d, m)
}

func resourceStaticIPv4AddressRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceStaticIPv4AddressUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceStaticIPv4AddressRead(d, m)
}

func resourceStaticIPv4AddressDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
