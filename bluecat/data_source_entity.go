package bluecat

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/umich-vci/golang-bluecat"
)

func dataSourceEntity() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceEntityRead,
		Schema: map[string]*schema.Schema{
			"parent_id": &schema.Schema{
				Type:     schema.TypeInt,
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
				ValidateFunc: validation.StringInSlice(bam.ObjectTypes, false),
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
	parentID := d.Get("parent_id").(int)
	name := d.Get("name").(string)
	objType := d.Get("type").(string)
	resp, err := client.GetEntityByName(int64(parentID), name, objType)
	if err = bam.LogoutClientIfError(client, err, "Failed to get entity by name: %s"); err != nil {
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
