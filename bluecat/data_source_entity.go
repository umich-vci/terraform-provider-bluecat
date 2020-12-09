package bluecat

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/umich-vci/gobam"
)

func dataSourceEntity() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceEntityRead,
		Schema: map[string]*schema.Schema{
			"parent_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  0,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"type": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(gobam.ObjectTypes, false),
			},
			"properties": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceEntityRead(d *schema.ResourceData, meta interface{}) error {
	mutex.Lock()
	config := meta.(*Config)
	client, err := meta.(*Config).Client()
	if err != nil {
		mutex.Unlock()
		return err
	}
	err = client.Login(config.Username, config.Password)
	if err != nil {
		return fmt.Errorf("Login error: %s", err)
	}
	parentID, err := strconv.ParseInt(d.Get("parent_id").(string), 10, 64)
	if err = gobam.LogoutClientIfError(client, err, "Unable to convert parent_id from string to int64"); err != nil {
		mutex.Unlock()
		return err
	}
	name := d.Get("name").(string)
	objType := d.Get("type").(string)
	resp, err := client.GetEntityByName(parentID, name, objType)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get entity by name: %s"); err != nil {
		mutex.Unlock()
		return err
	}

	d.SetId(strconv.FormatInt(*resp.Id, 10))
	d.Set("properties", resp.Properties)

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return err
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return nil
}
