package bluecat

import (
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/umich-vci/gobam"
)

func dataSourceIP4Address() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceIP4AddressRead,
		Schema: map[string]*schema.Schema{
			"container_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"address": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"properties": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"mac_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_properties": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func dataSourceIP4AddressRead(d *schema.ResourceData, meta interface{}) error {
	mutex.Lock()
	client, err := meta.(*Config).Client()
	if err != nil {
		mutex.Unlock()
		return err
	}

	containerID, err := strconv.ParseInt(d.Get("container_id").(string), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert container_id from string to int64"); err != nil {
		mutex.Unlock()
		return err
	}
	address := d.Get("address").(string)

	resp, err := client.GetIP4Address(containerID, address)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get IP4 Address"); err != nil {
		mutex.Unlock()
		return err
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
		return err
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
