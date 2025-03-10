package provider

import (
	"context"
	"terraform-provider-binarylane/internal/binarylane"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
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
	client *binarylane.ClientWithResponses
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
				MarkdownDescription: "Binary Lane API endpoint. Defaults to `https://api.binarylane.com.au/v2`, but can be " +
					"overridden by setting this attribute or the `BINARYLANE_API_ENDPOINT` environment variable.",
				Optional: true,
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: "Binary Lane API token. If not defined, will default to `BINARYLANE_API_TOKEN` environment variable.",
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

	client, err := binarylane.NewClientWithConfig(
		config.Endpoint.ValueString(),
		config.Token.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Binary Lane API client", err.Error())
		return
	}

	binarylaneClient := BinarylaneClient{
		client: client,
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
		NewServerFirewallRulesDataSource,
		NewSshKeyDataSource,
		NewVpcDataSource,
		NewVpcRouteEntriesDataSource,
		NewLoadBalancerDataSource,
		NewImagesDataSource,
		NewRegionsDataSource,
		NewSizesDataSource,
	}
}

func (p *binarylaneProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewServerResource,
		NewServerFirewallRulesResource,
		NewSshKeyResource,
		NewVpcResource,
		NewVpcRouteEntriesResource,
		NewLoadBalancerResource,
	}
}
