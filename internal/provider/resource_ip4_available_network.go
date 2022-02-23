package provider

import (
	"context"
	"hash/crc64"
	"log"
	"math/rand"
	"strconv"
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
				Description: "A list of Network IDs to search for a free IP address. By default, the network with the most free addresses will be returned. See the `random` argument for another selection method. The resource will be recreated if the network_id_list is changed. You may want to use a `lifecycle` customization to ignore changes to the list after resource creation so that a new network is not selected if the list is changed.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"network_name_list": {
				Description: "A list of Network Names to search for a free IP address. By default, the network with the most free addresses will be returned. See the `random` argument for another selection method. The resource will be recreated if the network_id_list is changed. You may want to use a `lifecycle` customization to ignore changes to the list after resource creation so that a new network is not selected if the list is changed.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"container_id": {
				Description: "The object ID of a container that contains the specified IPv4 network.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"keepers": {
				Description: "An arbitrary map of values. If this argument is changed, then the resource will be recreated.",
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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
	client.Login(meta.(*apiClient).Username, meta.(*apiClient).Password)

	result := -1

	networkIDList := d.Get("network_id_list").(*schema.Set).List()
	networkNameList := d.Get("network_name_list").(*schema.Set).List()

	if len(networkIDList) == 0 && len(networkNameList) == 0 {
		err := gobam.LogoutClientWithError(client, "one of network_id_list or network_name_list must be set.")
		mutex.Unlock()
		return diag.FromErr(err)
	}

	if len(networkIDList) > 0 && len(networkNameList) > 0 {
		err := gobam.LogoutClientWithError(client, "only one of network_id_list or network_name_list can be set.")
		mutex.Unlock()
		return diag.FromErr(err)
	}

	if len(networkNameList) > 0 {
		if cid, ok := d.GetOk("container_id"); ok {
			containerID, err := strconv.ParseInt(cid.(string), 10, 64)
			if err = gobam.LogoutClientIfError(client, err, "Unable to convert container_id from string to int64"); err != nil {
				mutex.Unlock()
				return diag.FromErr(err)
			}

			for _, name := range networkNameList {
				options := "hint=" + name.(string)
				resp, err := client.GetIP4NetworksByHint(containerID, 0, 1, options)
				if err = gobam.LogoutClientIfError(client, err, "Failed to get IP4 Networks by hint"); err != nil {
					mutex.Unlock()
					return diag.FromErr(err)
				}

				if len(resp.Item) > 1 || len(resp.Item) == 0 {
					var diags diag.Diagnostics
					err := gobam.LogoutClientWithError(client, "Network lookup error")
					mutex.Unlock()
					diags = append(diags, diag.FromErr(err)...)
					diags = append(diags, diag.Errorf("network with name %s not found", name)...)
					return diags
				}

				networkIDList = append(networkIDList, resp.Item[0].Id)
			}
		} else {
			err := gobam.LogoutClientWithError(client, "container_id must be set when using network_name_list")
			mutex.Unlock()
			return diag.FromErr(err)
		}
	}

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
