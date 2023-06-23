// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/umich-vci/gobam"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &IP4AddressDataSource{}

func NewIP4AddressDataSource() datasource.DataSource {
	return &IP4AddressDataSource{}
}

// IP4AddressDataSource defines the data source implementation.
type IP4AddressDataSource struct {
	client *loginClient
}

// IP4AddressDataSourceModel describes the data source data model.
type IP4AddressDataSourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	Address          types.String `tfsdk:"address"`
	ContainerID      types.Int64  `tfsdk:"container_id"`
	CustomProperties types.Map    `tfsdk:"custom_properties"`
	MACAddress       types.String `tfsdk:"mac_address"`
	Name             types.String `tfsdk:"name"`
	Properties       types.String `tfsdk:"properties"`
	State            types.String `tfsdk:"state"`
	Type             types.String `tfsdk:"type"`
}

func (d *IP4AddressDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip4_address"
}

func (d *IP4AddressDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Data source to access the attributes of an IPv4 address.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "IP4 Address identifier",
				Computed:            true,
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "The IPv4 address to get data for.",
				Required:            true,
			},
			"container_id": schema.Int64Attribute{
				MarkdownDescription: "The object ID of the container that has the specified `address`.  This can be a Configuration, IPv4 Block, IPv4 Network, or DHCP range.",
				Required:            true,
			},
			"custom_properties": schema.MapAttribute{
				MarkdownDescription: "A map of all custom properties associated with the IPv4 address.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"mac_address": schema.StringAttribute{
				MarkdownDescription: "The MAC address associated with the IPv4 address.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name assigned to the IPv4 address.  This is not related to DNS.",
				Computed:            true,
			},
			"properties": schema.StringAttribute{
				MarkdownDescription: "The properties of the IPv4 address as returned by the API (pipe delimited).",
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The state of the IPv4 address.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the resource.",
				Computed:            true,
			},
		},
	}
}

func (d *IP4AddressDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*loginClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *IP4AddressDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IP4AddressDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := d.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }

	mutex.Lock()
	client := d.client.Client
	client.Login(d.client.Username, d.client.Password)

	containerID := data.ContainerID.ValueInt64()
	address := data.Address.ValueString()

	ip4Address, err := client.GetIP4Address(containerID, address)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get IP4 Address"); err != nil {
		mutex.Unlock()
		resp.Diagnostics.AddError(
			"Failed to get IP4 Address",
			fmt.Sprintf("Failed to get IP4 Address: %s", err.Error()),
		)
		return
	}

	data.ID = types.Int64PointerValue(ip4Address.Id)
	data.Name = types.StringPointerValue(ip4Address.Name)
	data.Properties = types.StringPointerValue(ip4Address.Properties)
	data.Type = types.StringPointerValue(ip4Address.Type)

	addressProperties, err := parseIP4AddressProperties(*ip4Address.Properties)
	if err != nil {
		gobam.LogoutClientWithError(d.client.Client, "Error parsing host record properties")
		mutex.Unlock()

		resp.Diagnostics.AddError(
			"Error parsing the host record properties",
			err.Error(),
		)
	}
	data.Address = addressProperties.address
	data.State = addressProperties.state
	data.MACAddress = addressProperties.macAddress
	data.CustomProperties = addressProperties.customProperties

	// logout client
	if err := d.client.Client.Logout(); err != nil {
		mutex.Unlock()
		resp.Diagnostics.AddError(
			"Failed logout client",
			fmt.Sprintf("Unexpected error logging out client: %s", err.Error()),
		)
		return
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type ip4AddressProperties struct {
	address          types.String
	state            types.String
	macAddress       types.String
	customProperties types.Map
}

func parseIP4AddressProperties(properties string) (ip4AddressProperties, error) {
	var ip4Properties ip4AddressProperties
	cpMap := make(map[string]attr.Value)

	props := strings.Split(properties, "|")
	for x := range props {
		if len(props[x]) > 0 {
			prop := strings.Split(props[x], "=")[0]
			val := strings.Split(props[x], "=")[1]

			switch prop {
			case "address":
				ip4Properties.address = types.StringValue(val)
			case "state":
				ip4Properties.state = types.StringValue(val)
			case "macAddress":
				ip4Properties.macAddress = types.StringValue(val)
			default:
				cpMap[prop] = types.StringValue(val)
			}
		}
	}

	customProperties, diag := types.MapValue(types.StringType, cpMap)
	if diag.HasError() {
		return ip4Properties, fmt.Errorf("error creating custom properties map")
	}
	ip4Properties.customProperties = customProperties
	return ip4Properties, nil
}
