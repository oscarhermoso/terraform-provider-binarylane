package provider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"terraform-provider-binarylane/internal/binarylane"
	"terraform-provider-binarylane/internal/resources"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &loadBalancerResource{}
	_ resource.ResourceWithConfigure   = &loadBalancerResource{}
	_ resource.ResourceWithImportState = &loadBalancerResource{}
)

func NewLoadBalancerResource() resource.Resource {
	return &loadBalancerResource{}
}

type loadBalancerResource struct {
	bc *BinarylaneClient
}

type loadBalancerResourceModel struct {
	resources.LoadBalancerModel
}

func (r *loadBalancerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer"
}

func (d *loadBalancerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *loadBalancerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resources.LoadBalancerResourceSchema(ctx)
	// resp.Schema.Description = "TODO"

	// Overrides
	id := resp.Schema.Attributes["id"]
	resp.Schema.Attributes["id"] = &schema.Int64Attribute{
		Description:         id.GetDescription(),
		MarkdownDescription: id.GetMarkdownDescription(),
		// read only
		Optional: false,
		Required: false,
		Computed: true,
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.UseStateForUnknown(),
		},
	}

	region := resp.Schema.Attributes["region"]
	resp.Schema.Attributes["region"] = &schema.StringAttribute{
		Description:         region.GetDescription(),
		MarkdownDescription: region.GetMarkdownDescription(),
		Optional:            true,
		Computed:            false, // region is not computed, defined at creation
		PlanModifiers: []planmodifier.String{ // region is not allowed to be changed
			stringplanmodifier.RequiresReplace(),
		},
	}
}

func (r *loadBalancerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data loadBalancerResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	forwardingRules := []binarylane.ForwardingRule{}
	diags := data.LoadBalancerModel.ForwardingRules.ElementsAs(ctx, &forwardingRules, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverIds := []int64{}
	diags = data.ServerIds.ElementsAs(ctx, &serverIds, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := binarylane.CreateLoadBalancerRequest{
		Name:            data.Name.ValueString(),
		ForwardingRules: &forwardingRules,
		ServerIds:       &serverIds,
	}

	if data.HealthCheck.Path.ValueStringPointer() != nil || data.HealthCheck.Protocol.ValueStringPointer() != nil {
		body.HealthCheck = &binarylane.HealthCheck{
			Path:     data.HealthCheck.Path.ValueStringPointer(),
			Protocol: data.HealthCheck.Protocol.ValueStringPointer(),
		}
	}

	tflog.Info(ctx, fmt.Sprintf("Creating Load Balancer: name=%s", data.Name.ValueString()))

	lbResp, err := r.bc.client.PostLoadBalancersWithResponse(ctx, body)
	if err != nil {
		tflog.Info(ctx, fmt.Sprintf("Attempted to create new load balancer: request=%+v", body))
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error sending request to create load balancer: name=%s", data.Name.ValueString()),
			err.Error(),
		)
		return
	}

	if lbResp.StatusCode() != http.StatusOK {
		tflog.Info(ctx, fmt.Sprintf("Attempted to create new load balancer: request=%+v", body))
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code creating load balancer",
			fmt.Sprintf("Received %s creating new load balancer: name=%s. Details: %s", lbResp.Status(), data.Name.ValueString(), lbResp.Body))
		return
	}

	diags = SetLoadBalancerModelState(ctx, &data.LoadBalancerModel, lbResp.JSON200.LoadBalancer)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// Wait for server to be ready
	var createActionId int64
	for _, action := range *lbResp.JSON200.Links.Actions {
		if *action.Rel == "create_load_balancer" {
			createActionId = *action.Id
			break
		}
	}
	if createActionId == 0 {
		resp.Diagnostics.AddError(
			"Unable to wait for load balancer to be created, links.actions with rel=create_load_balancer missing from response",
			fmt.Sprintf("Received %s creating new load balancer: name=%s. Details: %s", lbResp.Status(), data.Name.ValueString(), lbResp.Body))
		return
	}

	timeLimit := time.Now().Add(time.Duration(60 * time.Second))
	for {
		tflog.Info(ctx, "Waiting for load balancer to be ready...")

		readyResp, err := r.bc.client.GetActionsActionIdWithResponse(ctx, createActionId)
		if err != nil {
			resp.Diagnostics.AddError("Error waiting for load balancer to be ready", err.Error())
			return
		}
		if readyResp.StatusCode() == http.StatusOK && readyResp.JSON200.Action.CompletedAt != nil {
			tflog.Info(ctx, "Load balancer is ready")
			break
		}
		if time.Now().After(timeLimit) {
			resp.Diagnostics.AddError(
				"Timed out waiting for load balancer to be ready",
				fmt.Sprintf(
					"Timed out waiting for load balancer to be created, as `wait_for_create` was surpassed without "+
						"recieving a `completed_at` in response: name=%s, status=%s, details: %s",
					data.Name.ValueString(), readyResp.Status(), readyResp.Body,
				),
			)
			return
		}
		tflog.Debug(ctx, fmt.Sprintf("Waiting for load balancer to be ready: name=%s, status=%s, details: %s", data.Name.ValueString(), readyResp.Status(), readyResp.Body))
		time.Sleep(time.Second * 5)
	}
}

