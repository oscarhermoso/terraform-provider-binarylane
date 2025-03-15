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
	// AdvancedFeatures     types.Set    `tfsdk:"advanced_features"`
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
	resp.Schema.Description = "Retrieve details about a BinaryLane Server."

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

	// advFeatureDescripton := `By default, server will have some advanced features enabled. To only enable specific advance ` +
	// 	`features, provide as a list. Any currently enabled advanced features that aren't included in the list will be disabled.`
	// advFeatureDescriptonMarkdown := advFeatureDescripton + `

	// | Value | Description |
	// | ----- | ----------- |
	// | ` + "`" + `emulated-hyperv` + "`" + ` | Enable HyperV (a hypervisor produced by Microsoft) support. Enabled by default on Windows servers, generally of no value for non-Windows servers. |
	// | ` + "`" + `emulated-devices` + "`" + ` | When emulated devices is enabled, the KVM specific \"VirtIO\" disk drive and network devices are removed, and replaced with emulated versions of physical hardware: an old IDE HDD and an Intel E1000 network card.  Emulated devices are much slower than the VirtIO devices, and so this option should not be enabled unless absolutely necessary. |
	// | ` + "`" + `nested-virt` + "`" + ` | When this option is enabled the functionality necessary to run your own KVM servers within your server is enabled. Note that all the networking limits - one MAC address per VPS, restricted to specific IPs - still apply to public cloud so this is feature is generally only useful in combination with Virtual Private Cloud. |
	// | ` + "`" + `driver-disk` + "`" + ` | When this option is enabled a copy of the KVM driver disc for Windows (\"virtio-win.iso\") will be attached to your server as a virtual CD. This option can also be used in combination with your own attached backup when installing Windows. |
	// | ` + "`" + `unset-uuid` + "`" + ` | When this option is NOT enabled a 128-bit unique identifier is exposed to your server through the virtual BIOS. Each server receives a different UUID. Some propriety licensed software utilise this identifier to \"tie\" the license to a specific server. |
	// | ` + "`" + `local-rtc` + "`" + ` | When a server is booted the virtual BIOS receives the current date and time from the host node. The BIOS does not have an explicit timezone, so the timezone used is implicit and must be understood by the operating system. Most operating systems other than Windows expect the time to be UTC since it allows the operating system to control the timezone used when displaying the time. Our Windows installations have also been customized to use UTC, but when using your own installation of Windows this should be set to the host node's local timezone. |
	// | ` + "`" + `emulated-tpm` + "`" + ` | When enabled this provides an emulated TPM v1.2 device to your Cloud Server. Warning: the TPM state is not backed up. |
	// | ` + "`" + `cloud-init` + "`" + ` | (Read-Only) When this option is enabled the Cloud Server will be provided a datasource for the cloud-init service. |
	// | ` + "`" + `qemu-guest-agent` + "`" + ` | (Read-Only) When this option is enabled the server will allow QEMU Guest Agent to perform password reset without rebooting. |
	// | ` + "`" + `uefi-boot` + "`" + ` | (Read-Only) When this option is enabled the Cloud Server will use UEFI instead of legacy PC BIOS. |
	// `

	// resp.Schema.Attributes["advanced_features"] = &schema.SetAttribute{
	// 	ElementType:         basetypes.StringType{},
	// 	Description:         advFeatureDescripton,
	// 	MarkdownDescription: advFeatureDescriptonMarkdown,
	// 	Optional:            false,
	// 	Required:            false,
	// 	Computed:            true,
	// }
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

	// serverRespAdvancedFeatures, diag := types.SetValueFrom(ctx, basetypes.StringType{}, serverResp.JSON200.Server.AdvancedFeatures.EnabledAdvancedFeatures)
	// diags.Append(diag...)
	// data.AdvancedFeatures = serverRespAdvancedFeatures

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
