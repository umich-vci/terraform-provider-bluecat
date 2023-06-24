// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/umich-vci/gobam"
)

type loginClient struct {
	Client   gobam.ProteusAPI
	Username string
	Password string
}

// Ensure blueCatProvider satisfies various provider interfaces.
var _ provider.Provider = &blueCatProvider{}

var mutex = &sync.Mutex{}

// blueCatProvider defines the provider implementation.
type blueCatProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// bluecatProviderModel describes the provider data model.
type blueCatProviderModel struct {
	BlueCatEndpoint types.String `tfsdk:"bluecat_endpoint"`
	Username        types.String `tfsdk:"username"`
	Password        types.String `tfsdk:"password"`
	SSLVerify       types.Bool   `tfsdk:"ssl_verify"`
}

func (p *blueCatProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "bluecat"
	resp.Version = p.version
}

func (p *blueCatProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"bluecat_endpoint": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The BlueCat Address Manager endpoint hostname. Can also use the environment variable `BLUECAT_ENDPOINT`",
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "A BlueCat Address Manager username. Can also use the environment variable `BLUECAT_USERNAME`",
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "The BlueCat Address Manager password. Can also use the environment variable `BLUECAT_PASSWORD`",
			},
			"ssl_verify": schema.BoolAttribute{
				Optional:    true,
				Description: "Verify the SSL certificate of the BlueCat Address Manager endpoint?",
			},
		},
	}
}

func (p *blueCatProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config blueCatProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.BlueCatEndpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("bluecat_endpoint"),
			"Unknown BlueCat API Endpoint",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the HashiCups API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the BLUECAT_ENDPOINT environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown BlueCat API Username",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the HashiCups API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the BLUECAT_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown BlueCat API Password",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the HashiCups API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the BLUECAT_PASSWORD environment variable.",
		)
	}

	if config.SSLVerify.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("ssl_verify"),
			"Unknown BlueCat API",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the HashiCups API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the BLUECAT_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	endpoint := os.Getenv("BLUECAT_ENDPOINT")
	username := os.Getenv("BLUECAT_USERNAME")
	password := os.Getenv("BLUECAT_PASSWORD")
	sslVerify := true

	if !config.BlueCatEndpoint.IsNull() {
		endpoint = config.BlueCatEndpoint.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	if !config.SSLVerify.IsNull() {
		sslVerify = config.SSLVerify.ValueBool()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("bluecat_endpoint"),
			"Missing BlueCat API Endpoint",
			"The provider cannot create the HashiCups API client as there is a missing or empty value for the HashiCups API host. "+
				"Set the host value in the configuration or use the BLUECAT_ENDPOINT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing BlueCat API Username",
			"The provider cannot create the HashiCups API client as there is a missing or empty value for the HashiCups API username. "+
				"Set the username value in the configuration or use the BLUECAT_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing BlueCat API Password",
			"The provider cannot create the HashiCups API client as there is a missing or empty value for the HashiCups API password. "+
				"Set the password value in the configuration or use the BLUECAT_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client := gobam.NewClient(endpoint, sslVerify)
	loginClient := &loginClient{Client: client, Username: username, Password: password}
	// err := client.Login(username, password)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Unable to Create BlueCat API Client",
	// 		"An error occurred when creating the BlueCat API client. "+
	// 			"If the error is not clear, please contact the provider developers.\n\n"+
	// 			"BlueCat Client Error: "+err.Error(),
	// 	)
	// 	return
	// }

	// Make the BlueCat client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = loginClient
	resp.ResourceData = loginClient
}

func (p *blueCatProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// NewExampleResource,
	}
}

func (p *blueCatProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewEntityDataSource,
		NewHostRecordDataSource,
		NewIP4AddressDataSource,
		NewIP4NBRDataSource,
		NewIP4NetworkDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &blueCatProvider{
			version: version,
		}
	}
}
