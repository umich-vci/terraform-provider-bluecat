package provider

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/umich-vci/gobam"
)

func resourceHostRecord() *schema.Resource {
	return &schema.Resource{
		Description: "Resource create a host record.",

		CreateContext: resourceHostRecordCreate,
		ReadContext:   resourceHostRecordRead,
		UpdateContext: resourceHostRecordUpdate,
		DeleteContext: resourceHostRecordDelete,

		Schema: map[string]*schema.Schema{
			"addresses": {
				Description: "The address(es) to be associated with the host record.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"dns_zone": {
				Description: "The DNS zone to create the host record in. Combined with `name` to make the fqdn.  If changed, forces a new resource.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the host record to be created. Combined with `dns_zone` to make the fqdn.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"view_id": {
				Description: "The object ID of the View that host record should be created in. If changed, forces a new resource.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"comments": {
				Description: "Comments to be associated with the host record.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"custom_properties": {
				Description: "A map of all custom properties associated with the host record.",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"reverse_record": {
				Description: "If a reverse record should be created for addresses.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"ttl": {
				Description: "The TTL for the host record.  When set to -1, ignores the TTL.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     -1,
			},
			"absolute_name": {
				Description: "The absolute name (fqdn) of the host record.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"properties": {
				Description: "The properties of the host record as returned by the API (pipe delimited).",
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

func resourceHostRecordCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client

	viewID, err := strconv.ParseInt(d.Get("view_id").(string), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert view_id from string to int64"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	absoluteName := d.Get("name").(string) + "." + d.Get("dns_zone").(string)
	ttl := int64(d.Get("ttl").(int))
	rawAddresses := d.Get("addresses").(*schema.Set).List()
	addresses := []string{}
	for x := range rawAddresses {
		addresses = append(addresses, rawAddresses[x].(string))
	}
	reverseRecord := strconv.FormatBool(d.Get("reverse_record").(bool))
	comments := d.Get("comments").(string)
	properties := "reverseRecord=" + reverseRecord + "|comments=" + comments + "|"

	if customProperties, ok := d.GetOk("custom_properties"); ok {
		for k, v := range customProperties.(map[string]interface{}) {
			properties = properties + k + "=" + v.(string) + "|"
		}
	}

	resp, err := client.AddHostRecord(viewID, absoluteName, strings.Join(addresses, ","), ttl, properties)
	if err = gobam.LogoutClientIfError(client, err, "AddHostRecord failed"); err != nil {
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

	return resourceHostRecordRead(ctx, d, meta)
}

func resourceHostRecordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert id from string to int64"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	resp, err := client.GetEntityById(id)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get host record by Id"); err != nil {
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

	hostRecordProperties, err := parseHostRecordProperties(*resp.Properties)
	if err = gobam.LogoutClientIfError(client, err, "Error parsing host record properties"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	d.Set("absolute_name", hostRecordProperties.absoluteName)
	d.Set("reverse_record", hostRecordProperties.reverseRecord)
	d.Set("addresses", hostRecordProperties.addresses)
	d.Set("ttl", hostRecordProperties.ttl)
	d.Set("custom_properties", hostRecordProperties.customProperties)

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return nil
}

func resourceHostRecordUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert id from string to int64"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	name := d.Get("name").(string)
	otype := d.Get("type").(string)
	ttl := strconv.Itoa(d.Get("ttl").(int))
	rawAddresses := d.Get("addresses").(*schema.Set).List()
	addresses := []string{}
	for x := range rawAddresses {
		addresses = append(addresses, rawAddresses[x].(string))
	}
	reverseRecord := strconv.FormatBool(d.Get("reverse_record").(bool))
	comments := d.Get("comments").(string)
	properties := "reverseRecord=" + reverseRecord + "|comments=" + comments + "|ttl=" + ttl + "|addresses=" + strings.Join(addresses, ",") + "|"

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
	if err = gobam.LogoutClientIfError(client, err, "Host Record Update failed"); err != nil {
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

	return resourceHostRecordRead(ctx, d, meta)
}

func resourceHostRecordDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert id from string to int64"); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}

	resp, err := client.GetEntityById(id)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get host record by Id"); err != nil {
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
