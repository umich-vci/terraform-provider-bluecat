package bluecat

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/umich-vci/golang-bluecat"
)

func dataSourceIP4NetworkByName() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceIP4NetworkByNameRead,
		Schema: map[string]*schema.Schema{
			"container_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"start": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"result_count": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"cidr": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"allow_duplicate_host": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"inherit_allow_duplicate_host": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ping_before_assign": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"inherit_ping_before_assign": &schema.Schema{
				Type:     schema.TypeBool,
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

	containerID, err := strconv.ParseInt(d.Get("container_id").(string), 10, 64)
	if err = bam.LogoutClientIfError(client, err, "Unable to convert container_id from string to int64"); err != nil {
		mutex.Unlock()
		return err
	}
	start := d.Get("start").(int)
	count := d.Get("result_count").(int)
	name := d.Get("name").(string)

	options := "hint=" + name

	resp, err := client.GetIP4NetworksByHint(containerID, start, count, options)
	if err = bam.LogoutClientIfError(client, err, "Failed to get IP4 Networks by hint: %s"); err != nil {
		mutex.Unlock()
		return err
	}

	matches := 0
	matchLocation := -1
	for x := range resp.Item {
		if *resp.Item[x].Name == name {
			matches++
			matchLocation = x
		}
	}

	if matches == 0 || matches > 1 {
		err := fmt.Errorf("No exact IP4 network match found for name: %s", name)
		if err = bam.LogoutClientIfError(client, err, "No exact IP4 network match found for name"); err != nil {
			mutex.Unlock()
			return err
		}
	}

	d.SetId(strconv.FormatInt(*resp.Item[matchLocation].Id, 10))
	d.Set("type", resp.Item[matchLocation].Type)

	props := strings.Split(d.Get("properties").(string), "|")
	for x := range props {
		if len(props[x]) > 0 {
			prop := strings.Split(props[x], "=")[0]
			val := strings.Split(props[x], "=")[1]

			switch prop {
			case "CIDR":
				d.Set("cidr", val)
			case "allowDuplicateHost":
				d.Set("allow_duplicate_host", val)
			case "inheritAllowDuplicateHost":
				b, err := strconv.ParseBool(val)
				if err = bam.LogoutClientIfError(client, err, "Unable to parse inheritAllowDuplicateHost to bool"); err != nil {
					mutex.Unlock()
					return err
				}
				d.Set("inherit_allow_duplicate_host", b)
			case "pingBeforeAssign":
				d.Set("ping_before_assign", val)
			case "inheritPingBeforeAssign":
				b, err := strconv.ParseBool(val)
				if err = bam.LogoutClientIfError(client, err, "Unable to parse inheritPingBeforeAssign to bool"); err != nil {
					mutex.Unlock()
					return err
				}
				d.Set("inherit_ping_before_assign", b)
			default:
				err := fmt.Errorf("Unknown IP4 Network Property: %s", val)
				if err = bam.LogoutClientIfError(client, err, "Unknown IP4 Network Property"); err != nil {
					mutex.Unlock()
					return err
				}
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
