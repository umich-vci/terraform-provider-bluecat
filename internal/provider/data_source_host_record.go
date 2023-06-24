// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/umich-vci/gobam"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &HostRecordDataSource{}

func NewHostRecordDataSource() datasource.DataSource {
	return &HostRecordDataSource{}
}

// HostRecordDataSource defines the data source implementation.
type HostRecordDataSource struct {
	client *loginClient
}

// HostRecordDataSourceModel describes the data source data model.
type HostRecordDataSourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	AbsoluteName     types.String `tfsdk:"absolute_name"`
	Addresses        types.Set    `tfsdk:"addresses"`
	AddressIDs       types.Set    `tfsdk:"address_ids"`
	CustomProperties types.Map    `tfsdk:"custom_properties"`
	Name             types.String `tfsdk:"name"`
	ParentID         types.Int64  `tfsdk:"parent_id"`
	ParentType       types.String `tfsdk:"parent_type"`
	Properties       types.String `tfsdk:"properties"`
	ReverseRecord    types.Bool   `tfsdk:"reverse_record"`
	TTL              types.Int64  `tfsdk:"ttl"`
	Type             types.String `tfsdk:"type"`
}

func (d *HostRecordDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_example"
}

func (d *HostRecordDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Example data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Entity identifier",
				Computed:            true,
			},
			"absolute_name": schema.StringAttribute{
				MarkdownDescription: "The absolute name/fqdn of the host record.",
				Required:            true,
			},
			"addresses": schema.SetAttribute{
				MarkdownDescription: "A set of all addresses associated with the host record.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"address_ids": schema.SetAttribute{
				MarkdownDescription: "A set of all address ids associated with the host record.",
				Computed:            true,
				ElementType:         types.Int64Type,
			},
			"custom_properties": schema.MapAttribute{
				MarkdownDescription: "A map of all custom properties associated with the host record.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The short name of the host record.",
				Computed:            true,
			},
			"parent_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the parent of the host record.",
				Computed:            true,
			},
			"parent_type": schema.StringAttribute{
				MarkdownDescription: "The type of the parent of the host record.",
				Computed:            true,
			},
			"properties": schema.StringAttribute{
				MarkdownDescription: "The properties of the host record as returned by the API (pipe delimited).",
				Computed:            true,
			},
			"reverse_record": schema.BoolAttribute{
				MarkdownDescription: "A boolean that represents if the host record should set reverse records.",
				Computed:            true,
			},
			"ttl": schema.Int64Attribute{
				MarkdownDescription: "The TTL of the host record.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the resource.",
				Computed:            true,
			},
		},
	}
}

