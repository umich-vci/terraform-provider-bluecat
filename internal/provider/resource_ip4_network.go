package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
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
	// These are exposed for a generic entity object in bluecat
	ID         types.Int64  `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	Properties types.String `tfsdk:"properties"`

	// These are exposed via the entity properties field for objects of type IP4Network
	CIDR                      types.String `tfsdk:"cidr"`
	Template                  types.Int64  `tfsdk:"template"`
	Gateway                   types.String `tfsdk:"gateway"`
	DefaultDomains            types.Set    `tfsdk:"default_domains"`
	DefaultView               types.Int64  `tfsdk:"default_view"`
	DNSRestrictions           types.Set    `tfsdk:"dns_restrictions"`
	AllowDuplicateHost        types.Bool   `tfsdk:"allow_duplicate_host"`
	PingBeforeAssign          types.Bool   `tfsdk:"ping_before_assign"`
	InheritAllowDuplicateHost types.Bool   `tfsdk:"inherit_allow_duplicate_host"`
	InheritPingBeforeAssign   types.Bool   `tfsdk:"inherit_ping_before_assign"`
	InheritDNSRestrictions    types.Bool   `tfsdk:"inherit_dns_restrictions"`
	InheritDefaultDomains     types.Bool   `tfsdk:"inherit_default_domains"`
	InheritDefaultView        types.Bool   `tfsdk:"inherit_default_view"`
	LocationCode              types.String `tfsdk:"location_code"`
	LocationInherited         types.Bool   `tfsdk:"location_inherited"`
	SharedNetwork             types.String `tfsdk:"shared_network"`

	// these are user defined fields that are not built-in
	UserDefinedFields types.Map `tfsdk:"user_defined_fields"`

	// These fields are only used for creation
	IsLargerAllowed types.Bool   `tfsdk:"is_larger_allowed"`
	ParentID        types.Int64  `tfsdk:"parent_id"`
	Size            types.Int64  `tfsdk:"size"`
	TraversalMethod types.String `tfsdk:"traversal_method"`
}

func (r *IP4NetworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip4_network"
}

func (r *IP4NetworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Resource to create an IPv4 network.",

		Attributes: map[string]schema.Attribute{
			// These are exposed for Entity objects via the API
			"id": schema.Int64Attribute{
				MarkdownDescription: "IPv4 Network identifier.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The display name of the IPv4 network.",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the resource.",
				Computed:            true,
			},
			"properties": schema.StringAttribute{
				MarkdownDescription: "The properties of the resource as returned by the API (pipe delimited).",
				Computed:            true,
			},
			// These fields are only used for creation and are not exposed via the API entity
			"is_larger_allowed": schema.BoolAttribute{
				MarkdownDescription: "(Optional) Is it ok to return a network that is larger than the size specified?",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
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
			"traversal_method": schema.StringAttribute{
				MarkdownDescription: "The traversal method used to find the range to allocate the network. Must be one of \"NO_TRAVERSAL\", \"DEPTH_FIRST\", or \"BREADTH_FIRST\".",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("NO_TRAVERSAL"),
				Validators: []validator.String{
					stringvalidator.OneOf("NO_TRAVERSAL", "DEPTH_FIRST", "BREADTH_FIRST"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// These are exposed via the API properties field for objects of type IP4Network
			"cidr": schema.StringAttribute{
				MarkdownDescription: "The CIDR address of the IPv4 network.",
				Computed:            true,
			},
			"template": schema.Int64Attribute{
				MarkdownDescription: "The ID of the linked template",
				Computed:            true,
			},
			"gateway": schema.StringAttribute{
				MarkdownDescription: "The gateway of the IPv4 network.",
				Computed:            true,
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`), "Gateway must be a valid IPv4 address"),
				},
			},
			"default_domains": schema.SetAttribute{
				MarkdownDescription: "The object ids of the default DNS domains for the network.",
				Computed:            true,
				Optional:            true,
				ElementType:         types.Int64Type,
				Default:             nil,
			},
			"default_view": schema.Int64Attribute{
				MarkdownDescription: "The object id of the default DNS View for the network.",
				Computed:            true,
				Optional:            true,
				Default:             nil,
			},
			"dns_restrictions": schema.SetAttribute{
				MarkdownDescription: "The object ids of the DNS restrictions for the network.",
				Computed:            true,
				Optional:            true,
				ElementType:         types.Int64Type,
				Default:             nil,
			},
			"allow_duplicate_host": schema.BoolAttribute{
				MarkdownDescription: "Duplicate host names check.",
				Computed:            true,
				Optional:            true,
				Default:             nil,
			},
			"ping_before_assign": schema.BoolAttribute{
				MarkdownDescription: "The network pings an address before assignment.",
				Computed:            true,
				Optional:            true,
				Default:             nil,
			},
			"inherit_allow_duplicate_host": schema.BoolAttribute{
				MarkdownDescription: "Duplicate host names check is inherited.",
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(true),
			},
			"inherit_ping_before_assign": schema.BoolAttribute{
				MarkdownDescription: "The network pings an address before assignment is inherited.",
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(true),
			},
			"inherit_dns_restrictions": schema.BoolAttribute{
				MarkdownDescription: "DNS restrictions are inherited.",
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(true),
			},
			"inherit_default_domains": schema.BoolAttribute{
				MarkdownDescription: "Default domains are inherited.",
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(true),
			},
			"inherit_default_view": schema.BoolAttribute{
				MarkdownDescription: "The default DNS View is inherited.",
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(true),
			},
			"location_code": schema.StringAttribute{
				MarkdownDescription: "The location code of the network.",
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
			"shared_network": schema.StringAttribute{
				MarkdownDescription: "The name of the shared network tag associated with the IP4 Network.",
				Computed:            true,
			},
			"user_defined_fields": schema.MapAttribute{
				MarkdownDescription: "A map of all user-definied fields associated with the IP4 Network.",
				Computed:            true,
				ElementType:         types.StringType,
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
	data.Properties = types.StringPointerValue(network.Properties)
	data.Type = types.StringPointerValue(network.Type)

	// we have an ID at this point so save the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	properties = ""

	if !data.Gateway.IsUnknown() {
		properties = properties + "gateway=" + data.Gateway.ValueString() + "|"
	}

	if !data.DefaultDomains.IsUnknown() {
		var defaultDomains []string
		data.DefaultDomains.ElementsAs(ctx, &defaultDomains, false)
		properties = properties + "defaultDomains=" + strings.Join(defaultDomains, ",") + "|"
	}

	if !data.DefaultView.IsUnknown() {
		properties = properties + "defaultView=" + strconv.FormatInt(data.DefaultView.ValueInt64(), 10) + "|"
	}

	if !data.DNSRestrictions.IsUnknown() {
		var dnsRestrictions []string
		data.DNSRestrictions.ElementsAs(ctx, &dnsRestrictions, false)
		properties = properties + "dnsRestrictions=" + strings.Join(dnsRestrictions, ",") + "|"
	}

	if !data.AllowDuplicateHost.IsUnknown() {
		properties = properties + "allowDuplicateHost=" + boolToEnableDisable(data.AllowDuplicateHost.ValueBoolPointer()) + "|"
	}

	if !data.PingBeforeAssign.IsUnknown() {
		properties = properties + "pingBeforeAssign=" + boolToEnableDisable(data.PingBeforeAssign.ValueBoolPointer()) + "|"
	}

	if !data.InheritAllowDuplicateHost.IsUnknown() {
		properties = properties + "inheritAllowDuplicateHost=" + strconv.FormatBool(data.InheritAllowDuplicateHost.ValueBool()) + "|"
	}

	if !data.InheritPingBeforeAssign.IsUnknown() {
		properties = properties + "inheritPingBeforeAssign=" + strconv.FormatBool(data.InheritPingBeforeAssign.ValueBool()) + "|"
	}

	if !data.InheritDNSRestrictions.IsUnknown() {
		properties = properties + "inheritDNSRestrictions=" + strconv.FormatBool(data.InheritDNSRestrictions.ValueBool()) + "|"
	}

	if !data.InheritDefaultDomains.IsUnknown() {
		properties = properties + "inheritDefaultDomains=" + strconv.FormatBool(data.InheritDefaultDomains.ValueBool()) + "|"
	}

	if !data.InheritDefaultView.IsUnknown() {
		properties = properties + "inheritDefaultView=" + strconv.FormatBool(data.InheritDefaultView.ValueBool()) + "|"
	}

	if !data.LocationCode.IsUnknown() {
		properties = properties + "locationCode=" + data.LocationCode.ValueString() + "|"
	}

	setName := gobam.APIEntity{
		Id:         data.ID.ValueInt64Pointer(),
		Name:       data.Name.ValueStringPointer(),
		Properties: &properties,
		Type:       data.Type.ValueStringPointer(),
	}

	err = client.Update(&setName)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"Failed to update created IP4 Network",
			err.Error(),
		)

		return
	}

	entity, err := client.GetEntityById(data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"Failed to get IP4 Network by Id",
			err.Error(),
		)
		return
	}

	networkProperties, diag := flattenIP4NetworkProperties(entity)
	if diag.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.Append(diag...)
		return
	}

	data.Name = types.StringPointerValue(entity.Name)
	data.Properties = types.StringPointerValue(entity.Properties)
	data.Type = types.StringPointerValue(entity.Type)
	data.CIDR = networkProperties.CIDR
	data.Template = networkProperties.Template
	data.Gateway = networkProperties.Gateway
	data.DefaultDomains = networkProperties.DefaultDomains
	data.DefaultView = networkProperties.DefaultView
	data.DNSRestrictions = networkProperties.DNSRestrictions
	data.AllowDuplicateHost = networkProperties.AllowDuplicateHost
	data.PingBeforeAssign = networkProperties.PingBeforeAssign
	data.InheritAllowDuplicateHost = networkProperties.InheritAllowDuplicateHost
	data.InheritPingBeforeAssign = networkProperties.InheritPingBeforeAssign
	data.InheritDNSRestrictions = networkProperties.InheritDNSRestrictions
	data.InheritDefaultDomains = networkProperties.InheritDefaultDomains
	data.InheritDefaultView = networkProperties.InheritDefaultView
	data.LocationCode = networkProperties.LocationCode
	data.LocationInherited = networkProperties.LocationInherited
	data.SharedNetwork = networkProperties.SharedNetwork
	data.UserDefinedFields = networkProperties.UserDefinedFields

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
		tflog.Trace(ctx, "IP4 Network was deleted outside terraform")
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringPointerValue(entity.Name)
	data.Properties = types.StringPointerValue(entity.Properties)
	data.Type = types.StringPointerValue(entity.Type)

	networkProperties, diag := flattenIP4NetworkProperties(entity)
	if diag.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.Append(diag...)
		return
	}

	data.CIDR = networkProperties.CIDR
	data.Template = networkProperties.Template
	data.Gateway = networkProperties.Gateway
	data.DefaultDomains = networkProperties.DefaultDomains
	data.DefaultView = networkProperties.DefaultView
	data.DNSRestrictions = networkProperties.DNSRestrictions
	data.AllowDuplicateHost = networkProperties.AllowDuplicateHost
	data.PingBeforeAssign = networkProperties.PingBeforeAssign
	data.InheritAllowDuplicateHost = networkProperties.InheritAllowDuplicateHost
	data.InheritPingBeforeAssign = networkProperties.InheritPingBeforeAssign
	data.InheritDNSRestrictions = networkProperties.InheritDNSRestrictions
	data.InheritDefaultDomains = networkProperties.InheritDefaultDomains
	data.InheritDefaultView = networkProperties.InheritDefaultView
	data.LocationCode = networkProperties.LocationCode
	data.LocationInherited = networkProperties.LocationInherited
	data.SharedNetwork = networkProperties.SharedNetwork
	data.UserDefinedFields = networkProperties.UserDefinedFields

	resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IP4NetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state *IP4NetworkResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, diag := clientLogin(ctx, r.client, mutex)
	if diag.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

		return
	}

	properties := ""

	if !data.Gateway.IsUnknown() && !data.Gateway.Equal(state.Gateway) {
		properties = properties + fmt.Sprintf("gateway=%s|", data.Gateway.ValueString())
	}

	if !data.DefaultDomains.IsUnknown() && !data.DefaultDomains.Equal(state.DefaultDomains) {
		var domains []string
		data.DefaultDomains.ElementsAs(ctx, &domains, false)
		if domains != nil {
			properties = properties + fmt.Sprintf("defaultDomains=%s|", strings.Join(domains, ","))
		}
	}

	if !data.DefaultView.IsUnknown() && !data.DefaultView.Equal(state.DefaultView) {

		properties = properties + fmt.Sprintf("defaultView=%s|", strconv.FormatInt(data.DefaultView.ValueInt64(), 10))

	}

	if !data.DNSRestrictions.IsUnknown() && !data.DNSRestrictions.Equal(state.DNSRestrictions) {
		var dns []string
		data.DNSRestrictions.ElementsAs(ctx, &dns, false)
		if dns != nil {
			properties = properties + fmt.Sprintf("dnsRestrictions=%s|", dns)
		}

	}

	if !data.AllowDuplicateHost.IsUnknown() && !data.AllowDuplicateHost.Equal(state.AllowDuplicateHost) {
		properties = properties + fmt.Sprintf("allowDuplicateHost=%s|", boolToEnableDisable(data.AllowDuplicateHost.ValueBoolPointer()))

	}

	if !data.PingBeforeAssign.IsUnknown() && !data.PingBeforeAssign.Equal(state.PingBeforeAssign) {
		properties = properties + fmt.Sprintf("pingBeforeAssign=%s|", boolToEnableDisable(data.PingBeforeAssign.ValueBoolPointer()))
	}

	if !data.InheritAllowDuplicateHost.Equal(state.InheritAllowDuplicateHost) {
		properties = properties + fmt.Sprintf("inheritAllowDuplicateHost=%s|", strconv.FormatBool(data.InheritAllowDuplicateHost.ValueBool()))
	}

	if !data.InheritPingBeforeAssign.Equal(state.InheritPingBeforeAssign) {
		properties = properties + fmt.Sprintf("inheritPingBeforeAssign=%s|", strconv.FormatBool(data.InheritPingBeforeAssign.ValueBool()))
	}

	if !data.InheritDNSRestrictions.Equal(state.InheritDNSRestrictions) {
		properties = properties + fmt.Sprintf("inheritDNSRestrictions=%s|", strconv.FormatBool(data.InheritDNSRestrictions.ValueBool()))
	}

	if !data.InheritDefaultDomains.Equal(state.InheritDefaultDomains) {
		properties = properties + fmt.Sprintf("inheritDefaultDomains=%s|", strconv.FormatBool(data.InheritDefaultDomains.ValueBool()))

	}

	if !data.InheritDefaultView.Equal(state.InheritDefaultView) {
		properties = properties + fmt.Sprintf("inheritDefaultView=%s|", strconv.FormatBool(data.InheritDefaultView.ValueBool()))
	}

	if !data.LocationCode.IsUnknown() && !data.LocationCode.Equal(state.LocationCode) {
		properties = properties + fmt.Sprintf("locationCode=%s|", data.LocationCode.ValueString())
	}

	update := gobam.APIEntity{
		Id:         state.ID.ValueInt64Pointer(),
		Name:       data.Name.ValueStringPointer(),
		Properties: &properties,
		Type:       state.Type.ValueStringPointer(),
	}

	tflog.Debug(ctx, fmt.Sprintf("Attempting to update IP4Network with properties: %s", properties))

	err := client.Update(&update)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"IP4 Network Update failed",
			err.Error(),
		)
		return
	}

	entity, err := client.GetEntityById(data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"Failed to get IP4 Network by Id",
			err.Error(),
		)
		return
	}

	data.Name = types.StringPointerValue(entity.Name)
	data.Properties = types.StringPointerValue(entity.Properties)
	data.Type = types.StringPointerValue(entity.Type)

	networkProperties, diag := flattenIP4NetworkProperties(entity)
	if diag.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.Append(diag...)
		return
	}

	data.CIDR = networkProperties.CIDR
	data.Template = networkProperties.Template
	data.Gateway = networkProperties.Gateway
	data.DefaultDomains = networkProperties.DefaultDomains
	data.DefaultView = networkProperties.DefaultView
	data.DNSRestrictions = networkProperties.DNSRestrictions
	data.AllowDuplicateHost = networkProperties.AllowDuplicateHost
	data.PingBeforeAssign = networkProperties.PingBeforeAssign
	data.InheritAllowDuplicateHost = networkProperties.InheritAllowDuplicateHost
	data.InheritPingBeforeAssign = networkProperties.InheritPingBeforeAssign
	data.InheritDNSRestrictions = networkProperties.InheritDNSRestrictions
	data.InheritDefaultDomains = networkProperties.InheritDefaultDomains
	data.InheritDefaultView = networkProperties.InheritDefaultView
	data.LocationCode = networkProperties.LocationCode
	data.LocationInherited = networkProperties.LocationInherited
	data.SharedNetwork = networkProperties.SharedNetwork
	data.UserDefinedFields = networkProperties.UserDefinedFields

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

