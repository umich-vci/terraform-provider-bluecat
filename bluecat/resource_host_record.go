package bluecat

import (
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/umich-vci/gobam"
)

func resourceHostRecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceHostRecordCreate,
		Read:   resourceHostRecordRead,
		Update: resourceHostRecordUpdate,
		Delete: resourceHostRecordDelete,
		Schema: map[string]*schema.Schema{
			"view_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"dns_zone": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"addresses": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ttl": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},
			"reverse_record": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"comments": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"properties": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"absolute_name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_properties": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceHostRecordCreate(d *schema.ResourceData, meta interface{}) error {
	mutex.Lock()
	client, err := meta.(*Config).Client()
	if err != nil {
		mutex.Unlock()
		return err
	}

	viewID, err := strconv.ParseInt(d.Get("view_id").(string), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert view_id from string to int64"); err != nil {
		mutex.Unlock()
		return err
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
		return err
	}

	d.SetId(strconv.FormatInt(resp, 10))

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return err
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return resourceHostRecordRead(d, meta)
}

func resourceHostRecordRead(d *schema.ResourceData, meta interface{}) error {
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
	if err = gobam.LogoutClientIfError(client, err, "Failed to get host record by Id"); err != nil {
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

	hostRecordProperties, err := parseHostRecordProperties(*resp.Properties)
	if err = gobam.LogoutClientIfError(client, err, "Error parsing host record properties"); err != nil {
		mutex.Unlock()
		return err
	}

	d.Set("absolute_name", hostRecordProperties.absoluteName)
	d.Set("reverse_record", hostRecordProperties.reverseRecord)
	d.Set("addresses", hostRecordProperties.addresses)
	d.Set("ttl", hostRecordProperties.ttl)
	d.Set("custom_properties", hostRecordProperties.customProperties)

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return err
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return nil
}

func resourceHostRecordUpdate(d *schema.ResourceData, meta interface{}) error {
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
		return err
	}

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return err
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return resourceHostRecordRead(d, meta)
}

func resourceHostRecordDelete(d *schema.ResourceData, meta interface{}) error {
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
	if err = gobam.LogoutClientIfError(client, err, "Failed to get host record by Id"); err != nil {
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
