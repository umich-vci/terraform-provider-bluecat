package provider

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/umich-vci/gobam"
	"golang.org/x/exp/maps"
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

// IP4AddressResourceModel describes the resource data model.
type IP4AddressResourceModel struct {
	// These are exposed for a generic entity object in bluecat
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	Properties types.String `tfsdk:"properties"`

	// These are exposed via the entity properties field for objects of type IP4Address
	Address               types.String `tfsdk:"address"`
	State                 types.String `tfsdk:"state"`
	MACAddress            types.String `tfsdk:"mac_address"`
	RouterPortInfo        types.String `tfsdk:"router_port_info"`
	SwitchPortInfo        types.String `tfsdk:"switch_port_info"`
	VLANInfo              types.String `tfsdk:"vlan_info"`
	LeaseTime             types.String `tfsdk:"lease_time"`
	ExpiryTime            types.String `tfsdk:"expiry_time"`
	ParameterRequestList  types.String `tfsdk:"parameter_request_list"`
	VendorClassIdentifier types.String `tfsdk:"vendor_class_identifier"`
	LocationCode          types.String `tfsdk:"location_code"`
	LocationInherited     types.Bool   `tfsdk:"location_inherited"`

	// these are user defined fields that are not built-in
	UserDefinedFields types.Map `tfsdk:"user_defined_fields"`

	// These fields are only used for creation
	Action          types.String `tfsdk:"action"`
	ConfigurationID types.Int64  `tfsdk:"configuration_id"`
	ParentID        types.Int64  `tfsdk:"parent_id"`
}

func (r *IP4AddressResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip4_address"
}

