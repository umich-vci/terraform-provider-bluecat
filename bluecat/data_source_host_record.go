package bluecat

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/umich-vci/golang-bluecat"
)

func dataSourceHostRecord() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHostRecordRead,
		Schema: map[string]*schema.Schema{
			"start": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"result_count": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  10,
			},
			"absolute_name": &schema.Schema{
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
			"parent_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"reverse_record": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"addresses": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"address_ids": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceHostRecordRead(d *schema.ResourceData, meta interface{}) error {
	mutex.Lock()
	client, err := meta.(*Config).Client()
	if err != nil {
		mutex.Unlock()
		return err
	}

	start := d.Get("start").(int)
	count := d.Get("result_count").(int)
	absoluteName := d.Get("absolute_name").(string)
	options := "hint=^" + absoluteName + "$|"

	resp, err := client.GetHostRecordsByHint(start, count, options)
	if err = bam.LogoutClientIfError(client, err, "Failed to get Host Records by hint"); err != nil {
		mutex.Unlock()
		return err
	}

	log.Printf("[INFO] GetHostRecordsByHint returned %s results", strconv.Itoa(len(resp.Item)))

	matches := 0
	matchLocation := -1
	for x := range resp.Item {
		properties := *resp.Item[x].Properties
		props := strings.Split(properties, "|")
		for y := range props {
			if len(props[y]) > 0 {
				prop := strings.Split(props[y], "=")[0]
				val := strings.Split(props[y], "=")[1]
				if prop == "absoluteName" && val == absoluteName {
					matches++
					matchLocation = x
				}
			}
		}
	}

	if matches == 0 || matches > 1 {
		err := fmt.Errorf("No exact host record match found for: %s", absoluteName)
		if err = bam.LogoutClientIfError(client, err, "No exact host record match found for hint"); err != nil {
			mutex.Unlock()
			return err
		}
	}

	d.SetId(strconv.FormatInt(*resp.Item[matchLocation].Id, 10))
	d.Set("name", *resp.Item[matchLocation].Name)
	d.Set("properties", *resp.Item[matchLocation].Properties)
	d.Set("type", resp.Item[matchLocation].Type)

	props := strings.Split(*resp.Item[matchLocation].Properties, "|")
	for x := range props {
		if len(props[x]) > 0 {
			prop := strings.Split(props[x], "=")[0]
			val := strings.Split(props[x], "=")[1]

			switch prop {
			case "absoluteName":
				d.Set("absolute_name", val)
			case "parentId":
				d.Set("parent_id", val)
			case "parentType":
				d.Set("parent_type", val)
			case "reverseRecord":
				b, err := strconv.ParseBool(val)
				if err = bam.LogoutClientIfError(client, err, "Unable to parse reverseRecord to bool"); err != nil {
					mutex.Unlock()
					return err
				}
				d.Set("reverse_record", b)
			case "addresses":
				d.Set("addresses", val)
			case "addressIds":
				d.Set("address_ids", val)
			default:
				log.Printf("[WARN] Unknown Host Record Property: %s", prop)
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
