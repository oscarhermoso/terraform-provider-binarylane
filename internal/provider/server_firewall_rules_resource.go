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
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &serverFirewallRulesResource{}
	_ resource.ResourceWithConfigure   = &serverFirewallRulesResource{}
	_ resource.ResourceWithImportState = &serverFirewallRulesResource{}
)

func NewServerFirewallRulesResource() resource.Resource {
	return &serverFirewallRulesResource{}
}

type serverFirewallRulesResource struct {
	bc *BinarylaneClient
}

type serverFirewallRulesResourceModel struct {
	resources.ServerFirewallRulesModel
}

func (d *serverFirewallRulesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serverFirewallRulesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_firewall_rules"
}

func (r *serverFirewallRulesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resources.ServerFirewallRulesResourceSchema(ctx)
	resp.Schema.Description = "Retrieve details about the External Firewall Rules assigned to a BinaryLane server."

	// Overrides
	serverId := resp.Schema.Attributes["server_id"]
	resp.Schema.Attributes["server_id"] = &schema.Int64Attribute{
		Description:         serverId.GetDescription(),
		MarkdownDescription: serverId.GetMarkdownDescription(),
		Required:            true, // Server ID is required to define the firewall rules
		Validators: []validator.Int64{
			int64validator.AtLeast(1),
		},
	}
}

func (r *serverFirewallRulesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data serverFirewallRulesResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	firewallRules := []binarylane.AdvancedFirewallRule{}
	diags := data.FirewallRules.ElementsAs(ctx, &firewallRules, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(firewallRules) == 0 {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Creating server firewall rules: server_id=%s", data.ServerId.String()))
	fwResp, err := r.bc.client.PostServersServerIdActionsChangeAdvancedFirewallRulesWithResponse(
		ctx,
		data.ServerId.ValueInt64(),
		binarylane.PostServersServerIdActionsChangeAdvancedFirewallRulesJSONRequestBody{
			Type:          "change_advanced_firewall_rules",
			FirewallRules: firewallRules,
		})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error creating server firewall rules: server_id=%s", data.ServerId.String()),
			err.Error(),
		)
		return
	}
	if fwResp.StatusCode() == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf("Server firewall rules not found, removing from state: server_id=%s", data.ServerId.String()))
		resp.State.RemoveResource(ctx)
		return
	}
	if fwResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code creating server firewall rules",
			fmt.Sprintf("Received %s creating server firewall rules: server_id=%s. Details: %s", fwResp.Status(), data.ServerId.String(), fwResp.Body))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverFirewallRulesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data serverFirewallRulesResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	tflog.Debug(ctx, fmt.Sprintf("Reading server firewall rules: server_id=%d", data.ServerId.ValueInt64()))
	fwResp, err := r.bc.client.GetServersServerIdAdvancedFirewallRulesWithResponse(ctx, data.ServerId.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading server firewall rules: server_id=%d", data.ServerId.ValueInt64()),
			err.Error(),
		)
		return
	}
	if fwResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code reading server firewall rules",
			fmt.Sprintf("Received %s reading server firewall rules: server_id=%d. Details: %s", fwResp.Status(), data.ServerId.ValueInt64(), fwResp.Body))
		return
	}

	firewallRulesValue, diags := GetFirewallRulesState(ctx, &fwResp.JSON200.FirewallRules)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.FirewallRules = firewallRulesValue

	// Save into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverFirewallRulesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data serverFirewallRulesResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	firewallRules := []binarylane.AdvancedFirewallRule{}
	diags := data.FirewallRules.ElementsAs(ctx, &firewallRules, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(firewallRules) == 0 {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Updating server firewall rules: server_id=%s", data.ServerId.String()))
	serverResp, err := r.bc.client.PostServersServerIdActionsChangeAdvancedFirewallRulesWithResponse(
		ctx,
		data.ServerId.ValueInt64(),
		binarylane.PostServersServerIdActionsChangeAdvancedFirewallRulesJSONRequestBody{
			Type:          "change_advanced_firewall_rules",
			FirewallRules: firewallRules,
		})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updating server firewall rules: server_id=%s", data.ServerId.String()),
			err.Error(),
		)
		return
	}
	if serverResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code updating server firewall rules",
			fmt.Sprintf("Received %s updating server firewall rules: server_id=%s. Details: %s", serverResp.Status(), data.ServerId.String(), serverResp.Body))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverFirewallRulesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data serverFirewallRulesResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete API call logic
	tflog.Debug(ctx, fmt.Sprintf("Deleting server firewall rules: server_id=%s", data.ServerId.String()))
	serverResp, err := r.bc.client.PostServersServerIdActionsChangeAdvancedFirewallRulesWithResponse(
		ctx,
		data.ServerId.ValueInt64(),
		binarylane.PostServersServerIdActionsChangeAdvancedFirewallRulesJSONRequestBody{
			Type:          "change_advanced_firewall_rules",
			FirewallRules: []binarylane.AdvancedFirewallRule{},
		})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error deleting server firewall rules: server_id=%s", data.ServerId.String()),
			err.Error(),
		)
		return
	}
	if serverResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code deleting server firewall rules",
			fmt.Sprintf("Received %s deleting server firewall rules: server_id=%s. Details: %s", serverResp.Status(), data.ServerId.String(), serverResp.Body))
		return
	}
}

