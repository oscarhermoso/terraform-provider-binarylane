package provider

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-binarylane/internal/binarylane"
	"terraform-provider-binarylane/internal/data_sources"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &regionsDataSource{}
	_ datasource.DataSourceWithConfigure = &regionsDataSource{}
)

func NewRegionsDataSource() datasource.DataSource {
	return &regionsDataSource{}
}

type regionsDataSource struct {
	bc *BinarylaneClient
}

type regionsDataSourceModel struct {
	data_sources.RegionsModel
}

func (d *regionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_regions"
}

func (d *regionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *regionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = data_sources.RegionsDataSourceSchema(ctx)
}

func (d *regionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data regionsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	var page int32 = 1
	perPage := int32(200)
	var nextPage bool = true
	var regResults []binarylane.Region

	for nextPage {
		params := binarylane.GetRegionsParams{
			Page:    &page,
			PerPage: &perPage,
		}
		listResp, err := d.bc.client.GetRegionsWithResponse(ctx, &params)
		if err != nil {
			resp.Diagnostics.AddError("Error listing regions", err.Error())
			return
		}
		if listResp.StatusCode() != http.StatusOK {
			resp.Diagnostics.AddError("Unexpected status code listing regions", string(listResp.Body))
			return
		}
		regResults = append(regResults, listResp.JSON200.Regions...)

		if listResp.JSON200.Links == nil || listResp.JSON200.Links.Pages == nil || listResp.JSON200.Links.Pages.Next == nil {
			nextPage = false
			break
		}
		page++
	}

	regions, diags := types.ListValueFrom(ctx, data.Regions.ElementType(ctx), regResults)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Regions = regions

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
