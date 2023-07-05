// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/umich-vci/gobam"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &HostRecordResource{}
var _ resource.ResourceWithImportState = &HostRecordResource{}

func NewHostRecordResource() resource.Resource {
	return &HostRecordResource{}
}

// HostRecordResource defines the resource implementation.
type HostRecordResource struct {
	client *loginClient
}

// ExampleResourceModel describes the resource data model.
type HostRecordResourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	Addresses        types.Set    `tfsdk:"addresses"`
	DNSZone          types.String `tfsdk:"dns_zone"`
	Name             types.String `tfsdk:"name"`
	ViewID           types.Int64  `tfsdk:"view_id"`
	Comments         types.String `tfsdk:"comments"`
	CustomProperties types.Map    `tfsdk:"custom_properties"`
	ReverseRecord    types.Bool   `tfsdk:"reverse_record"`
	TTL              types.Int64  `tfsdk:"ttl"`
	AbsoluteName     types.String `tfsdk:"absolute_name"`
	Properties       types.String `tfsdk:"properties"`
	Type             types.String `tfsdk:"type"`
}

func (r *HostRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host_record"
}

func (r *HostRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Resource create a host record.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Host Record identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"addresses": schema.SetAttribute{
				MarkdownDescription: "The address(es) to be associated with the host record.",
				Required:            true,
				ElementType:         types.StringType,
			},
			"dns_zone": schema.StringAttribute{
				MarkdownDescription: "The DNS zone to create the host record in. Combined with `name` to make the fqdn.  If changed, forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the host record to be created. Combined with `dns_zone` to make the fqdn.",
				Required:            true,
			},
			"view_id": schema.Int64Attribute{
				MarkdownDescription: "The object ID of the View that host record should be created in. If changed, forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"comments": schema.StringAttribute{
				MarkdownDescription: "Comments to be associated with the host record.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"custom_properties": schema.MapAttribute{
				MarkdownDescription: "A map of all custom properties associated with the host record.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"reverse_record": schema.BoolAttribute{
				MarkdownDescription: "If a reverse record should be created for addresses.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"ttl": schema.Int64Attribute{
				MarkdownDescription: "The TTL for the host record.  When set to -1, ignores the TTL.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(-1),
			},
			"absolute_name": schema.StringAttribute{
				MarkdownDescription: "The absolute name (fqdn) of the host record.",
				Computed:            true,
			},
			"properties": schema.StringAttribute{
				MarkdownDescription: "The properties of the host record as returned by the API (pipe delimited).",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the resource.",
				Computed:            true,
			},
		},
	}
}

