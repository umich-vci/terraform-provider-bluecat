package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/umich-vci/gobam"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &IP4NetworkResource{}
var _ resource.ResourceWithImportState = &IP4NetworkResource{}

func NewIP4NetworkResource() resource.Resource {
	return &IP4NetworkResource{}
}

// IP4NetworkResource defines the resource implementation.
type IP4NetworkResource struct {
	client *loginClient
}

// IP4NetworkResourceModel describes the resource data model.
type IP4NetworkResourceModel struct {
	ID                        types.Int64  `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	ParentID                  types.Int64  `tfsdk:"parent_id"`
	Size                      types.Int64  `tfsdk:"size"`
	IsLargerAllowed           types.Bool   `tfsdk:"is_larger_allowed"`
	TraversalMethod           types.String `tfsdk:"traversal_method"`
	AddressesInUse            types.Int64  `tfsdk:"addresses_in_use"`
	AddressesFree             types.Int64  `tfsdk:"addresses_free"`
	AllowDuplicateHost        types.String `tfsdk:"allow_duplicate_host"`
	CIDR                      types.String `tfsdk:"cidr"`
	CustomProperties          types.Map    `tfsdk:"custom_properties"`
	DefaultView               types.Int64  `tfsdk:"default_view"`
	Gateway                   types.String `tfsdk:"gateway"`
	InheritAllowDuplicateHost types.Bool   `tfsdk:"inherit_allow_duplicate_host"`
	InheritDefaultDomains     types.Bool   `tfsdk:"inherit_default_domains"`
	InheritDefaultView        types.Bool   `tfsdk:"inherit_default_view"`
	InheritDNSRestrictions    types.Bool   `tfsdk:"inherit_dns_restrictions"`
	InheritPingBeforeAssign   types.Bool   `tfsdk:"inherit_ping_before_assign"`
	PingBeforeAssign          types.String `tfsdk:"ping_before_assign"`
	Properties                types.String `tfsdk:"properties"`
	Type                      types.String `tfsdk:"type"`
}

func (r *IP4NetworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip4_network"
}

func (r *IP4NetworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Resource to create an IPv4 network.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "IPv4 Network identifier.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The display name of the IPv4 network.",
				Required:            true,
			},
			"parent_id": schema.Int64Attribute{
				MarkdownDescription: "The object ID of the parent object that will contain the new IPv4 network. If this argument is changed, then the resource will be recreated.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "The size of the IPv4 network expressed as a power of 2. For example, 256 would create a /24. If this argument is changed, then the resource will be recreated.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"is_larger_allowed": schema.BoolAttribute{
				MarkdownDescription: "(Optional) Is it ok to return a network that is larger than the size specified?",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"traversal_method": schema.StringAttribute{
				MarkdownDescription: "The traversal method used to find the range to allocate the network. Must be one of \"NO_TRAVERSAL\", \"DEPTH_FIRST\", or \"BREADTH_FIRST\".",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("NO_TRAVERSAL"),
				Validators: []validator.String{
					stringvalidator.OneOf("NO_TRAVERSAL", "DEPTH_FIRST", "BREADTH_FIRST"),
				},
			},
			"addresses_in_use": schema.Int64Attribute{
				MarkdownDescription: "The number of addresses allocated/in use on the network.",
				Computed:            true,
			},
			"addresses_free": schema.Int64Attribute{
				MarkdownDescription: "The number of addresses unallocated/free on the network.",
				Computed:            true,
			},
			"allow_duplicate_host": schema.StringAttribute{
				MarkdownDescription: "Duplicate host names check.",
				Computed:            true,
			},
			"cidr": schema.StringAttribute{
				MarkdownDescription: "The CIDR address of the IPv4 network.",
				Computed:            true,
			},
			"custom_properties": schema.MapAttribute{
				MarkdownDescription: "A map of all custom properties associated with the IPv4 network.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"default_view": schema.Int64Attribute{
				MarkdownDescription: "The object id of the default DNS View for the network.",
				Computed:            true,
			},
			"gateway": schema.StringAttribute{
				MarkdownDescription: "The gateway of the IPv4 network.",
				Computed:            true,
			},
			"inherit_allow_duplicate_host": schema.BoolAttribute{
				MarkdownDescription: "Duplicate host names check is inherited.",
				Computed:            true,
			},
			"inherit_default_domains": schema.BoolAttribute{
				MarkdownDescription: "Default domains are inherited.",
				Computed:            true,
			},
			"inherit_default_view": schema.BoolAttribute{
				MarkdownDescription: "The default DNS Viewis inherited.",
				Computed:            true,
			},
			"inherit_dns_restrictions": schema.BoolAttribute{
				MarkdownDescription: "DNS restrictions are inherited.",
				Computed:            true,
			},
			"inherit_ping_before_assign": schema.BoolAttribute{
				MarkdownDescription: "The network pings an address before assignment is inherited.",
				Computed:            true,
			},
			"ping_before_assign": schema.StringAttribute{
				MarkdownDescription: "The network pings an address before assignment.",
				Computed:            true,
			},
			"properties": schema.StringAttribute{
				MarkdownDescription: "The properties of the resource as returned by the API (pipe delimited).",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the resource.",
				Computed:            true,
			},
		},
	}
}

func (r *IP4NetworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IP4NetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *IP4NetworkResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, diag := clientLogin(ctx, r.client, mutex)
	if diag.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

		return
	}

	parentID := data.ParentID.ValueInt64()
	size := data.Size.ValueInt64()
	isLargerAllowed := data.IsLargerAllowed.ValueBool()
	traversalMethod := data.TraversalMethod.ValueString()
	autoCreate := true     //we always want to create since this is a resource after all
	reuseExisting := false //we never want to use an existing network created outside terraform
	Type := "IP4Network"   //Since this is the ip4_network resource we are setting the type
	properties := "reuseExisting=" + strconv.FormatBool(reuseExisting) + "|"
	properties = properties + "isLargerAllowed=" + strconv.FormatBool(isLargerAllowed) + "|"
	properties = properties + "autoCreate=" + strconv.FormatBool(autoCreate) + "|"
	properties = properties + "traversalMethod=" + traversalMethod + "|"

	network, err := client.GetNextAvailableIPRange(parentID, size, Type, properties)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"Failed to create IP4 Network",
			err.Error(),
		)
		return
	}

	data.ID = types.Int64PointerValue(network.Id)
	name := data.Name.ValueString()
	id := *network.Id
	properties = ""
	otype := "IP4Network"

	setName := gobam.APIEntity{
		Id:         &id,
		Name:       &name,
		Properties: &properties,
		Type:       &otype,
	}

	err = client.Update(&setName)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"Failed to update created IP4 Network with name",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IP4NetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *IP4NetworkResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, diag := clientLogin(ctx, r.client, mutex)
	if diag.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

		return
	}

	id := data.ID.ValueInt64()

	entity, err := client.GetEntityById(id)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"Failed to get IP4 Network by Id",
			err.Error(),
		)
		return
	}

	if *entity.Id == 0 {
		data.ID = types.Int64Null()
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"Failed to create IP4 Network",
			"ID returned was 0",
		)

		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

		return
	}

	data.Name = types.StringPointerValue(entity.Name)
	data.Properties = types.StringPointerValue(entity.Properties)
	data.Type = types.StringPointerValue(entity.Type)

	networkProperties, diag := parseIP4NetworkProperties(*entity.Properties)
	if diag.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.Append(diag...)
		return
	}

	data.CIDR = networkProperties.cidr
	data.AllowDuplicateHost = networkProperties.allowDuplicateHost
	data.InheritAllowDuplicateHost = networkProperties.inheritAllowDuplicateHost
	data.InheritPingBeforeAssign = networkProperties.inheritPingBeforeAssign
	data.PingBeforeAssign = networkProperties.pingBeforeAssign
	data.Gateway = networkProperties.gateway
	data.InheritDefaultDomains = networkProperties.inheritDefaultDomains
	data.DefaultView = networkProperties.defaultView
	data.InheritDefaultView = networkProperties.inheritDefaultView
	data.InheritDNSRestrictions = networkProperties.inheritDNSRestrictions
	data.CustomProperties = networkProperties.customProperties

	addressesInUse, addressesFree, err := getIP4NetworkAddressUsage(*entity.Id, networkProperties.cidr.ValueString(), client)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"Error calculating network usage",
			err.Error(),
		)
		return
	}

	data.AddressesInUse = types.Int64Value(addressesInUse)
	data.AddressesFree = types.Int64Value(addressesFree)

	resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IP4NetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *IP4NetworkResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, diag := clientLogin(ctx, r.client, mutex)
	if diag.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

		return
	}

	id := data.ID.ValueInt64()
	name := data.Name.ValueString()
	properties := ""
	otype := "IP4Network"

	update := gobam.APIEntity{
		Id:         &id,
		Name:       &name,
		Properties: &properties,
		Type:       &otype,
	}

	err := client.Update(&update)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"IP4 Network Update failed",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IP4NetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *IP4NetworkResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, diag := clientLogin(ctx, r.client, mutex)
	if diag.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

		return
	}

	id := data.ID.ValueInt64()

	entity, err := client.GetEntityById(id)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"Failed to get IP4 Network by Id",
			err.Error(),
		)
		return
	}

	if *entity.Id == 0 {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		return
	}

	err = client.Delete(id)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"Delete failed",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
}

func (r *IP4NetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
