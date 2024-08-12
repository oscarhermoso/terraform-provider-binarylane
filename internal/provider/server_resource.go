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
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &serverResource{}
	_ resource.ResourceWithConfigure = &serverResource{}
	// _ resource.ResourceWithImportState = &serverResource{}
)

// Helper function to simplify the provider implementation.
func NewServerResource() resource.Resource {
	return &serverResource{}
}

type serverResource struct {
	bc *BinarylaneClient
}

func (d *serverResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serverResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *serverResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resources.ServerResourceSchema(ctx)
	resp.Schema.Description = "Provides a Binary Lane Server resource. This can be used to create and delete servers."

	// Overrides
	pw := resp.Schema.Attributes["password"]
	resp.Schema.Attributes["password"] = &schema.StringAttribute{
		Description:         pw.GetDescription(),
		MarkdownDescription: pw.GetMarkdownDescription(),
		Optional:            true,
		Computed:            true,
		Sensitive:           true, // Add sensitive to password
	}

	backups := resp.Schema.Attributes["backups"]
	resp.Schema.Attributes["backups"] = &schema.BoolAttribute{
		Description:         backups.GetDescription(),
		MarkdownDescription: backups.GetMarkdownDescription(),
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false), // Add default to backups
	}
}

func (r *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resources.ServerModel

	// Read Terraform plan data into the model
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	tflog.Debug(ctx, fmt.Sprintf("Creating server: name=%s", plan.Name.ValueString()))

	body := binarylane.CreateServerRequest{
		Name:   plan.Name.ValueStringPointer(),
		Image:  plan.Image.ValueString(),
		Region: plan.Region.ValueString(),
		Size:   plan.Size.ValueString(),
	}

	if plan.Password.IsNull() {
		plan.Password = types.StringUnknown()
	} else {
		body.Password = plan.Password.ValueStringPointer()
	}

	serverResp, err := r.bc.client.PostServersWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error creating server: name=%s", plan.Name.ValueString()),
			err.Error(),
		)
		return
	}

	if serverResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code creating server",
			fmt.Sprintf("Received %s creating new server: name=%s. Details: %s", serverResp.Status(), plan.Name.ValueString(), serverResp.Body))
		return
	}

	assignInt64(serverResp.JSON200.Server.Id, &plan.Id)
	assignStr(serverResp.JSON200.Server.Name, &plan.Name)
	assignStr(serverResp.JSON200.Server.Image.Slug, &plan.Image)
	assignStr(serverResp.JSON200.Server.Region.Slug, &plan.Region)
	assignStr(serverResp.JSON200.Server.Size.Slug, &plan.Size)
	plan.Backups = types.BoolValue(serverResp.JSON200.Server.NextBackupWindow != nil)
	// assignBool(&serverResp.JSON200.Server.Networks.PortBlocking, &plan.PortBlocking)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resources.ServerModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	tflog.Debug(ctx, fmt.Sprintf("Reading server: name=%s", data.Id.String()))

	serverResp, err := r.bc.client.GetServersServerIdWithResponse(ctx, data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading server: name=%s", data.Id.String()),
			err.Error(),
		)
		return
	}

	if serverResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code creating server",
			fmt.Sprintf("Received %s reading server: name=%s. Details: %s", serverResp.Status(), data.Id.String(), serverResp.Body))
		return
	}

	assignInt64(serverResp.JSON200.Server.Id, &data.Id)
	assignStr(serverResp.JSON200.Server.Name, &data.Name)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resources.ServerModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resources.ServerModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete API call logic
	tflog.Debug(ctx, fmt.Sprintf("Deleting server: name=%s", data.Id.String()))

	reason := "Terraform deletion"
	params := binarylane.DeleteServersServerIdParams{
		Reason: &reason,
	}

	delResp, err := r.bc.client.DeleteServersServerIdWithResponse(ctx, data.Id.ValueInt64(), &params)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error deleting server: name=%s, server_id=%s", data.Name.ValueString(), data.Id.String()),
			err.Error(),
		)
		return
	}

	if delResp.StatusCode() != http.StatusNoContent {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code deleting server",
			fmt.Sprintf("Received %s deleting server: name=%s, server_id=%s. Details: %s", delResp.Status(), data.Name.ValueString(), data.Id.String(), delResp.Body))
		return
	}
}
