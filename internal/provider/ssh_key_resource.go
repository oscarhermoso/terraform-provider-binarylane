package provider

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-binarylane/internal/binarylane"
	"terraform-provider-binarylane/internal/resources"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &sshKeyResource{}
	_ resource.ResourceWithConfigure = &sshKeyResource{}
	// _ resource.ResourceWithImportState = &serverResource{}
)

func NewSshKeyResource() resource.Resource {
	return &sshKeyResource{}
}

type sshKeyResource struct {
	bc *BinarylaneClient
}

func (d *sshKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *sshKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

func (r *sshKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resources.SshKeyResourceSchema(ctx)
	resp.Schema.Description = "TODO"

	// Overrides
	defaultAttr := resp.Schema.Attributes["default"]
	resp.Schema.Attributes["default"] = schema.BoolAttribute{
		Optional:            true,
		Computed:            true,
		Description:         defaultAttr.GetDescription(),
		MarkdownDescription: defaultAttr.GetMarkdownDescription(),
		Default:             booldefault.StaticBool(false), // Add default to backups
	}
}

func (r *sshKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resources.SshKeyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	sshResp, err := r.bc.client.PostAccountKeysWithResponse(ctx, binarylane.SshKeyRequest{
		Name:      data.Name.ValueString(),
		Default:   data.Default.ValueBoolPointer(),
		PublicKey: data.PublicKey.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error creating SSH Key: name=%s", data.Name.ValueString()),
			err.Error(),
		)
		return
	}
	if sshResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code creating server",
			fmt.Sprintf("Received %s creating new server: name=%s. Details: %s", sshResp.Status(), data.Name.ValueString(), sshResp.Body),
		)
		return
	}

	// Set data values
	data.Id = types.Int64Value(*sshResp.JSON200.SshKey.Id)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sshKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resources.SshKeyModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	sshResp, err := r.bc.client.GetAccountKeysKeyIdWithResponse(ctx, int(data.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error creating SSH Key: name=%s", data.Name.ValueString()),
			err.Error(),
		)
		return
	}
	if sshResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code creating server",
			fmt.Sprintf("Received %s creating new server: name=%s. Details: %s", sshResp.Status(), data.Name.ValueString(), sshResp.Body),
		)
		return
	}

	// Set data values
	data.Id = types.Int64Value(*sshResp.JSON200.SshKey.Id)
	data.Default = types.BoolValue(*sshResp.JSON200.SshKey.Default)
	data.Name = types.StringValue(*sshResp.JSON200.SshKey.Name)
	data.PublicKey = types.StringValue(*sshResp.JSON200.SshKey.PublicKey)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sshKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resources.SshKeyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	// TODO

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sshKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resources.SshKeyModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete API call logic
	sshResp, err := r.bc.client.DeleteAccountKeysKeyIdWithResponse(ctx, int(data.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error deleting SSH Key: name=%s", data.Name.ValueString()),
			err.Error(),
		)
		return
	}
	if sshResp.StatusCode() != http.StatusNoContent {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code deleting SSH Key",
			fmt.Sprintf("Received %s deleting SSH Key: name=%s. Details: %s", sshResp.Status(), data.Name.ValueString(), sshResp.Body),
		)
		return
	}
}