func (r *IP4AddressResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Resource to reserve an IPv4 address.",

		Attributes: map[string]schema.Attribute{
			// These are exposed for Entity objects via the API
			"id": schema.StringAttribute{
				MarkdownDescription: "IPv4 Address identifier.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The display name of the IPv4 address.",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"properties": schema.StringAttribute{
				MarkdownDescription: "The properties of the resource as returned by the API (pipe delimited).",
				Computed:            true,
			},
			// These fields are only used for creation and are not exposed via the API entity
			"action": schema.StringAttribute{
				MarkdownDescription: "The action to take on the next available IPv4 address.  Must be one of: \"MAKE_STATIC\", \"MAKE_RESERVED\", or \"MAKE_DHCP_RESERVED\". If changed, forces a new resource.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("MAKE_STATIC"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(ip4AddressActionPlanModifier, ip4AddressActionPlanModifierDescription, ip4AddressActionPlanModifierDescription),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(gobam.IPAssignmentActions...),
				},
			},
			"configuration_id": schema.Int64Attribute{
				MarkdownDescription: "The object ID of the Configuration that will hold the new address. If changed, forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplaceIf(ip4AddressConfigurationIDPlanModifier, ip4AddressConfigurationIDPlanModifierDescription, ip4AddressConfigurationIDPlanModifierDescription),
				},
			},
			"parent_id": schema.Int64Attribute{
				MarkdownDescription: "The object ID of the Configuration, Block, or Network to find the next available IPv4 address in. If changed, forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			// These are exposed via the API properties field for objects of type IP4Address
			"address": schema.StringAttribute{
				MarkdownDescription: "The IPv4 address that was allocated.",
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The state of the IPv4 address.",
				Computed:            true,
			},
			"mac_address": schema.StringAttribute{
				MarkdownDescription: "The MAC address to associate with the IPv4 address.",
				Optional:            true,
			},
			"router_port_info": schema.StringAttribute{
				MarkdownDescription: "Connected router port information of the IPv4 address.",
				Computed:            true,
			},
			"switch_port_info": schema.StringAttribute{
				MarkdownDescription: "Connected switch port information of the IPv4 address.",
				Computed:            true,
			},
			"vlan_info": schema.StringAttribute{
				MarkdownDescription: "VLAN information of the IPv4 address.",
				Computed:            true,
			},
			"lease_time": schema.StringAttribute{
				MarkdownDescription: "Time that IPv4 address was leased.",
				Computed:            true,
			},
			"expiry_time": schema.StringAttribute{
				MarkdownDescription: "Time that IPv4 address lease expires.",
				Computed:            true,
			},
			"parameter_request_list": schema.StringAttribute{
				MarkdownDescription: "Time that IPv4 address lease expires.",
				Computed:            true,
			},
			"vendor_class_identifier": schema.StringAttribute{
				MarkdownDescription: "Time that IPv4 address lease expires.",
				Computed:            true,
			},
			"location_code": schema.StringAttribute{
				MarkdownDescription: "The location code of the address.",
				Computed:            true,
				Optional:            true,
				Default:             nil,
				Validators:          []validator.String{
					// The code is case-sensitive and must be in uppercase letters. The country code and child location code should be alphanumeric strings.
				},
			},
			"location_inherited": schema.BoolAttribute{
				MarkdownDescription: "The location is inherited.",
				Computed:            true,
			},
			"user_defined_fields": schema.MapAttribute{
				MarkdownDescription: "A map of all user-definied fields associated with the IPv4 address.",
				Computed:            true,
				Optional:            true,
				Default:             mapdefault.StaticValue(basetypes.NewMapValueMust(types.StringType, nil)),
				ElementType:         types.StringType,
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

	if !data.LocationCode.IsUnknown() && !data.LocationCode.IsNull() {
		properties = properties + fmt.Sprintf("locationCode=%s|", data.LocationCode.ValueString())
	}

	var udfs map[string]string
	data.UserDefinedFields.ElementsAs(ctx, &udfs, false)
	for k, v := range udfs {
		properties = properties + k + "=" + v + "|"
	}

	ip, err := client.AssignNextAvailableIP4Address(configID, parentID, macAddress, hostInfo, action, properties)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("AssignNextAvailableIP4Address failed", err.Error())
		return
	}

	data.ID = types.StringValue(strconv.FormatInt(*ip.Id, 10))

	entity, err := client.GetEntityById(*ip.Id)
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

	addressProperties, diag := flattenIP4AddressProperties(entity)
	if diag.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.Append(diag...)
		return
	}

	data.Address = addressProperties.Address
	data.State = addressProperties.State
	data.MACAddress = addressProperties.MACAddress
	data.RouterPortInfo = addressProperties.RouterPortInfo
	data.SwitchPortInfo = addressProperties.SwitchPortInfo
	data.VLANInfo = addressProperties.VLANInfo
	data.LeaseTime = addressProperties.LeaseTime
	data.ExpiryTime = addressProperties.ExpiryTime
	data.ParameterRequestList = addressProperties.ParameterRequestList
	data.VendorClassIdentifier = addressProperties.VendorClassIdentifier
	data.LocationCode = addressProperties.LocationCode
	data.LocationInherited = addressProperties.LocationInherited
	data.UserDefinedFields = addressProperties.UserDefinedFields

	resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

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

	id, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Failed to parse ID", err.Error())
		return
	}

	entity, err := client.GetEntityById(id)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Failed to get IP4 Address by Id", err.Error())
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

	addressProperties, diag := flattenIP4AddressProperties(entity)
	if diag.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.Append(diag...)
		return
	}

	data.Address = addressProperties.Address
	data.State = addressProperties.State
	data.MACAddress = addressProperties.MACAddress
	data.RouterPortInfo = addressProperties.RouterPortInfo
	data.SwitchPortInfo = addressProperties.SwitchPortInfo
	data.VLANInfo = addressProperties.VLANInfo
	data.LeaseTime = addressProperties.LeaseTime
	data.ExpiryTime = addressProperties.ExpiryTime
	data.ParameterRequestList = addressProperties.ParameterRequestList
	data.VendorClassIdentifier = addressProperties.VendorClassIdentifier
	data.LocationCode = addressProperties.LocationCode
	data.LocationInherited = addressProperties.LocationInherited
	data.UserDefinedFields = addressProperties.UserDefinedFields

	// get the parent id of the address so we can set it in the state so import works
	parent, err := client.GetParent(id)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Failed to get parent entity of IP4 address", err.Error())
		return
	}
	data.ParentID = types.Int64Value(*parent.Id)

	resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IP4AddressResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state *IP4AddressResourceModel

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

	properties := ""

	if !data.MACAddress.Equal(state.MACAddress) {
		properties = properties + fmt.Sprintf("macAddress=%s|", data.MACAddress.ValueString())
	}

	if !data.LocationCode.Equal(state.LocationCode) {
		properties = properties + fmt.Sprintf("locationCode=%s|", data.LocationCode.ValueString())
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

	id, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Failed to parse ID", err.Error())
		return
	}

	update := gobam.APIEntity{
		Id:         &id,
		Name:       data.Name.ValueStringPointer(),
		Properties: &properties,
		Type:       state.Type.ValueStringPointer(),
	}

	err = client.Update(&update)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Failed to update IP4 Address", err.Error())
		return
	}

	entity, err := client.GetEntityById(id)
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

	addressProperties, diag := flattenIP4AddressProperties(entity)
	if diag.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.Append(diag...)
		return
	}

	data.Address = addressProperties.Address
	data.State = addressProperties.State
	data.MACAddress = addressProperties.MACAddress
	data.RouterPortInfo = addressProperties.RouterPortInfo
	data.SwitchPortInfo = addressProperties.SwitchPortInfo
	data.VLANInfo = addressProperties.VLANInfo
	data.LeaseTime = addressProperties.LeaseTime
	data.ExpiryTime = addressProperties.ExpiryTime
	data.ParameterRequestList = addressProperties.ParameterRequestList
	data.VendorClassIdentifier = addressProperties.VendorClassIdentifier
	data.LocationCode = addressProperties.LocationCode
	data.LocationInherited = addressProperties.LocationInherited
	data.UserDefinedFields = addressProperties.UserDefinedFields

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

	id, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Failed to parse ID", err.Error())
		return
	}

	err = client.Delete(id)
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

const ip4AddressActionPlanModifierDescription string = "action is required for creation and cannot be changed. Null values in the state are ignored to allow for import."

func ip4AddressActionPlanModifier(ctx context.Context, p planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	var state *IP4AddressResourceModel
	resp.Diagnostics.Append(p.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Action.IsNull() {
		// Since this is a required field with required values, it should only be null when doing an import
		resp.RequiresReplace = false
		return
	}

	resp.RequiresReplace = true
}

const ip4AddressConfigurationIDPlanModifierDescription string = "configuration_id is required for creation and cannot be changed. Null values in the state are ignored to allow for import."

func ip4AddressConfigurationIDPlanModifier(ctx context.Context, p planmodifier.Int64Request, resp *int64planmodifier.RequiresReplaceIfFuncResponse) {
	var state *IP4AddressResourceModel
	resp.Diagnostics.Append(p.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ConfigurationID.IsNull() {
		// Since this is a required field with required values, it should only be null when doing an import
		resp.RequiresReplace = false
		return
	}

	resp.RequiresReplace = true
}
