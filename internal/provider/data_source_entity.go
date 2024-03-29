package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/umich-vci/gobam"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &entityDataSource{}

func NewEntityDataSource() datasource.DataSource {
	return &entityDataSource{}
}

// EntityDataSource defines the data source implementation.
type entityDataSource struct {
	client *loginClient
}

// ExampleDataSourceModel describes the data source data model.
type EntityDataSourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	ParentID   types.Int64  `tfsdk:"parent_id"`
	Properties types.String `tfsdk:"properties"`
}

func (d *entityDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entity"
}

func (d *entityDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Data source to access the attributes of a BlueCat entity.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Entity identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the entity to find.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the entity you want to retrieve.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(gobam.ObjectTypes...),
				},
			},
			"parent_id": schema.Int64Attribute{
				MarkdownDescription: "The object ID of the parent object that contains the entity. Configurations are stored in ID `0`.",
				Required:            true,
			},
			"properties": schema.StringAttribute{
				MarkdownDescription: "The properties of the entity as returned by the API (pipe delimited).",
				Computed:            true,
			},
		},
	}
}

func (d *entityDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *entityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EntityDataSourceModel

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

	parentID := data.ParentID.ValueInt64()

	name := data.Name.ValueString()
	objType := data.Type.ValueString()

	entity, err := client.GetEntityByName(parentID, name, objType)
	if err != nil {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Failed to get entity by name", err.Error())
		return
	}

	if *entity.Id == 0 {
		resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)
		resp.Diagnostics.AddError("Entity not found", "Entity ID returned was 0")

		return
	}

	data.Id = types.StringValue(strconv.FormatInt(*entity.Id, 10))
	data.Properties = types.StringValue(*entity.Properties)

	resp.Diagnostics.Append(clientLogout(ctx, &client, mutex)...)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
