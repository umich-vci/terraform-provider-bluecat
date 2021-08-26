package provider

import (
	"context"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/umich-vci/gobam"
)

func resourceIP4Network() *schema.Resource {
	return &schema.Resource{
		Description: "Resource to create an IPv4 network.",

		CreateContext: resourceIP4NetworkCreate,
		ReadContext:   resourceIP4NetworkRead,
		UpdateContext: resourceIP4NetworkUpdate,
		DeleteContext: resourceIP4NetworkDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The display name of the IPv4 network.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"parent_id": {
				Description: "The object ID of the parent object that will contain the new IPv4 network. If this argument is changed, then the resource will be recreated.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"size": {
				Description: "The size of the IPv4 network expressed as a power of 2. For example, 256 would create a /24. If this argument is changed, then the resource will be recreated.",
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
			},
			"is_larger_allowed": {
				Description: "(Optional) Is it ok to return a network that is larger than the size specified?",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"traversal_method": {
				Description:  "The traversal method used to find the range to allocate the network. Must be one of \"NO_TRAVERSAL\", \"DEPTH_FIRST\", or \"BREADTH_FIRST\".",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "NO_TRAVERSAL",
				ValidateFunc: validation.StringInSlice([]string{"NO_TRAVERSAL", "DEPTH_FIRST", "BREADTH_FIRST"}, false),
			},
			"addresses_in_use": {
				Description: "The number of addresses allocated/in use on the network.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"addresses_free": {
				Description: "The number of addresses unallocated/free on the network.",
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
			"default_view": {
				Description: "The object id of the default DNS View for the network.",
				Type:        schema.TypeString,
				Computed:    true,
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
				Description: "The default DNS Viewis inherited.",
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
			"type": {
				Description: "The type of the resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}
func resourceIP4NetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client

	parentID, err := strconv.ParseInt(d.Get("parent_id").(string), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert parent_id from string to int64"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	size := int64(d.Get("size").(int))
	isLargerAllowed := d.Get("is_larger_allowed").(bool)
	traversalMethod := d.Get("traversal_method").(string)
	autoCreate := true     //we always want to create since this is a resource after all
	reuseExisting := false //we never want to use an existing network created outside terraform
	Type := "IP4Network"   //Since this is the ip4_network resource we are setting the type
	properties := "reuseExisting=" + strconv.FormatBool(reuseExisting) + "|"
	properties = properties + "isLargerAllowed=" + strconv.FormatBool(isLargerAllowed) + "|"
	properties = properties + "autoCreate=" + strconv.FormatBool(autoCreate) + "|"
	properties = properties + "traversalMethod=" + traversalMethod + "|"

	resp, err := client.GetNextAvailableIPRange(parentID, size, Type, properties)
	if err = gobam.LogoutClientIfError(client, err, "Failed on GetNextAvailableIP4Network"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(*resp.Id, 10))

	id := *resp.Id
	name := d.Get("name").(string)
	properties = ""
	otype := "IP4Network"

	setName := gobam.APIEntity{
		Id:         &id,
		Name:       &name,
		Properties: &properties,
		Type:       &otype,
	}

	client.Update(&setName)
	if err = gobam.LogoutClientIfError(client, err, "Failed to update new IP4 Network"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return resourceIP4NetworkRead(ctx, d, meta)
}

func resourceIP4NetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert id from string to int64"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	resp, err := client.GetEntityById(id)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get IP4 Address by Id"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	if *resp.Id == 0 {
		d.SetId("")

		if err := client.Logout(); err != nil {
			mutex.Unlock()
			return diag.FromErr(err)
		}

		mutex.Unlock()
		return nil
	}

	d.Set("name", *resp.Name)
	d.Set("properties", *resp.Properties)
	d.Set("type", resp.Type)

	networkProperties, err := gobam.ParseIP4NetworkProperties(*resp.Properties)
	if err = gobam.LogoutClientIfError(client, err, "Error parsing IPv4 network properties"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	d.Set("cidr", networkProperties.CIDR)
	d.Set("allow_duplicate_host", networkProperties.AllowDuplicateHost)
	d.Set("inherit_allow_duplicate_host", networkProperties.InheritAllowDuplicateHost)
	d.Set("inherit_ping_before_assign", networkProperties.InheritPingBeforeAssign)
	d.Set("ping_before_assign", networkProperties.PingBeforeAssign)
	d.Set("gateway", networkProperties.Gateway)
	d.Set("inherit_default_domains", networkProperties.InheritDefaultDomains)
	d.Set("default_view", networkProperties.DefaultView)
	d.Set("inherit_default_view", networkProperties.InheritDefaultView)
	d.Set("inherit_dns_restrictions", networkProperties.InheritDNSRestrictions)
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

func resourceIP4NetworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert id from string to int64"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	name := d.Get("name").(string)
	properties := ""
	otype := "IP4Network"

	update := gobam.APIEntity{
		Id:         &id,
		Name:       &name,
		Properties: &properties,
		Type:       &otype,
	}

	client.Update(&update)
	if err = gobam.LogoutClientIfError(client, err, "IP4 Network Update failed"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return resourceIP4NetworkRead(ctx, d, meta)
}

func resourceIP4NetworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert id from string to int64"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	resp, err := client.GetEntityById(id)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get IP4 Network by Id"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	if *resp.Id == 0 {
		if err := client.Logout(); err != nil {
			mutex.Unlock()
			return diag.FromErr(err)
		}

		mutex.Unlock()
		return nil
	}

	err = client.Delete(id)
	if err = gobam.LogoutClientIfError(client, err, "Delete failed"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return nil
}
