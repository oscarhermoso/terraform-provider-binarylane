package provider

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"terraform-provider-binarylane/internal/resources"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &serverDataSource{}
	_ datasource.DataSourceWithConfigure = &serverDataSource{}
)

func NewServerDataSource() datasource.DataSource {
	return &serverDataSource{}
}

type serverDataSource struct {
	bc *BinarylaneClient
}

type serverDataModel struct {
	resources.ServerModel
	PublicIpv4Addresses       types.List   `tfsdk:"public_ipv4_addresses"`
	PrivateIPv4Addresses      types.List   `tfsdk:"private_ipv4_addresses"`
	Permalink                 types.String `tfsdk:"permalink"`
	Memory                    types.Int32  `tfsdk:"memory"`
	Disk                      types.Int32  `tfsdk:"disk"`
	SourceAndDestinationCheck types.Bool   `tfsdk:"source_and_destination_check"`
}

func (d *serverDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serverDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (d *serverDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	ds, err := convertResourceSchemaToDataSourceSchema(
		serverSchema(ctx),
		AttributeConfig{
			RequiredAttributes: &[]string{"id"},
			ExcludedAttributes: &[]string{"password", "public_ipv4_count", "password", "password_change_supported", "timeouts"},
		})
	if err != nil {
		resp.Diagnostics.AddError("Failed to convert resource schema to data source schema", err.Error())
		return
	}
	resp.Schema = *ds
	resp.Schema.Description = "Retrieve details about a BinaryLane Server."
}

func (d *serverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var diag diag.Diagnostics
	var data serverDataModel

	// Read Terraform configuration data into the model
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read server
	serverResp, err := d.bc.client.GetServersServerIdWithResponse(ctx, data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading server: name=%s, id=%s", data.Name.ValueString(), data.Id.String()),
			err.Error(),
		)
		return
	}
	if serverResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unexpected HTTP status code reading server: name=%s, id=%s", data.Name.ValueString(), data.Id.String()),
			string(serverResp.Body),
		)
		return
	}

	data.Id = types.Int64Value(*serverResp.JSON200.Server.Id)
	data.Name = types.StringValue(*serverResp.JSON200.Server.Name)
	data.Image = types.StringValue(*serverResp.JSON200.Server.Image.Slug)
	data.Region = types.StringValue(*serverResp.JSON200.Server.Region.Slug)
	data.Size = types.StringValue(*serverResp.JSON200.Server.Size.Slug)
	data.Backups = types.BoolValue(serverResp.JSON200.Server.NextBackupWindow != nil)
	data.PortBlocking = types.BoolValue(serverResp.JSON200.Server.Networks.PortBlocking)
	data.VpcId = types.Int64PointerValue(serverResp.JSON200.Server.VpcId)
	data.Permalink = types.StringValue(*serverResp.JSON200.Server.Permalink)
	data.Memory = types.Int32Value(*serverResp.JSON200.Server.Memory)
	data.Disk = types.Int32Value(*serverResp.JSON200.Server.Disk)
	data.SourceAndDestinationCheck = types.BoolPointerValue(serverResp.JSON200.Server.Networks.SourceAndDestinationCheck)

	advFeat := *serverResp.JSON200.Server.AdvancedFeatures.EnabledAdvancedFeatures
	data.AdvancedFeatures, diags = resources.NewAdvancedFeaturesValue(
		resources.AdvancedFeaturesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"emulated_hyperv":  types.BoolValue(slices.Contains(advFeat, "emulated-hyperv")),
			"emulated_devices": types.BoolValue(slices.Contains(advFeat, "emulated-devices")),
			"nested_virt":      types.BoolValue(slices.Contains(advFeat, "nested-virt")),
			"driver_disk":      types.BoolValue(slices.Contains(advFeat, "driver-disk")),
			"unset_uuid":       types.BoolValue(slices.Contains(advFeat, "unset-uuid")),
			"local_rtc":        types.BoolValue(slices.Contains(advFeat, "local-rtc")),
			"emulated_tpm":     types.BoolValue(slices.Contains(advFeat, "emulated-tpm")),
			"cloud_init":       types.BoolValue(slices.Contains(advFeat, "cloud-init")),
			"qemu_guest_agent": types.BoolValue(slices.Contains(advFeat, "qemu-guest-agent")),
			"uefi_boot":        types.BoolValue(slices.Contains(advFeat, "uefi-boot")),
		})
	if diags.HasError() {
		data.AdvancedFeatures = resources.NewAdvancedFeaturesValueUnknown()
	}

	publicIpv4Addresses := []string{}
	privateIpv4Addresses := []string{}

	for _, v4address := range serverResp.JSON200.Server.Networks.V4 {
		if v4address.Type == "public" {
			publicIpv4Addresses = append(publicIpv4Addresses, v4address.IpAddress)
		} else {
			privateIpv4Addresses = append(privateIpv4Addresses, v4address.IpAddress)
		}
	}

	tfPublicIpv4Addresses, diag := types.ListValueFrom(ctx, types.StringType, publicIpv4Addresses)
	diags.Append(diag...)
	if diag.HasError() {
		data.PublicIpv4Addresses = types.ListUnknown(data.PublicIpv4Addresses.ElementType(ctx))
	} else {
		data.PublicIpv4Addresses = tfPublicIpv4Addresses
	}

	tfPrivateIpv4Addresses, diag := types.ListValueFrom(ctx, types.StringType, privateIpv4Addresses)
	diags.Append(diag...)
	if diag.HasError() {
		data.PrivateIPv4Addresses = types.ListUnknown(data.PrivateIPv4Addresses.ElementType(ctx))
	} else {
		data.PrivateIPv4Addresses = tfPrivateIpv4Addresses
	}

	// Get user data script
	userDataResp, err := d.bc.client.GetServersServerIdUserDataWithResponse(ctx, data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading server user data: id=%s, name=%s", data.Id.String(), data.Name.ValueString()),
			err.Error(),
		)
		return
	}
	if userDataResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unexpected HTTP status %d reading server user data: name=%s, id=%s", userDataResp.StatusCode(), data.Name.ValueString(), data.Id.String()),
			string(userDataResp.Body),
		)
		return
	}
	data.UserData = types.StringPointerValue(userDataResp.JSON200.UserData)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
