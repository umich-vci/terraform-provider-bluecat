package provider

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/umich-vci/gobam"
	"golang.org/x/exp/maps"
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

// HostRecordResourceModel describes the resource data model.
type HostRecordResourceModel struct {
	// These are exposed for a generic entity object in bluecat
	ID         types.Int64  `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	Properties types.String `tfsdk:"properties"`

	// These are exposed via the entity properties field for objects of type IP4Address
	TTL           types.Int64  `tfsdk:"ttl"`
	AbsoluteName  types.String `tfsdk:"absolute_name"`
	Addresses     types.Set    `tfsdk:"addresses"`
	ReverseRecord types.Bool   `tfsdk:"reverse_record"`

	// this is returned by the API but do not appear in the documentation
	AddressIDs types.Set `tfsdk:"address_ids"`

	// these are user defined fields that are not built-in
	UserDefinedFields types.Map `tfsdk:"user_defined_fields"`

	// These fields are only used for creation
	DNSZone types.String `tfsdk:"dns_zone"`
	ViewID  types.Int64  `tfsdk:"view_id"`
}

func (r *HostRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host_record"
}

func (r *HostRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Resource create a host record.",

		Attributes: map[string]schema.Attribute{
			// These are exposed for Entity objects via the API
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Host Record identifier",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the host record to be created. Combined with `dns_zone` to make the fqdn.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the resource.",
				Computed:            true,
			},
			"properties": schema.StringAttribute{
				MarkdownDescription: "The properties of the host record as returned by the API (pipe delimited).",
				Computed:            true,
			},
			// These fields are only used for creation and are not exposed via the API entity
			"dns_zone": schema.StringAttribute{
				MarkdownDescription: "The DNS zone to create the host record in. Combined with `name` to make the fqdn.  If changed, forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"view_id": schema.Int64Attribute{
				MarkdownDescription: "The object ID of the View that host record should be created in. If changed, forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			// These are exposed via the API properties field for objects of type Host Record
			"addresses": schema.SetAttribute{
				MarkdownDescription: "The address(es) to be associated with the host record.",
				Required:            true,
				ElementType:         types.StringType,
			},
			"address_ids": schema.SetAttribute{
				MarkdownDescription: "A set of all address ids associated with the host record.",
				Computed:            true,
				ElementType:         types.Int64Type,
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
			"user_defined_fields": schema.MapAttribute{
				MarkdownDescription: "A map of all user-definied fields associated with the Host Record.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             mapdefault.StaticValue(basetypes.NewMapValueMust(types.StringType, nil)),
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

	client, diag := clientLogin(ctx, r.client, mutex)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	viewID := data.ViewID.ValueInt64()
	absoluteName := data.Name.ValueString() + "." + data.DNSZone.ValueString()
	ttl := data.TTL.ValueInt64()

	var addresses []string
	diag = data.Addresses.ElementsAs(ctx, &addresses, false)
	if diag.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.Append(diag...)
		return
	}

	properties := ""
	properties = properties + fmt.Sprintf("reverseRecord=%s|", strconv.FormatBool(data.ReverseRecord.ValueBool()))

	var udfs map[string]string
	resp.Diagnostics.Append(data.UserDefinedFields.ElementsAs(ctx, &udfs, false)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.Append(diag...)
		return
	}
	for k, v := range udfs {
		properties = properties + fmt.Sprintf("%s=%s|", k, v)
	}

	host, err := client.AddHostRecord(viewID, absoluteName, strings.Join(addresses, ","), ttl, properties)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("AddHostRecord failed", err.Error())
		return
	}

	data.ID = types.Int64Value(host)

	entity, err := client.GetEntityById(data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"Failed to get IP4 Address by Id after creation",
			err.Error(),
		)
		return
	}

	data.Name = types.StringPointerValue(entity.Name)
	data.Properties = types.StringPointerValue(entity.Properties)
	data.Type = types.StringPointerValue(entity.Type)

	hrProperties, diag := flattenHostRecordProperties(entity)
	if diag.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.Append(diag...)
		return
	}

	data.AbsoluteName = hrProperties.AbsoluteName
	data.Addresses = hrProperties.Addresses
	data.AddressIDs = hrProperties.AddressIDs
	data.TTL = hrProperties.TTL
	data.ReverseRecord = hrProperties.ReverseRecord
	data.UserDefinedFields = hrProperties.UserDefinedFields

	resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

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

	client, diag := clientLogin(ctx, r.client, mutex)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	id := data.ID.ValueInt64()

	entity, err := client.GetEntityById(id)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Failed to get host record by Id", err.Error())
		return
	}

	if *entity.Id == 0 {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringPointerValue(entity.Name)
	data.Properties = types.StringPointerValue(entity.Properties)
	data.Type = types.StringPointerValue(entity.Type)

	hostRecordProperties, diag := flattenHostRecordProperties(entity)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		return
	}

	data.AbsoluteName = hostRecordProperties.AbsoluteName
	data.Addresses = hostRecordProperties.Addresses
	data.AddressIDs = hostRecordProperties.AddressIDs
	data.ReverseRecord = hostRecordProperties.ReverseRecord
	data.TTL = hostRecordProperties.TTL
	data.UserDefinedFields = hostRecordProperties.UserDefinedFields

	resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state *HostRecordResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, diag := clientLogin(ctx, r.client, mutex)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.Append(diag...)
		return
	}

	properties := ""

	// addresses must always be set
	var addresses []string
	resp.Diagnostics.Append(data.Addresses.ElementsAs(ctx, &addresses, false)...)
	properties = properties + fmt.Sprintf("addresses=%s|", strings.Join(addresses, ","))

	if !data.ReverseRecord.Equal(state.ReverseRecord) {
		properties = properties + fmt.Sprintf("reverseRecord=%s|", strconv.FormatBool(data.ReverseRecord.ValueBool()))
	}

	if !data.TTL.Equal(state.TTL) {
		properties = properties + fmt.Sprintf("ttl=%d|", data.TTL.ValueInt64())
	}

	if !data.UserDefinedFields.Equal(state.UserDefinedFields) {
		var udfs, oldudfs map[string]string
		resp.Diagnostics.Append(data.UserDefinedFields.ElementsAs(ctx, &udfs, false)...)
		resp.Diagnostics.Append(state.UserDefinedFields.ElementsAs(ctx, &oldudfs, false)...)

		for k, v := range udfs {
			properties = properties + fmt.Sprintf("%s=%s|", k, v)
		}

		// set keys that no longer exist to empty string
		oldkeys := maps.Keys(oldudfs)
		keys := maps.Keys(udfs)
		for _, x := range oldkeys {
			if !slices.Contains(keys, x) {
				properties = properties + fmt.Sprintf("%s=|", x)
			}
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Attempting to update HostRecord with properties: %s", properties))

	update := gobam.APIEntity{
		Id:         data.ID.ValueInt64Pointer(),
		Name:       data.Name.ValueStringPointer(),
		Properties: &properties,
		Type:       state.Type.ValueStringPointer(),
	}

	err := client.Update(&update)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Host Record Update failed", err.Error())
		return
	}

	entity, err := client.GetEntityById(data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"Failed to get host record by Id after update",
			err.Error(),
		)
		return
	}

	data.Name = types.StringPointerValue(entity.Name)
	data.Properties = types.StringPointerValue(entity.Properties)
	data.Type = types.StringPointerValue(entity.Type)

	hrProperties, diag := flattenHostRecordProperties(entity)
	if diag.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.Append(diag...)
		return
	}

	data.AbsoluteName = hrProperties.AbsoluteName
	data.Addresses = hrProperties.Addresses
	data.AddressIDs = hrProperties.AddressIDs
	data.TTL = hrProperties.TTL
	data.ReverseRecord = hrProperties.ReverseRecord
	data.UserDefinedFields = hrProperties.UserDefinedFields

	resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

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

	client, diag := clientLogin(ctx, r.client, mutex)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	id := data.ID.ValueInt64()

	entity, err := client.GetEntityById(id)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Failed to get host record by id", err.Error())
		return
	}

	if *entity.Id == 0 {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

		return
	}

	err = client.Delete(id)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Host Record Delete failed", err.Error())
		return
	}

	resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
}

func (r *HostRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
