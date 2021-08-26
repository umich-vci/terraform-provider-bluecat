package provider

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/umich-vci/gobam"
)

func dataSourceIP4Address() *schema.Resource {
	return &schema.Resource{
		Description: "",

		ReadContext: dataSourceIP4AddressRead,

		Schema: map[string]*schema.Schema{
			"container_id": {
				Description: "",
				Type:        schema.TypeString,
				Required:    true,
			},
			"address": {
				Description: "",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"properties": {
				Description: "",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"type": {
				Description: "",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"mac_address": {
				Description: "",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"custom_properties": {
				Description: "",
				Type:        schema.TypeMap,
				Computed:    true,
			},
		},
	}
}

func dataSourceIP4AddressRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client

	containerID, err := strconv.ParseInt(d.Get("container_id").(string), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert container_id from string to int64"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	address := d.Get("address").(string)

	resp, err := client.GetIP4Address(containerID, address)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get IP4 Address"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(*resp.Id, 10))
	d.Set("name", resp.Name)
	d.Set("properties", resp.Properties)
	d.Set("type", resp.Type)

	addressProperties := parseIP4AddressProperties(*resp.Properties)
	d.Set("address", addressProperties.address)
	d.Set("state", addressProperties.state)
	d.Set("mac_address", addressProperties.macAddress)
	d.Set("custom_properties", addressProperties.customProperties)

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return nil
}

type ip4AddressProperties struct {
	address          string
	state            string
	macAddress       string
	customProperties map[string]string
}

func parseIP4AddressProperties(properties string) ip4AddressProperties {
	var ip4Properties ip4AddressProperties
	ip4Properties.customProperties = make(map[string]string)

	props := strings.Split(properties, "|")
	for x := range props {
		if len(props[x]) > 0 {
			prop := strings.Split(props[x], "=")[0]
			val := strings.Split(props[x], "=")[1]

			switch prop {
			case "address":
				ip4Properties.address = val
			case "state":
				ip4Properties.state = val
			case "macAddress":
				ip4Properties.macAddress = val
			default:
				ip4Properties.customProperties[prop] = val
			}
		}
	}

	return ip4Properties
}
