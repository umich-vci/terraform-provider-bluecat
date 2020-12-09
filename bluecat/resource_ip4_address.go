package bluecat

import (
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/umich-vci/gobam"
)

func resourceIP4Address() *schema.Resource {
	return &schema.Resource{
		Create: resourceIP4AddressCreate,
		Read:   resourceIP4AddressRead,
		Update: resourceIP4AddressUpdate,
		Delete: resourceIP4AddressDelete,

		Schema: map[string]*schema.Schema{
			"configuration_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"parent_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mac_address": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			// host records should be created in a separate resource
			// "host_info": &schema.Schema{
			// 	Type:     schema.TypeString,
			// 	Optional: true,
			// 	Default:  "",
			// },
			"action": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "MAKE_STATIC",
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(gobam.IPAssignmentActions, false),
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"properties": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_properties": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
	}
}

func resourceIP4AddressCreate(d *schema.ResourceData, meta interface{}) error {
	mutex.Lock()
	client, err := meta.(*Config).Client()
	if err != nil {
		mutex.Unlock()
		return err
	}

	configID, err := strconv.ParseInt(d.Get("configuration_id").(string), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert configuration_id from string to int64"); err != nil {
		mutex.Unlock()
		return err
	}

	parentIDString := d.Get("parent_id").(string)

	parentID, err := strconv.ParseInt(parentIDString, 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert parent_id from string to int64"); err != nil {
		mutex.Unlock()
		return err
	}
	macAddress := d.Get("mac_address").(string)
	hostInfo := "" // host records should be created as a separate resource
	action := d.Get("action").(string)
	name := d.Get("name").(string)
	properties := "name=" + name + "|"
	customProperties := make(map[string]string)
	if rawCustomProperties, ok := d.GetOk("custom_properties"); ok {
		for k, v := range rawCustomProperties.(map[string]interface{}) {
			customProperties[k] = v.(string)
			properties = properties + k + "=" + v.(string) + "|"
		}
	}

	resp, err := client.AssignNextAvailableIP4Address(configID, parentID, macAddress, hostInfo, action, properties)
	if err = gobam.LogoutClientIfError(client, err, "AssignNextAvailableIP4Address failed"); err != nil {
		mutex.Unlock()
		return err
	}

	d.SetId(strconv.FormatInt(*resp.Id, 10))

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return err
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return resourceIP4AddressRead(d, meta)
}

func resourceIP4AddressRead(d *schema.ResourceData, meta interface{}) error {
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

	d.Set("name", resp.Name)
	d.Set("properties", resp.Properties)
	d.Set("type", resp.Type)

	addressProperties := parseIP4AddressProperties(*resp.Properties)
	d.Set("address", addressProperties.address)
	d.Set("state", addressProperties.state)
	d.Set("mac_address", addressProperties.macAddress)
	d.Set("custom_properties", addressProperties.customProperties)

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return err
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return nil
}

func resourceIP4AddressUpdate(d *schema.ResourceData, meta interface{}) error {
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

	macAddress := d.Get("mac_address").(string)
	name := d.Get("name").(string)
	otype := d.Get("type").(string)
	properties := "name=" + name + "|"

	if macAddress != "" {
		properties = properties + "macAddress=" + macAddress + "|"
	}

	if customProperties, ok := d.GetOk("custom_properties"); ok {
		for k, v := range customProperties.(map[string]string) {
			properties = properties + k + "=" + v + "|"
		}
	}

	update := gobam.APIEntity{
		Id:         &id,
		Name:       &name,
		Properties: &properties,
		Type:       &otype,
	}

	err = client.Update(&update)
	if err = gobam.LogoutClientIfError(client, err, "IP4 Address Update failed"); err != nil {
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

	return resourceIP4AddressRead(d, meta)
}

func resourceIP4AddressDelete(d *schema.ResourceData, meta interface{}) error {
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
