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

func resourceIP4Address() *schema.Resource {
	return &schema.Resource{
		Description: "Resource to reserve an IPv4 address.",

		CreateContext: resourceIP4AddressCreate,
		ReadContext:   resourceIP4AddressRead,
		UpdateContext: resourceIP4AddressUpdate,
		DeleteContext: resourceIP4AddressDelete,

		Schema: map[string]*schema.Schema{
			"configuration_id": {
				Description: "The object ID of the Configuration that will hold the new address. If changed, forces a new resource.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The name assigned to the IPv4 address. This is not related to DNS.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"parent_id": {
				Description: "The object ID of the Configuration, Block, or Network to find the next available IPv4 address in. If changed, forces a new resource.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"action": {
				Description:  "The action to take on the next available IPv4 address.  Must be one of: \"MAKE_STATIC\", \"MAKE_RESERVED\", or \"MAKE_DHCP_RESERVED\". If changed, forces a new resource.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "MAKE_STATIC",
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(gobam.IPAssignmentActions, false),
			},
			"custom_properties": {
				Description: "A map of all custom properties associated with the IPv4 address.",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"mac_address": {
				Description: "The MAC address to associate with the IPv4 address.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"address": {
				Description: "The IPv4 address that was allocated.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"properties": {
				Description: "The properties of the IPv4 address as returned by the API (pipe delimited).",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the IPv4 address.",
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

func resourceIP4AddressCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client

	configID, err := strconv.ParseInt(d.Get("configuration_id").(string), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert configuration_id from string to int64"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	parentIDString := d.Get("parent_id").(string)

	parentID, err := strconv.ParseInt(parentIDString, 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert parent_id from string to int64"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
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
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(*resp.Id, 10))

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return resourceIP4AddressRead(ctx, d, meta)
}

func resourceIP4AddressRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return diag.FromErr(err)
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return nil
}

func resourceIP4AddressUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert id from string to int64"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
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
		return diag.FromErr(err)
	}

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return resourceIP4AddressRead(ctx, d, meta)
}

func resourceIP4AddressDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
