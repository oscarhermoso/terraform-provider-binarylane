package provider

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-binarylane/internal/binarylane"
	"terraform-provider-binarylane/internal/resources"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

type serverModel struct {
	resources.ServerModel
	WaitForCreateSeconds int32 `tfsdk:"wait_for_create"`
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
		Optional:            pw.IsOptional(),
		Computed:            false, // Computed must be false to allow server to be created without password
		Sensitive:           true,  // Mark password as sensitive
	}

	backups := resp.Schema.Attributes["backups"]
	resp.Schema.Attributes["backups"] = &schema.BoolAttribute{
		Description:         backups.GetDescription(),
		MarkdownDescription: backups.GetMarkdownDescription(),
		Optional:            backups.IsComputed(),
		Computed:            backups.IsOptional(),
		Default:             booldefault.StaticBool(false), // Add default to backups
	}

	user_data := resp.Schema.Attributes["user_data"]
	resp.Schema.Attributes["user_data"] = &schema.StringAttribute{
		Description:         user_data.GetDescription(),
		MarkdownDescription: user_data.GetMarkdownDescription(),
		Optional:            user_data.IsComputed(),
		Computed:            user_data.IsOptional(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}

	// Additional attributes
	waitDescription := "The number of seconds to wait for the server to be created, after which, a timeout error will " +
		"be reported. If `wait_seconds` is left empty or set to 0, Terraform will succeed without waiting for the " +
		"server creation to complete."
	resp.Schema.Attributes["wait_for_create"] = &schema.Int32Attribute{
		Description:         waitDescription,
		MarkdownDescription: waitDescription,
		Optional:            true,
		Computed:            true,
		Default:             int32default.StaticInt32(0),
	}
}

func (r *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data serverModel

	// Read Terraform plan data into the model
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	tflog.Debug(ctx, fmt.Sprintf("Creating server: name=%s", data.Name.ValueString()))

	body := binarylane.CreateServerRequest{
		Name:     data.Name.ValueStringPointer(),
		Image:    data.Image.ValueString(),
		Region:   data.Region.ValueString(),
		Size:     data.Size.ValueString(),
		UserData: data.UserData.ValueStringPointer(),
	}

	if data.Password.IsNull() {
		data.Password = types.StringNull()
	} else {
		body.Password = data.Password.ValueStringPointer()
	}

	serverResp, err := r.bc.client.PostServersWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error creating server: name=%s", data.Name.ValueString()),
			err.Error(),
		)
		return
	}

	if serverResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code creating server",
			fmt.Sprintf("Received %s creating new server: name=%s. Details: %s", serverResp.Status(), data.Name.ValueString(), serverResp.Body))
		return
	}

	data.Id = types.Int64Value(*serverResp.JSON200.Server.Id)
	data.Name = types.StringValue(*serverResp.JSON200.Server.Name)
	data.Image = types.StringValue(*serverResp.JSON200.Server.Image.Slug)
	data.Region = types.StringValue(*serverResp.JSON200.Server.Region.Slug)
	data.Size = types.StringValue(*serverResp.JSON200.Server.Size.Slug)
	data.Backups = types.BoolValue(serverResp.JSON200.Server.NextBackupWindow != nil)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if data.WaitForCreateSeconds <= 0 {
		return
	}

	// Wait for server to be ready
	var createActionId int64
	for _, action := range *serverResp.JSON200.Links.Actions {
		if *action.Rel == "create" {
			createActionId = *action.Id
			break
		}
	}
	if createActionId == 0 {
		resp.Diagnostics.AddError(
			"Unable to wait for server to be created, links.actions with rel=create missing from response",
			fmt.Sprintf("Received %s creating new server: name=%s. Details: %s", serverResp.Status(), data.Name.ValueString(), serverResp.Body))
		return
	}

	timeLimit := time.Now().Add(time.Duration(data.WaitForCreateSeconds) * time.Second)
	for {
		tflog.Info(ctx, "Waiting for server to be ready...")

		readyResp, err := r.bc.client.GetServersServerIdActionsActionIdWithResponse(ctx, data.Id.ValueInt64(), createActionId)
		if err != nil {
			resp.Diagnostics.AddError("Error waiting for server to be ready", err.Error())
			return
		}
		if readyResp.StatusCode() == http.StatusOK && readyResp.JSON200.Action.CompletedAt != nil {
			tflog.Info(ctx, "Server is ready")
			break
		}
		if time.Now().After(timeLimit) {
			resp.Diagnostics.AddError(
				"Timed out waiting for server to be ready",
				fmt.Sprintf(
					"Timed out waiting for server to be created, as `wait_for_create` was surpassed without "+
						"recieving a `completed_at` in response: name=%s, status=%s, details: %s",
					data.Name.ValueString(), readyResp.Status(), readyResp.Body,
				),
			)
			return
		}
		tflog.Debug(ctx, fmt.Sprintf("Waiting for server to be ready: name=%s, status=%s, details: %s", data.Name.ValueString(), readyResp.Status(), readyResp.Body))
		time.Sleep(time.Second * 5)
	}
}

func (r *serverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data serverModel

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
			"Unexpected HTTP status code reading server",
			fmt.Sprintf("Received %s reading server: name=%s. Details: %s", serverResp.Status(), data.Id.String(), serverResp.Body))
		return
	}

	data.Id = types.Int64Value(*serverResp.JSON200.Server.Id)
	data.Name = types.StringValue(*serverResp.JSON200.Server.Name)
	data.Image = types.StringValue(*serverResp.JSON200.Server.Image.Slug)
	data.Region = types.StringValue(*serverResp.JSON200.Server.Region.Slug)
	data.Size = types.StringValue(*serverResp.JSON200.Server.Size.Slug)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data serverModel

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

func (r *serverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data serverModel

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

	serverResp, err := r.bc.client.DeleteServersServerIdWithResponse(ctx, data.Id.ValueInt64(), &params)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error deleting server: name=%s, server_id=%s", data.Name.ValueString(), data.Id.String()),
			err.Error(),
		)
		return
	}

	if serverResp.StatusCode() != http.StatusNoContent {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code deleting server",
			fmt.Sprintf("Received %s deleting server: name=%s, server_id=%s. Details: %s", serverResp.Status(), data.Name.ValueString(), data.Id.String(), serverResp.Body))
		return
	}
}

// func (r *serverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
// 	// Retrieve import ID and save to id attribute
// 	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
// }
