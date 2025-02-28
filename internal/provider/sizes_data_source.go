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
	_ datasource.DataSource              = &sizesDataSource{}
	_ datasource.DataSourceWithConfigure = &sizesDataSource{}
)

func NewSizesDataSource() datasource.DataSource {
	return &sizesDataSource{}
}

type sizesDataSource struct {
	bc *BinarylaneClient
}

type sizesDataSourceModel struct {
	data_sources.SizesModel
}

func (d *sizesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sizes"
}

func (d *sizesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sizesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = data_sources.SizesDataSourceSchema(ctx)
}

func (d *sizesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data sizesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	var page int32 = 1
	perPage := int32(200)
	var nextPage bool = true
	var regResults []binarylane.Size

	for nextPage {
		params := binarylane.GetSizesParams{
			Page:    &page,
			PerPage: &perPage,
		}
		listResp, err := d.bc.client.GetSizesWithResponse(ctx, &params)
		if err != nil {
			resp.Diagnostics.AddError("Error listing sizes", err.Error())
			return
		}
		if listResp.StatusCode() != http.StatusOK {
			resp.Diagnostics.AddError("Unexpected status code listing sizes", string(listResp.Body))
			return
		}
		regResults = append(regResults, listResp.JSON200.Sizes...)

		if listResp.JSON200.Links == nil || listResp.JSON200.Links.Pages == nil || listResp.JSON200.Links.Pages.Next == nil {
			nextPage = false
			break
		}
		page++
	}

	sizes, diags := types.ListValueFrom(ctx, data.Sizes.ElementType(ctx), regResults)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Sizes = sizes

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
