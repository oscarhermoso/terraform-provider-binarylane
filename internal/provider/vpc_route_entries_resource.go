package provider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"terraform-provider-binarylane/internal/binarylane"
	"terraform-provider-binarylane/internal/resources"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                = &vpcRouteEntriesResource{}
	_ resource.ResourceWithConfigure   = &vpcRouteEntriesResource{}
	_ resource.ResourceWithImportState = &vpcRouteEntriesResource{}
)

func NewVpcRouteEntriesResource() resource.Resource {
	return &vpcRouteEntriesResource{}
}

type vpcRouteEntriesResource struct {
	bc *BinarylaneClient
}

type vpcRouteEntriesResourceModel struct {
	resources.VpcRouteEntriesModel
}

func (r *vpcRouteEntriesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_route_entries"
}

func (r *vpcRouteEntriesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resources.VpcRouteEntriesResourceSchema(ctx)
	// resp.Schema.Description = "TODO"

	// Overrides
	vpcId := resp.Schema.Attributes["vpc_id"]
	resp.Schema.Attributes["vpc_id"] = &schema.Int64Attribute{
		Description:         vpcId.GetDescription(),
		MarkdownDescription: vpcId.GetMarkdownDescription(),
		Required:            true, // vpc_id is required to define the route entries
		Validators: []validator.Int64{
			int64validator.AtLeast(1),
		},
	}
}

func (r *vpcRouteEntriesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data vpcRouteEntriesResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	routeEntries := []binarylane.RouteEntryRequest{}
	diags := data.VpcRouteEntriesModel.RouteEntries.ElementsAs(ctx, &routeEntries, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	vpcResp, err := r.bc.client.PatchVpcsVpcIdWithResponse(ctx, data.VpcId.ValueInt64(), binarylane.PatchVpcRequest{
		RouteEntries: &routeEntries,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error creating VPC route entries: vpc_id=%d", data.VpcId.ValueInt64()),
			err.Error(),
		)
		return
	}

	if vpcResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code creating VPC route entries",
			fmt.Sprintf("Received %s creating VPC route entries: vpc_id=%d. Details: %s", vpcResp.Status(), data.VpcId.ValueInt64(), vpcResp.Body))
		return
	}

	routeEntriesState, diags := GetRouteEntriesState(ctx, vpcResp.JSON200.Vpc.RouteEntries)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.RouteEntries = routeEntriesState

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vpcRouteEntriesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data vpcRouteEntriesResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	vpcResp, err := r.bc.client.GetVpcsVpcIdWithResponse(ctx, data.VpcId.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading VPC route entries: vpc_id=%d", data.VpcId.ValueInt64()),
			err.Error(),
		)
		return
	}
	if vpcResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code reading VPC route entries",
			fmt.Sprintf("Received %s reading VPC route entries: vpc_id=%d. Details: %s", vpcResp.Status(), data.VpcId.ValueInt64(), vpcResp.Body))
		return
	}

	routeEntries, routeEntriesDiags := GetRouteEntriesState(ctx, vpcResp.JSON200.Vpc.RouteEntries)
	resp.Diagnostics.Append(routeEntriesDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.RouteEntries = routeEntries

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vpcRouteEntriesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state vpcRouteEntriesResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	routeEntries := []binarylane.RouteEntryRequest{}
	diags := plan.VpcRouteEntriesModel.RouteEntries.ElementsAs(ctx, &routeEntries, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	vpcResp, err := r.bc.client.PatchVpcsVpcIdWithResponse(ctx, state.VpcId.ValueInt64(), binarylane.PatchVpcRequest{
		RouteEntries: &routeEntries,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updating VPC route entries: vpc_id=%d", state.VpcId.ValueInt64()),
			err.Error(),
		)
		return
	}

	if vpcResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code updating VPC route entries",
			fmt.Sprintf("Received %s updating VPC route entries: vpc_id=%d. Details: %s", vpcResp.Status(), state.VpcId.ValueInt64(), vpcResp.Body))
		return
	}

	routeEntriesState, diags := GetRouteEntriesState(ctx, vpcResp.JSON200.Vpc.RouteEntries)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.RouteEntries = routeEntriesState

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *vpcRouteEntriesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data vpcRouteEntriesResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete API call logic
	var routeEntries []binarylane.RouteEntryRequest
	vpcResp, err := r.bc.client.PatchVpcsVpcIdWithResponse(ctx, data.VpcId.ValueInt64(), binarylane.PatchVpcRequest{
		RouteEntries: &routeEntries,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error deleting VPC route entries: vpc_id=%d", data.VpcId.ValueInt64()),
			err.Error(),
		)
		return
	}

	if vpcResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code deleting VPC route entries",
			fmt.Sprintf("Received %s deleting VPC route entries: vpc_id=%d. Details: %s", vpcResp.Status(), data.VpcId.ValueInt64(), vpcResp.Body))
		return
	}
}

func (d *vpcRouteEntriesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func GetRouteEntriesState(ctx context.Context, routeEntries *[]binarylane.RouteEntry) (basetypes.ListValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	routeEntriesValue := resources.RouteEntriesValue{}
	var routeEntriesValues []resources.RouteEntriesValue

	if *routeEntries == nil || len(*routeEntries) == 0 {
		routeEntriesValues = []resources.RouteEntriesValue{}
	} else {
		for _, route := range *routeEntries {
			// TODO: This panics if invalid, should be handled better
			r := resources.NewRouteEntriesValueMust(routeEntriesValue.AttributeTypes(ctx), map[string]attr.Value{
				"description": types.StringValue(*route.Description),
				"destination": types.StringValue(*route.Destination),
				"router":      types.StringValue(*route.Router),
			})
			routeEntriesValues = append(routeEntriesValues, r)
		}
	}

	routeEntriesListValue, diag := types.ListValueFrom(ctx, routeEntriesValue.Type(ctx), routeEntriesValues)
	diags.Append(diag...)
	if diags.HasError() {
		return basetypes.NewListUnknown(routeEntriesValue.Type(ctx)), diags
	}

	return routeEntriesListValue, diags
}

func (r *vpcRouteEntriesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing VPC route entries",
			"Could not import VPC route entries, unexpected error (ID should be an integer): "+err.Error(),
		)
		return
	}

	diags := resp.State.SetAttribute(ctx, path.Root("vpc_id"), id)
	resp.Diagnostics.Append(diags...)
}
