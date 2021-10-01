package provider

import (
	"context"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/umich-vci/gobam"
)

func dataSourceIP4Network() *schema.Resource {
	return &schema.Resource{
		Description: "Data source to access the attributes of an IPv4 network from a hint based search.",

		ReadContext: dataSourceIP4NetworkRead,

		Schema: map[string]*schema.Schema{
			"hint": {
				Description: "Hint to find the IP4Network",
				Type:        schema.TypeString,
				Required:    true,
			},
			"container_id": {
				Description: "The object ID of a container that contains the specified IPv4 network.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description: "The type of the IP4Network",
				Type:        schema.TypeString,
				Computed:    true,
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

func dataSourceIP4NetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client

	containerID, err := strconv.ParseInt(d.Get("container_id").(string), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert container_id from string to int64"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	hint := d.Get("hint").(string)
	options := "hint=" + hint

	resp, err := client.GetIP4NetworksByHint(containerID, 0, 0, options)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get IP4 Networks by hint"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	if len(resp.Item) > 1 || len(resp.Item) == 0 {
		var diags diag.Diagnostics
		err := gobam.LogoutClientWithError(client, "Network lookup error")
		mutex.Unlock()

		diags = append(diags, diag.FromErr(err)...)
		diags = append(diags, diag.Errorf("Hint %s returned %d networks", hint, len(resp.Item))...)

		return diags
	}

	d.SetId(strconv.FormatInt(*resp.Item[0].Id, 10))
	d.Set("name", resp.Item[0].Name)
	d.Set("properties", resp.Item[0].Properties)
	d.Set("type", resp.Item[0].Type)

	networkProperties, err := gobam.ParseIP4NetworkProperties(*resp.Item[0].Properties)
	if err = gobam.LogoutClientIfError(client, err, "Error parsing network properties"); err != nil {
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

	addressesInUse, addressesFree, err := getIP4NetworkAddressUsage(*resp.Item[0].Id, networkProperties.CIDR, client)
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