func (d *HostRecordDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*loginClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *loginClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *HostRecordDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HostRecordDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	mutex.Lock()
	client := d.client.Client
	client.Login(d.client.Username, d.client.Password)

	start := 0
	count := 10
	absoluteName := data.AbsoluteName.String()
	options := "hint=^" + absoluteName + "$|"

	hostRecords, err := client.GetHostRecordsByHint(start, count, options)
	if err = gobam.LogoutClientIfError(client, err, "Failed to get Host Records by hint"); err != nil {
		mutex.Unlock()
		resp.Diagnostics.AddError(
			"Failed to get Host Records by hint",
			fmt.Sprintf("Failed to get Host Records by hint: %s", err.Error()),
		)
		return
	}

	log.Printf("[INFO] GetHostRecordsByHint returned %s results", strconv.Itoa(len(hostRecords.Item)))

	matches := 0
	matchLocation := -1
	for x := range hostRecords.Item {
		properties := *hostRecords.Item[x].Properties
		props := strings.Split(properties, "|")
		for y := range props {
			if len(props[y]) > 0 {
				prop := strings.Split(props[y], "=")[0]
				val := strings.Split(props[y], "=")[1]
				if prop == "absoluteName" && val == absoluteName {
					matches++
					matchLocation = x
				}
			}
		}
	}

	if matches == 0 || matches > 1 {
		err := fmt.Errorf("no exact host record match found for: %s", absoluteName)
		if err = gobam.LogoutClientIfError(client, err, "No exact host record match found for hint"); err != nil {
			mutex.Unlock()
			resp.Diagnostics.AddError(
				"No exact host record match found for hint",
				fmt.Sprintf("No exact host record match found for hint: %s", err.Error()),
			)
			return
		}
	}

	data.ID = types.Int64Value(*hostRecords.Item[matchLocation].Id)
	data.Name = types.StringValue(*hostRecords.Item[matchLocation].Name)
	data.Properties = types.StringValue(*hostRecords.Item[matchLocation].Properties)
	data.Type = types.StringValue(*hostRecords.Item[matchLocation].Type)

	hostRecordProperties, err := parseHostRecordProperties(*hostRecords.Item[matchLocation].Properties)
	if err != nil {
		gobam.LogoutClientWithError(client, "Error parsing host record properties")
		mutex.Unlock()

		resp.Diagnostics.AddError(
			"Error parsing the host record properties",
			err.Error(),
		)
	}

	data.AbsoluteName = hostRecordProperties.absoluteName
	data.ParentID = hostRecordProperties.parentID
	data.ParentType = hostRecordProperties.parentType
	data.ReverseRecord = hostRecordProperties.reverseRecord
	data.Addresses = hostRecordProperties.addresses
	data.AddressIDs = hostRecordProperties.addressIDs
	data.CustomProperties = hostRecordProperties.customProperties
	data.TTL = hostRecordProperties.ttl

	if err := client.Logout(); err != nil {
		mutex.Unlock()
		resp.Diagnostics.AddError(
			"Failed logout client",
			fmt.Sprintf("Unexpected error logging out client: %s", err.Error()),
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

type hostRecordProperties struct {
	absoluteName     types.String
	parentID         types.Int64
	parentType       types.String
	ttl              types.Int64
	reverseRecord    types.Bool
	addresses        types.Set
	addressIDs       types.Set
	customProperties types.Map
}

func parseHostRecordProperties(properties string) (hostRecordProperties, error) {
	var hrProperties hostRecordProperties

	cpMap := make(map[string]attr.Value)

	// if ttl isn't returned as a property it will remain set at -1
	hrProperties.ttl = types.Int64Value(-1)

	props := strings.Split(properties, "|")
	for x := range props {
		if len(props[x]) > 0 {
			prop := strings.Split(props[x], "=")[0]
			val := strings.Split(props[x], "=")[1]

			switch prop {
			case "absoluteName":
				hrProperties.absoluteName = types.StringValue(val)
			case "parentId":
				pID, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					return hrProperties, fmt.Errorf("error parsing parentId to int64")
				}
				hrProperties.parentID = types.Int64Value(pID)
			case "parentType":
				hrProperties.parentType = types.StringValue(val)
			case "reverseRecord":
				b, err := strconv.ParseBool(val)
				if err != nil {
					return hrProperties, fmt.Errorf("error parsing reverseRecord to bool")
				}
				hrProperties.reverseRecord = types.BoolValue(b)
			case "addresses":
				addresses := strings.Split(val, ",")
				addressList := []attr.Value{}
				for i := range addresses {
					addressList = append(addressList, types.StringValue(addresses[i]))
				}
				addressSet, diag := types.SetValue(types.StringType, addressList)
				if diag.HasError() {
					return hrProperties, fmt.Errorf("error creating address set")
				}
				hrProperties.addresses = addressSet
			case "addressIds":
				addressIDs := strings.Split(val, ",")
				aidList := []attr.Value{}
				for i := range addressIDs {
					aID, err := strconv.ParseInt(addressIDs[i], 10, 64)
					if err != nil {
						return hrProperties, fmt.Errorf("error parsing addressIds to int64")
					}
					aidList = append(aidList, types.Int64Value(aID))
				}
				aidSet, diag := types.SetValue(types.StringType, aidList)
				if diag.HasError() {
					return hrProperties, fmt.Errorf("error creating address id set")
				}
				hrProperties.addressIDs = aidSet
			case "ttl":
				ttlval, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					return hrProperties, fmt.Errorf("error parsing ttl to int")
				}
				hrProperties.ttl = types.Int64Value(ttlval)
			default:
				cpMap[prop] = types.StringValue(val)
			}
		}
	}

	customProperties, diag := types.MapValue(types.StringType, cpMap)
	if diag.HasError() {
		return hrProperties, fmt.Errorf("error creating custom properties map")
	}
	hrProperties.customProperties = customProperties

	return hrProperties, nil
}
