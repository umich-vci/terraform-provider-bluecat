package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	// These are exposed for a generic entity object in bluecat
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	Properties types.String `tfsdk:"properties"`

	// This is used to help find the IP4Address
	ContainerID types.Int64 `tfsdk:"container_id"`

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
}

func (d *IP4AddressDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip4_address"
}

func (d *IP4AddressDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Data source to access the attributes of an IPv4 address.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
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

	client, diag := clientLogin(ctx, d.client, mutex)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	containerID := data.ContainerID.ValueInt64()
	address := data.Address.ValueString()

	ip4Address, err := client.GetIP4Address(containerID, address)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Failed to get IP4 Address", err.Error())
		return
	}

	data.ID = types.StringValue(strconv.FormatInt(*ip4Address.Id, 10))
	data.Name = types.StringPointerValue(ip4Address.Name)
	data.Properties = types.StringPointerValue(ip4Address.Properties)
	data.Type = types.StringPointerValue(ip4Address.Type)

	addressProperties, diag := flattenIP4AddressProperties(ip4Address)
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

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
