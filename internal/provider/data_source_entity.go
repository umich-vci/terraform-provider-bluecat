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

func dataSourceEntity() *schema.Resource {
	return &schema.Resource{
		Description: "",

		ReadContext: dataSourceEntityRead,

		Schema: map[string]*schema.Schema{
			"parent_id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     0,
			},
			"name": {
				Description: "",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description:  "",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(gobam.ObjectTypes, false),
			},
			"properties": {
				Description: "",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceEntityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client

	parentID, err := strconv.ParseInt(d.Get("parent_id").(string), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert parent_id from string to int64"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	name := d.Get("name").(string)
	objType := d.Get("type").(string)
	resp, err := client.GetEntityByName(parentID, name, objType)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get entity by name: %s"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(*resp.Id, 10))
	d.Set("properties", resp.Properties)

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return nil
}
