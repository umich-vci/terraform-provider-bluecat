package bluecat

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/umich-vci/gobam"
)

func dataSourceIP4Network() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceIP4NetworkRead,
		Schema: map[string]*schema.Schema{
			"container_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"type": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"IP4Block", "IP4Network", "DHCP4Range", ""}, false),
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
			"cidr": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"template": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"gateway": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_domains": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"default_view": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_restrictions": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"allow_duplicate_host": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"ping_before_assign": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"inherit_allow_duplicate_host": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"inherit_ping_before_assign": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"inherit_dns_restrictions": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"inherit_default_domains": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"inherit_default_view": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"location_code": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"location_inherited": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"addresses_in_use": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"addresses_free": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"custom_properties": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func dataSourceIP4NetworkRead(d *schema.ResourceData, meta interface{}) error {
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
	otype := d.Get("type").(string)
	address := d.Get("address").(string)

	resp, err := client.GetIPRangedByIP(containerID, otype, address)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get IP4 Networks by hint"); err != nil {
		mutex.Unlock()
		return err
	}

	d.SetId(strconv.FormatInt(*resp.Id, 10))
	d.Set("name", *resp.Name)
	d.Set("properties", *resp.Properties)
	d.Set("type", *resp.Type)

	networkProperties, err := gobam.ParseIP4NetworkProperties(*resp.Properties)
	if err = gobam.LogoutClientIfError(client, err, "Error parsing host record properties"); err != nil {
		mutex.Unlock()
		return err
	}

	d.Set("cidr", networkProperties.CIDR)
	d.Set("template", networkProperties.Template)
	d.Set("gateway", networkProperties.Gateway)
	d.Set("default_domains", networkProperties.DefaultDomains)
	d.Set("default_view", networkProperties.DefaultView)
	d.Set("dns_restrictions", networkProperties.DefaultDomains)
	d.Set("allow_duplicate_host", networkProperties.AllowDuplicateHost)
	d.Set("ping_before_assign", networkProperties.PingBeforeAssign)
	d.Set("inherit_allow_duplicate_host", networkProperties.InheritAllowDuplicateHost)
	d.Set("inherit_ping_before_assign", networkProperties.InheritPingBeforeAssign)
	d.Set("inherit_dns_restrictions", networkProperties.InheritDNSRestrictions)
	d.Set("inherit_default_domains", networkProperties.InheritDefaultDomains)
	d.Set("inherit_default_view", networkProperties.InheritDefaultView)
	d.Set("location_code", networkProperties.LocationCode)
	d.Set("location_inherited", networkProperties.LocationInherited)
	d.Set("custom_properties", networkProperties.CustomProperties)

	addressesInUse, addressesFree, err := getIP4NetworkAddressUsage(*resp.Id, networkProperties.CIDR, client)
	if err = gobam.LogoutClientIfError(client, err, "Error calculating network usage"); err != nil {
		mutex.Unlock()
		return err
	}

	d.Set("addresses_in_use", addressesInUse)
	d.Set("addresses_free", addressesFree)

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return err
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return nil
}

func getIP4NetworkAddressUsage(id int64, cidr string, client gobam.ProteusAPI) (int, int, error) {

	netmask, err := strconv.ParseFloat(strings.Split(cidr, "/")[1], 64)
	if err != nil {
		mutex.Unlock()
		return 0, 0, fmt.Errorf("Error parsing netmask from cidr string")
	}
	addressCount := int(math.Pow(2, (32 - netmask)))

	resp, err := client.GetEntities(id, "IP4Address", 0, addressCount)
	if err != nil {
		return 0, 0, err
	}

	addressesInUse := len(resp.Item)
	addressesFree := addressCount - addressesInUse

	return addressesInUse, addressesFree, nil
}