func GetFirewallRulesState(ctx context.Context, firewallRules *[]binarylane.AdvancedFirewallRule) (basetypes.ListValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	firewallRulesValue := resources.FirewallRulesValue{}
	var firewallRulesValues []resources.FirewallRulesValue

	if *firewallRules == nil || len(*firewallRules) == 0 {
		firewallRulesValues = []resources.FirewallRulesValue{}
	} else {
		for _, rule := range *firewallRules {
			destinationAddresses, diag := types.ListValueFrom(ctx, types.StringType, rule.DestinationAddresses)
			diags.Append(diag...)

			var destinationPorts basetypes.ListValue
			if rule.DestinationPorts != nil {
				destinationPorts, diag = types.ListValueFrom(ctx, types.StringType, rule.DestinationPorts)
				diags.Append(diag...)
			}

			sourceAddresses, diag := types.ListValueFrom(ctx, types.StringType, rule.SourceAddresses)
			diags.Append(diag...)

			if diags.HasError() {
				return basetypes.NewListUnknown(firewallRulesValue.Type(ctx)), diags
			}

			// TODO: This panics if invalid, should be handled better
			ruleValue := resources.NewFirewallRulesValueMust(firewallRulesValue.AttributeTypes(ctx), map[string]attr.Value{
				"description":           types.StringPointerValue(rule.Description),
				"action":                types.StringValue(string(rule.Action)),
				"protocol":              types.StringValue(string(rule.Protocol)),
				"destination_addresses": destinationAddresses,
				"destination_ports":     destinationPorts,
				"source_addresses":      sourceAddresses,
			})
			firewallRulesValues = append(firewallRulesValues, ruleValue)
		}
	}

	firewallRulesListValue, diag := types.ListValueFrom(ctx, firewallRulesValue.Type(ctx), firewallRulesValues)
	diags.Append(diag...)
	if diags.HasError() {
		return basetypes.NewListUnknown(firewallRulesValue.Type(ctx)), diags
	}

	return firewallRulesListValue, diags
}

func (r *serverFirewallRulesResource) ImportState(
	ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	id, err := strconv.ParseInt(req.ID, 10, 32)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing server firewall rules",
			"Could not import server firewall rules, unexpected error (ID should be an integer): "+err.Error(),
		)
		return
	}

	diags := resp.State.SetAttribute(ctx, path.Root("server_id"), id)
	resp.Diagnostics.Append(diags...)
}
