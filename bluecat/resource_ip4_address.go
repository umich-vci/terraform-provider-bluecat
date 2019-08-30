package bluecat

import (
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/umich-vci/golang-bluecat"
)

func resourceIP4Address() *schema.Resource {
	return &schema.Resource{
		Create: resourceIP4AddressCreate,
		Read:   resourceIP4AddressRead,
		Update: resourceIP4AddressUpdate,
		Delete: resourceIP4AddressDelete,

		Schema: map[string]*schema.Schema{
			"configuration_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"parent_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mac_address": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			// host records should be created in a separate resource
			// "host_info": &schema.Schema{
			// 	Type:     schema.TypeString,
			// 	Optional: true,
			// 	Default:  "",
			// },
			"action": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "MAKE_STATIC",
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(bam.IPAssignmentActions, false),
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"assigned_date": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"requested_by": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"notes": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"properties": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceIP4AddressCreate(d *schema.ResourceData, meta interface{}) error {
	mutex.Lock()
	client, err := meta.(*Config).Client()
	if err != nil {
		mutex.Unlock()
		return err
	}

	configID, err := strconv.ParseInt(d.Get("configuration_id").(string), 10, 64)
	if err = bam.LogoutClientIfError(client, err, "Unable to convert configuration_id from string to int64"); err != nil {
		mutex.Unlock()
		return err
	}
	parentID, err := strconv.ParseInt(d.Get("parent_id").(string), 10, 64)
	if err = bam.LogoutClientIfError(client, err, "Unable to convert parent_id from string to int64"); err != nil {
		mutex.Unlock()
		return err
	}
	macAddress := d.Get("mac_address").(string)
	//hostInfo := d.Get("host_info").(string)
	hostInfo := "" // host records should be created as a separate resource
	action := d.Get("action").(string)
	requestedBy := d.Get("requested_by").(string)
	assignedDate := d.Get("assigned_date").(string)
	notes := d.Get("notes").(string)
	name := d.Get("name").(string)
	properties := "Requested_by=" + requestedBy + "|Assigned_Date=" + assignedDate + "|Notes=" + notes + "|name=" + name

	resp, err := client.AssignNextAvailableIP4Address(configID, parentID, macAddress, hostInfo, action, properties)
	if err = bam.LogoutClientIfError(client, err, "AssignNextAvailableIP4Address failed"); err != nil {
		mutex.Unlock()
		return err
	}

	d.SetId(strconv.FormatInt(*resp.Id, 10))
	d.Set("name", resp.Name)
	d.Set("properties", resp.Properties)
	d.Set("type", resp.Type)

	props := strings.Split(*resp.Properties, "|")
	for x := range props {
		if len(props[x]) > 0 {
			prop := strings.Split(props[x], "=")[0]
			val := strings.Split(props[x], "=")[1]

			switch prop {
			case "Assigned_Date":
				d.Set("assigned_date", val)
			case "Requested_by":
				d.Set("requested_by", val)
			case "Notes":
				d.Set("notes", val)
			case "address":
				d.Set("address", val)
			case "state":
				d.Set("state", val)
			default:
				log.Printf("[WARN]Unknown IP4 Address Property: %s", prop)
			}
		}
	}

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return err
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return resourceIP4AddressRead(d, meta)
}

func resourceIP4AddressRead(d *schema.ResourceData, meta interface{}) error {
	mutex.Lock()
	client, err := meta.(*Config).Client()
	if err != nil {
		mutex.Unlock()
		return err
	}
	parentID, err := strconv.ParseInt(d.Get("parent_id").(string), 10, 64)
	if err = bam.LogoutClientIfError(client, err, "Unable to convert parent_id from string to int64"); err != nil {
		mutex.Unlock()
		return err
	}
	address := d.Get("address").(string)

	resp, err := client.GetIP4Address(parentID, address)
	if err = bam.LogoutClientIfError(client, err, "Failed to get IP4 Address"); err != nil {
		mutex.Unlock()
		return err
	}

	d.SetId(strconv.FormatInt(*resp.Id, 10))
	d.Set("name", resp.Name)
	d.Set("properties", resp.Properties)
	d.Set("type", resp.Type)

	props := strings.Split(*resp.Properties, "|")
	for x := range props {
		if len(props[x]) > 0 {
			prop := strings.Split(props[x], "=")[0]
			val := strings.Split(props[x], "=")[1]

			switch prop {
			case "Assigned_Date":
				d.Set("assigned_date", val)
			case "Requested_by":
				d.Set("requested_by", val)
			case "Notes":
				d.Set("notes", val)
			case "address":
				// since we have to pass in an address to read it we don't really need this
				// 	d.Set("address", val)
			case "state":
				d.Set("state", val)
			default:
				log.Printf("[WARN]Unknown IP4 Address Property: %s", prop)
			}
		}
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

func resourceIP4AddressUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceIP4AddressRead(d, meta)
}

func resourceIP4AddressDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
