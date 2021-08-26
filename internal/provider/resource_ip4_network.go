package provider

import (
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/umich-vci/gobam"
)

func resourceIP4Network() *schema.Resource {
	return &schema.Resource{
		Create: resourceIP4NetworkCreate,
		Read:   resourceIP4NetworkRead,
		Update: resourceIP4NetworkUpdate,
		Delete: resourceIP4NetworkDelete,
		Schema: map[string]*schema.Schema{
			"parent_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"is_larger_allowed": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			// We don't want to touch resources created outside of Terraform so always assume false
			// "reuse_existing": &schema.Schema{
			// 	Type:     schema.TypeBool,
			// 	Optional: true,
			// 	Default:  false,
			// },
			// We don't use auto_create since we will always want to create a network
			// "auto_create": &schema.Schema{
			// 	Type:     schema.TypeBool,
			// 	Optional: true,
			// 	Default:  true,
			// },
			"traversal_method": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "NO_TRAVERSAL",
				ValidateFunc: validation.StringInSlice([]string{"NO_TRAVERSAL", "DEPTH_FIRST", "BREADTH_FIRST"}, false),
			},
			"properties": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"allow_duplicate_host": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"inherit_allow_duplicate_host": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ping_before_assign": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"inherit_ping_before_assign": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"gateway": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"inherit_default_domains": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"default_view": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"inherit_default_view": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"inherit_dns_restrictions": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"addresses_in_use": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"addresses_free": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}
func resourceIP4NetworkCreate(d *schema.ResourceData, meta interface{}) error {
	mutex.Lock()
	client := meta.(*apiClient).Client

	parentID, err := strconv.ParseInt(d.Get("parent_id").(string), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert parent_id from string to int64"); err != nil {
		mutex.Unlock()
		return err
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
		return err
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
		return err
	}

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return err
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return resourceIP4NetworkRead(d, meta)
}

func resourceIP4NetworkRead(d *schema.ResourceData, meta interface{}) error {
	mutex.Lock()
	client := meta.(*apiClient).Client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert id from string to int64"); err != nil {
		mutex.Unlock()
		return err
	}

	resp, err := client.GetEntityById(id)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get IP4 Address by Id"); err != nil {
		mutex.Unlock()
		return err
	}

	if *resp.Id == 0 {
		d.SetId("")

		if err := client.Logout(); err != nil {
			mutex.Unlock()
			return err
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
		return err
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

func resourceIP4NetworkUpdate(d *schema.ResourceData, meta interface{}) error {
	mutex.Lock()
	client := meta.(*apiClient).Client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert id from string to int64"); err != nil {
		mutex.Unlock()
		return err
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
		return err
	}

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return err
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return resourceIP4NetworkRead(d, meta)
}

func resourceIP4NetworkDelete(d *schema.ResourceData, meta interface{}) error {
	mutex.Lock()
	client := meta.(*apiClient).Client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert id from string to int64"); err != nil {
		mutex.Unlock()
		return err
	}

	resp, err := client.GetEntityById(id)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get IP4 Network by Id"); err != nil {
		mutex.Unlock()
		return err
	}

	if *resp.Id == 0 {
		if err := client.Logout(); err != nil {
			mutex.Unlock()
			return err
		}

		mutex.Unlock()
		return nil
	}

	err = client.Delete(id)
	if err = gobam.LogoutClientIfError(client, err, "Delete failed"); err != nil {
		mutex.Unlock()
		return err
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
