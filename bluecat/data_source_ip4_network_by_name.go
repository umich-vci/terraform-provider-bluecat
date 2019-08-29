package bluecat

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/umich-vci/golang-bluecat"
)

func dataSourceIP4NetworkByName() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceIP4NetworkByNameRead,
		Schema: map[string]*schema.Schema{
			"container_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
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

func dataSourceIP4NetworkByNameRead(d *schema.ResourceData, meta interface{}) error {
	mutex.Lock()
	client, err := meta.(*Config).Client()
	if err != nil {
		mutex.Unlock()
		return err
	}

	containerID := d.Get("container_id").(int)
	start := 0
	count := 10
	name := d.Get("name").(string)

	options := "hint=" + name

	resp, err := client.GetIP4NetworksByHint(int64(containerID), start, count, options)
	if err = bam.LogoutClientIfError(client, err, "Failed to get IP4 Networks by hint: %s"); err != nil {
		mutex.Unlock()
		return err
	}

	matches := 0

	for x := range resp.Item {
		if *resp.Item[x].Name == name {
			d.SetId(strconv.FormatInt(*resp.Item[x].Id, 10))
			d.Set("properties", resp.Item[x].Properties)
			d.Set("type", resp.Item[x].Type)
			matches++
		}
	}

	if matches == 0 || matches > 1 {
		err := fmt.Errorf("No exact IP4 network match found for name: %s", name)
		if err = bam.LogoutClientIfError(client, err, "No exact IP4 network match found for name"); err != nil {
			mutex.Unlock()
			return err
		}
	}

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return err
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return nil
}
