// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/umich-vci/gobam"
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
	ID                        types.Int64  `tfsdk:"id"`
	ContainerID               types.Int64  `tfsdk:"container_id"`
	Hint                      types.String `tfsdk:"hint"`
	Type                      types.String `tfsdk:"type"`
	AddressesFree             types.Int64  `tfsdk:"addresses_free"`
	AddressesInUse            types.Int64  `tfsdk:"addresses_in_use"`
	AllowDuplicateHost        types.String `tfsdk:"allow_duplicate_host"`
	CIDR                      types.String `tfsdk:"cidr"`
	CustomProperties          types.Map    `tfsdk:"custom_properties"`
	DefaultDomains            types.Set    `tfsdk:"default_domain"`
	DefaultView               types.Int64  `tfsdk:"default_view"`
	DNSRestrictions           types.Set    `tfsdk:"dns_restrictions"`
	Gateway                   types.String `tfsdk:"gateway"`
	InheritAllowDuplicateHost types.Bool   `tfsdk:"inherit_allow_duplicate_host"`
	InheritDefaultDomains     types.Bool   `tfsdk:"inherit_default_domain"`
	InheritDefaultView        types.Bool   `tfsdk:"inherit_default_view"`
	InheritDNSRestrictions    types.Bool   `tfsdk:"inherit_dns_restrictions"`
	InheritPingBeforeAssign   types.Bool   `tfsdk:"inherit_ping_before_assign"`
	LocationCode              types.String `tfsdk:"location_code"`
	LocationInherited         types.Bool   `tfsdk:"location_inherited"`
	Name                      types.String `tfsdk:"name"`
	PingBeforeAssign          types.String `tfsdk:"ping_before_assign"`
	Properties                types.String `tfsdk:"properties"`
	Template                  types.Int64  `tfsdk:"template"`
}

func (d *IP4NetworkDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip4_network"
}

func (d *IP4NetworkDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Data source to access the attributes of an IPv4 network from a hint based search.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Example identifier",
				Computed:            true,
			},
			"hint": schema.StringAttribute{
				MarkdownDescription: "Hint to find the IP4Network",
				Required:            true,
			},
			"container_id": schema.Int64Attribute{
				MarkdownDescription: "The object ID of a container that contains the specified IPv4 network.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the IP4Network",
				Computed:            true,
			},
			"addresses_free": schema.StringAttribute{
				MarkdownDescription: "The number of addresses unallocated/free on the network.",
				Computed:            true,
			},
			"addresses_in_use": schema.Int64Attribute{
				MarkdownDescription: "The number of addresses allocated/in use on the network.",
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
			"default_domains": schema.SetAttribute{
				MarkdownDescription: "TODO",
				Computed:            true,
				ElementType:         types.Int64Type,
			},
			"default_view": schema.Int64Attribute{
				MarkdownDescription: "The object id of the default DNS View for the network.",
				Computed:            true,
			},
			"dns_restrictions": schema.SetAttribute{
				MarkdownDescription: "TODO",
				Computed:            true,
				ElementType:         types.Int64Type,
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
				MarkdownDescription: "The default DNS View is inherited.",
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
			"location_code": schema.StringAttribute{
				MarkdownDescription: "TODO",
				Computed:            true,
			},
			"location_inherited": schema.BoolAttribute{
				MarkdownDescription: "TODO",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name assigned the resource.",
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
			"template": schema.Int64Attribute{
				MarkdownDescription: "TODO",
				Computed:            true,
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

	mutex.Lock()
	client := d.client.Client
	client.Login(d.client.Username, d.client.Password)

	containerID := data.ContainerID.ValueInt64()
	hint := data.Hint.ValueString()
	options := "hint=" + hint

	hintResp, err := client.GetIP4NetworksByHint(containerID, 0, 1, options)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get IP4 Networks by hint"); err != nil {
		mutex.Unlock()
		resp.Diagnostics.AddError(
			"Failed to get IP4 Networks by hint",
			err.Error(),
		)
		return
	}

	if len(hintResp.Item) > 1 || len(hintResp.Item) == 0 {
		err := gobam.LogoutClientWithError(client, "Network lookup error")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error logging out",
				err.Error(),
			)
		}
		mutex.Unlock()

		resp.Diagnostics.AddError(
			"Network lookup error",
			fmt.Sprintf("Hint %s returned %d networks but the data source only supports 1", hint, len(hintResp.Item)),
		)
		return
	}

	data.ID = types.Int64PointerValue(hintResp.Item[0].Id)

	// GetIP4NetworksByHint doesn't seem to return all properties so use the ID returned by it to call GetEntityById
	entity, err := client.GetEntityById(*hintResp.Item[0].Id)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get IP4 Network via Entity ID"); err != nil {
		mutex.Unlock()
		resp.Diagnostics.AddError(
			"Failed to get IP4 Network via Entity ID",
			err.Error(),
		)
		return
	}

	data.Name = types.StringPointerValue(entity.Name)
	data.Properties = types.StringPointerValue(entity.Properties)
	data.Type = types.StringPointerValue(entity.Type)

	networkProperties, diag := parseIP4NetworkProperties(*entity.Properties)
	if diag.HasError() {
		clientLogout(&client, mutex, resp.Diagnostics)
		resp.Diagnostics.Append(diag...)
		return
	}

	data.CIDR = networkProperties.cidr
	data.Template = networkProperties.template

	data.Gateway = networkProperties.gateway
	data.DefaultDomains = networkProperties.defaultDomains
	data.DefaultView = networkProperties.defaultView
	data.DNSRestrictions = networkProperties.dnsRestrictions
	data.AllowDuplicateHost = networkProperties.allowDuplicateHost
	data.PingBeforeAssign = networkProperties.pingBeforeAssign
	data.InheritAllowDuplicateHost = networkProperties.inheritAllowDuplicateHost
	data.InheritPingBeforeAssign = networkProperties.inheritPingBeforeAssign
	data.InheritDNSRestrictions = networkProperties.inheritDNSRestrictions
	data.InheritDefaultDomains = networkProperties.inheritDefaultDomains
	data.InheritDefaultView = networkProperties.inheritDefaultView
	data.LocationCode = networkProperties.locationCode
	data.LocationInherited = networkProperties.locationInherited
	data.CustomProperties = networkProperties.customProperties

	addressesInUse, addressesFree, err := getIP4NetworkAddressUsage(*entity.Id, networkProperties.cidr.ValueString(), client)
	if err = gobam.LogoutClientIfError(client, err, "Error calculating network usage"); err != nil {
		mutex.Unlock()
		resp.Diagnostics.AddError(
			"Error calculating network usage",
			err.Error(),
		)
		return
	}
	data.AddressesInUse = types.Int64Value(addressesInUse)
	data.AddressesFree = types.Int64Value(addressesFree)

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

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
