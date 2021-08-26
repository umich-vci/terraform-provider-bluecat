package provider

import (
	"context"
	"hash/crc64"
	"log"
	"math/rand"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/umich-vci/gobam"
)

func resourceIP4AvailableNetwork() *schema.Resource {
	return &schema.Resource{
		Description: "Resource to select an IPv4 network from a list of networks based on availability of IP addresses.",

		CreateContext: resourceIP4AvailableNetworkCreate,
		ReadContext:   schema.NoopContext,
		DeleteContext: resourceIP4AvailableNetworkDelete,

		Schema: map[string]*schema.Schema{
			"network_id_list": {
				Description: "A list of Network IDs to search for a free IP address. By default, the address with the most free addresses will be returned. See the `random` argument for another selection method. The resource will be recreated if the network_id_list is changed. You may want to use a `lifecycle` customization to ignore changes to the list after resource creation so that a new network is not selected if the list is changed.",
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"keepers": {
				Description: "An arbitrary map of values. If this argument is changed, then the resource will be recreated.",
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
			},
			"random": {
				Description: "By default, the network with the most free IP addresses is returned. By setting this to `true` a random network from the list will be returned instead. The network will be validated to have at least 1 free IP address.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
			},
			"seed": {
				Description: "A seed for the `random` argument's generator. Can be used to try to get more predictable results from the random selection. The results will not be fixed however.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"network_id": {
				Description: "The network ID of the network selected by the resource.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}
func resourceIP4AvailableNetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex.Lock()
	client := meta.(*apiClient).Client

	result := -1

	networkIDList := d.Get("network_id_list").([]interface{})
	seed := d.Get("seed").(string)
	random := d.Get("random").(bool)

	if len(networkIDList) == 0 {
		err := gobam.LogoutClientWithError(client, "network_id_list cannot be empty")
		mutex.Unlock()
		return diag.FromErr(err)
	}

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
					return diag.FromErr(err)
				}

				networkProperties, err := gobam.ParseIP4NetworkProperties(*resp.Properties)
				if err = gobam.LogoutClientIfError(client, err, "Error parsing IP4 network properties"); err != nil {
					mutex.Unlock()
					return diag.FromErr(err)
				}

				_, addressesFree, err := getIP4NetworkAddressUsage(*resp.Id, networkProperties.CIDR, client)
				if err = gobam.LogoutClientIfError(client, err, "Error calculating network usage"); err != nil {
					mutex.Unlock()
					return diag.FromErr(err)
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
				return diag.FromErr(err)
			}

			networkProperties, err := gobam.ParseIP4NetworkProperties(*resp.Properties)
			if err = gobam.LogoutClientIfError(client, err, "Error parsing IP4 network properties"); err != nil {
				mutex.Unlock()
				return diag.FromErr(err)
			}

			_, addressesFree, err := getIP4NetworkAddressUsage(*resp.Id, networkProperties.CIDR, client)
			if err = gobam.LogoutClientIfError(client, err, "Error calculating network usage"); err != nil {
				mutex.Unlock()
				return diag.FromErr(err)
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
		return diag.FromErr(err)
	}

	d.SetId("-")
	d.Set("network_id", result)

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		return diag.FromErr(err)
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	return nil
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

func resourceIP4AvailableNetworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}
