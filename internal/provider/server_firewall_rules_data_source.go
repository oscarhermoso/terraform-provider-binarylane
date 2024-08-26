package provider

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-binarylane/internal/resources"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &serverFirewallRulesDataSource{}
	_ datasource.DataSourceWithConfigure = &serverFirewallRulesDataSource{}
)

func NewServerFirewallRulesDataSource() datasource.DataSource {
	return &serverFirewallRulesDataSource{}
}

type serverFirewallRulesDataSource struct {
	bc *BinarylaneClient
}

type serverFirewallRulesDataSourceModel struct {
	resources.ServerFirewallRulesModel
}

func (d *serverFirewallRulesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_firewall_rules"
}

func (d *serverFirewallRulesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	ds, err := convertResourceSchemaToDataSourceSchema(resources.ServerFirewallRulesResourceSchema(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to convert resource schema to data source schema", err.Error())
		return
	}
	resp.Schema = *ds
	resp.Schema.Description = "TODO"
}

func (d *serverFirewallRulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data serverFirewallRulesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	tflog.Debug(ctx, fmt.Sprintf("Reading server firewall rules: server_id=%d", data.ServerId.ValueInt64()))
	fwResp, err := d.bc.client.GetServersServerIdAdvancedFirewallRulesWithResponse(ctx, data.ServerId.ValueInt64())
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

func (d *serverFirewallRulesDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	bc, ok := req.ProviderData.(BinarylaneClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *BinarylaneClient, got: %T.", req.ProviderData))
		return
	}

	d.bc = &bc
}
