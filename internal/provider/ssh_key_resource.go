package provider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"terraform-provider-binarylane/internal/binarylane"
	"terraform-provider-binarylane/internal/resources"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &sshKeyResource{}
	_ resource.ResourceWithConfigure   = &sshKeyResource{}
	_ resource.ResourceWithImportState = &sshKeyResource{}
)

func NewSshKeyResource() resource.Resource {
	return &sshKeyResource{}
}

type sshKeyResource struct {
	bc *BinarylaneClient
}

type sshKeyModel struct {
	resources.SshKeyModel
	Fingerprint types.String `tfsdk:"fingerprint"`
}

func (d *sshKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	bc, ok := req.ProviderData.(BinarylaneClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
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
	// resp.Schema.Description = "TODO"

	// Overrides
	default_ := resp.Schema.Attributes["default"]
	resp.Schema.Attributes["default"] = schema.BoolAttribute{
		Optional:            true,
		Computed:            true,
		Description:         default_.GetDescription(),
		MarkdownDescription: default_.GetMarkdownDescription(),
		Default:             booldefault.StaticBool(false), // Add default to backups
	}

	public_key := resp.Schema.Attributes["public_key"]
	resp.Schema.Attributes["public_key"] = schema.StringAttribute{
		Required:            true,
		Description:         public_key.GetDescription(),
		MarkdownDescription: public_key.GetMarkdownDescription(),
		PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
	}

	// Additional attributes
	fingerprintDescription := "The fingerprint of the SSH key."
	resp.Schema.Attributes["fingerprint"] = &schema.StringAttribute{
		Description:         fingerprintDescription,
		MarkdownDescription: fingerprintDescription,
		Optional:            false,
		Required:            false,
		Computed:            true,
	}
}

func (r *sshKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data sshKeyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	const maxRetries = 3
	var sshResp *binarylane.PostAccountKeysResponse

retryLoop:
	for i := 0; i < maxRetries; i++ {
		var err error
		sshResp, err = r.bc.client.PostAccountKeysWithResponse(ctx, binarylane.SshKeyRequest{
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

		switch sshResp.StatusCode() {

		case http.StatusOK:
			break retryLoop

		case http.StatusInternalServerError:
			if i < maxRetries-1 {
				tflog.Warn(ctx, "Received 500 creating SSH key, retrying...")
				time.Sleep(time.Second * 5)
				continue
			}

		default:
			resp.Diagnostics.AddError(
				"Unexpected HTTP status code creating SSH key",
				fmt.Sprintf("Received %s creating SSH key: name=%s. Details: %s", sshResp.Status(), data.Name.ValueString(), sshResp.Body),
			)
			return
		}
	}

	// Check if retries exceeded
	if sshResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Failed to create SSH key after retries",
			fmt.Sprintf("Final status code: %d", sshResp.StatusCode()),
		)
		return
	}

	// Set data values
	data.Id = types.Int64Value(*sshResp.JSON200.SshKey.Id)
	data.Fingerprint = types.StringValue(*sshResp.JSON200.SshKey.Fingerprint)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sshKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data sshKeyModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	data.PublicKey = types.StringValue(data.PublicKey.ValueString())

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
			"Unexpected HTTP status code getting SSH key",
			fmt.Sprintf("Received %s getting SSH key: name=%s. Details: %s", sshResp.Status(), data.Name.ValueString(), sshResp.Body),
		)
		return
	}

	// Set data values
	data.Id = types.Int64Value(*sshResp.JSON200.SshKey.Id)
	data.Default = types.BoolValue(*sshResp.JSON200.SshKey.Default)
	data.Name = types.StringValue(*sshResp.JSON200.SshKey.Name)
	data.PublicKey = types.StringValue(*sshResp.JSON200.SshKey.PublicKey)
	data.Fingerprint = types.StringValue(*sshResp.JSON200.SshKey.Fingerprint)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sshKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data sshKeyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	sshResp, err := r.bc.client.PutAccountKeysKeyIdWithResponse(ctx, int(data.Id.ValueInt64()), binarylane.UpdateSshKeyRequest{
		Name:    data.Name.ValueString(),
		Default: data.Default.ValueBoolPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updating SSH Key: name=%s", data.Name.ValueString()),
			err.Error(),
		)
		return
	}
	if sshResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code updating SSH Key",
			fmt.Sprintf("Received %s updating new SSH Key: name=%s. Details: %s", sshResp.Status(), data.Name.ValueString(), sshResp.Body),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sshKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data sshKeyModel

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

func (r *sshKeyResource) ImportState(
	ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	// Import by ID
	id, err := strconv.ParseInt(req.ID, 10, 32)
	if err == nil {
		diags := resp.State.SetAttribute(ctx, path.Root("id"), int32(id))
		resp.Diagnostics.Append(diags...)
		return
	}

	// Import by name
	var page int32 = 1
	perPage := int32(200)
	var sshKey binarylane.SshKey
	var nextPage bool = true

	for nextPage { // Need to paginate because the API does not support filtering by fingerprint
		params := binarylane.GetAccountKeysParams{
			Page:    &page,
			PerPage: &perPage,
		}

		sshResp, err := r.bc.client.GetAccountKeysWithResponse(ctx, &params)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Error getting SSH key for import: fingerprint=%s", req.ID), err.Error())
			return
		}

		if sshResp.StatusCode() != http.StatusOK {
			resp.Diagnostics.AddError(
				"Unexpected HTTP status code getting SSH key for import",
				fmt.Sprintf("Received %s getting SSH key for import: fingerprint=%s. Details: %s", sshResp.Status(), req.ID,
					sshResp.Body))
			return
		}

		sshKeys := sshResp.JSON200.SshKeys
		for _, key := range sshKeys {
			if *key.Fingerprint == req.ID {
				sshKey = key
				nextPage = false
				break
			}
		}
		if sshResp.JSON200.Links == nil || sshResp.JSON200.Links.Pages == nil || sshResp.JSON200.Links.Pages.Next == nil {
			nextPage = false
			break
		}

		page++
	}

	if sshKey.Id == nil {
		resp.Diagnostics.AddError(
			"Could not find SSH key by fingerprint",
			fmt.Sprintf("Error finding SSH key: fingerprint=%s", req.ID),
		)
		return
	}

	diags := resp.State.SetAttribute(ctx, path.Root("id"), *sshKey.Id)
	resp.Diagnostics.Append(diags...)
}
