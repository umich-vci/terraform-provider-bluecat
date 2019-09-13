package bluecat

import (
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/umich-vci/gobam"
)

func resourceIP4Network() *schema.Resource {
	return &schema.Resource{
		Create: resourceIP4NetworkCreate,
		Read:   resourceIP4NetworkRead,
		Update: resourceIP4NetworkUpdate,
		Delete: resourceIP4NetworkDelete,
		Schema: map[string]*schema.Schema{
			"parent_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"size": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"is_larger_allowed": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			// We don't use auto_create since we will always want to create a network
			// "auto_create": &schema.Schema{
			// 	Type:     schema.TypeBool,
			// 	Optional: true,
			// 	Default:  true,
			// },
			"properties": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
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
			"gateway": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"inherit_default_domains": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"default_view": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"inherit_default_view": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"inherit_dns_restrictions": &schema.Schema{
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
		},
	}
}
func resourceIP4NetworkCreate(d *schema.ResourceData, meta interface{}) error {
	mutex.Lock()
	client, err := meta.(*Config).Client()
	if err != nil {
		mutex.Unlock()
		return err
	}

	parentID, err := strconv.ParseInt(d.Get("parent_id").(string), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert parent_id from string to int64"); err != nil {
		mutex.Unlock()
		return err
	}
	size := int64(d.Get("size").(int))
	isLargerAllowed := d.Get("is_larger_allowed").(bool)
	autoCreate := true //we always want to create since this is a resource after all

	resp, err := client.GetNextAvailableIP4Network(parentID, size, isLargerAllowed, autoCreate)
	if err = gobam.LogoutClientIfError(client, err, "Failed on GetNextAvailableIP4Network"); err != nil {
		mutex.Unlock()
		return err
	}

	d.SetId(strconv.FormatInt(resp, 10))

	id := resp
	name := d.Get("name").(string)
	properties := ""
	otype := "IP4Network"

	setName := gobam.APIEntity{
		Id:         &id,
		Name:       &name,
		Properties: &properties,
		Type:       &otype,
	}

	client.Update(&setName)
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

func resourceIP4NetworkRead(d *schema.ResourceData, meta interface{}) error {
	mutex.Lock()
	client, err := meta.(*Config).Client()
	if err != nil {
		mutex.Unlock()
		return err
	}

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

	props := strings.Split(*resp.Properties, "|")
	for x := range props {
		if len(props[x]) > 0 {
			prop := strings.Split(props[x], "=")[0]
			val := strings.Split(props[x], "=")[1]

			switch prop {
			case "CIDR":
				netmask, err := strconv.ParseFloat(strings.Split(val, "/")[1], 64)
				if err = gobam.LogoutClientIfError(client, err, "Failed to get IP4 Network netmask"); err != nil {
					mutex.Unlock()
					return err
				}
				addressCount := int(math.Pow(2, (32 - netmask)))

				resp, err := client.GetEntities(*resp.Id, "IP4Address", 0, addressCount)
				if err = gobam.LogoutClientIfError(client, err, "Failed to get child IP4 Addresses"); err != nil {
					mutex.Unlock()
					return err
				}

				addressesInUse := len(resp.Item)
				addressesFree := addressCount - addressesInUse

				d.Set("addresses_in_use", addressesInUse)
				d.Set("addresses_free", addressesFree)
				d.Set("cidr", val)
			case "allowDuplicateHost":
				d.Set("allow_duplicate_host", val)
			case "inheritAllowDuplicateHost":
				b, err := strconv.ParseBool(val)
				if err = gobam.LogoutClientIfError(client, err, "Unable to parse inheritAllowDuplicateHost to bool"); err != nil {
					mutex.Unlock()
					return err
				}
				d.Set("inherit_allow_duplicate_host", b)
			case "pingBeforeAssign":
				d.Set("ping_before_assign", val)
			case "inheritPingBeforeAssign":
				b, err := strconv.ParseBool(val)
				if err = gobam.LogoutClientIfError(client, err, "Unable to parse inheritPingBeforeAssign to bool"); err != nil {
					mutex.Unlock()
					return err
				}
				d.Set("inherit_ping_before_assign", b)
			case "gateway":
				d.Set("gateway", val)
			case "inheritDefaultDomains":
				b, err := strconv.ParseBool(val)
				if err = gobam.LogoutClientIfError(client, err, "Unable to parse inheritDefaultDomains to bool"); err != nil {
					mutex.Unlock()
					return err
				}
				d.Set("inherit_default_domains", b)
			case "defaultView":
				d.Set("default_view", val)
			case "inheritDefaultView":
				b, err := strconv.ParseBool(val)
				if err = gobam.LogoutClientIfError(client, err, "Unable to parse inheritDefaultView to bool"); err != nil {
					mutex.Unlock()
					return err
				}
				d.Set("inherit_default_view", b)
			case "inheritDNSRestrictions":
				b, err := strconv.ParseBool(val)
				if err = gobam.LogoutClientIfError(client, err, "Unable to parse inheritDNSRestrictions to bool"); err != nil {
					mutex.Unlock()
					return err
				}
				d.Set("inherit_dns_restrictions", b)
			default:
				log.Printf("[WARN] Unknown IP4 Address Property: %s", prop)
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

func resourceIP4NetworkUpdate(d *schema.ResourceData, meta interface{}) error {
	mutex.Lock()
	client, err := meta.(*Config).Client()
	if err != nil {
		mutex.Unlock()
		return err
	}

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
	client, err := meta.(*Config).Client()
	if err != nil {
		mutex.Unlock()
		return err
	}

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
