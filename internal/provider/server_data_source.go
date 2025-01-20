package provider

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-binarylane/internal/resources"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
	PublicIpv4Addresses  types.List   `tfsdk:"public_ipv4_addresses"`
	PrivateIPv4Addresses types.List   `tfsdk:"private_ipv4_addresses"`
	Permalink            types.String `tfsdk:"permalink"`
	Memory               types.Int32  `tfsdk:"memory"`
	Disk                 types.Int32  `tfsdk:"disk"`
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
		resources.ServerResourceSchema(ctx),
		AttributeConfig{
			RequiredAttributes: &[]string{"id"},
			ExcludedAttributes: &[]string{"password"},
		})
	if err != nil {
		resp.Diagnostics.AddError("Failed to convert resource schema to data source schema", err.Error())
		return
	}
	resp.Schema = *ds
	resp.Schema.Description = "Retrieve details about a BinaryLane Load Balancer."

	// Overrides
	id := resp.Schema.Attributes["id"]
	resp.Schema.Attributes["id"] = schema.Int64Attribute{
		Description:         id.GetDescription(),
		MarkdownDescription: id.GetMarkdownDescription(),
		Required:            true, // ID is required to find the server
	}

	nameDescription := "The hostname of your server, such as vps01.yourcompany.com."
	resp.Schema.Attributes["name"] = &schema.StringAttribute{
		Description:         nameDescription,
		MarkdownDescription: nameDescription,
		Optional:            false,
		Required:            false,
		Computed:            true,
	}

	backupsDescription := "If `true`, the server will be backed up twice per day. By default, backups are disabled."
	resp.Schema.Attributes["backups"] = &schema.BoolAttribute{
		Description:         backupsDescription,
		MarkdownDescription: backupsDescription,
		Optional:            false,
		Required:            false,
		Computed:            true,
	}

	sshKeysDescription := "This is a list of SSH key ids that were added to the server during creation."
	resp.Schema.Attributes["ssh_keys"] = &schema.ListAttribute{
		ElementType:         types.Int64Type,
		Description:         sshKeysDescription,
		MarkdownDescription: sshKeysDescription,
		Optional:            false,
		Required:            false,
		Computed:            true,
	}

	userDataDescription := "A script or cloud-config YAML file to configure the server."
	resp.Schema.Attributes["user_data"] = &schema.StringAttribute{
		Description:         userDataDescription,
		MarkdownDescription: userDataDescription,
		Optional:            false,
		Required:            false,
		Computed:            true,
	}

	vpcIdDescription := "ID of the Virtual Private Cloud (VPC) the server is connected to."
	resp.Schema.Attributes["vpc_id"] = &schema.Int64Attribute{
		Description:         vpcIdDescription,
		MarkdownDescription: vpcIdDescription,
		Optional:            false,
		Required:            false,
		Computed:            true,
	}

	// Additional attributes
	resp.Schema.Attributes["permalink"] = &schema.StringAttribute{
		Description:         "A randomly generated two-word identifier assigned to servers in regions that support this feature",
		MarkdownDescription: "A randomly generated two-word identifier assigned to servers in regions that support this feature",
		Optional:            false,
		Required:            false,
		Computed:            true,
	}

	publicIpv4AddressesDescription := "The public IPv4 addresses assigned to the server."
	resp.Schema.Attributes["public_ipv4_addresses"] = &schema.ListAttribute{
		Description:         publicIpv4AddressesDescription,
		MarkdownDescription: publicIpv4AddressesDescription,
		ElementType:         types.StringType,
		Optional:            false,
		Required:            false,
		Computed:            true,
	}

	privateIpv4AddressesDescription := "The private IPv4 addresses assigned to the server."
	resp.Schema.Attributes["private_ipv4_addresses"] = &schema.ListAttribute{
		Description:         privateIpv4AddressesDescription,
		MarkdownDescription: privateIpv4AddressesDescription,
		ElementType:         types.StringType,
		Optional:            false,
		Required:            false,
		Computed:            true,
	}

	memoryDescription := "The amount of memory in MB assigned to the server."
	resp.Schema.Attributes["memory"] = &schema.Int32Attribute{
		Description:         memoryDescription,
		MarkdownDescription: memoryDescription,
		Optional:            false,
		Required:            false,
		Computed:            true,
	}

	diskDescription := "The amount of storage in GB assigned to the server."
	resp.Schema.Attributes["disk"] = &schema.Int32Attribute{
		Description:         diskDescription,
		MarkdownDescription: diskDescription,
		Optional:            false,
		Required:            false,
		Computed:            true,
	}
}

func (d *serverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
