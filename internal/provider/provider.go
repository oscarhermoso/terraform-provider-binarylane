package provider

import (
	"context"
	"os"
	"terraform-provider-binarylane/internal/binarylane"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = (*binarylaneProvider)(nil)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &binarylaneProvider{
			version: version,
		}
	}
}

type BinarylaneClient struct {
	endpoint string
	client   *binarylane.ClientWithResponses
}

type binarylaneProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type binarylaneProviderModel struct {
	Endpoint types.String `tfsdk:"api_endpoint"`
	Token    types.String `tfsdk:"api_token"`
}

func (p *binarylaneProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_endpoint": schema.StringAttribute{
				MarkdownDescription: "Binary Lane API endpoint. If not set checks env for `BINARYLANE_API_ENDPOINT`. " +
					"Default: `https://api.binarylane.com.au/v2`.",
				Optional: true,
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: "Binary Lane API token. If not set checks env for `BINARYLANE_API_TOKEN`.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *binarylaneProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config binarylaneProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_endpoint"),
			"Unknown Binary Lane API endpoint",
			"The provider cannot create the Binary Lane API client as there is an unknown configuration value for the "+
				"Binary Lane API endpoint. Either target apply the source of the value first, set the value statically in "+
				"the configuration, or use the BINARYLANE_API_ENDPOINT environment variable.",
		)
	}
	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Unknown Binary Lane token",
			"The provider cannot create the Binary Lane API client as there is an unknown configuration value for the "+
				"Binary Lane API token. Either target apply the source of the value first, set the value statically in the "+
				"configuration, or use the BINARYLANE_API_TOKEN environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	endpoint := os.Getenv("BINARYLANE_API_ENDPOINT")
	token := os.Getenv("BINARYLANE_API_TOKEN")
	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}
	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}
	if endpoint == "" {
		endpoint = "https://api.binarylane.com.au/v2"
	}
	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Missing Binary Lane API token",
			"The provider cannot create the Binary Lane API client as there is a missing or empty value for the Binary "+
				"Lane API token. Set the token value in the configuration or use the BINARYLANE_API_TOKEN environment "+
				"variable. If either is already set, ensure the value is not empty.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	auth, err := securityprovider.NewSecurityProviderBearerToken(token)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create security provider with supplied token", err.Error())
		return
	}

	client, err := binarylane.NewClientWithResponses(endpoint, binarylane.WithRequestEditorFn(auth.Intercept))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Binary Lane API client", err.Error())
		return
	}

	binarylaneClient := BinarylaneClient{
		endpoint: endpoint,
		client:   client,
	}

	resp.DataSourceData = binarylaneClient
	resp.ResourceData = binarylaneClient
}

func (p *binarylaneProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "binarylane"
	resp.Version = p.version
}

func (p *binarylaneProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewServerDataSource,
		NewSshKeyDataSource,
		NewVpcDataSource,
	}
}

func (p *binarylaneProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewServerResource,
		NewSshKeyResource,
		NewVpcResource,
		NewVpcRouteEntriesResource,
	}
}