func (r *HostRecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *HostRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *HostRecordResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	mutex.Lock()
	client := r.client.Client
	client.Login(r.client.Username, r.client.Password)

	viewID := data.ViewID.ValueInt64()
	absoluteName := data.Name.ValueString() + "." + data.DNSZone.ValueString()
	ttl := data.TTL.ValueInt64()

	addresses := []string{}
	diag := data.Addresses.ElementsAs(ctx, addresses, false)
	if diag.HasError() {
		resp.Diagnostics.AddError(
			"Parsing addresses failed",
			"",
		)
		gobam.LogoutClientWithError(client, "arsing addresses failed")
		mutex.Unlock()
		return
	}

	reverseRecord := strconv.FormatBool(data.ReverseRecord.ValueBool())
	comments := data.Comments.ValueString()
	properties := "reverseRecord=" + reverseRecord + "|comments=" + comments + "|"

	customProperties := data.CustomProperties.Elements()
	for k, v := range customProperties {
		properties = properties + k + "=" + v.String() + "|"
	}

	host, err := client.AddHostRecord(viewID, absoluteName, strings.Join(addresses, ","), ttl, properties)
	if err = gobam.LogoutClientIfError(client, err, "AddHostRecord failed"); err != nil {
		mutex.Unlock()
		resp.Diagnostics.AddError(
			"AddHostRecord failed",
			err.Error(),
		)
		return
	}

	data.ID = types.Int64Value(host)

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		resp.Diagnostics.AddError(
			"Failed to logout client",
			err.Error(),
		)
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *HostRecordResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	mutex.Lock()
	client := r.client.Client
	client.Login(r.client.Username, r.client.Password)

	id := data.ID.ValueInt64()

	entity, err := client.GetEntityById(id)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get host record by Id"); err != nil {
		mutex.Unlock()
		resp.Diagnostics.AddError(
			"Failed to get entity by name",
			fmt.Sprintf("Failed to get entity by name: %s", err.Error()),
		)
		return
	}

	if *entity.Id == 0 {
		data.ID = types.Int64Null()
		if err := client.Logout(); err != nil {
			mutex.Unlock()
			resp.Diagnostics.AddError(
				"Error logging out after API call",
				err.Error(),
			)
			return
		}

		mutex.Unlock()
		return
	}

	data.Name = types.StringPointerValue(entity.Name)
	data.Properties = types.StringPointerValue(entity.Properties)
	data.Type = types.StringPointerValue(entity.Type)

	hostRecordProperties := parseHostRecordProperties(*entity.Properties, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		clientLogout(&client, mutex, resp.Diagnostics)
		return
	}

	data.AbsoluteName = hostRecordProperties.absoluteName
	data.Addresses = hostRecordProperties.addresses
	data.CustomProperties = hostRecordProperties.customProperties
	data.ReverseRecord = hostRecordProperties.reverseRecord
	data.TTL = hostRecordProperties.ttl

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		resp.Diagnostics.AddError(
			"Failed to logout client",
			err.Error(),
		)
		return
	}
	log.Printf("[INFO] BlueCat Logout was successful")
	mutex.Unlock()

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *HostRecordResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	mutex.Lock()
	client := r.client.Client
	client.Login(r.client.Username, r.client.Password)

	id := data.ID.ValueInt64()
	name := data.Name.ValueString()
	otype := data.Type.ValueString()
	ttl := strconv.FormatInt(data.TTL.ValueInt64(), 10)

	addresses := []string{}
	diag := data.Addresses.ElementsAs(ctx, addresses, false)
	if diag.HasError() {
		resp.Diagnostics.AddError(
			"Parsing addresses failed",
			"",
		)
		gobam.LogoutClientWithError(client, "arsing addresses failed")
		mutex.Unlock()
		return
	}

	reverseRecord := strconv.FormatBool(data.ReverseRecord.ValueBool())
	comments := data.Comments.ValueString()
	properties := "reverseRecord=" + reverseRecord + "|comments=" + comments + "|ttl=" + ttl + "|addresses=" + strings.Join(addresses, ",") + "|"

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
	if err = gobam.LogoutClientIfError(client, err, "Host Record Update failed"); err != nil {
		mutex.Unlock()
		resp.Diagnostics.AddError(
			"Host Record Update failed",
			err.Error(),
		)
		return
	}

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		resp.Diagnostics.AddError(
			"Failed to logout client",
			err.Error(),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *HostRecordResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	mutex.Lock()
	client := r.client.Client
	client.Login(r.client.Username, r.client.Password)

	id := data.ID.ValueInt64()

	entity, err := client.GetEntityById(id)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get host record by Id"); err != nil {
		mutex.Unlock()
		resp.Diagnostics.AddError(
			"Failed to get host record by id",
			err.Error(),
		)
		return
	}

	if *entity.Id == 0 {
		if err := client.Logout(); err != nil {
			mutex.Unlock()
			resp.Diagnostics.AddError(
				"Failed to logout client",
				err.Error(),
			)
			return
		}

		mutex.Unlock()
		return
	}

	err = client.Delete(id)
	if err = gobam.LogoutClientIfError(client, err, "Delete failed"); err != nil {
		mutex.Unlock()
		resp.Diagnostics.AddError(
			"Host Record Delete failed",
			err.Error(),
		)
		return
	}

	// logout client
	if err := client.Logout(); err != nil {
		mutex.Unlock()
		resp.Diagnostics.AddError(
			"Failed to logout client",
			err.Error(),
		)
		return
	}
}

func (r *HostRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