func (r IP4NetworkResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data IP4NetworkResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// if inherit_allow_duplicate_host is true, allow_duplicate_host must be unset
	if data.InheritAllowDuplicateHost.ValueBool() && !data.AllowDuplicateHost.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("allow_duplicate_host"),
			"Attribute Conflict",
			"allow_duplicate_host cannot be configured if inherit_allow_duplicate_host is true.",
		)
	}

	// if inherit_allow_duplicate_host is false, allow_duplicate_host must be set
	if !data.InheritAllowDuplicateHost.ValueBool() && data.AllowDuplicateHost.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("allow_duplicate_host"),
			"Attribute Conflict",
			"allow_duplicate_host must be configured if inherit_allow_duplicate_host is false.",
		)
	}

	// if inherit_dns_restrictions is true, dns_restrictions must be unset
	if data.InheritDNSRestrictions.ValueBool() && !data.DNSRestrictions.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("dns_restrictions"),
			"Attribute Conflict",
			"dns_restrictions cannot be configured if inherit_dns_restrictions is true.",
		)
	}

	// if inherit_dns_restrictions is false, dns_restrictions must be set
	if !data.InheritDNSRestrictions.ValueBool() && data.DNSRestrictions.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("dns_restrictions"),
			"Attribute Conflict",
			"allow_duplicate_host must be configured if inherit_allow_duplicate_host is false.",
		)
	}

	// if inherit_default_domains is true, default_domains must be unset
	if data.InheritDefaultDomains.ValueBool() && !data.DefaultDomains.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("default_domains"),
			"Attribute Conflict",
			"default_domains cannot be configured if inherit_default_domains is true.",
		)
	}

	// if inherit_default_domains is false, default_domains must be set
	if !data.InheritDefaultDomains.ValueBool() && data.DefaultDomains.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("default_domains"),
			"Attribute Conflict",
			"default_domains must be configured if inherit_default_domains is false.",
		)
	}

	// if inherit_default_view is true, default_view must be unset
	if data.InheritDefaultView.ValueBool() && !data.DefaultView.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("default_view"),
			"Attribute Conflict",
			"default_view cannot be configured if inherit_default_view is true.",
		)
	}

	// if inherit_default_view is false, default_view must be set
	if !data.InheritDefaultView.ValueBool() && data.DefaultView.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("default_view"),
			"Attribute Conflict",
			"default_view must be configured if inherit_default_view is false.",
		)
	}

	// if inherit_ping_before_assign is true, ping_before_assign must be unset
	if data.InheritPingBeforeAssign.ValueBool() && !data.PingBeforeAssign.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("ping_before_assign"),
			"Attribute Conflict",
			"ping_before_assign cannot be configured if inherit_ping_before_assign is true.",
		)
	}

	// if inherit_ping_before_assign is false, ping_before_assign must be set
	if !data.InheritPingBeforeAssign.ValueBool() && data.PingBeforeAssign.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("ping_before_assign"),
			"Attribute Conflict",
			"ping_before_assign must be configured if inherit_ping_before_assign is false.",
		)
	}
}
