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
	_ datasource.DataSource              = &imagesDataSource{}
	_ datasource.DataSourceWithConfigure = &imagesDataSource{}
)

func NewImagesDataSource() datasource.DataSource {
	return &imagesDataSource{}
}

type imagesDataSource struct {
	bc *BinarylaneClient
}

type imagesDataSourceModel struct {
	data_sources.ImagesModel
}

func (d *imagesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_images"
}

func (d *imagesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *imagesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = data_sources.ImagesDataSourceSchema(ctx)
}

func (d *imagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data imagesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	var page int32 = 1
	perPage := int32(200)
	var nextPage bool = true
	var imgResults []binarylane.Image

	for nextPage {
		params := binarylane.GetImagesParams{
			Page:    &page,
			PerPage: &perPage,
			Type:    data.Type.ValueStringPointer(),
		}
		listResp, err := d.bc.client.GetImagesWithResponse(ctx, &params)
		if err != nil {
			resp.Diagnostics.AddError("Error listing images", err.Error())
			return
		}
		if listResp.StatusCode() != http.StatusOK {
			resp.Diagnostics.AddError("Unexpected status code listing images", string(listResp.Body))
			return
		}
		imgResults = append(imgResults, *listResp.JSON200.Images...)

		if listResp.JSON200.Links == nil || listResp.JSON200.Links.Pages == nil || listResp.JSON200.Links.Pages.Next == nil {
			nextPage = false
			break
		}
		page++
	}

	images, diags := types.ListValueFrom(ctx, data.Images.ElementType(ctx), imgResults)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Images = images

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
