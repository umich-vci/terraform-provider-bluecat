package bluecat

import (
	"log"
	"strconv"
	"strings"

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
			"assigned_date": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"requested_by": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"notes": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": &schema.Schema{
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
	address := d.Get("address").(string)

	resp, err := client.GetIP4Address(containerID, address)
	if err = bam.LogoutClientIfError(client, err, "Failed to get IP4 Address"); err != nil {
		mutex.Unlock()
		return err
	}

	d.SetId(strconv.FormatInt(*resp.Id, 10))
	d.Set("name", resp.Name)
	d.Set("properties", resp.Properties)
	d.Set("type", resp.Type)

	props := strings.Split(*resp.Properties, "|")
	for x := range props {
		if len(props[x]) > 0 {
			prop := strings.Split(props[x], "=")[0]
			val := strings.Split(props[x], "=")[1]

			switch prop {
			case "Assigned_Date":
				d.Set("assigned_date", val)
			case "Requested_by":
				d.Set("requested_by", val)
			case "Notes":
				d.Set("notes", val)
			case "address":
				// since we have to pass in an address to read it we don't really need this
				// 	d.Set("address", val)
			case "state":
				d.Set("state", val)
			default:
				log.Printf("[WARN]Unknown IP4 Address Property: %s", prop)
			}
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
