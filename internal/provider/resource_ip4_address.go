// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/umich-vci/gobam"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &IP4AddressResource{}
var _ resource.ResourceWithImportState = &IP4AddressResource{}

func NewIP4AddressResource() resource.Resource {
	return &IP4AddressResource{}
}

// IP4AddressResource defines the resource implementation.
type IP4AddressResource struct {
	client *loginClient
}

// ExampleResourceModel describes the resource data model.
type IP4AddressResourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	ConfigurationID  types.Int64  `tfsdk:"configuration_id"`
	Name             types.String `tfsdk:"name"`
	ParentID         types.Int64  `tfsdk:"parent_id"`
	Action           types.String `tfsdk:"action"`
	CustomProperties types.Map    `tfsdk:"custom_properties"`
	MACAddress       types.String `tfsdk:"mac_address"`
	Address          types.String `tfsdk:"address"`
	Properties       types.String `tfsdk:"properties"`
	State            types.String `tfsdk:"state"`
	Type             types.String `tfsdk:"type"`
}

func (r *IP4AddressResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip4_address"
}

func (r *IP4AddressResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Resource to reserve an IPv4 address.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "IP4 address identifier",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"configuration_id": schema.Int64Attribute{
				MarkdownDescription: "The object ID of the Configuration that will hold the new address. If changed, forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name assigned to the IPv4 address. This is not related to DNS.",
				Required:            true,
			},
			"parent_id": schema.Int64Attribute{
				MarkdownDescription: "The object ID of the Configuration, Block, or Network to find the next available IPv4 address in. If changed, forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"action": schema.StringAttribute{
				MarkdownDescription: "The action to take on the next available IPv4 address.  Must be one of: \"MAKE_STATIC\", \"MAKE_RESERVED\", or \"MAKE_DHCP_RESERVED\". If changed, forces a new resource.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("MAKE_STATIC"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(gobam.IPAssignmentActions...),
				},
			},
			"custom_properties": schema.MapAttribute{
				MarkdownDescription: "A map of all custom properties associated with the IPv4 address.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"mac_address": schema.StringAttribute{
				MarkdownDescription: "The MAC address to associate with the IPv4 address.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "The IPv4 address that was allocated.",
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

func (r *IP4AddressResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IP4AddressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *IP4AddressResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, diag := clientLogin(ctx, r.client, mutex)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	configID := data.ConfigurationID.ValueInt64()
	parentID := data.ParentID.ValueInt64()
	macAddress := data.MACAddress.ValueString()
	hostInfo := "" // host records should be created as a separate resource
	action := data.Action.ValueString()
	name := data.Name.ValueString()
	properties := "name=" + name + "|"
	customProperties := data.CustomProperties.Elements()
	for k, v := range customProperties {
		properties = properties + k + "=" + v.String() + "|"
	}

	ip, err := client.AssignNextAvailableIP4Address(configID, parentID, macAddress, hostInfo, action, properties)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("AssignNextAvailableIP4Address failed", err.Error())
		return
	}

	data.ID = types.Int64PointerValue(ip.Id)

	resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IP4AddressResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *IP4AddressResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, diag := clientLogin(ctx, r.client, mutex)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	id := data.ID.ValueInt64()
	entity, err := client.GetEntityById(id)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Failed to get IP4 Address by Id", err.Error())
		return
	}

	if *entity.Id == 0 {
		data.ID = types.Int64Null()
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	data.Name = types.StringPointerValue(entity.Name)
	data.Properties = types.StringPointerValue(entity.Properties)
	data.Type = types.StringPointerValue(entity.Type)

	addressProperties, err := parseIP4AddressProperties(*entity.Properties)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Failed to parse IP4 Address properties", err.Error())
		return
	}

	data.Address = addressProperties.address
	data.State = addressProperties.state
	data.MACAddress = addressProperties.macAddress
	data.CustomProperties = addressProperties.customProperties

	resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IP4AddressResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *IP4AddressResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, diag := clientLogin(ctx, r.client, mutex)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	id := data.ID.ValueInt64()
	macAddress := data.MACAddress.ValueString()
	name := data.Name.ValueString()
	otype := data.Type.ValueString()
	properties := "name=" + name + "|"

	if macAddress != "" {
		properties = properties + "macAddress=" + macAddress + "|"
	}

	customProperties := data.CustomProperties.Elements()
	for k, v := range customProperties {
		properties = properties + k + "=" + v.String() + "|"
	}

	update := gobam.APIEntity{
		Id:         &id,
		Name:       &name,
		Properties: &properties,
		Type:       &otype,
	}

	err := client.Update(&update)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Failed to update IP4 Address", err.Error())
		return
	}

	resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IP4AddressResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *IP4AddressResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, diag := clientLogin(ctx, r.client, mutex)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	id := data.ID.ValueInt64()

	err := client.Delete(id)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Failed to delete IP4 Address", err.Error())
		return
	}

	resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
}

func (r *IP4AddressResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
