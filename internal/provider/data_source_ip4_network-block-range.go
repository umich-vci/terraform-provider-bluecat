package provider

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/umich-vci/gobam"
)

func dataSourceIP4NBR() *schema.Resource {
	return &schema.Resource{
		Description: "Data source to access the attributes of an IPv4 network, IPv4 Block, or DHCPv4 Range.",

		ReadContext: dataSourceIP4NBRRead,

		Schema: map[string]*schema.Schema{
			"address": {
				Description: "IP address to find the IPv4 network, IPv4 Block, or DHCPv4 Range of.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"container_id": {
				Description: "The object ID of a container that contains the specified IPv4 network, block, or range.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description:  "Must be \"IP4Block\", \"IP4Network\", \"DHCP4Range\", or \"\". \"\" will find the most specific container.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"IP4Block", "IP4Network", "DHCP4Range", ""}, false),
			},
			"addresses_free": {
				Description: "The number of addresses unallocated/free on the network.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"addresses_in_use": {
				Description: "The number of addresses allocated/in use on the network.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"allow_duplicate_host": {
				Description: "Duplicate host names check.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cidr": {
				Description: "The CIDR address of the IPv4 network.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"custom_properties": {
				Description: "A map of all custom properties associated with the IPv4 network.",
				Type:        schema.TypeMap,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"default_domains": {
				Description: "TODO",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"default_view": {
				Description: "The object id of the default DNS View for the network.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"dns_restrictions": {
				Description: "TODO",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"gateway": {
				Description: "The gateway of the IPv4 network.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"inherit_allow_duplicate_host": {
				Description: "Duplicate host names check is inherited.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"inherit_default_domains": {
				Description: "Default domains are inherited.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"inherit_default_view": {
				Description: "The default DNS View is inherited.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"inherit_dns_restrictions": {
				Description: "DNS restrictions are inherited.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"inherit_ping_before_assign": {
				Description: "The network pings an address before assignment is inherited.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"location_code": {
				Description: "TODO",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"location_inherited": {
				Description: "TODO",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"name": {
				Description: "The name assigned the resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ping_before_assign": {
				Description: "The network pings an address before assignment.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"properties": {
				Description: "The properties of the resource as returned by the API (pipe delimited).",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"template": {
				Description: "TODO",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceIP4NBRRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client

	containerID, err := strconv.ParseInt(d.Get("container_id").(string), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert container_id from string to int64"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	otype := d.Get("type").(string)
	address := d.Get("address").(string)

	resp, err := client.GetIPRangedByIP(containerID, otype, address)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get IP4 Networks by hint"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(*resp.Id, 10))
	d.Set("name", resp.Name)
	d.Set("properties", resp.Properties)
	d.Set("type", resp.Type)

	networkProperties, err := gobam.ParseIP4NetworkProperties(*resp.Properties)
	if err = gobam.LogoutClientIfError(client, err, "Error parsing host record properties"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
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
		return diag.FromErr(err)
	}

	d.Set("addresses_in_use", addressesInUse)
	d.Set("addresses_free", addressesFree)

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return nil
}

func getIP4NetworkAddressUsage(id int64, cidr string, client gobam.ProteusAPI) (int, int, error) {

	netmask, err := strconv.ParseFloat(strings.Split(cidr, "/")[1], 64)
	if err != nil {
		mutex.Unlock()
		return 0, 0, fmt.Errorf("error parsing netmask from cidr string")
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
