package provider

import (
	"context"
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
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
var _ resource.Resource = &IP4BlockResource{}
var _ resource.ResourceWithImportState = &IP4BlockResource{}

func NewIP4BlockResource() resource.Resource {
	return &IP4BlockResource{}
}

// IP4BlockResource defines the resource implementation.
type IP4BlockResource struct {
	client *loginClient
}

// IP4BlockResourceModel describes the resource data model.
type IP4BlockResourceModel struct {
	// These are exposed for a generic entity object in bluecat
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	Properties types.String `tfsdk:"properties"`

	// These are exposed via the entity properties field for objects of type IP4Block
	CIDR                      types.String `tfsdk:"cidr"`
	DefaultDomains            types.Set    `tfsdk:"default_domains"`
	Start                     types.String `tfsdk:"start"`
	End                       types.String `tfsdk:"end"`
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

	// these are user defined fields that are not built-in
	UserDefinedFields types.Map `tfsdk:"user_defined_fields"`

	// These fields are only used for creation
	IsLargerAllowed types.Bool   `tfsdk:"is_larger_allowed"`
	ParentID        types.Int64  `tfsdk:"parent_id"`
	Size            types.Int64  `tfsdk:"size"`
	TraversalMethod types.String `tfsdk:"traversal_method"`
}

func (r *IP4BlockResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip4_block"
}

func (r *IP4BlockResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Resource to create an IPv4 block.",

		Attributes: map[string]schema.Attribute{
			// These are exposed for Entity objects via the API
			"id": schema.StringAttribute{
				MarkdownDescription: "IPv4 Block identifier.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The display name of the IPv4 block.",
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
			"is_larger_allowed": schema.BoolAttribute{
				MarkdownDescription: "(Optional) Is it ok to return a block that is larger than the size specified?",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIf(ip4BlockIsLargerAllowedPlanModifier, ip4BlockIsLargerAllowedPlanModifierDescription, ip4BlockIsLargerAllowedPlanModifierDescription),
				},
			},
			"parent_id": schema.Int64Attribute{
				MarkdownDescription: "The object ID of the parent object that will contain the new IPv4 block. If this argument is changed, then the resource will be recreated.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "The size of the IPv4 block expressed as a power of 2. For example, 256 would create a /24. If this argument is changed, then the resource will be recreated.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"traversal_method": schema.StringAttribute{
				MarkdownDescription: "The traversal method used to find the range to allocate the block. Must be one of \"NO_TRAVERSAL\", \"DEPTH_FIRST\", or \"BREADTH_FIRST\".",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("NO_TRAVERSAL"),
				Validators: []validator.String{
					stringvalidator.OneOf("NO_TRAVERSAL", "DEPTH_FIRST", "BREADTH_FIRST"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(ip4BlockTraversalMethodPlanModifier, ip4BlockTraversalMethodPlanModifierDescription, ip4BlockTraversalMethodPlanModifierDescription),
				},
			},

			// These are exposed via the API properties field for objects of type IP4Block
			"cidr": schema.StringAttribute{
				MarkdownDescription: "The CIDR value of the block (if it forms a valid CIDR).",
				Computed:            true,
			},
			"default_domains": schema.SetAttribute{
				MarkdownDescription: "The object ids of the default DNS domains.",
				Computed:            true,
				Optional:            true,
				ElementType:         types.Int64Type,
				Default:             nil,
			},
			"start": schema.StringAttribute{
				MarkdownDescription: "The start of the block (if it does not form a valid CIDR).",
				Computed:            true,
			},
			"end": schema.StringAttribute{
				MarkdownDescription: "The end of the block (if it does not form a valid CIDR).",
				Computed:            true,
			},
			"default_view": schema.Int64Attribute{
				MarkdownDescription: "The object id of the default DNS View for the block.",
				Computed:            true,
				Optional:            true,
				Default:             nil,
			},
			"dns_restrictions": schema.SetAttribute{
				MarkdownDescription: "The object ids of the DNS restrictions for the block.",
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
				MarkdownDescription: "Option to ping check. The possible values are enable and disable.",
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
				MarkdownDescription: "PingBeforeAssign option inheritance check option property.",
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
				MarkdownDescription: "The location code of the block.",
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
				MarkdownDescription: "A map of all user-definied fields associated with the IP4 Block.",
				Computed:            true,
				Optional:            true,
				Default:             mapdefault.StaticValue(basetypes.NewMapValueMust(types.StringType, nil)),
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *IP4BlockResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IP4BlockResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *IP4BlockResourceModel

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
	reuseExisting := false //we never want to use an existing block created outside terraform
	Type := "IP4Block"   //Since this is the ip4_block resource we are setting the type
	properties := "reuseExisting=" + strconv.FormatBool(reuseExisting) + "|"
	properties = properties + "isLargerAllowed=" + strconv.FormatBool(isLargerAllowed) + "|"
	properties = properties + "autoCreate=" + strconv.FormatBool(autoCreate) + "|"
	properties = properties + "traversalMethod=" + traversalMethod + "|"

	block, err := client.GetNextAvailableIPRange(parentID, size, Type, properties)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"Failed to create IP4 Block",
			err.Error(),
		)
		return
	}

	data.ID = types.StringValue(strconv.FormatInt(*block.Id, 10))
	data.Properties = types.StringPointerValue(block.Properties)
	data.Type = types.StringPointerValue(block.Type)

	// we have an ID at this point so save the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	properties = ""

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

	var udfs map[string]string
	data.UserDefinedFields.ElementsAs(ctx, &udfs, false)
	for k, v := range udfs {
		properties = properties + k + "=" + v + "|"
	}

	setName := gobam.APIEntity{
		Id:         block.Id,
		Name:       data.Name.ValueStringPointer(),
		Properties: &properties,
		Type:       data.Type.ValueStringPointer(),
	}

	err = client.Update(&setName)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"Failed to update created IP4 Block",
			err.Error(),
		)

		return
	}

	entity, err := client.GetEntityById(*block.Id)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"Failed to get IP4 Block by Id",
			err.Error(),
		)
		return
	}

	blockProperties, diag := flattenIP4BlockProperties(entity)
	if diag.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.Append(diag...)
		return
	}

	data.Name = types.StringPointerValue(entity.Name)
	data.Properties = types.StringPointerValue(entity.Properties)
	data.Type = types.StringPointerValue(entity.Type)
	data.CIDR = blockProperties.CIDR
	data.DefaultDomains = blockProperties.DefaultDomains
	data.Start = blockProperties.Start
	data.End = blockProperties.End
	data.DefaultView = blockProperties.DefaultView
	data.DNSRestrictions = blockProperties.DNSRestrictions
	data.AllowDuplicateHost = blockProperties.AllowDuplicateHost
	data.PingBeforeAssign = blockProperties.PingBeforeAssign
	data.InheritAllowDuplicateHost = blockProperties.InheritAllowDuplicateHost
	data.InheritPingBeforeAssign = blockProperties.InheritPingBeforeAssign
	data.InheritDNSRestrictions = blockProperties.InheritDNSRestrictions
	data.InheritDefaultDomains = blockProperties.InheritDefaultDomains
	data.InheritDefaultView = blockProperties.InheritDefaultView
	data.LocationCode = blockProperties.LocationCode
	data.LocationInherited = blockProperties.LocationInherited
	data.UserDefinedFields = blockProperties.UserDefinedFields

	resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IP4BlockResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *IP4BlockResourceModel

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

	id, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Failed to parse ID", err.Error())
		return
	}

	entity, err := client.GetEntityById(id)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"Failed to get IP4 Block by Id",
			err.Error(),
		)
		return
	}

	if *entity.Id == 0 {
		tflog.Trace(ctx, "IP4 Block was deleted outside terraform")
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringPointerValue(entity.Name)
	data.Properties = types.StringPointerValue(entity.Properties)
	data.Type = types.StringPointerValue(entity.Type)

	blockProperties, diag := flattenIP4BlockProperties(entity)
	if diag.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.Append(diag...)
		return
	}

	data.CIDR = blockProperties.CIDR
	data.DefaultDomains = blockProperties.DefaultDomains
	data.Start = blockProperties.Start
	data.End = blockProperties.End
	data.DefaultView = blockProperties.DefaultView
	data.DNSRestrictions = blockProperties.DNSRestrictions
	data.AllowDuplicateHost = blockProperties.AllowDuplicateHost
	data.PingBeforeAssign = blockProperties.PingBeforeAssign
	data.InheritAllowDuplicateHost = blockProperties.InheritAllowDuplicateHost
	data.InheritPingBeforeAssign = blockProperties.InheritPingBeforeAssign
	data.InheritDNSRestrictions = blockProperties.InheritDNSRestrictions
	data.InheritDefaultDomains = blockProperties.InheritDefaultDomains
	data.InheritDefaultView = blockProperties.InheritDefaultView
	data.LocationCode = blockProperties.LocationCode
	data.LocationInherited = blockProperties.LocationInherited
	data.UserDefinedFields = blockProperties.UserDefinedFields

	// calculate the size of the block so we can set it in the state so import works
	cidrNetmask, err := strconv.ParseInt(strings.Split(blockProperties.CIDR.ValueString(), "/")[1], 10, 64)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Failed to parse CIDR netmask to integer", err.Error())
		return
	}
	var size, e = big.NewInt(2), big.NewInt(32 - cidrNetmask)
	size.Exp(size, e, nil)
	data.Size = types.Int64Value(size.Int64())

	// get the parent id of the block so we can set it in the state so import works
	parent, err := client.GetParent(id)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Failed to get parent entity of IP4 Block", err.Error())
		return
	}
	data.ParentID = types.Int64Value(*parent.Id)

	resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IP4BlockResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state *IP4BlockResourceModel

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

	tflog.Debug(ctx, fmt.Sprintf("Attempting to update IP4Block with properties: %s", properties))

	err = client.Update(&update)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"IP4 Block Update failed",
			err.Error(),
		)
		return
	}

	entity, err := client.GetEntityById(id)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"Failed to get IP4 Block by Id",
			err.Error(),
		)
		return
	}

	data.Name = types.StringPointerValue(entity.Name)
	data.Properties = types.StringPointerValue(entity.Properties)
	data.Type = types.StringPointerValue(entity.Type)

	blockProperties, diag := flattenIP4BlockProperties(entity)
	if diag.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.Append(diag...)
		return
	}

	data.CIDR = blockProperties.CIDR
	data.DefaultDomains = blockProperties.DefaultDomains
	data.Start = blockProperties.Start
	data.End = blockProperties.End
	data.DefaultView = blockProperties.DefaultView
	data.DNSRestrictions = blockProperties.DNSRestrictions
	data.AllowDuplicateHost = blockProperties.AllowDuplicateHost
	data.PingBeforeAssign = blockProperties.PingBeforeAssign
	data.InheritAllowDuplicateHost = blockProperties.InheritAllowDuplicateHost
	data.InheritPingBeforeAssign = blockProperties.InheritPingBeforeAssign
	data.InheritDNSRestrictions = blockProperties.InheritDNSRestrictions
	data.InheritDefaultDomains = blockProperties.InheritDefaultDomains
	data.InheritDefaultView = blockProperties.InheritDefaultView
	data.LocationCode = blockProperties.LocationCode
	data.LocationInherited = blockProperties.LocationInherited
	data.UserDefinedFields = blockProperties.UserDefinedFields

	resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IP4BlockResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *IP4BlockResourceModel

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

	id, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Failed to parse ID", err.Error())
		return
	}

	entity, err := client.GetEntityById(id)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"Failed to get IP4 Block by Id",
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

