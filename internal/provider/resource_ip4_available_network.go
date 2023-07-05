// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"hash/crc64"
	"math/rand"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &IP4AvailableNetworkResource{}
var _ resource.ResourceWithImportState = &IP4AvailableNetworkResource{}

func NewIP4AvailableNetworkResource() resource.Resource {
	return &IP4AvailableNetworkResource{}
}

// IP4AvailableNetworkResource defines the resource implementation.
type IP4AvailableNetworkResource struct {
	client *loginClient
}

// IP4AvailableNetworkResourceModel describes the resource data model.
type IP4AvailableNetworkResourceModel struct {
	ID            types.String `tfsdk:"id"`
	NetworkIDList types.List   `tfsdk:"network_id_list"`
	Keepers       types.Map    `tfsdk:"keepers"`
	Random        types.Bool   `tfsdk:"random"`
	Seed          types.String `tfsdk:"seed"`
	NetworkID     types.Int64  `tfsdk:"network_id"`
}

func (r *IP4AvailableNetworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip4_available_network"
}

func (r *IP4AvailableNetworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Resource to select an IPv4 network from a list of networks based on availability of IP addresses.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Example identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_id_list": schema.ListAttribute{
				MarkdownDescription: "A list of Network IDs to search for a free IP address. By default, the address with the most free addresses will be returned. See the `random` argument for another selection method. The resource will be recreated if the network_id_list is changed. You may want to use a `lifecycle` customization to ignore changes to the list after resource creation so that a new network is not selected if the list is changed.",
				Required:            true,
				ElementType:         types.Int64Type,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"keepers": schema.MapAttribute{
				MarkdownDescription: "An arbitrary map of values. If this argument is changed, then the resource will be recreated.",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"random": schema.BoolAttribute{
				MarkdownDescription: "By default, the network with the most free IP addresses is returned. By setting this to `true` a random network from the list will be returned instead. The network will be validated to have at least 1 free IP address.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"seed": schema.StringAttribute{
				MarkdownDescription: "A seed for the `random` argument's generator. Can be used to try to get more predictable results from the random selection. The results will not be fixed however.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_id": schema.Int64Attribute{
				MarkdownDescription: "The network ID of the network selected by the resource.",
				Computed:            true,
			},
		},
	}
}

func (r *IP4AvailableNetworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*loginClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *IP4AvailableNetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *IP4AvailableNetworkResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client := *clientLogin(r.client, mutex, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result := int64(-1)

	networkIDList := []int64{}
	diag := data.NetworkIDList.ElementsAs(ctx, networkIDList, false)
	if diag.HasError() {
		resp.Diagnostics.AddError(
			"Parsing network ids failed",
			"",
		)
		clientLogout(&client, mutex, resp.Diagnostics)

		return
	}

	seed := data.Seed.ValueString()
	random := data.Random.ValueBool()

	if len(networkIDList) == 0 {
		resp.Diagnostics.AddError(
			"network_id_list cannot be empty",
			"",
		)
		clientLogout(&client, mutex, resp.Diagnostics)

		return
	}

	if random {
		rand := NewRand(seed)

		// Keep producing permutations until we fill our result
	Batches:
		for {
			perm := rand.Perm(len(networkIDList))

			for _, i := range perm {
				id := networkIDList[i]

				entity, err := client.GetEntityById(id)
				if err != nil {
					resp.Diagnostics.AddError(
						"Failed to get IP4 Network by Id",
						err.Error(),
					)
					clientLogout(&client, mutex, resp.Diagnostics)

					return
				}

				networkProperties, diag := parseIP4NetworkProperties(*entity.Properties)
				if diag.HasError() {
					clientLogout(&client, mutex, resp.Diagnostics)
					resp.Diagnostics.Append(diag...)
					return
				}

				_, addressesFree, err := getIP4NetworkAddressUsage(*entity.Id, networkProperties.cidr.ValueString(), client)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error calculating network usage",
						err.Error(),
					)
					clientLogout(&client, mutex, resp.Diagnostics)

					return
				}

				if addressesFree > 0 {
					result = networkIDList[i]
					break Batches
				}
			}
		}

	} else {

		freeAddressMap := make(map[int64]int64)
		for i := range networkIDList {
			id := networkIDList[i]

			entity, err := client.GetEntityById(id)
			if err != nil {
				resp.Diagnostics.AddError(
					"Failed to get IP4 Network by Id",
					err.Error(),
				)
				clientLogout(&client, mutex, resp.Diagnostics)
				return
			}

			networkProperties, diag := parseIP4NetworkProperties(*entity.Properties)
			if diag.HasError() {
				clientLogout(&client, mutex, resp.Diagnostics)
				resp.Diagnostics.Append(diag...)
				return
			}

			_, addressesFree, err := getIP4NetworkAddressUsage(*entity.Id, networkProperties.cidr.ValueString(), client)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error calculating network usage",
					err.Error(),
				)
				clientLogout(&client, mutex, resp.Diagnostics)
				return
			}

			if addressesFree > 0 {
				freeAddressMap[id] = addressesFree
			}

		}

		freeCount := int64(0)
		for k, v := range freeAddressMap {
			if v > freeCount {
				freeCount = v
				result = int64(k)
			}
		}
	}

	if result == -1 {
		resp.Diagnostics.AddError(
			"No networks had a free address",
			"",
		)
		clientLogout(&client, mutex, resp.Diagnostics)
		return
	}

	data.ID = types.StringValue("-")
	data.NetworkID = types.Int64Value(result)

	clientLogout(&client, mutex, resp.Diagnostics)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IP4AvailableNetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *IP4AvailableNetworkResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IP4AvailableNetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *IP4AvailableNetworkResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IP4AvailableNetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *IP4AvailableNetworkResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// d.SetId("")
	// return nil
}

func (r *IP4AvailableNetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
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
