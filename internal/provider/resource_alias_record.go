package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/umich-vci/gobam"
	"log"
	"strconv"
)

func resourceAliasRecord() *schema.Resource {
	return &schema.Resource{
		Description: "Resource create an alias record.",

		CreateContext: resourceAliasRecordCreate,
		ReadContext:   resourceAliasRecordRead,
		UpdateContext: resourceAliasRecordUpdate,
		DeleteContext: resourceAliasRecordDelete,

		Schema: map[string]*schema.Schema{
			"linked_record_name": {
				Description: "The record to which this alias should point.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"dns_zone": {
				Description: "The DNS zone to create the alias record in. Combined with `name` to make the fqdn.  If changed, forces a new resource.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the alias record to be created. Combined with `dns_zone` to make the fqdn.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"view_id": {
				Description: "The object ID of the View that alias record should be created in. If changed, forces a new resource.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"custom_properties": {
				Description: "A map of all custom properties associated with the alias record.",
				Type:        schema.TypeMap,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ttl": {
				Description: "The TTL for the alias record.  When set to -1, ignores the TTL.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     -1,
			},
			"absolute_name": {
				Description: "The absolute name (fqdn) of the alias record.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"properties": {
				Description: "The properties of the alias record as returned by the API (pipe delimited).",
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

func resourceAliasRecordCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client
	client.Login(meta.(*apiClient).Username, meta.(*apiClient).Password)

	viewID, err := strconv.ParseInt(d.Get("view_id").(string), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert view_id from string to int64"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	absoluteName := d.Get("name").(string) + "." + d.Get("dns_zone").(string)
	ttl := int64(d.Get("ttl").(int))
	linkedRecordName := d.Get("linked_record_name").(string)
	properties := ""

	if customProperties, ok := d.GetOk("custom_properties"); ok {
		for k, v := range customProperties.(map[string]interface{}) {
			properties = properties + k + "=" + v.(string) + "|"
		}
	}

	resp, err := client.AddAliasRecord(viewID, absoluteName, linkedRecordName, ttl, properties)
	if err = gobam.LogoutClientIfError(client, err, "AddAliasRecord failed"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(resp, 10))

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return resourceAliasRecordRead(ctx, d, meta)
}

func resourceAliasRecordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client
	client.Login(meta.(*apiClient).Username, meta.(*apiClient).Password)

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert id from string to int64"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	resp, err := client.GetEntityById(id)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get alias record by Id"); err != nil {
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

	aliasRecordProperties, err := parseAliasRecordProperties(*resp.Properties)
	if err = gobam.LogoutClientIfError(client, err, "Error parsing alias record properties"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	d.Set("absolute_name", aliasRecordProperties.absoluteName)
	d.Set("linked_record_name", aliasRecordProperties.linkedRecordName)
	d.Set("ttl", aliasRecordProperties.ttl)
	d.Set("custom_properties", aliasRecordProperties.customProperties)

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return nil
}

func resourceAliasRecordUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client
	client.Login(meta.(*apiClient).Username, meta.(*apiClient).Password)

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert id from string to int64"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	name := d.Get("name").(string)
	otype := d.Get("type").(string)
	ttl := strconv.Itoa(d.Get("ttl").(int))
	linkedRecordName := d.Get("linked_record_name").(string)
	properties := "linkedRecordName=" + linkedRecordName + "|ttl=" + ttl + "|"

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
	if err = gobam.LogoutClientIfError(client, err, "Alias Record Update failed"); err != nil {
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

	return resourceAliasRecordRead(ctx, d, meta)
}

func resourceAliasRecordDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client
	client.Login(meta.(*apiClient).Username, meta.(*apiClient).Password)

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert id from string to int64"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	resp, err := client.GetEntityById(id)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get alias record by Id"); err != nil {
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
