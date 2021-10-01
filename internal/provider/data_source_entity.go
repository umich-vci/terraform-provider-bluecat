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
		Description: "Data source to access the attributes of a BlueCat entity.",

		ReadContext: dataSourceEntityRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the entity to find.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description:  "The type of the entity you want to retrieve.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(gobam.ObjectTypes, false),
			},
			"parent_id": {
				Description: "The object ID of the parent object that contains the entity. Configurations are stored in ID `0`.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     0,
			},
			"properties": {
				Description: "The properties of the entity as returned by the API (pipe delimited).",
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

	if *resp.Id == 0 {
		var diags diag.Diagnostics
		err := gobam.LogoutClientWithError(client, "Entity not found")
		mutex.Unlock()

		diags = append(diags, diag.FromErr(err)...)
		diags = append(diags, diag.Errorf("Entity ID returned was 0")...)

		return diags
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