func (r *loadBalancerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data loadBalancerResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	lbResp, err := r.bc.client.GetLoadBalancersLoadBalancerIdWithResponse(ctx, data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading load balancer: name=%s", data.Name.ValueString()),
			err.Error(),
		)
		return
	}

	if lbResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code reading load balancer",
			fmt.Sprintf("Received %s reading load balancer: name=%s. Details: %s", lbResp.Status(), data.Name.ValueString(), lbResp.Body),
		)
		return
	}

	diags := SetLoadBalancerModelState(ctx, &data.LoadBalancerModel, lbResp.JSON200.LoadBalancer)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *loadBalancerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data loadBalancerResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	forwardingRules := &[]binarylane.ForwardingRule{}
	diags := data.LoadBalancerModel.ForwardingRules.ElementsAs(ctx, forwardingRules, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverIds := &[]int64{}
	diags = data.ServerIds.ElementsAs(ctx, serverIds, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Updating Load Balancer: name=%s", data.Name.ValueString()))

	body := binarylane.UpdateLoadBalancerRequest{
		Name:            data.Name.ValueString(),
		ForwardingRules: forwardingRules,
		HealthCheck: &binarylane.HealthCheck{
			Path:     data.HealthCheck.Path.ValueStringPointer(),
			Protocol: data.HealthCheck.Protocol.ValueStringPointer(),
		},
		ServerIds: serverIds,
	}

	lbResp, err := r.bc.client.PutLoadBalancersLoadBalancerIdWithResponse(ctx, data.Id.ValueInt64(), body)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading load balancer: name=%s", data.Name.ValueString()),
			err.Error(),
		)
		return
	}

	if lbResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code reading load balancer",
			fmt.Sprintf("Received %s reading load balancer: name=%s. Details: %s", lbResp.Status(), data.Name.ValueString(), lbResp.Body),
		)
		return
	}

	diags = SetLoadBalancerModelState(ctx, &data.LoadBalancerModel, lbResp.JSON200.LoadBalancer)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *loadBalancerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data loadBalancerResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete API call logic
	tflog.Debug(ctx, fmt.Sprintf("Deleting Load Balancer: name=%s", data.Name.ValueString()))

	lbResp, err := r.bc.client.DeleteLoadBalancersLoadBalancerIdWithResponse(ctx, data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error deleting load balancer: name=%s", data.Name.ValueString()),
			err.Error(),
		)
		return
	}

	if lbResp.StatusCode() != http.StatusNoContent {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code deleting load balancer",
			fmt.Sprintf("Received %s deleting load balancer: name=%s. Details: %s", lbResp.Status(), data.Name.ValueString(), lbResp.Body))
		return
	}
}

func (r *loadBalancerResource) ImportState(
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
	var loadBalancer binarylane.LoadBalancer
	var nextPage bool = true

	for nextPage { // Need to paginate because the API does not support filtering by name
		params := binarylane.GetLoadBalancersParams{
			Page:    &page,
			PerPage: &perPage,
		}

		lbResp, err := r.bc.client.GetLoadBalancersWithResponse(ctx, &params)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Error getting load balancer for import: name=%s", req.ID), err.Error())
			return
		}

		if lbResp.StatusCode() != http.StatusOK {
			resp.Diagnostics.AddError(
				"Unexpected HTTP status code getting load balancer for import",
				fmt.Sprintf("Received %s reading load balancer: name=%s. Details: %s", lbResp.Status(), req.ID,
					lbResp.Body))
			return
		}

		loadBalancers := *lbResp.JSON200.LoadBalancers
		for _, lb := range loadBalancers {
			if *lb.Name == req.ID {
				loadBalancer = lb
				nextPage = false
				break
			}
		}
		if lbResp.JSON200.Links == nil || lbResp.JSON200.Links.Pages == nil || lbResp.JSON200.Links.Pages.Next == nil {
			nextPage = false
			break
		}

		page++
	}

	if loadBalancer.Id == nil {
		resp.Diagnostics.AddError(
			"Could not find load balancer by name",
			fmt.Sprintf("Error finding load balancer: name=%s", req.ID),
		)
		return
	}

	diags := resp.State.SetAttribute(ctx, path.Root("id"), *loadBalancer.Id)
	resp.Diagnostics.Append(diags...)
}

func SetLoadBalancerModelState(ctx context.Context, data *resources.LoadBalancerModel, lb *binarylane.LoadBalancer) diag.Diagnostics {
	var diags, diag diag.Diagnostics

	data.Id = types.Int64Value(*lb.Id)
	data.Name = types.StringValue(*lb.Name)

	if lb.Region == nil {
		data.Region = types.StringNull()
	} else {
		data.Region = types.StringValue(*lb.Region.Slug)
	}

	data.ServerIds, diags = types.ListValueFrom(ctx, types.Int64Type, lb.ServerIds)

	if lb.HealthCheck == nil {
		data.HealthCheck = resources.NewHealthCheckValueNull()
	} else {
		data.HealthCheck, diag = resources.NewHealthCheckValue(
			resources.HealthCheckValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"path":     types.StringValue(*lb.HealthCheck.Path),
				"protocol": types.StringValue(*lb.HealthCheck.Protocol),
			},
		)
		diags.Append(diag...)
	}

	data.ForwardingRules, diag = types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: resources.ForwardingRulesValue{}.AttributeTypes(ctx)},
		lb.ForwardingRules,
	)
	diags.Append(diag...)

	return diags
}
