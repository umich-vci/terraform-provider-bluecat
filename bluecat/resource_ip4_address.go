package bluecat

import (
	"log"
	"strconv"

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
				Optional: true,
				ForceNew: true,
			},
			"parent_id_list": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
			"custom_properties": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
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

	parentIDString := ""
	if pid, ok := d.GetOk("parent_id"); ok {
		parentIDString = pid.(string)

		if _, ok := d.GetOk("parent_id_list"); ok {
			err := bam.LogoutClientWithError(client, "Cannot specify both parent_id and parent_id_list")
			mutex.Unlock()
			return err
		}

	} else {
		if pidList, ok := d.GetOk("parent_id_list"); ok {
			list := pidList.(*schema.Set).List()
			freeAddressMap := make(map[string]int)
			for i := range list {
				id, err := strconv.ParseInt(list[i].(string), 10, 64)
				if err = bam.LogoutClientIfError(client, err, "Unable to convert parent_id from string to int64"); err != nil {
					mutex.Unlock()
					return err
				}
				resp, err := client.GetEntityById(id)
				if err = bam.LogoutClientIfError(client, err, "Failed to get IP4 Network by Id"); err != nil {
					mutex.Unlock()
					return err
				}

				networkProperties, err := parseIP4NetworkProperties(*resp.Properties)
				if err = bam.LogoutClientIfError(client, err, "Error parsing IP4 network properties"); err != nil {
					mutex.Unlock()
					return err
				}

				_, addressesFree, err := getIP4NetworkAddressUsage(*resp.Id, networkProperties.cidr, client)
				if err = bam.LogoutClientIfError(client, err, "Error calculating network usage"); err != nil {
					mutex.Unlock()
					return err
				}

				if addressesFree > 0 {
					freeAddressMap[strconv.FormatInt(id, 10)] = addressesFree
				}

			}

			parentIDMostFree := ""
			freeCount := 0
			for k, v := range freeAddressMap {
				if v > freeCount {
					freeCount = v
					parentIDMostFree = k
				}
			}

			if freeCount == 0 {
				err := bam.LogoutClientWithError(client, "No networks had a free address")
				mutex.Unlock()
				return err
			}

			parentIDString = parentIDMostFree
		} else {
			err := bam.LogoutClientWithError(client, "One of parent_id or parent_id_list must be specified")
			mutex.Unlock()
			return err
		}
	}

	parentID, err := strconv.ParseInt(parentIDString, 10, 64)
	if err = bam.LogoutClientIfError(client, err, "Unable to convert parent_id from string to int64"); err != nil {
		mutex.Unlock()
		return err
	}
	macAddress := d.Get("mac_address").(string)
	hostInfo := "" // host records should be created as a separate resource
	action := d.Get("action").(string)
	name := d.Get("name").(string)
	properties := "name=" + name + "|"

	if customProperties, ok := d.GetOk("custom_properties"); ok {
		for k, v := range customProperties.(map[string]interface{}) {
			properties = properties + k + "=" + v.(string) + "|"
		}
	}

	resp, err := client.AssignNextAvailableIP4Address(configID, parentID, macAddress, hostInfo, action, properties)
	if err = bam.LogoutClientIfError(client, err, "AssignNextAvailableIP4Address failed"); err != nil {
		mutex.Unlock()
		return err
	}

	d.SetId(strconv.FormatInt(*resp.Id, 10))
	//need to set parent ID here in case we selected one from parent_id_list
	d.Set("parent_id", parentIDString)

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

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err = bam.LogoutClientIfError(client, err, "Unable to convert id from string to int64"); err != nil {
		mutex.Unlock()
		return err
	}

	resp, err := client.GetEntityById(id)
	if err = bam.LogoutClientIfError(client, err, "Failed to get IP4 Address by Id"); err != nil {
		mutex.Unlock()
		return err
	}

	d.Set("name", resp.Name)
	d.Set("properties", resp.Properties)
	d.Set("type", resp.Type)

	addressProperties := parseIP4AddressProperties(*resp.Properties)
	d.Set("address", addressProperties.address)
	d.Set("state", addressProperties.state)
	d.Set("mac_address", addressProperties.macAddress)
	d.Set("custom_properties", addressProperties.customProperties)

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
	mutex.Lock()
	client, err := meta.(*Config).Client()
	if err != nil {
		mutex.Unlock()
		return err
	}

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err = bam.LogoutClientIfError(client, err, "Unable to convert id from string to int64"); err != nil {
		mutex.Unlock()
		return err
	}

	macAddress := d.Get("mac_address").(string)
	name := d.Get("name").(string)
	otype := d.Get("type").(string)
	properties := "name=" + name + "|"

	if macAddress != "" {
		properties = properties + "macAddress=" + macAddress + "|"
	}

	if customProperties, ok := d.GetOk("custom_properties"); ok {
		for k, v := range customProperties.(map[string]string) {
			properties = properties + k + "=" + v + "|"
		}
	}

	update := bam.APIEntity{
		Id:         &id,
		Name:       &name,
		Properties: &properties,
		Type:       &otype,
	}

	err = client.Update(&update)
	if err = bam.LogoutClientIfError(client, err, "IP4 Address Update failed"); err != nil {
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

	return resourceIP4AddressRead(d, meta)
}

func resourceIP4AddressDelete(d *schema.ResourceData, meta interface{}) error {
	mutex.Lock()
	client, err := meta.(*Config).Client()
	if err != nil {
		mutex.Unlock()
		return err
	}

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err = bam.LogoutClientIfError(client, err, "Unable to convert id from string to int64"); err != nil {
		mutex.Unlock()
		return err
	}

	resp, err := client.GetEntityById(id)
	if err = bam.LogoutClientIfError(client, err, "Failed to get IP4 Address by Id"); err != nil {
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
	if err = bam.LogoutClientIfError(client, err, "Delete failed"); err != nil {
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
