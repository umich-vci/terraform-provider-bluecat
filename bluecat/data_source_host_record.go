package bluecat

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/umich-vci/gobam"
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
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"address_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"custom_properties": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
			"ttl": &schema.Schema{
				Type:     schema.TypeInt,
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
	if err = gobam.LogoutClientIfError(client, err, "Failed to get Host Records by hint"); err != nil {
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
		if err = gobam.LogoutClientIfError(client, err, "No exact host record match found for hint"); err != nil {
			mutex.Unlock()
			return err
		}
	}

	d.SetId(strconv.FormatInt(*resp.Item[matchLocation].Id, 10))
	d.Set("name", *resp.Item[matchLocation].Name)
	d.Set("properties", *resp.Item[matchLocation].Properties)
	d.Set("type", resp.Item[matchLocation].Type)

	hostRecordProperties, err := parseHostRecordProperties(*resp.Item[matchLocation].Properties)
	if err = gobam.LogoutClientIfError(client, err, "Error parsing host record properties"); err != nil {
		mutex.Unlock()
		return err
	}

	d.Set("absolute_name", hostRecordProperties.absoluteName)
	d.Set("parent_id", hostRecordProperties.parentID)
	d.Set("parent_type", hostRecordProperties.parentType)
	d.Set("reverse_record", hostRecordProperties.reverseRecord)
	d.Set("addresses", hostRecordProperties.addresses)
	d.Set("address_ids", hostRecordProperties.addressIDs)
	d.Set("custom_properties", hostRecordProperties.customProperties)
	d.Set("ttl", hostRecordProperties.ttl)

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return err
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return nil
}

type hostRecordProperties struct {
	absoluteName     string
	parentID         string
	parentType       string
	ttl              int
	reverseRecord    bool
	addresses        []string
	addressIDs       []string
	customProperties map[string]string
}

func parseHostRecordProperties(properties string) (hostRecordProperties, error) {
	var hrProperties hostRecordProperties
	hrProperties.customProperties = make(map[string]string)

	// if ttl isn't returned as a property it will remain set at -1
	hrProperties.ttl = -1

	props := strings.Split(properties, "|")
	for x := range props {
		if len(props[x]) > 0 {
			prop := strings.Split(props[x], "=")[0]
			val := strings.Split(props[x], "=")[1]

			switch prop {
			case "absoluteName":
				hrProperties.absoluteName = val
			case "parentId":
				hrProperties.parentID = val
			case "parentType":
				hrProperties.parentType = val
			case "reverseRecord":
				b, err := strconv.ParseBool(val)
				if err != nil {
					return hrProperties, fmt.Errorf("Error parsing reverseRecord to bool")
				}
				hrProperties.reverseRecord = b
			case "addresses":
				addresses := strings.Split(val, ",")
				for i := range addresses {
					hrProperties.addresses = append(hrProperties.addresses, addresses[i])
				}
			case "addressIds":
				addressIDs := strings.Split(val, ",")
				for i := range addressIDs {
					hrProperties.addressIDs = append(hrProperties.addressIDs, addressIDs[i])
				}
			case "ttl":
				ttlval, err := strconv.Atoi(val)
				if err != nil {
					return hrProperties, fmt.Errorf("Error parsing ttl to int")
				}
				hrProperties.ttl = ttlval
			default:
				hrProperties.customProperties[prop] = val
			}
		}
	}

	return hrProperties, nil
}
