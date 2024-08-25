package provider

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-binarylane/internal/resources"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &sshKeyDataSource{}
	_ datasource.DataSourceWithConfigure = &sshKeyDataSource{}
)

func NewVpcDataSource() datasource.DataSource {
	return &vpcDataSource{}
}

type vpcDataSource struct {
	bc *BinarylaneClient
}

type vpcDataSourceModel struct {
	resources.VpcModel
}

func (d *vpcDataSource) Configure(
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

func (d *vpcDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}

func (d *vpcDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	ds, err := convertResourceSchemaToDataSourceSchema(resources.VpcResourceSchema(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to convert resource schema to data source schema", err.Error())
		return
	}

	resp.Schema = *ds
	resp.Schema.Description = "TODO"
}

func (d *vpcDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data vpcDataSourceModel

	// Read Terraform configuration data into the model
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	vpcResp, err := d.bc.client.GetVpcsVpcIdWithResponse(ctx, data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading VPC: id=%d", data.Id.ValueInt64()),
			err.Error(),
		)
		return
	}
	if vpcResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code reading VPC",
			fmt.Sprintf("Received %s reading VPC: id=%d. Details: %s", vpcResp.Status(), data.Id.ValueInt64(), vpcResp.Body))
		return
	}
	data.IpRange = types.StringValue(*vpcResp.JSON200.Vpc.IpRange)
	data.Name = types.StringValue(*vpcResp.JSON200.Vpc.Name)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