func (r *IP4BlockResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r IP4BlockResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data IP4BlockResourceModel

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
	if !data.InheritPingBeforeAssign.ValueBool() && data.PingBeforeAssign.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("ping_before_assign"),
			"Attribute Conflict",
			"ping_before_assign must be configured if inherit_ping_before_assign is false.",
		)
	}
}

const ip4BlockIsLargerAllowedPlanModifierDescription string = "is_larger_allowed is required for creation and cannot be changed. Null values in the state are ignored to allow for import."

func ip4BlockIsLargerAllowedPlanModifier(ctx context.Context, p planmodifier.BoolRequest, resp *boolplanmodifier.RequiresReplaceIfFuncResponse) {
	var state *IP4BlockResourceModel
	resp.Diagnostics.Append(p.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.IsLargerAllowed.IsNull() {
		// Since this is an optional field with a default value, it should only be null when doing an import
		resp.RequiresReplace = false
		return
	}

	resp.RequiresReplace = true
}

const ip4BlockTraversalMethodPlanModifierDescription string = "traversal_method is required for creation and cannot be changed. Null values in the state are ignored to allow for import."

func ip4BlockTraversalMethodPlanModifier(ctx context.Context, p planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	var state *IP4BlockResourceModel
	resp.Diagnostics.Append(p.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.TraversalMethod.IsNull() {
		// Since this is a required field with required values, it should only be null when doing an import
		resp.RequiresReplace = false
		return
	}

	resp.RequiresReplace = true
}


