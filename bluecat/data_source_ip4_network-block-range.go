package bluecat

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/umich-vci/golang-bluecat"
)

func dataSourceIP4Network() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceIP4NetworkRead,
		Schema: map[string]*schema.Schema{
			"container_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"type": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"IP4Block", "IP4Network", "DHCP4Range", ""}, false),
			},
			"address": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"properties": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"cidr": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"allow_duplicate_host": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"inherit_allow_duplicate_host": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ping_before_assign": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"inherit_ping_before_assign": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"reference": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"gateway": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"inherit_default_domains": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"default_view": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"inherit_default_view": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"inherit_dns_restrictions": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"addresses_in_use": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"addresses_free": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"custom_properties": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func dataSourceIP4NetworkRead(d *schema.ResourceData, meta interface{}) error {
	mutex.Lock()
	client, err := meta.(*Config).Client()
	if err != nil {
		mutex.Unlock()
		return err
	}

	containerID, err := strconv.ParseInt(d.Get("container_id").(string), 10, 64)
	if err = bam.LogoutClientIfError(client, err, "Unable to convert container_id from string to int64"); err != nil {
		mutex.Unlock()
		return err
	}
	otype := d.Get("type").(string)
	address := d.Get("address").(string)

	resp, err := client.GetIPRangedByIP(containerID, otype, address)
	if err = bam.LogoutClientIfError(client, err, "Failed to get IP4 Networks by hint"); err != nil {
		mutex.Unlock()
		return err
	}

	d.SetId(strconv.FormatInt(*resp.Id, 10))
	d.Set("name", *resp.Name)
	d.Set("properties", *resp.Properties)
	d.Set("type", *resp.Type)

	networkProperties, err := parseIP4NetworkProperties(*resp.Properties)
	if err = bam.LogoutClientIfError(client, err, "Error parsing host record properties"); err != nil {
		mutex.Unlock()
		return err
	}

	d.Set("cidr", networkProperties.cidr)
	d.Set("allow_duplicate_host", networkProperties.allowDuplicateHost)
	d.Set("inherit_allow_duplicate_host", networkProperties.inheritAllowDuplicateHost)
	d.Set("inherit_ping_before_assign", networkProperties.inheritPingBeforeAssign)
	d.Set("reference", networkProperties.reference)
	d.Set("ping_before_assign", networkProperties.pingBeforeAssign)
	d.Set("gateway", networkProperties.gateway)
	d.Set("inherit_default_domains", networkProperties.inheritDefaultDomains)
	d.Set("default_view", networkProperties.defaultView)
	d.Set("inherit_default_view", networkProperties.inheritDefaultView)
	d.Set("inherit_dns_restrictions", networkProperties.inheritDNSRestrictions)
	d.Set("custom_properties", networkProperties.customProperties)

	addressesInUse, addressesFree, err := getIP4NetworkAddressUsage(*resp.Id, networkProperties.cidr, client)
	d.Set("addresses_in_use", addressesInUse)
	d.Set("addresses_free", addressesFree)

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return err
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return nil
}

type ip4NetworkProperties struct {
	cidr                      string
	allowDuplicateHost        string
	inheritAllowDuplicateHost bool
	pingBeforeAssign          string
	inheritPingBeforeAssign   bool
	reference                 string
	gateway                   string
	inheritDefaultDomains     bool
	defaultView               string
	inheritDefaultView        bool
	inheritDNSRestrictions    bool
	customProperties          map[string]string
}

func parseIP4NetworkProperties(properties string) (ip4NetworkProperties, error) {
	var networkProperties ip4NetworkProperties

	props := strings.Split(properties, "|")
	for x := range props {
		if len(props[x]) > 0 {
			prop := strings.Split(props[x], "=")[0]
			val := strings.Split(props[x], "=")[1]

			switch prop {
			case "CIDR":
				networkProperties.cidr = val
			case "allowDuplicateHost":
				networkProperties.allowDuplicateHost = val
			case "inheritAllowDuplicateHost":
				b, err := strconv.ParseBool(val)
				if err != nil {
					return networkProperties, fmt.Errorf("Error parsing inheritAllowDuplicateHost to bool")
				}
				networkProperties.inheritAllowDuplicateHost = b
			case "pingBeforeAssign":
				networkProperties.pingBeforeAssign = val
			case "inheritPingBeforeAssign":
				b, err := strconv.ParseBool(val)
				if err != nil {
					return networkProperties, fmt.Errorf("Error parsing inheritPingBeforeAssign to bool")
				}
				networkProperties.inheritPingBeforeAssign = b
			case "reference":
				networkProperties.reference = val
			case "gateway":
				networkProperties.gateway = val
			case "inheritDefaultDomains":
				b, err := strconv.ParseBool(val)
				if err != nil {
					return networkProperties, fmt.Errorf("Error parsing inheritDefaultDomains to bool")
				}
				networkProperties.inheritDefaultDomains = b
			case "defaultView":
				networkProperties.defaultView = val
			case "inheritDefaultView":
				b, err := strconv.ParseBool(val)
				if err != nil {
					return networkProperties, fmt.Errorf("Error parsing inheritDefaultView to bool")
				}
				networkProperties.inheritDefaultView = b
			case "inheritDNSRestrictions":
				b, err := strconv.ParseBool(val)
				if err != nil {
					return networkProperties, fmt.Errorf("Error parsing inheritDNSRestrictions to bool")
				}
				networkProperties.inheritDNSRestrictions = b
			default:
				networkProperties.customProperties[prop] = val
			}
		}
	}

	return networkProperties, nil
}

func getIP4NetworkAddressUsage(id int64, cidr string, client bam.ProteusAPI) (int, int, error) {

	netmask, err := strconv.ParseFloat(strings.Split(cidr, "/")[1], 64)
	if err != nil {
		mutex.Unlock()
		return 0, 0, fmt.Errorf("Error parsing netmask from cidr string")
	}
	addressCount := int(math.Pow(2, (32 - netmask)))

	resp, err := client.GetEntities(id, "IP4Address", 0, addressCount)
	if err != nil {
		return 0, 0, err
	}

	addressesInUse := len(resp.Item)
	addressesFree := addressCount - addressesInUse

	return addressesInUse, addressesFree, nil
}
