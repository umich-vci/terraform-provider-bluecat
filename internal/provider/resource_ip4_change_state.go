package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/umich-vci/gobam"
	"log"
	"strconv"
)

func resourceIP4ChangeState() *schema.Resource {
	return &schema.Resource{
		Description: "Resource change the state of a IP4 resource that was created outside of the context of this provider (i.e. a DHCP Reservation).",

		CreateContext: resourceIP4ChangeStateCreate,
		ReadContext:   schema.NoopContext,
		DeleteContext: resourceIP4ChangeStateDelete,

		Schema: map[string]*schema.Schema{
			"address_id": {
				Description: "The object ID of the Address that will have its state changed. If changed, forces a new resource.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"action": {
				Description:  "The action to take on the provided IPv4 address.  Must be one of: \"MAKE_STATIC\", \"MAKE_RESERVED\", or \"MAKE_DHCP_RESERVED\". If changed, forces a new resource.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(gobam.IPAssignmentActions, false),
			},
			"mac_address": {
				Description: "The MAC address to associate with the IPv4 address.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "",
			},
			"state": {
				Description: "The state of the IPv4 address.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceIP4ChangeStateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client
	client.Login(meta.(*apiClient).Username, meta.(*apiClient).Password)

	addressId, err := strconv.ParseInt(d.Get("address_id").(string), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert address_id from string to int64"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	macAddress := d.Get("mac_address").(string)
	action := d.Get("action").(string)

	if err := client.ChangeStateIP4Address(addressId, action, macAddress); err != nil {
		if err = gobam.LogoutClientIfError(client, err, "ChangeStateIP4Address failed"); err != nil {
			mutex.Unlock()
			return diag.FromErr(err)
		}
	}

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	d.SetId(d.Get("address_id").(string))
	d.Set("state", action)

	return nil
}

func resourceIP4ChangeStateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")

	return nil
}
