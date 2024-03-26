package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &IP4NetworkDataSource{}

func NewIP4NetworkDataSource() datasource.DataSource {
	return &IP4NetworkDataSource{}
}

// IP4NetworkDataSource defines the data source implementation.
type IP4NetworkDataSource struct {
	client *loginClient
}

// IP4NetworkDataSourceModel describes the data source data model.
type IP4NetworkDataSourceModel struct {
	// These are exposed for a generic entity object in bluecat
	ID         types.String `tfsdk:"id"`
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

	// these exist only for the data source to find the network
	ContainerID types.Int64  `tfsdk:"container_id"`
	Hint        types.String `tfsdk:"hint"`
}

func (d *IP4NetworkDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip4_network"
}

func (d *IP4NetworkDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Data source to access the attributes of an IPv4 network from a hint based search.",

		Attributes: map[string]schema.Attribute{
			"container_id": schema.Int64Attribute{
				MarkdownDescription: "The object ID of a container that contains the specified IPv4 network.",
				Required:            true,
			},
			"hint": schema.StringAttribute{
				MarkdownDescription: "Hint to find the IP4Network",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID assigned to the IP4Network.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name assigned to the IP4Network.",
				Computed:            true,
			},
			"properties": schema.StringAttribute{
				MarkdownDescription: "The properties of the IP4Network (pipe delimited).",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the entity.",
				Computed:            true,
			},
			"cidr": schema.StringAttribute{
				MarkdownDescription: "The CIDR address of the IP4Network.",
				Computed:            true,
			},
			"template": schema.Int64Attribute{
				MarkdownDescription: "The ID of the linked template",
				Computed:            true,
			},
			"gateway": schema.StringAttribute{
				MarkdownDescription: "The gateway of the IP4Network.",
				Computed:            true,
			},
			"default_domains": schema.SetAttribute{
				MarkdownDescription: "The object ids of the default DNS domains for the network.",
				Computed:            true,
				ElementType:         types.Int64Type,
			},
			"default_view": schema.Int64Attribute{
				MarkdownDescription: "The object id of the default DNS View for the network.",
				Computed:            true,
			},
			"dns_restrictions": schema.SetAttribute{
				MarkdownDescription: "The object ids of the DNS restrictions for the network.",
				Computed:            true,
				ElementType:         types.Int64Type,
			},
			"allow_duplicate_host": schema.BoolAttribute{
				MarkdownDescription: "Duplicate host names check.",
				Computed:            true,
			},
			"ping_before_assign": schema.BoolAttribute{
				MarkdownDescription: "The network pings an address before assignment.",
				Computed:            true,
			},
			"inherit_allow_duplicate_host": schema.BoolAttribute{
				MarkdownDescription: "Duplicate host names check is inherited.",
				Computed:            true,
			},
			"inherit_ping_before_assign": schema.BoolAttribute{
				MarkdownDescription: "The network pings an address before assignment is inherited.",
				Computed:            true,
			},
			"inherit_dns_restrictions": schema.BoolAttribute{
				MarkdownDescription: "DNS restrictions are inherited.",
				Computed:            true,
			},
			"inherit_default_domains": schema.BoolAttribute{
				MarkdownDescription: "Default domains are inherited.",
				Computed:            true,
			},
			"inherit_default_view": schema.BoolAttribute{
				MarkdownDescription: "The default DNS View is inherited.",
				Computed:            true,
			},
			"location_code": schema.StringAttribute{
				MarkdownDescription: "The location code of the network.",
				Computed:            true,
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
				MarkdownDescription: "A map of all user-definied fields associated with the entity.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *IP4NetworkDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IP4NetworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IP4NetworkDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, diag := clientLogin(ctx, d.client, mutex)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	containerID := data.ContainerID.ValueInt64()
	hint := data.Hint.ValueString()
	options := "hint=" + hint

	hintResp, err := client.GetIP4NetworksByHint(containerID, 0, 1, options)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Failed to get IP4 Networks by hint", err.Error())
		return
	}

	if len(hintResp.Item) > 1 || len(hintResp.Item) == 0 {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"Network lookup error",
			fmt.Sprintf("Hint %s returned %d networks but the data source only supports 1", hint, len(hintResp.Item)),
		)
		return
	}

	data.ID = types.StringValue(strconv.FormatInt(*hintResp.Item[0].Id, 10))

	// GetIP4NetworksByHint doesn't seem to return all properties so use the ID returned by it to call GetEntityById
	entity, err := client.GetEntityById(*hintResp.Item[0].Id)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError(
			"Failed to get IP4 Network via Entity ID",
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

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
