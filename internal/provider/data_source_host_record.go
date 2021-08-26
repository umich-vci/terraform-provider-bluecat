package provider

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/umich-vci/gobam"
)

func dataSourceHostRecord() *schema.Resource {
	return &schema.Resource{
		Description: "Data source to access the attributes of a host record. If the API returns more than one host record that matches, an error will be returned.",

		ReadContext: dataSourceHostRecordRead,

		Schema: map[string]*schema.Schema{
			"absolute_name": {
				Description: "The absolute name/fqdn of the host record.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"result_count": {
				Description: "The number of results the API should return. This must be between 1 and 10.  You most likely want to leave this alone.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     10,
			},
			"start": {
				Description: "The start index of the search results the API should return. You most likely want to leave this alone.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
			},
			"addresses": {
				Description: "A set of all addresses associated with the host record.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"address_ids": {
				Description: "A set of all address ids associated with the host record.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"custom_properties": {
				Description: "A map of all custom properties associated with the host record.",
				Type:        schema.TypeMap,
				Computed:    true,
			},
			"name": {
				Description: "The short name of the host record.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"parent_id": {
				Description: "The ID of the parent of the host record.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"parent_type": {
				Description: "The type of the parent of the host record.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"properties": {
				Description: "The properties of the host record as returned by the API (pipe delimited).",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"reverse_record": {
				Description: "A boolean that represents if the host record should set reverse records.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"ttl": {
				Description: "The TTL of the host record.",
				Type:        schema.TypeInt,
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

func dataSourceHostRecordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client

	start := d.Get("start").(int)
	count := d.Get("result_count").(int)
	absoluteName := d.Get("absolute_name").(string)
	options := "hint=^" + absoluteName + "$|"

	resp, err := client.GetHostRecordsByHint(start, count, options)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get Host Records by hint"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
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
		err := fmt.Errorf("no exact host record match found for: %s", absoluteName)
		if err = gobam.LogoutClientIfError(client, err, "No exact host record match found for hint"); err != nil {
			mutex.Unlock()
			return diag.FromErr(err)
		}
	}

	d.SetId(strconv.FormatInt(*resp.Item[matchLocation].Id, 10))
	d.Set("name", *resp.Item[matchLocation].Name)
	d.Set("properties", *resp.Item[matchLocation].Properties)
	d.Set("type", resp.Item[matchLocation].Type)

	hostRecordProperties, err := parseHostRecordProperties(*resp.Item[matchLocation].Properties)
	if err = gobam.LogoutClientIfError(client, err, "Error parsing host record properties"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
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
		return diag.FromErr(err)
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
					return hrProperties, fmt.Errorf("error parsing reverseRecord to bool")
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
					return hrProperties, fmt.Errorf("error parsing ttl to int")
				}
				hrProperties.ttl = ttlval
			default:
				hrProperties.customProperties[prop] = val
			}
		}
	}

	return hrProperties, nil
}
