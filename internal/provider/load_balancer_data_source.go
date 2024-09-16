package provider

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-binarylane/internal/resources"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &loadBalancerDataSource{}
	_ datasource.DataSourceWithConfigure = &loadBalancerDataSource{}
)

func NewLoadBalancerDataSource() datasource.DataSource {
	return &loadBalancerDataSource{}
}

type loadBalancerDataSource struct {
	bc *BinarylaneClient
}

type loadBalancerDataSourceModel struct {
	resources.LoadBalancerModel
}

func (d *loadBalancerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer"
}

func (d *loadBalancerDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	bc, ok := req.ProviderData.(BinarylaneClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *BinarylaneClient, got: %T.", req.ProviderData))
		return
	}

	d.bc = &bc
}

func (d *loadBalancerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	ds, err := convertResourceSchemaToDataSourceSchema(
		resources.LoadBalancerResourceSchema(ctx),
		AttributeConfig{
			RequiredAttributes: &[]string{"id"},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to convert resource schema to data source schema", err.Error())
		return
	}
	resp.Schema = *ds
	// resp.Schema.Description = "TODO"
}

func (d *loadBalancerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data loadBalancerDataSourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	lbResp, err := d.bc.client.GetLoadBalancersLoadBalancerIdWithResponse(ctx, data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading load balancer: name=%s", data.Name.ValueString()),
			err.Error(),
		)
		return
	}

	if lbResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code reading load balancer",
			fmt.Sprintf("Received %s reading load balancer: name=%s. Details: %s", lbResp.Status(), data.Name.ValueString(), lbResp.Body),
		)
		return
	}

	diags := SetLoadBalancerModelState(ctx, &data.LoadBalancerModel, lbResp.JSON200.LoadBalancer)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
