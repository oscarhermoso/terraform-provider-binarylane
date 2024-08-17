package provider

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-binarylane/internal/resources"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	d_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &serverDataSource{}
	_ datasource.DataSourceWithConfigure = &serverDataSource{}
)

func NewServerDataSource() datasource.DataSource {
	return &serverDataSource{}
}

type serverDataSource struct {
	bc *BinarylaneClient
}

func (d *serverDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	bc, ok := req.ProviderData.(BinarylaneClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *BinarylaneClient, got: %T.", req.ProviderData),
		)

		return
	}

	d.bc = &bc
}

func (d *serverDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (d *serverDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = *convertResourceSchemaToDataSourceSchema(ctx, resources.ServerResourceSchema(ctx))
	resp.Schema.Description = "TODO"

	// Overrides
	id := resp.Schema.Attributes["id"]
	resp.Schema.Attributes["id"] = d_schema.Int64Attribute{
		Description:         id.GetDescription(),
		MarkdownDescription: id.GetMarkdownDescription(),
		Required:            true, // ID is required to find the server
	}
}

func (d *serverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data resources.ServerModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: Use go routines here to make requests concurrently

	// Read server
	serverResp, err := d.bc.client.GetServersServerIdWithResponse(ctx, data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading server: name=%s", data.Id.String()),
			err.Error(),
		)
		return
	}
	if serverResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unexpected HTTP status code reading server: name=%s", data.Id.String()),
			fmt.Sprintf("Unexpected HTTP status code reading server: %s", serverResp.Body),
		)
		return
	}

	// TODO: Fetch user data

	// Read user data
	// userDataResp, err := d.bc.client.GetServersServerIdUserDataWithResponse(ctx, data.Id.ValueInt64())
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		fmt.Sprintf("Error reading server user data: name=%s", data.Id.String()),
	// 		err.Error(),
	// 	)
	// 	return
	// }
	// if userDataResp.StatusCode() != http.StatusOK {
	// 	resp.Diagnostics.AddError(
	// 		fmt.Sprintf("Unexpected HTTP status code reading server user data: name=%s", data.Id.String()),
	// 		fmt.Sprintf("Unexpected HTTP status code reading server user data: %s", serverResp.Body),
	// 	)
	// 	return
	// }

	// Set data values
	data.Id = types.Int64Value(*serverResp.JSON200.Server.Id)
	data.Name = types.StringValue(*serverResp.JSON200.Server.Name)
	data.Image = types.StringValue(*serverResp.JSON200.Server.Image.Slug)
	data.Region = types.StringValue(*serverResp.JSON200.Server.Region.Slug)
	data.Size = types.StringValue(*serverResp.JSON200.Server.Size.Slug)
	// data.UserData = types.StringValue(*userDataResp.JSON200.UserData)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
