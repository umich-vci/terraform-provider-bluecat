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
		Description: "Data source to access the attributes of an IPv4 address.",

		ReadContext: dataSourceIP4AddressRead,

		Schema: map[string]*schema.Schema{
			"address": {
				Description: "The IPv4 address to get data for.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"container_id": {
				Description: "The object ID of the container that has the specified `address`.  This can be a Configuration, IPv4 Block, IPv4 Network, or DHCP range.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"custom_properties": {
				Description: "A map of all custom properties associated with the IPv4 address.",
				Type:        schema.TypeMap,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"mac_address": {
				Description: "The MAC address associated with the IPv4 address.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "The name assigned to the IPv4 address.  This is not related to DNS.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"properties": {
				Description: "The properties of the IPv4 address as returned by the API (pipe delimited).",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the IPv4 address.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"type": {
				Description: "The type of the resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceIP4AddressRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client
	client.Login(meta.(*apiClient).Username, meta.(*apiClient).Password)

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
