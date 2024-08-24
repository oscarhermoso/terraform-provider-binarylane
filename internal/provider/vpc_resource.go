package provider

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-binarylane/internal/binarylane"
	"terraform-provider-binarylane/internal/resources"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &vpcResource{}
	_ resource.ResourceWithConfigure = &vpcResource{}
	// _ resource.ResourceWithImportState = &exampleResource{}
)

func NewVpcResource() resource.Resource {
	return &vpcResource{}
}

type vpcResource struct {
	bc *BinarylaneClient
}

type vpcResourceModel struct {
	resources.VpcModel
}

func (r *vpcResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	bc, ok := req.ProviderData.(BinarylaneClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *BinarylaneClient, got: %T.", req.ProviderData))
		return
	}

	r.bc = &bc
}

func (r *vpcResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}

func (r *vpcResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resources.VpcResourceSchema(ctx)
	resp.Schema.Description = "TODO"
}

func (r *vpcResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data vpcResourceModel

	// Read Terraform plan data into the model
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	tflog.Debug(ctx, fmt.Sprintf("Creating VPC: name=%s", data.Name.ValueString()))

	var routeEntries []binarylane.RouteEntryRequest
	diags = data.VpcModel.RouteEntries.ElementsAs(ctx, &routeEntries, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := binarylane.CreateVpcRequest{
		Name:         data.Name.ValueString(),
		IpRange:      data.IpRange.ValueStringPointer(),
		RouteEntries: &routeEntries,
	}

	vpcResp, err := r.bc.client.PostVpcsWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error creating VPC: name=%s", data.Name.ValueString()),
			err.Error(),
		)
		return
	}

	if vpcResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code creating VPC",
			fmt.Sprintf("Received %s creating new VPC: name=%s. Details: %s", vpcResp.Status(), data.Name.ValueString(), vpcResp.Body))
		return
	}

	data.Id = types.Int64Value(*vpcResp.JSON200.Vpc.Id)
	data.Name = types.StringValue(*vpcResp.JSON200.Vpc.Name)

	routeEntriesState, diags := GetRouteEntriesState(ctx, vpcResp.JSON200.Vpc.RouteEntries)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.RouteEntries = routeEntriesState

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vpcResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data vpcResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	vpcResp, err := r.bc.client.GetVpcsVpcIdWithResponse(ctx, data.Id.ValueInt64())
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

	routeEntries, routeEntriesDiags := GetRouteEntriesState(ctx, vpcResp.JSON200.Vpc.RouteEntries)
	resp.Diagnostics.Append(routeEntriesDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.RouteEntries = routeEntries

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vpcResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data vpcResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append()
	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	tflog.Debug(ctx, fmt.Sprintf("Updating VPC: id=%d", data.Id.ValueInt64()))

	var routeEntries []binarylane.RouteEntryRequest
	diags := data.VpcModel.RouteEntries.ElementsAs(ctx, &routeEntries, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	vpcResp, err := r.bc.client.PutVpcsVpcIdWithResponse(ctx, data.Id.ValueInt64(), binarylane.UpdateVpcRequest{
		Name:         data.Name.ValueString(),
		RouteEntries: &routeEntries,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updating VPC: id=%d", data.Id.ValueInt64()),
			err.Error(),
		)
		return
	}

	if vpcResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code updating VPC",
			fmt.Sprintf("Received %s updating VPC: id=%d. Details: %s", vpcResp.Status(), data.Id.ValueInt64(), vpcResp.Body))
		return
	}

	data.Id = types.Int64Value(*vpcResp.JSON200.Vpc.Id)
	data.Name = types.StringValue(*vpcResp.JSON200.Vpc.Name)

	routeEntriesState, diags := GetRouteEntriesState(ctx, vpcResp.JSON200.Vpc.RouteEntries)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.RouteEntries = routeEntriesState

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vpcResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data vpcResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete API call logic
	tflog.Debug(ctx, fmt.Sprintf("Deleting VPC: id=%d", data.Id.ValueInt64()))
	vpcResp, err := r.bc.client.DeleteVpcsVpcIdWithResponse(ctx, data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error deleting VPC: id=%d", data.Id.ValueInt64()),
			err.Error(),
		)
		return
	}

	if vpcResp.StatusCode() != http.StatusNoContent {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code deleting VPC",
			fmt.Sprintf("Received %s deleting VPC: id=%d. Details: %s", vpcResp.Status(), data.Id.ValueInt64(), vpcResp.Body))
		return
	}
}
