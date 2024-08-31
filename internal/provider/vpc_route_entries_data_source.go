package provider

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-binarylane/internal/resources"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
)

var (
	_ datasource.DataSource              = &vpcRouteEntriesDataSource{}
	_ datasource.DataSourceWithConfigure = &vpcRouteEntriesDataSource{}
)

func NewVpcRouteEntriesDataSource() datasource.DataSource {
	return &vpcRouteEntriesDataSource{}
}

type vpcRouteEntriesDataSource struct {
	bc *BinarylaneClient
}

type vpcRouteEntriesDataSourceModel struct {
	resources.VpcRouteEntriesModel
}

func (d *vpcRouteEntriesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_route_entries"
}

func (d *vpcRouteEntriesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	ds, err := convertResourceSchemaToDataSourceSchema(resources.VpcRouteEntriesResourceSchema(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to convert resource schema to data source schema", err.Error())
		return
	}
	resp.Schema = *ds
	resp.Schema.Description = "TODO"

	// Overrides
	vpcId := resp.Schema.Attributes["vpc_id"]
	resp.Schema.Attributes["vpc_id"] = &schema.Int64Attribute{
		Description:         vpcId.GetDescription(),
		MarkdownDescription: vpcId.GetMarkdownDescription(),
		Required:            true, // VPC ID is required to define the route entries
	}
}

func (d *vpcRouteEntriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data vpcRouteEntriesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	vpcResp, err := d.bc.client.GetVpcsVpcIdWithResponse(ctx, data.VpcId.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading VPC: vpc_id=%d", data.VpcId.ValueInt64()),
			err.Error(),
		)
		return
	}
	if vpcResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code reading VPC",
			fmt.Sprintf("Received %s reading VPC: vpc_id=%d. Details: %s", vpcResp.Status(), data.VpcId.ValueInt64(), vpcResp.Body))
		return
	}

	routeEntries, routeEntriesDiags := GetRouteEntriesState(ctx, vpcResp.JSON200.Vpc.RouteEntries)
	resp.Diagnostics.Append(routeEntriesDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.RouteEntries = routeEntries

	// Example data value setting

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *vpcRouteEntriesDataSource) Configure(
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