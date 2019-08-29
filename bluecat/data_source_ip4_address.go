package bluecat

import (
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/umich-vci/golang-bluecat"
)

func dataSourceIP4Address() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceIP4AddressRead,
		Schema: map[string]*schema.Schema{
			"container_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"ip4_address": &schema.Schema{
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
	if err = bam.LogoutClientIfError(client, err, "Unable to convert container_id from string to int64"); err != nil {
		mutex.Unlock()
		return err
	}
	address := d.Get("ip4_address").(string)

	resp, err := client.GetIP4Address(containerID, address)
	if err = bam.LogoutClientIfError(client, err, "Failed to get IP4 Address"); err != nil {
		mutex.Unlock()
		return err
	}

	d.SetId(strconv.FormatInt(*resp.Id, 10))
	d.Set("name", resp.Name)
	d.Set("properties", resp.Properties)
	d.Set("type", resp.Type)

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return err
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return nil
}
