package bluecat

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/tiaguinho/gosoap"
)

// getEntityByNameResponse will hold the SOAP response
type getEntityByNameResponse struct {
	ID         string `xml:"return.id"`
	Name       string `xml:"return.name"`
	Properties string `xml:"return.properties"`
	Type       string `xml:"return.type"`
}

func dataSourceEntityByName() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceEntityByNameRead,
		Schema: map[string]*schema.Schema{
			"parent_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				Default:  0,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"type": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ObjectTypes, false),
			},
			"properties": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceEntityByNameRead(d *schema.ResourceData, meta interface{}) error {
	mutex.Lock()

	client, err := meta.(*Config).Client()
	if err != nil {
		mutex.Unlock()
		return err
	}

	params := gosoap.Params{
		"parentId": d.Get("parentId").(int),
		"name":     d.Get("name").(string),
		"type":     d.Get("type").(string),
	}

	resp, err := client.Call("getEntityByName", params)

	if err = logoutClientIfError(client, err, "Failed to get entity by name: %s"); err != nil {
		mutex.Unlock()
		return err
	}

	var r getEntityByNameResponse
	resp.Unmarshal(&r)

	d.SetId(r.ID)
	d.Set("properties", r.Properties)

	// logout client
	if _, err := client.Call("logout", params); err != nil {
		mutex.Unlock()
		return err
	}

	mutex.Unlock()

	return nil
}
