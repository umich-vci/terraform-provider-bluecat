package provider

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/umich-vci/gobam"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &IP4NBRDataSource{}

func NewIP4NBRDataSource() datasource.DataSource {
	return &IP4NBRDataSource{}
}

// IP4NBRDataSource defines the data source implementation.
type IP4NBRDataSource struct {
	client *loginClient
}

// IP4NBRDataSourceModel describes the data source data model.
type IP4NBRDataSourceModel struct {
	ID                        types.Int64  `tfsdk:"id"`
	Address                   types.String `tfsdk:"address"`
	ContainerID               types.Int64  `tfsdk:"container_id"`
	Type                      types.String `tfsdk:"type"`
	AddressesFree             types.Int64  `tfsdk:"addresses_free"`
	AddressesInUse            types.Int64  `tfsdk:"addresses_in_use"`
	AllowDuplicateHost        types.String `tfsdk:"allow_duplicate_host"`
	CIDR                      types.String `tfsdk:"cidr"`
	CustomProperties          types.Map    `tfsdk:"custom_properties"`
	DefaultDomains            types.Set    `tfsdk:"default_domains"`
	DefaultView               types.Int64  `tfsdk:"default_view"`
	DNSRestrictions           types.Set    `tfsdk:"dns_restrictions"`
	Gateway                   types.String `tfsdk:"gateway"`
	InheritAllowDuplicateHost types.Bool   `tfsdk:"inherit_allow_duplicate_host"`
	InheritDefaultDomains     types.Bool   `tfsdk:"inherit_default_domains"`
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

func (d *IP4NBRDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip4_nbr"
}

func (d *IP4NBRDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Data source to access the attributes of an IPv4 network, IPv4 Block, or DHCPv4 Range from an IPv4 address.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Example identifier",
				Computed:            true,
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "IP address to find the IPv4 network, IPv4 Block, or DHCPv4 Range of.",
				Required:            true,
			},
			"container_id": schema.Int64Attribute{
				MarkdownDescription: "The object ID of a container that contains the specified IPv4 network, block, or range.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Must be \"IP4Block\", \"IP4Network\", \"DHCP4Range\", or \"\". \"\" will find the most specific container.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("IP4Block", "IP4Network", "DHCP4Range", ""),
				},
			},
			"addresses_free": schema.Int64Attribute{
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

func (d *IP4NBRDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IP4NBRDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IP4NBRDataSourceModel

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
	otype := data.Type.ValueString()
	address := data.Address.ValueString()

	ipRange, err := client.GetIPRangedByIP(containerID, otype, address)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Failed to get IP4 Networks by hint", err.Error())
		return
	}

	data.ID = types.Int64PointerValue(ipRange.Id)
	data.Name = types.StringPointerValue(ipRange.Name)
	data.Properties = types.StringPointerValue(ipRange.Properties)
	data.Type = types.StringPointerValue(ipRange.Type)

	tflog.Info(ctx, fmt.Sprintf("parsing properties: %s", *ipRange.Properties))
	networkProperties, diag := parseIP4NetworkProperties(*ipRange.Properties)
	if diag.HasError() {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
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

	addressesInUse, addressesFree, err := getIP4NetworkAddressUsage(*ipRange.Id, networkProperties.cidr.ValueString(), client)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Error calculating network usage", err.Error())
		return
	}
	data.AddressesInUse = types.Int64Value(addressesInUse)
	data.AddressesFree = types.Int64Value(addressesFree)

	resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ip4NetworkProperties contains all properties returned by an IP4Network.
type ip4NetworkProperties struct {
	name                      types.String
	cidr                      types.String
	template                  types.Int64
	gateway                   types.String
	defaultDomains            types.Set
	defaultView               types.Int64
	dnsRestrictions           types.Set
	allowDuplicateHost        types.String
	pingBeforeAssign          types.String
	inheritAllowDuplicateHost types.Bool
	inheritPingBeforeAssign   types.Bool
	inheritDNSRestrictions    types.Bool
	inheritDefaultDomains     types.Bool
	inheritDefaultView        types.Bool
	locationCode              types.String
	locationInherited         types.Bool
	customProperties          types.Map
}

func parseIP4NetworkProperties(properties string) (ip4NetworkProperties, diag.Diagnostics) {
	networkProperties := ip4NetworkProperties{
		defaultDomains:   basetypes.NewSetNull(types.Int64Type),
		dnsRestrictions:  basetypes.NewSetNull(types.Int64Type),
		customProperties: basetypes.NewMapNull(types.StringType),
	}
	var diag diag.Diagnostics
	cpMap := make(map[string]attr.Value)

	props := strings.Split(properties, "|")
	for x := range props {
		if len(props[x]) > 0 {
			prop := strings.Split(props[x], "=")[0]
			val := strings.Split(props[x], "=")[1]

			switch prop {
			case "name":
				networkProperties.name = types.StringValue(val)
			case "CIDR":
				networkProperties.cidr = types.StringValue(val)
			case "template":
				t, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					diag.AddError("error parsing template to int64", err.Error())
					break
				}
				networkProperties.template = types.Int64Value(t)
			case "gateway":
				networkProperties.gateway = types.StringValue(val)
			case "defaultDomains":
				defaultDomains := strings.Split(val, ",")
				defaultDomainsList := []attr.Value{}
				for i := range defaultDomains {
					dID, err := strconv.ParseInt(defaultDomains[i], 10, 64)
					if err != nil {
						diag.AddError("error parsing defaultDomains to int64", err.Error())
						break
					}
					defaultDomainsList = append(defaultDomainsList, types.Int64Value(dID))
				}
				defaultDomainsSet, d := basetypes.NewSetValue(types.Int64Type, defaultDomainsList)
				if d.HasError() {
					diag.Append(d...)
					break
				}
				networkProperties.defaultDomains = defaultDomainsSet
			case "defaultView":
				dv, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					diag.AddError("error parsing defaultView to int64", err.Error())
					break
				}
				networkProperties.defaultView = types.Int64Value(dv)
			case "dnsRestrictions":
				dnsRestrictions := strings.Split(val, ",")
				didList := []attr.Value{}
				for i := range dnsRestrictions {
					dID, err := strconv.ParseInt(dnsRestrictions[i], 10, 64)
					if err != nil {
						diag.AddError("error parsing dnsRestrictions to int64", err.Error())
						break
					}
					didList = append(didList, types.Int64Value(dID))
					var didSet basetypes.SetValue
					didSet, diag = basetypes.NewSetValue(types.Int64Type, didList)
					if diag.HasError() {
						break
					}
					networkProperties.dnsRestrictions = didSet
				}
			case "allowDuplicateHost":
				networkProperties.allowDuplicateHost = types.StringValue(val)
			case "pingBeforeAssign":
				networkProperties.pingBeforeAssign = types.StringValue(val)
			case "inheritAllowDuplicateHost":
				b, err := strconv.ParseBool(val)
				if err != nil {
					diag.AddError("error parsing inheritAllowDuplicateHost to bool", err.Error())
					break
				}
				networkProperties.inheritAllowDuplicateHost = types.BoolValue(b)
			case "inheritPingBeforeAssign":
				b, err := strconv.ParseBool(val)
				if err != nil {
					diag.AddError("error parsing inheritPingBeforeAssign to bool", err.Error())
					break
				}
				networkProperties.inheritAllowDuplicateHost = types.BoolValue(b)
			case "inheritDNSRestrictions":
				b, err := strconv.ParseBool(val)
				if err != nil {
					diag.AddError("error parsing inheritDNSRestrictions to bool", err.Error())
					break
				}
				networkProperties.inheritDNSRestrictions = types.BoolValue(b)
			case "inheritDefaultDomains":
				b, err := strconv.ParseBool(val)
				if err != nil {
					diag.AddError("error parsing inheritDefaultDomains to bool", err.Error())
					break
				}
				networkProperties.inheritDefaultDomains = types.BoolValue(b)
			case "inheritDefaultView":
				b, err := strconv.ParseBool(val)
				if err != nil {
					diag.AddError("error parsing inheritDefaultView to bool", err.Error())
					break
				}
				networkProperties.inheritDefaultView = types.BoolValue(b)
			case "locationCode":
				networkProperties.locationCode = types.StringValue(val)
			case "locationInherited":
				b, err := strconv.ParseBool(val)
				if err != nil {
					diag.AddError("error parsing locationInherited to bool", err.Error())
					break
				}
				networkProperties.locationInherited = types.BoolValue(b)
			default:
				cpMap[prop] = types.StringValue(val)
			}
		}
	}

	var customProperties basetypes.MapValue
	customProperties, d := basetypes.NewMapValue(types.StringType, cpMap)
	if d.HasError() {
		diag.Append(d...)
	}
	networkProperties.customProperties = customProperties
	return networkProperties, diag
}

func getIP4NetworkAddressUsage(id int64, cidr string, client gobam.ProteusAPI) (int64, int64, error) {

	netmask, err := strconv.ParseFloat(strings.Split(cidr, "/")[1], 64)
	if err != nil {
		mutex.Unlock()
		return 0, 0, fmt.Errorf("error parsing netmask from cidr string")
	}
	addressCount := int(math.Pow(2, (32 - netmask)))

	resp, err := client.GetEntities(id, "IP4Address", 0, addressCount)
	if err != nil {
		return 0, 0, err
	}

	addressesInUse := int64(len(resp.Item))
	addressesFree := int64(addressCount) - addressesInUse

	return addressesInUse, addressesFree, nil
}
