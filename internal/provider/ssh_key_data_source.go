package provider

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-binarylane/internal/resources"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &sshKeyDataSource{}
	_ datasource.DataSourceWithConfigure = &sshKeyDataSource{}
)

func NewSshKeyDataSource() datasource.DataSource {
	return &sshKeyDataSource{}
}

type sshKeyDataSource struct {
	bc *BinarylaneClient
}

func (d *sshKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sshKeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

func (d *sshKeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = *convertResourceSchemaToDataSourceSchema(ctx, resources.SshKeyResourceSchema(ctx))
	resp.Schema.Description = "TODO"
}

func (d *sshKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data resources.SshKeyModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	sshResp, err := d.bc.client.GetAccountKeysKeyIdWithResponse(ctx, int(data.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading SSH Key: name=%s", data.Name.ValueString()),
			err.Error(),
		)
		return
	}
	if sshResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code reading SSH Key",
			fmt.Sprintf("Received %s reading SSH Key: name=%s. Details: %s", sshResp.Status(), data.Name.ValueString(), sshResp.Body),
		)
		return
	}

	// Example data value setting
	data.Id = types.Int64Value(*sshResp.JSON200.SshKey.Id)
	data.Default = types.BoolValue(*sshResp.JSON200.SshKey.Default)
	data.Name = types.StringValue(*sshResp.JSON200.SshKey.Name)
	data.PublicKey = types.StringValue(*sshResp.JSON200.SshKey.PublicKey)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}