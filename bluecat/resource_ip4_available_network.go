package bluecat

import (
	"hash/crc64"
	"log"
	"math/rand"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/umich-vci/gobam"
)

func resourceIP4AvailableNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceIP4AvailableNetworkCreate,
		Read:   schema.Noop,
		Delete: schema.RemoveFromState,
		Schema: map[string]*schema.Schema{
			"network_id_list": &schema.Schema{
				Type:         schema.TypeList,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"keepers": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
			"random": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"seed": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"network_id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}
func resourceIP4AvailableNetworkCreate(d *schema.ResourceData, meta interface{}) error {
	mutex.Lock()
	client, err := meta.(*Config).Client()
	if err != nil {
		mutex.Unlock()
		return err
	}

	result := -1

	networkIDList := d.Get("network_id_list").([]interface{})
	seed := d.Get("seed").(string)
	random := d.Get("random").(bool)

	if random {
		rand := NewRand(seed)

		// Keep producing permutations until we fill our result
	Batches:
		for {
			perm := rand.Perm(len(networkIDList))

			for _, i := range perm {
				id := int64(networkIDList[i].(int))

				resp, err := client.GetEntityById(id)
				if err = gobam.LogoutClientIfError(client, err, "Failed to get IP4 Network by Id"); err != nil {
					mutex.Unlock()
					return err
				}

				networkProperties, err := gobam.ParseIP4NetworkProperties(*resp.Properties)
				if err = gobam.LogoutClientIfError(client, err, "Error parsing IP4 network properties"); err != nil {
					mutex.Unlock()
					return err
				}

				_, addressesFree, err := getIP4NetworkAddressUsage(*resp.Id, networkProperties.CIDR, client)
				if err = gobam.LogoutClientIfError(client, err, "Error calculating network usage"); err != nil {
					mutex.Unlock()
					return err
				}

				if addressesFree > 0 {
					result = networkIDList[i].(int)
					break Batches
				}
			}
		}

	} else {

		freeAddressMap := make(map[int64]int)
		for i := range networkIDList {
			id := int64(networkIDList[i].(int))

			resp, err := client.GetEntityById(id)
			if err = gobam.LogoutClientIfError(client, err, "Failed to get IP4 Network by Id"); err != nil {
				mutex.Unlock()
				return err
			}

			networkProperties, err := gobam.ParseIP4NetworkProperties(*resp.Properties)
			if err = gobam.LogoutClientIfError(client, err, "Error parsing IP4 network properties"); err != nil {
				mutex.Unlock()
				return err
			}

			_, addressesFree, err := getIP4NetworkAddressUsage(*resp.Id, networkProperties.CIDR, client)
			if err = gobam.LogoutClientIfError(client, err, "Error calculating network usage"); err != nil {
				mutex.Unlock()
				return err
			}

			if addressesFree > 0 {
				freeAddressMap[id] = addressesFree
			}

		}

		freeCount := 0
		for k, v := range freeAddressMap {
			if v > freeCount {
				freeCount = v
				result = int(k)
			}
		}
	}

	if result == -1 {
		err := gobam.LogoutClientWithError(client, "No networks had a free address")
		mutex.Unlock()
		return err
	}

	d.SetId("-")
	d.Set("network_id", result)

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return err
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return resourceIP4NetworkRead(d, meta)
}

// NewRand returns a seeded random number generator, using a seed derived
// from the provided string.
//
// If the seed string is empty, the current time is used as a seed.
func NewRand(seed string) *rand.Rand {
	var seedInt int64
	if seed != "" {
		crcTable := crc64.MakeTable(crc64.ISO)
		seedInt = int64(crc64.Checksum([]byte(seed), crcTable))
	} else {
		seedInt = time.Now().UnixNano()
	}

	randSource := rand.NewSource(seedInt)
	return rand.New(randSource)
}
