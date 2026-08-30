package provider

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"terraform-provider-binarylane/internal/binarylane"
	"terraform-provider-binarylane/internal/resources"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &serverResource{}
	_ resource.ResourceWithConfigure   = &serverResource{}
	_ resource.ResourceWithImportState = &serverResource{}
	_ resource.ResourceWithModifyPlan  = &serverResource{}
)

// Helper function to simplify the provider implementation.
func NewServerResource() resource.Resource {
	return &serverResource{}
}

type serverResource struct {
	bc *BinarylaneClient
}

type serverResourceModel struct {
	serverDataModel

	PublicIpv4Count         types.Int32    `tfsdk:"public_ipv4_count"`
	Password                types.String   `tfsdk:"password"`
	PasswordChangeSupported types.Bool     `tfsdk:"password_change_supported"`
	Timeouts                timeouts.Value `tfsdk:"timeouts"`
}

func (d *serverResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serverResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func serverSchema(ctx context.Context) schema.Schema {
	s := resources.ServerResourceSchema(ctx)

	// Overrides
	id := s.Attributes["id"]
	s.Attributes["id"] = schema.Int64Attribute{
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

	imageDescription := "The slug of the selected operating system, such as `debian-12`. You can fetch a full list of images from the BinaryLane API."
	image := s.Attributes["image"]
	s.Attributes["image"] = schema.StringAttribute{
		Description:         imageDescription,
		MarkdownDescription: imageDescription,
		Required:            image.IsRequired(),
		Optional:            image.IsOptional(),
		Computed:            image.IsComputed(),
		Validators: []validator.String{
			stringvalidator.LengthAtLeast(1),
		},
	}

	backupsDescription := "If `true` this will enable two daily backups for the server. By default, backups are disabled."
	s.Attributes["backups"] = schema.BoolAttribute{
		Description:         backupsDescription,
		MarkdownDescription: backupsDescription,
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false), // Add default to backups
	}

	user_data := s.Attributes["user_data"]
	s.Attributes["user_data"] = schema.StringAttribute{
		Description:         user_data.GetDescription(),
		MarkdownDescription: user_data.GetMarkdownDescription(),
		Optional:            true,  // Optional as not all servers have an initialization script
		Computed:            false, // User defined
	}

	vpcId := s.Attributes["vpc_id"]
	s.Attributes["vpc_id"] = schema.Int64Attribute{
		Description:         vpcId.GetDescription(),
		MarkdownDescription: vpcId.GetMarkdownDescription(),
		Optional:            vpcId.IsOptional(),
		Computed:            false, // vpc_id is not computed, defined at creation
	}

	portBlocking := s.Attributes["port_blocking"]
	s.Attributes["port_blocking"] = schema.BoolAttribute{
		Description:         portBlocking.GetDescription(),
		MarkdownDescription: portBlocking.GetMarkdownDescription(),
		Optional:            portBlocking.IsOptional(),
		Computed:            portBlocking.IsComputed(),
		Default:             booldefault.StaticBool(true), // Add default to port_blocking
	}

	region := s.Attributes["region"].(schema.StringAttribute)
	s.Attributes["region"] = schema.StringAttribute{
		Description:         region.GetDescription(),
		MarkdownDescription: region.GetMarkdownDescription(),
		Optional:            region.IsOptional(),
		Computed:            region.IsComputed(),
		Required:            region.IsRequired(),
		Validators:          region.StringValidators(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}

	sshKeys := s.Attributes["ssh_keys"]
	s.Attributes["ssh_keys"] = schema.ListAttribute{
		ElementType:         types.Int64Type,
		Description:         sshKeys.GetMarkdownDescription(),
		MarkdownDescription: sshKeys.GetDescription(),
		Optional:            sshKeys.IsOptional(),
		Computed:            false, // SSH keys are not computed, defined at creation
		Validators: []validator.List{
			listvalidator.ValueInt64sAre(int64validator.AtLeast(1)),
		},
	}

	userDataDescription := "A script or cloud-config YAML file to configure the server. Can only be specified if the OS image supports UserData (i.e. not Windows)." +
		" See more: https://cloudinit.readthedocs.io/en/latest/explanation/format.html#user-data-script"
	userData := s.Attributes["user_data"]
	s.Attributes["user_data"] = schema.StringAttribute{
		Description:         userDataDescription,
		MarkdownDescription: userDataDescription,
		Required:            userData.IsRequired(),
		Optional:            userData.IsOptional(),
		Computed:            userData.IsComputed(),
		Validators: []validator.String{
			stringvalidator.LengthAtMost(65536),
		},
	}

	// Additional attributes
	pwDescription := "If this is provided the specified or default remote user's account password will be set to this value. " +
		"Only valid if the server supports password change actions. If omitted and the server supports password " +
		"change actions a random password will be generated and emailed to the account email address."
	s.Attributes["password"] = schema.StringAttribute{
		Description:         pwDescription,
		MarkdownDescription: pwDescription,
		Optional:            true,  // Password optional, if not set will be emailed to user
		Computed:            false, // Computed must be false to allow server to be created without password
		Sensitive:           true,  // Mark password as sensitive
	}

	publicIpv4CountDescription := "The number of public IPv4 addresses to assign to the server."
	s.Attributes["public_ipv4_count"] = schema.Int32Attribute{
		Description:         publicIpv4CountDescription,
		MarkdownDescription: publicIpv4CountDescription,
		Required:            true,
		Optional:            false,
		Computed:            false,
		Validators: []validator.Int32{
			int32validator.AtLeast(0),
			int32validator.AtMost(8),
		},
	}

	publicIpv4AddressesDescription := "The public IPv4 addresses assigned to the server."
	s.Attributes["public_ipv4_addresses"] = schema.ListAttribute{
		Description:         publicIpv4AddressesDescription,
		MarkdownDescription: publicIpv4AddressesDescription,
		ElementType:         types.StringType,
		// read only
		Optional: false,
		Required: false,
		Computed: true,
	}

	sourceDestCheckDescription := "This attribute can only be set if your server also has a `vpc_id` attribute set. " +
		"When enabled (which is `true` by default), your server will only be able to send or receive " +
		"packets that are directly addressed to one of the IP addresses associated with the Cloud Server. Generally, " +
		"this is desirable behaviour because it prevents IP conflicts and other hard-to-diagnose networking faults due " +
		"to incorrect network configuration. When `source_and_destination_check` is `false`, your Cloud Server will be able " +
		"to send and receive packets addressed to any server. This is typically used when you want to use " +
		"your Cloud Server as a VPN endpoint, a NAT server to provide internet access, or IP forwarding."
	s.Attributes["source_and_destination_check"] = schema.BoolAttribute{
		Description:         sourceDestCheckDescription,
		MarkdownDescription: sourceDestCheckDescription,
		Optional:            true,
		Required:            false,
		Computed:            true,
		Validators: []validator.Bool{
			boolvalidator.AlsoRequires(path.Expressions{
				path.MatchRoot("vpc_id"),
			}...),
		},
	}

	separatePrivateNicDescription := "This attribute can only be set if your server also has a `vpc_id` attribute set. " +
		"When enabled, a separate private network interface is provided for the server's VPC traffic."
	s.Attributes["separate_private_network_interface"] = schema.BoolAttribute{
		Description:         separatePrivateNicDescription,
		MarkdownDescription: separatePrivateNicDescription,
		Optional:            true,
		Computed:            true,
		Validators: []validator.Bool{
			boolvalidator.AlsoRequires(path.Expressions{
				path.MatchRoot("vpc_id"),
			}...),
		},
	}

	privateIpv4AddressesDescription := "The private IPv4 addresses assigned to the server."
	s.Attributes["private_ipv4_addresses"] = schema.ListAttribute{
		Description:         privateIpv4AddressesDescription,
		MarkdownDescription: privateIpv4AddressesDescription,
		ElementType:         types.StringType,
		// read only
		Optional: false,
		Required: false,
		Computed: true,
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
	}

	ipv6Description := "If `true` this will add a public and private IPv6 address to the server. By default, IPv6 is disabled."
	s.Attributes["ipv6"] = schema.BoolAttribute{
		Description:         ipv6Description,
		MarkdownDescription: ipv6Description,
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false), // Add default to ipv6
	}

	publicIpv6AddressesDescription := "The public IPv6 addresses assigned to the server."
	s.Attributes["public_ipv6_addresses"] = schema.ListAttribute{
		Description:         publicIpv6AddressesDescription,
		MarkdownDescription: publicIpv6AddressesDescription,
		ElementType:         types.StringType,
		// read only
		Optional: false,
		Required: false,
		Computed: true,
	}

	privateIpv6AddressesDescription := "The private IPv6 addresses assigned to the server."
	s.Attributes["private_ipv6_addresses"] = schema.ListAttribute{
		Description:         privateIpv6AddressesDescription,
		MarkdownDescription: privateIpv6AddressesDescription,
		ElementType:         types.StringType,
		// read only
		Optional: false,
		Required: false,
		Computed: true,
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
	}

	s.Attributes["permalink"] = schema.StringAttribute{
		Description:         "A randomly generated two-word identifier assigned to servers in regions that support this feature",
		MarkdownDescription: "A randomly generated two-word identifier assigned to servers in regions that support this feature",
		// read only
		Optional: false,
		Required: false,
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	pwChangeDescription := "If this is true then the `password` attribute can be changed with Terraform. " +
		"If this is false then the `password` attribute can only be replaced with a null/empty value, which will clear " +
		"the root/administrator password allowing the password to be changed via the web console."
	s.Attributes["password_change_supported"] = schema.BoolAttribute{
		Description:         pwChangeDescription,
		MarkdownDescription: pwChangeDescription,
		// read only
		Optional: false,
		Required: false,
		Computed: true,
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.UseStateForUnknown(),
		},
	}

	memoryDescription := `The total memory in MB for this server. Leave null to accept the default size.`
	memoryValidValues := "Valid values must be a multiple of 128. If the value is greater than 2048 MB, it must be a " +
		"multiple of 1024. If the value is greater than 16384 MB, it must be a multiple of 2048. If the value is greater " +
		"than 24576 MB, it must be a multiple of 4096."
	memoryValidValuesMarkdown := ` Valid values:
  - must be a multiple of 128
  - \> 2048 MB must be a multiple of 1024
  - \> 16384 MB must be a multiple of 2048
  - \> 24576 MB must be a multiple of 4096`

	s.Attributes["memory"] = schema.Int32Attribute{
		Description:         memoryDescription + memoryValidValues,
		MarkdownDescription: memoryDescription + memoryValidValuesMarkdown,
		Optional:            true,
		Required:            false,
		Computed:            true,
		Validators: []validator.Int32{
			int32validator.AtLeast(128),
			MultipleOfValidator{Multiple: 128},
			MultipleOfValidator{Multiple: 1024, RangeFrom: 2048, RangeTo: 16384},
			MultipleOfValidator{Multiple: 2048, RangeFrom: 16384, RangeTo: 24576},
			MultipleOfValidator{Multiple: 4096, RangeFrom: 24576},
		},
	}

	diskDescription := "The total storage in GB for this server. Leave null to accept the default for the size."
	diskValidValues := " Valid values must be a multiple of 5. If the value is greater than 60 GB, it must be a multiple of 10. " +
		"if the value is greater than 200 GB, it must be a multiple of 100. "
	diskValidValuesMarkdown := ` Valid values:
  - must be a multiple of 5
  - \> 60 GB must be a multiple of 10
  - \> 200 GB must be a multiple of 100`

	s.Attributes["disk"] = schema.Int32Attribute{
		Description:         diskDescription + diskValidValues,
		MarkdownDescription: diskDescription + diskValidValuesMarkdown,
		Optional:            true,
		Required:            false,
		Computed:            true,
		Validators: []validator.Int32{
			int32validator.AtLeast(20),
			MultipleOfValidator{Multiple: 5},
			MultipleOfValidator{Multiple: 10, RangeFrom: 60, RangeTo: 200},
			MultipleOfValidator{Multiple: 100, RangeFrom: 200},
		},
	}

	s.Attributes["timeouts"] = timeouts.Attributes(ctx, timeouts.Opts{
		Create: true,
		Update: true,
	})

	vpcIpv4Addr := s.Attributes["vpc_ipv4_address"]
	vpcIpv4AddrDescription := "If provided this will be the IPv4 address for the server's private VPC network adapter." +
		" If this is unspecified, then an unused IPv4 address will be assigned. This field is only valid when `vpc_id`" +
		" is provided."
	s.Attributes["vpc_ipv4_address"] = schema.StringAttribute{
		Description:         vpcIpv4AddrDescription,
		MarkdownDescription: vpcIpv4AddrDescription,
		Optional:            vpcIpv4Addr.IsOptional(),
		Computed:            vpcIpv4Addr.IsComputed(),
		Validators: []validator.String{
			stringvalidator.AlsoRequires(path.Expressions{
				path.MatchRoot("vpc_id"),
			}...),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	return s
}

func (r *serverResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = serverSchema(ctx)
}

func (r *serverResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var config, plan, state serverResourceModel

	if req.Plan.Raw.IsNull() {
		// Destruction plan, no modification needed
		return
	}

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// `disks` includes the primary, so more than one element means an additional disk.
	warnDisksMayRaceUserData := func() {
		if plan.UserData.ValueString() == "" || plan.Disks.IsNull() || plan.Disks.IsUnknown() || len(plan.Disks.Elements()) < 2 {
			return
		}
		resp.Diagnostics.AddAttributeWarning(
			path.Root("disks"),
			"Additional disks may race with user_data on first boot",
			"Additional disks are attached after the server boots into a fresh OS install (on create, or a "+
				"rebuild triggered by this plan), which can interrupt `user_data` (cloud-init) while it is "+
				"still running on first boot. cloud-init only runs user_data once per instance, so any "+
				"scripts that did not complete before the disk operations will not re-run on the next boot.",
		)
	}

	if plan.SourceAndDestinationCheck.IsUnknown() {
		if plan.VpcId.IsNull() {
			plan.SourceAndDestinationCheck = types.BoolNull()
		} else {
			plan.SourceAndDestinationCheck = types.BoolPointerValue(Pointer(true))
		}
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
	}

	if plan.SeparatePrivateNetworkInterface.IsUnknown() {
		if plan.VpcId.IsNull() {
			plan.SeparatePrivateNetworkInterface = types.BoolNull()
		} else {
			plan.SeparatePrivateNetworkInterface = types.BoolPointerValue(Pointer(false))
		}
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
	}

	if plan.VpcIpv4Address.IsUnknown() {
		if plan.VpcId.IsNull() {
			plan.VpcIpv4Address = types.StringNull()
		} else {
			plan.VpcIpv4Address = types.StringUnknown()
		}
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
	}

	if req.State.Raw.IsNull() {
		warnDisksMayRaceUserData()

		// Creation plan, no further modification needed
		return
	}

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.VpcId.Equal(state.VpcId) || !plan.VpcIpv4Address.Equal(state.VpcIpv4Address) {
		if config.VpcIpv4Address.IsNull() {
			plan.VpcIpv4Address = types.StringUnknown()
		}
		plan.PrivateIPv4Addresses = types.ListUnknown(plan.PrivateIPv4Addresses.ElementType(ctx))
	}

	// When IP count is changed, plan should show addition/removal of public IPs
	plannedPublicIpV4Count := int(plan.PublicIpv4Count.ValueInt32())
	plannedPublicIpAddresses := make([]attr.Value, plannedPublicIpV4Count)
	stateIpV4Addresses := []*string{}
	diags := state.PublicIpv4Addresses.ElementsAs(ctx, &stateIpV4Addresses, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	for i := range plannedPublicIpAddresses {
		if i < len(stateIpV4Addresses) {
			plannedPublicIpAddresses[i] = types.StringValue(*stateIpV4Addresses[i])
		} else {
			plannedPublicIpAddresses[i] = types.StringUnknown()
		}
	}
	plan.PublicIpv4Addresses, diags = types.ListValueFrom(ctx, types.StringType, plannedPublicIpAddresses)
	resp.Diagnostics.Append(diags...)

	// Add warning if rebuild is required
	attrsRequiringRebuild := attrsRequiringRebuild(&plan, &state)
	if len(attrsRequiringRebuild) > 0 {
		resp.Diagnostics.AddWarning(
			"Server Rebuild Required",
			fmt.Sprintf(
				"Server %d will lose all data if this Terraform plan is applied, because of modified attribute(s): %s",
				state.Id.ValueInt64(),
				strings.Join(attrsRequiringRebuild, ", "),
			),
		)
		warnDisksMayRaceUserData()
	}

	if !plan.Ipv6.Equal(state.Ipv6) {
		if plan.Ipv6.ValueBool() {
			plan.PublicIpv6Addresses = types.ListUnknown(state.PublicIpv6Addresses.ElementType(ctx))
			plan.PrivateIpv6Addresses = types.ListUnknown(state.PrivateIpv6Addresses.ElementType(ctx))
		} else {
			plan.PublicIpv6Addresses = types.ListNull(state.PublicIpv6Addresses.ElementType(ctx))
			plan.PrivateIpv6Addresses = types.ListNull(state.PrivateIpv6Addresses.ElementType(ctx))
		}
	} else {
		// Computed list attributes aren't carried forward from state by default, so without
		// this these would plan as unknown on every apply, even when ipv6 didn't change.
		plan.PublicIpv6Addresses = state.PublicIpv6Addresses
		plan.PrivateIpv6Addresses = state.PrivateIpv6Addresses
	}

	// Use state for unknown disk/memory values, as long as server size is the same
	if (plan.Memory.IsNull() || plan.Memory.IsUnknown()) && plan.Size.Equal(state.Size) {
		plan.Memory = state.Memory
	}
	if (plan.Disk.IsNull() || plan.Disk.IsUnknown()) && plan.Size.Equal(state.Size) {
		plan.Disk = state.Disk
	}

	// A configured `disks` is planned by the attribute's own plan modifier, which matches each
	// disk with the one it corresponds to in prior state. Left unset, the disks are Binary Lane's
	// to manage, and what happens to them depends on `disk` — resolved just above, after that
	// plan modifier has already run.
	if config.Disks.IsNull() || config.Disks.IsUnknown() {
		// Computed list attributes aren't carried forward from state by default, so without this
		// `disks` would plan as unknown on every apply, even when nothing about it changed.
		plan.Disks = state.Disks

		if state.Disks.IsNull() || state.Disks.IsUnknown() {
			plan.Disks = types.ListUnknown(resources.DisksValue{}.Type(ctx))
		} else if len(state.Disks.Elements()) == 1 && !plan.Disk.Equal(state.Disk) {
			// A resize resizes the primary disk to fill the new total, but only when it is the
			// server's only disk: with an additional disk it leaves every disk alone.
			plan.Disks = types.ListUnknown(resources.DisksValue{}.Type(ctx))

			statePrimary, _, diags := splitDisks(ctx, state.Disks)
			resp.Diagnostics.Append(diags...)
			if statePrimary != nil && !plan.Disk.IsUnknown() &&
				float64(plan.Disk.ValueInt32()) < statePrimary.SizeGigabytes {
				resp.Diagnostics.Append(resources.DiskShrinkWarning(path.Root("disk"), "The primary disk",
					int64(statePrimary.SizeGigabytes), int64(plan.Disk.ValueInt32())))
			}
		}
	}

	if isAdvFeatChanged(&config.AdvancedFeatures, &state.AdvancedFeatures) {
		advFeatResp, err := r.bc.client.GetServersServerIdAvailableAdvancedFeaturesWithResponse(ctx, state.Id.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error fetching available advanced features for server: id=%s", state.Id.String()),
				err.Error(),
			)
		} else if advFeatResp.StatusCode() != http.StatusOK {
			resp.Diagnostics.AddError(
				"Unexpected HTTP status code fetching available advanced features for server",
				fmt.Sprintf("Received %s fetching available advanced features for server: id=%s. Details: %s", advFeatResp.Status(), state.Id.String(), advFeatResp.Body))
		} else {
			debugAdvFeatResp, _ := json.Marshal(advFeatResp.JSON200)
			tflog.Debug(ctx, fmt.Sprintf("Recieved response for available advanced features: %s", debugAdvFeatResp))

			availableAdvFeat := advFeatResp.JSON200.AvailableAdvancedServerFeatures.AdvancedFeatures

			if plan.AdvancedFeatures.EmulatedHyperv.ValueBool() && !slices.Contains(availableAdvFeat, "emulated-hyperv") {
				resp.Diagnostics.AddAttributeError(
					path.Root("advanced_features").AtName("emulated_hyperv"),
					"Emulated Hyper-V not available",
					fmt.Sprintf("Emulated Hyper-V is not available for server %s", state.Name.String()),
				)
			} else if plan.AdvancedFeatures.EmulatedDevices.ValueBool() && !slices.Contains(availableAdvFeat, "emulated-devices") {
				resp.Diagnostics.AddAttributeError(
					path.Root("advanced_features").AtName("emulated_devices"),
					"Emulated Devices not available",
					fmt.Sprintf("Emulated Devices is not available for server %s", state.Name.String()),
				)
			} else if plan.AdvancedFeatures.EmulatedTpm.ValueBool() && !slices.Contains(availableAdvFeat, "emulated-tpm") {
				resp.Diagnostics.AddAttributeError(
					path.Root("advanced_features").AtName("emulated_tpm"),
					"Emulated TPM not available",
					fmt.Sprintf("Emulated TPM is not available for server %s", state.Name.String()),
				)
			} else if plan.AdvancedFeatures.NestedVirt.ValueBool() && !slices.Contains(availableAdvFeat, "nested-virt") {
				resp.Diagnostics.AddAttributeError(
					path.Root("advanced_features").AtName("nested_virt"),
					"Nested Virtualization not available",
					fmt.Sprintf("Nested Virtualization is not available for server %s", state.Name.String()),
				)
			} else if plan.AdvancedFeatures.DriverDisk.ValueBool() && !slices.Contains(availableAdvFeat, "driver-disk") {
				resp.Diagnostics.AddAttributeError(
					path.Root("advanced_features").AtName("driver_disk"),
					"Driver Disk not available",
					fmt.Sprintf("Driver Disk is not available for server %s", state.Name.String()),
				)
			} else if plan.AdvancedFeatures.UnsetUuid.ValueBool() && !slices.Contains(availableAdvFeat, "unset-uuid") {
				resp.Diagnostics.AddAttributeError(
					path.Root("advanced_features").AtName("unset_uuid"),
					"Unset UUID not available",
					fmt.Sprintf("Unset UUID is not available for server %s", state.Name.String()),
				)
			} else if plan.AdvancedFeatures.LocalRtc.ValueBool() && !slices.Contains(availableAdvFeat, "local-rtc") {
				resp.Diagnostics.AddAttributeError(
					path.Root("advanced_features").AtName("local_rtc"),
					"Local RTC not available",
					fmt.Sprintf("Local RTC is not available for server %s", state.Name.String()),
				)
			}
		}
	} else {
		// Computed object attributes aren't carried forward from state by default, so without
		// this `advanced_features` would plan as unknown on every apply, even when none of its
		// writable fields changed.
		plan.AdvancedFeatures = state.AdvancedFeatures
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

func (r *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var config, data serverResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := data.Timeouts.Create(ctx, 20*time.Minute)
	resp.Diagnostics.Append(diags...)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create API call logic
	tflog.Debug(ctx, fmt.Sprintf("Creating server: name=%s", data.Name.ValueString()))

	sshKeys := []int{}
	diags = data.SshKeys.ElementsAs(ctx, &sshKeys, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := binarylane.CreateServerRequest{
		Name:         data.Name.ValueStringPointer(),
		Image:        data.Image.ValueString(),
		Region:       data.Region.ValueString(),
		Size:         data.Size.ValueString(),
		UserData:     data.UserData.ValueStringPointer(),
		VpcId:        data.VpcId.ValueInt64Pointer(),
		PortBlocking: data.PortBlocking.ValueBoolPointer(),
		SshKeys:      &sshKeys,
		Options: &binarylane.SizeOptionsRequest{
			Ipv4Addresses: data.PublicIpv4Count.ValueInt32Pointer(),
		},
		Backups:                         data.Backups.ValueBoolPointer(),
		Ipv6:                            data.Ipv6.ValueBoolPointer(),
		SeparatePrivateNetworkInterface: data.SeparatePrivateNetworkInterface.ValueBoolPointer(),
	}

	// The planned `disks` is unknown when it isn't configured, so the disks to create are read
	// from config, which is always known.
	configPrimary, configAdditional, diskDiags := splitDisks(ctx, config.Disks)
	resp.Diagnostics.Append(diskDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Memory.IsNull() && !data.Memory.IsUnknown() {
		body.Options.Memory = data.Memory.ValueInt32Pointer()
	}
	if !data.Disk.IsNull() && !data.Disk.IsUnknown() {
		body.Options.Disk = data.Disk.ValueInt32Pointer()
	}
	if !data.VpcIpv4Address.IsNull() && !data.VpcIpv4Address.IsUnknown() {
		body.VpcIpv4Address = data.VpcIpv4Address.ValueStringPointer()
	}
	if data.Password.IsNull() {
		data.Password = types.StringNull()
	} else {
		body.Password = data.Password.ValueStringPointer()
		ctx = tflog.MaskMessageStrings(ctx, data.Password.String())
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
			fmt.Sprintf("Received %s creating new server: name=%s. Details: %s", serverResp.Status(), data.Name.ValueString(), serverResp.Body),
		)
		return
	}

	// Wait for server to be ready
	var createActionId int64
	for _, action := range serverResp.JSON200.Links.Actions {
		if action.Rel == "create" {
			createActionId = action.Id
			break
		}
	}
	if createActionId == 0 {
		resp.Diagnostics.AddError(
			"Unable to wait for server to be created, links.actions with rel=create missing from response",
			fmt.Sprintf("Received %s creating new server: name=%s. Details: %s", serverResp.Status(), data.Name.ValueString(), serverResp.Body))
		return
	}
	err = r.waitForServerAction(ctx, serverResp.JSON200.Server.Id, createActionId)
	if err != nil {
		resp.Diagnostics.AddError("Error waiting for server to be created", err.Error())
	}

	data.Id = types.Int64Value(serverResp.JSON200.Server.Id)
	data.Name = types.StringValue(serverResp.JSON200.Server.Name)
	data.Image = types.StringValue(*serverResp.JSON200.Server.Image.Slug)
	data.Region = types.StringValue(serverResp.JSON200.Server.Region.Slug)
	data.Size = types.StringValue(serverResp.JSON200.Server.Size.Slug)
	data.Backups = types.BoolValue(serverResp.JSON200.Server.NextBackupWindow != nil)
	data.Ipv6 = types.BoolValue(len(serverResp.JSON200.Server.Networks.V6) > 0)
	data.PortBlocking = types.BoolValue(serverResp.JSON200.Server.Networks.PortBlocking)
	data.VpcId = types.Int64PointerValue(serverResp.JSON200.Server.VpcId)
	data.Permalink = types.StringValue(*serverResp.JSON200.Server.Permalink)
	data.PasswordChangeSupported = types.BoolValue(serverResp.JSON200.Server.PasswordChangeSupported)
	data.Memory = types.Int32Value(serverResp.JSON200.Server.Memory)
	data.Disk = types.Int32Value(serverResp.JSON200.Server.Disk)
	data.Disks, diags = serverDisksToList(ctx, serverResp.JSON200.Server.Disks, config.Disks)
	resp.Diagnostics.Append(diags...)
	plannedSourceDestCheck := data.SourceAndDestinationCheck
	serverRespSourceDestCheck := types.BoolPointerValue(serverResp.JSON200.Server.Networks.SourceAndDestinationCheck)
	data.SourceAndDestinationCheck = serverRespSourceDestCheck
	data.SeparatePrivateNetworkInterface = types.BoolPointerValue(serverResp.JSON200.Server.Networks.SeparatePrivateNetworkInterface)

	if serverResp.JSON200.Server.VpcId == nil {
		data.VpcIpv4Address = types.StringNull()
	} else if len(serverResp.JSON200.Server.Networks.V4) > 0 {
		for _, v4address := range serverResp.JSON200.Server.Networks.V4 {
			// Skip addresses in 172.21.0.0/16, these are BL internal addresses that are not part of the user's VPC
			if v4address.Type == "private" && !strings.HasPrefix(v4address.IpAddress, "172.21.") {
				data.VpcIpv4Address = types.StringValue(v4address.IpAddress)
				break
			}
		}
	}

	advFeat := serverResp.JSON200.Server.AdvancedFeatures.EnabledAdvancedFeatures
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
		resp.Diagnostics.Append(diags...)
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
	data.PublicIpv4Addresses, diags = types.ListValueFrom(ctx, types.StringType, publicIpv4Addresses)
	resp.Diagnostics.Append(diags...)
	data.PrivateIPv4Addresses, diags = types.ListValueFrom(ctx, types.StringType, privateIpv4Addresses)
	resp.Diagnostics.Append(diags...)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update advanced features
	err = r.updateAdvancedFeatures(ctx, data.Id.ValueInt64(), &config.AdvancedFeatures, &data.AdvancedFeatures)
	if err != nil {
		resp.Diagnostics.AddError("Error updating advanced features", err.Error())
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update source_and_destination_check if needed
	if plannedSourceDestCheck.Equal(types.BoolPointerValue(Pointer(false))) {
		err := r.updateSourceDestCheck(ctx, data.Id.ValueInt64(), false)
		if err != nil {
			resp.Diagnostics.AddError("Error updating source and destination check", err.Error())
			return
		}
		data.SourceAndDestinationCheck = plannedSourceDestCheck
	}

	// Carve the additional disks out of the server's total: it is created with a single primary
	// disk spanning all of it, so shrink that to make room, as Binary Lane's own UI does
	// (https://support.binarylane.com.au/support/solutions/articles/11000133468). If cloud-init's
	// growpart has already expanded the filesystem to fill the original primary, this can
	// corrupt it, and a disk can only be resized in place.
	if configPrimary != nil {
		finalDisks, err := r.reconcileServerDisks(
			ctx, data.Id.ValueInt64(), serverResp.JSON200.Server.Disks,
			configPrimary.SizeGigabytes, configAdditional, true,
		)
		data.Disks, diags = serverDisksToList(ctx, finalDisks, config.Disks)
		resp.Diagnostics.Append(diags...)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		if err != nil {
			resp.Diagnostics.AddError("Error creating additional disks", err.Error())
			return
		}
	}

	// One extra read to check the final state of enabled_advanced_features, needed because
	// some flags (like "cloud-init") are not set until the server is fully created. See #13
	diag := r.fetchServerResourceState(ctx, &data)
	resp.Diagnostics.Append(diag...)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data serverResourceModel

	// Read Terraform prior state data into the model
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	tflog.Debug(ctx, fmt.Sprintf("Reading server: id=%s, name=%s", data.Id.String(), data.Name.ValueString()))
	diag := r.fetchServerResourceState(ctx, &data)
	resp.Diagnostics.Append(diag...)

	// Get user data script
	userDataResp, err := r.bc.client.GetServersServerIdUserDataWithResponse(ctx, data.Id.ValueInt64())
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

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var config, plan, state serverResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Update(ctx, 20*time.Minute)
	resp.Diagnostics.Append(diags...)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	rebuildNeeded := len(attrsRequiringRebuild(&plan, &state)) > 0
	refreshNeeded := false

	defer (func() {
		if !refreshNeeded {
			return
		}

		diag := r.fetchServerResourceState(ctx, &state)
		resp.Diagnostics.Append(diag...)

		// Save updated data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	})()

	// Rename
	if !plan.Name.Equal(state.Name) && !rebuildNeeded {
		renameResp, err := r.bc.client.PostServersServerIdActionsRenameWithResponse(
			ctx,
			state.Id.ValueInt64(),
			binarylane.PostServersServerIdActionsRenameJSONRequestBody{
				Type: "rename",
				Name: plan.Name.ValueString(),
			},
		)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error renaming server: server_id=%s", state.Id.String()),
				err.Error(),
			)
			return
		} else if renameResp.StatusCode() != http.StatusOK && renameResp.StatusCode() != http.StatusAccepted {
			resp.Diagnostics.AddError(
				"Unexpected HTTP status code renaming server",
				fmt.Sprintf("Received %s renaming server: server_id=%s. Details: %s", renameResp.Status(), state.Id.String(), renameResp.Body))
			return
		}

		// TODO - Currently, the API does not support polling for the rename to complete, because there is no action ID in 202 accepted response (see #13)

		state.Name = types.StringValue(plan.Name.ValueString())
	}

	// Change network
	if !plan.VpcId.Equal(state.VpcId) {
		networkResp, err := r.bc.client.PostServersServerIdActionsChangeNetworkWithResponse(
			ctx,
			state.Id.ValueInt64(),
			binarylane.PostServersServerIdActionsChangeNetworkJSONRequestBody{
				Type:  "change_network",
				VpcId: plan.VpcId.ValueInt64Pointer(),
			},
		)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error changing network for server: server_id=%s", state.Id.String()),
				err.Error())
			return
		}
		if networkResp.StatusCode() != http.StatusOK {
			resp.Diagnostics.AddError(
				"Unexpected HTTP status code changing network for server",
				fmt.Sprintf("Received %s changing network for server: server_id=%s. Details: %s", networkResp.Status(), state.Id.String(), networkResp.Body))
			return
		}
		err = r.waitForServerAction(ctx, state.Id.ValueInt64(), networkResp.JSON200.Action.Id)
		if err != nil {
			resp.Diagnostics.AddError("Error waiting for server to change network", err.Error())
			return
		}
		state.VpcId = plan.VpcId
		refreshNeeded = true // Refresh to get new private IP addresses
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Change VPC IPv4 address
	if !plan.VpcIpv4Address.Equal(state.VpcIpv4Address) && !plan.VpcIpv4Address.IsNull() && !plan.VpcIpv4Address.IsUnknown() {
		// If VPC was just changed, IP address needs to be fetched so it can be sent in the request
		if refreshNeeded {
			diag := r.fetchServerResourceState(ctx, &state)
			resp.Diagnostics.Append(diag...)
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			if resp.Diagnostics.HasError() {
				return
			}
		}

		// We might have got lucky - the new randomly assigned IP address might be the one we want, so check before making API call
		if state.VpcIpv4Address.ValueString() != plan.VpcIpv4Address.ValueString() {
			vpcIpv4Resp, err := r.bc.client.PostServersServerIdActionsChangeVpcIpv4WithResponse(
				ctx,
				state.Id.ValueInt64(),
				binarylane.PostServersServerIdActionsChangeVpcIpv4JSONRequestBody{
					Type:               "change_vpc_ipv4",
					CurrentIpv4Address: state.VpcIpv4Address.ValueString(),
					NewIpv4Address:     plan.VpcIpv4Address.ValueString(),
				},
			)
			if err != nil {
				resp.Diagnostics.AddError(
					fmt.Sprintf("Error changing VPC IPv4 address for server: server_id=%s", state.Id.String()),
					err.Error())
				return
			}
			if vpcIpv4Resp.StatusCode() != http.StatusOK {
				resp.Diagnostics.AddError(
					"Unexpected HTTP status code changing VPC IPv4 address for server",
					fmt.Sprintf("Received %s changing VPC IPv4 address for server: server_id=%s. Details: %s", vpcIpv4Resp.Status(), state.Id.String(), vpcIpv4Resp.Body))
				return
			}
			err = r.waitForServerAction(ctx, state.Id.ValueInt64(), vpcIpv4Resp.JSON200.Action.Id)
			if err != nil {
				resp.Diagnostics.AddError("Error waiting for VPC IPv4 address to change", err.Error())
				return
			}
			state.VpcIpv4Address = plan.VpcIpv4Address
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			refreshNeeded = true // Refresh to populate private_ipv4_addresses
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}

	// The disks to reconcile come from config, which describes every disk when `disks` is set.
	// Left unset, the disks are Binary Lane's to manage and there is nothing to reconcile.
	configPrimary, configAdditional, diags := splitDisks(ctx, config.Disks)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	planDiskKnown := !plan.Disk.IsNull() && !plan.Disk.IsUnknown()
	diskChanged := planDiskKnown && !plan.Disk.Equal(state.Disk)
	sizeChanged := !plan.Size.Equal(state.Size)

	// A `disk` or `size` change resizes the server, which can resize the primary disk with it, so
	// the disks are reconciled back to config after one of those too.
	reconcileDisks := configPrimary != nil &&
		(!plan.Disks.Equal(state.Disks) || diskChanged || sizeChanged)

	persistDisks := func(disks []binarylane.Disk) {
		if disks == nil {
			return
		}
		var diags diag.Diagnostics
		state.Disks, diags = serverDisksToList(ctx, disks, plan.Disks)
		resp.Diagnostics.Append(diags...)
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}

	// Free space before the resize below, which is where space to grow or add a disk comes from.
	var currentDisks []binarylane.Disk
	if reconcileDisks {
		var err error
		currentDisks, err = r.reconcileServerDisks(
			ctx, state.Id.ValueInt64(), nil, configPrimary.SizeGigabytes, configAdditional, false)
		persistDisks(currentDisks)
		if err != nil {
			resp.Diagnostics.AddError("Error reconciling server disks", err.Error())
			return
		}
	}

	// Resize operation
	memoryChanged := !plan.Memory.IsNull() && !plan.Memory.IsUnknown() && !plan.Memory.Equal(state.Memory)
	if sizeChanged || memoryChanged ||
		!plan.Image.Equal(state.Image) ||
		!plan.PublicIpv4Count.Equal(state.PublicIpv4Count) ||
		diskChanged {

		resizeReq := &binarylane.PostServersServerIdActionsResizeJSONRequestBody{
			Type:    "resize",
			Options: &binarylane.ChangeSizeOptionsRequest{},
		}

		if sizeChanged || memoryChanged || diskChanged {
			if sizeChanged {
				resizeReq.Size = plan.Size.ValueStringPointer()
			}
			if !plan.Memory.IsUnknown() && !plan.Memory.IsNull() {
				resizeReq.Options.Memory = plan.Memory.ValueInt32Pointer()
			}
			if planDiskKnown {
				disk := plan.Disk.ValueInt32()
				resizeReq.Options.Disk = &disk
			}
			state.Size = plan.Size
			state.Memory = plan.Memory
			state.Disk = plan.Disk
		}

		if !plan.Image.Equal(state.Image) {
			resizeReq.ChangeImage = &binarylane.ChangeImage{
				Image: plan.Image.ValueStringPointer(),
			}
			state.Image = plan.Image
		}

		if !plan.PublicIpv4Count.Equal(state.PublicIpv4Count) {
			resizeReq.Options.Ipv4Addresses = plan.PublicIpv4Count.ValueInt32Pointer()
			if plan.PublicIpv4Count.ValueInt32() < state.PublicIpv4Count.ValueInt32() {
				currentIps := []string{}
				diags := state.PublicIpv4Addresses.ElementsAs(ctx, &currentIps, false)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				removedIps := currentIps[plan.PublicIpv4Count.ValueInt32():state.PublicIpv4Count.ValueInt32()]
				resizeReq.Options.Ipv4AddressesToRemove = &removedIps
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			}
			state.PublicIpv4Count = plan.PublicIpv4Count
			state.PublicIpv4Addresses = plan.PublicIpv4Addresses
		}

		tflog.Info(ctx, fmt.Sprintf("Resizing server: server_id=%s", state.Id.String()))

		resizeResp, err := r.bc.client.PostServersServerIdActionsResizeWithResponse(
			ctx,
			state.Id.ValueInt64(),
			*resizeReq,
		)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error resizing server: server_id=%s", state.Id.String()),
				err.Error(),
			)
			return
		}
		if resizeResp.StatusCode() != http.StatusOK {
			resp.Diagnostics.AddError(
				"Unexpected HTTP status code resizing server",
				fmt.Sprintf("Received %s resizing server: server_id=%s. Details: %s", resizeResp.Status(), state.Id.String(), resizeResp.Body))
			return
		}

		err = r.waitForServerAction(ctx, state.Id.ValueInt64(), resizeResp.JSON200.Action.Id)
		if err != nil {
			resp.Diagnostics.AddError("Error waiting for server to be resized", err.Error())
			return
		}

		// The resize can resize the primary disk, so anything read before it is stale.
		currentDisks = nil

		if state.PublicIpv4Addresses.IsUnknown() || listContainsUnknown(ctx, state.PublicIpv4Addresses) || state.Memory.IsNull() || state.Memory.IsUnknown() {
			refreshNeeded = true
		}

		// Save updated data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Grow the primary back to its target size, then grow and add the rest, now that the resize
	// has made room. currentDisks is nil if that resize ran, so they are re-read.
	if reconcileDisks {
		finalDisks, err := r.reconcileServerDisks(
			ctx, state.Id.ValueInt64(), currentDisks, configPrimary.SizeGigabytes, configAdditional, true)
		persistDisks(finalDisks)
		if err != nil {
			resp.Diagnostics.AddError("Error reconciling server disks", err.Error())
			return
		}
	}

	// Change IPv6
	if !plan.Ipv6.Equal(state.Ipv6) {
		ipv6Resp, err := r.bc.client.PostServersServerIdActionsChangeIpv6WithResponse(
			ctx,
			state.Id.ValueInt64(),
			binarylane.PostServersServerIdActionsChangeIpv6JSONRequestBody{
				Type:    "change_ipv6",
				Enabled: plan.Ipv6.ValueBool(),
			},
		)
		if err != nil {
			resp.Diagnostics.AddError("Error changing IPv6", err.Error())
			return
		}

		err = r.waitForServerAction(ctx, state.Id.ValueInt64(), ipv6Resp.JSON200.Action.Id)
		if err != nil {
			resp.Diagnostics.AddError("Error waiting for changing IPv6", err.Error())
			return
		}

		state.Ipv6 = plan.Ipv6
		state.PublicIpv6Addresses = plan.PublicIpv6Addresses
		state.PrivateIpv6Addresses = plan.PrivateIpv6Addresses

		if state.PublicIpv6Addresses.IsUnknown() || state.PrivateIpv6Addresses.IsUnknown() || listContainsUnknown(ctx, state.PublicIpv6Addresses) || listContainsUnknown(ctx, state.PrivateIpv6Addresses) {
			refreshNeeded = true
		}

		// Save updated data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}

	}

	// Rebuild operation
	if !plan.SshKeys.Equal(state.SshKeys) || !plan.UserData.IsNull() && !plan.UserData.Equal(state.UserData) {
		var rebuildReq *binarylane.PostServersServerIdActionsRebuildJSONRequestBody

		sshKeys := []int{}
		diags := plan.SshKeys.ElementsAs(ctx, &sshKeys, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		rebuildReq = &binarylane.PostServersServerIdActionsRebuildJSONRequestBody{
			Type: "rebuild",
			Options: &binarylane.ImageOptions{
				Name:     plan.Name.ValueStringPointer(),
				Password: plan.Password.ValueStringPointer(),
				UserData: plan.UserData.ValueStringPointer(),
				SshKeys:  &sshKeys,
			},
		}
		rebuildResp, err := r.bc.client.PostServersServerIdActionsRebuildWithResponse(
			ctx,
			state.Id.ValueInt64(),
			*rebuildReq,
		)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error rebuilding server: server_id=%s", state.Id.String()),
				err.Error(),
			)
			return
		}
		if rebuildResp.StatusCode() != http.StatusOK {
			resp.Diagnostics.AddError(
				"Unexpected HTTP status code rebuilding server",
				fmt.Sprintf("Received %s rebuilding server: server_id=%s. Details: %s", rebuildResp.Status(), state.Id.String(), rebuildResp.Body))
			return
		}
		err = r.waitForServerAction(ctx, state.Id.ValueInt64(), rebuildResp.JSON200.Action.Id)
		if err != nil {
			resp.Diagnostics.AddError("Error waiting for server to be rebuilt", err.Error())
			return
		}
		state.Name = plan.Name
		state.Password = plan.Password
		state.UserData = plan.UserData
		state.SshKeys = plan.SshKeys

		// Save updated data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else
	// Reset Password (only needed if server didn't rebuild)
	if !plan.Password.Equal(state.Password) {
		passwordResp, err := r.bc.client.PostServersServerIdActionsPasswordResetWithResponse(ctx, state.Id.ValueInt64(),
			binarylane.PasswordReset{
				Type:     "password_reset",
				Password: plan.Password.ValueStringPointer(),
			},
		)
		if err != nil {
			resp.Diagnostics.AddError("Error resetting password", err.Error())
			return
		}
		err = r.waitForServerAction(ctx, state.Id.ValueInt64(), passwordResp.JSON200.Action.Id)
		if err != nil {
			resp.Diagnostics.AddError("Error waiting for password reset", err.Error())
			return
		}
		state.Password = plan.Password

		// Save updated data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	err := r.updateAdvancedFeatures(ctx, state.Id.ValueInt64(), &config.AdvancedFeatures, &state.AdvancedFeatures)
	if err != nil {
		resp.Diagnostics.AddError("Error updating advanced features", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check source_and_destination_check
	if !plan.SourceAndDestinationCheck.Equal(state.SourceAndDestinationCheck) {
		if !plan.SourceAndDestinationCheck.IsNull() {
			err := r.updateSourceDestCheck(ctx, state.Id.ValueInt64(), plan.SourceAndDestinationCheck.ValueBool())
			if err != nil {
				resp.Diagnostics.AddError("Error updating source and destination check", err.Error())
				return
			}
		}

		state.SourceAndDestinationCheck = plan.SourceAndDestinationCheck

		// Save updated data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Check separate_private_network_interface
	if !plan.SeparatePrivateNetworkInterface.Equal(state.SeparatePrivateNetworkInterface) {
		if !plan.SeparatePrivateNetworkInterface.IsNull() {
			err := r.updateSeparatePrivateNetworkInterface(ctx, state.Id.ValueInt64(), plan.SeparatePrivateNetworkInterface.ValueBool())
			if err != nil {
				resp.Diagnostics.AddError("Error updating separate private network interface", err.Error())
				return
			}
		}

		state.SeparatePrivateNetworkInterface = plan.SeparatePrivateNetworkInterface

		// Save updated data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Enable or disable backups
	if !plan.Backups.Equal(state.Backups) {
		if plan.Backups.ValueBool() {
			backupResp, err := r.bc.client.PostServersServerIdActionsEnableBackupsWithResponse(ctx, state.Id.ValueInt64(), binarylane.EnableBackups{
				Type: "enable_backups",
			})
			if err != nil {
				resp.Diagnostics.AddError("Error enabling backups", err.Error())
				return
			}
			err = r.waitForServerAction(ctx, state.Id.ValueInt64(), backupResp.JSON200.Action.Id)
			if err != nil {
				resp.Diagnostics.AddError("Error waiting for backups to be enabled", err.Error())
				return
			}
		} else {
			backupResp, err := r.bc.client.PostServersServerIdActionsDisableBackupsWithResponse(ctx, state.Id.ValueInt64(), binarylane.DisableBackups{
				Type: "disable_backups",
			})
			if err != nil {
				resp.Diagnostics.AddError("Error disabling backups", err.Error())
				return
			}
			err = r.waitForServerAction(ctx, state.Id.ValueInt64(), backupResp.JSON200.Action.Id)
			if err != nil {
				resp.Diagnostics.AddError("Error waiting for backups to be disabled", err.Error())
				return
			}
		}
		state.Backups = plan.Backups

		// Save updated data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Change port blocking
	if !plan.PortBlocking.Equal(state.PortBlocking) {
		portBlockingResp, err := r.bc.client.PostServersServerIdActionsChangePortBlockingWithResponse(ctx, state.Id.ValueInt64(),
			binarylane.ChangePortBlocking{
				Type:    "change_port_blocking",
				Enabled: plan.PortBlocking.ValueBool(),
			},
		)
		if err != nil {
			resp.Diagnostics.AddError("Error changing \"port_blocking\" attribute", err.Error())
			return
		}
		err = r.waitForServerAction(ctx, state.Id.ValueInt64(), portBlockingResp.JSON200.Action.Id)
		if err != nil {
			resp.Diagnostics.AddError("Error waiting for \"port_blocking\" attribute to change", err.Error())
			return
		}
		state.PortBlocking = plan.PortBlocking

		// Save updated data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}

func (r *serverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data serverResourceModel

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

func (r *serverResource) ImportState(
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

	name := req.ID
	params := binarylane.GetServersParams{
		Hostname: &name,
	}

	serverResp, err := r.bc.client.GetServersWithResponse(ctx, &params)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error getting server: hostname=%s", name), err.Error())
		return
	}

	if serverResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected HTTP status code getting server",
			fmt.Sprintf("Received %s reading server: hostname=%s. Details: %s", serverResp.Status(), name,
				serverResp.Body))
		return
	}

	servers := serverResp.JSON200.Servers
	idx := slices.IndexFunc(servers, func(s binarylane.Server) bool { return s.Name == name })
	if idx == -1 {
		resp.Diagnostics.AddError(
			"Could not find server by hostname",
			fmt.Sprintf("Error finding server: hostname=%s", name),
		)
		return
	}
	server := servers[idx]

	diags := resp.State.SetAttribute(ctx, path.Root("id"), int32(server.Id))
	resp.Diagnostics.Append(diags...)
}

func (r *serverResource) waitForServerAction(ctx context.Context, serverId int64, actionId int64) error {
	var lastReadyResp *binarylane.GetServersServerIdActionsActionIdResponse

	for {
		select {
		case <-ctx.Done():
			if lastReadyResp == nil {
				return fmt.Errorf("timed out waiting for server action: server_id=%d, action_id=%d", serverId, actionId)
			} else {
				return fmt.Errorf("timed out waiting for server action: server_id=%d, action_id=%d, last response was status=%s, body: %s",
					serverId, actionId, lastReadyResp.Status(), lastReadyResp.Body)
			}
		default:
			readyResp, err := r.bc.client.GetServersServerIdActionsActionIdWithResponse(ctx, serverId, actionId)
			if err != nil {
				return fmt.Errorf("unexpected error waiting for server action: server_id=%d, action_id=%d, error: %w", serverId, actionId, err)
			}
			if readyResp.StatusCode() == http.StatusOK && readyResp.JSON200.Action.Status == binarylane.Errored {
				return fmt.Errorf("server action failed to with error: server_id=%d, action_id=%d, error: %s", serverId, actionId, readyResp.Body)
			}
			if readyResp.StatusCode() == http.StatusOK && readyResp.JSON200.Action.CompletedAt != nil {
				return nil
			}
			lastReadyResp = readyResp
			tflog.Debug(ctx,
				fmt.Sprintf("waiting for server action for server_id=%d, action_id=%d: last response was status=%s, details: %s",
					serverId, actionId, readyResp.Status(), readyResp.Body,
				),
			)
		}
		time.Sleep(time.Second * 5)
	}
}

func attrsRequiringRebuild(plan *serverResourceModel, state *serverResourceModel) []string {
	attrs := []string{}

	if !plan.SshKeys.Equal(state.SshKeys) {
		attrs = append(attrs, "ssh_keys")
	}
	if !plan.Image.Equal(state.Image) {
		attrs = append(attrs, "image")
	}
	if !plan.UserData.IsNull() && !plan.UserData.Equal(state.UserData) {
		attrs = append(attrs, "user_data")
	}

	return attrs
}

func (r *serverResource) updateSourceDestCheck(
	ctx context.Context,
	serverId int64,
	sourceDestCheckEnabled bool,
) error {
	tflog.Info(ctx, fmt.Sprintf("Changing source and destination check for server: server_id=%d, enabled=%t",
		serverId, sourceDestCheckEnabled))

	sourceDestCheckResp, err := r.bc.client.PostServersServerIdActionsChangeSourceAndDestinationCheckWithResponse(
		ctx,
		serverId,
		binarylane.PostServersServerIdActionsChangeSourceAndDestinationCheckJSONRequestBody{
			Type:    "change_source_and_destination_check",
			Enabled: sourceDestCheckEnabled,
		},
	)
	if err != nil {
		return fmt.Errorf("error changing source and destination check for server: server_id=%d, error: %w", serverId, err)
	}
	if sourceDestCheckResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status code changing source and destination check for server: server_id=%d, details: %s", serverId, sourceDestCheckResp.Body)
	}

	err = r.waitForServerAction(ctx, serverId, sourceDestCheckResp.JSON200.Action.Id)
	if err != nil {
		return fmt.Errorf("error changing source and destination check: %w", err)
	}

	return nil
}

func (r *serverResource) updateSeparatePrivateNetworkInterface(
	ctx context.Context,
	serverId int64,
	enabled bool,
) error {
	tflog.Info(ctx, fmt.Sprintf("Changing separate private network interface for server: server_id=%d, enabled=%t",
		serverId, enabled))

	separatePrivateNicResp, err := r.bc.client.PostServersServerIdActionsChangeSeparatePrivateNetworkInterfaceWithResponse(
		ctx,
		serverId,
		binarylane.PostServersServerIdActionsChangeSeparatePrivateNetworkInterfaceJSONRequestBody{
			Type:    "change_separate_private_network_interface",
			Enabled: enabled,
		},
	)
	if err != nil {
		return fmt.Errorf("error changing separate private network interface for server: server_id=%d, error: %w", serverId, err)
	}
	if separatePrivateNicResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status code changing separate private network interface for server: server_id=%d, details: %s", serverId, separatePrivateNicResp.Body)
	}

	err = r.waitForServerAction(ctx, serverId, separatePrivateNicResp.JSON200.Action.Id)
	if err != nil {
		return fmt.Errorf("error changing separate private network interface: %w", err)
	}

	return nil
}

func (r *serverResource) fetchServerResourceState(ctx context.Context, state *serverResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	serverResp, err := r.bc.client.GetServersServerIdWithResponse(ctx, state.Id.ValueInt64())
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				fmt.Sprintf("Error reading server: id=%s, name=%s", state.Id.String(), state.Name.ValueString()),
				err.Error(),
			),
		}
	}
	if serverResp.StatusCode() != http.StatusOK {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				fmt.Sprintf("Unexpected HTTP status code %s reading server: name=%s, id=%s", serverResp.Status(), state.Name.ValueString(), state.Id.String()),
				string(serverResp.Body),
			),
		}
	}

	state.Id = types.Int64Value(serverResp.JSON200.Server.Id)
	state.Name = types.StringValue(serverResp.JSON200.Server.Name)
	state.Image = types.StringValue(*serverResp.JSON200.Server.Image.Slug)
	state.Region = types.StringValue(serverResp.JSON200.Server.Region.Slug)
	state.Size = types.StringValue(serverResp.JSON200.Server.Size.Slug)
	state.Backups = types.BoolValue(serverResp.JSON200.Server.NextBackupWindow != nil)
	state.PortBlocking = types.BoolValue(serverResp.JSON200.Server.Networks.PortBlocking)
	state.VpcId = types.Int64PointerValue(serverResp.JSON200.Server.VpcId)
	state.Permalink = types.StringValue(*serverResp.JSON200.Server.Permalink)
	state.PasswordChangeSupported = types.BoolValue(serverResp.JSON200.Server.PasswordChangeSupported)
	state.SourceAndDestinationCheck = types.BoolPointerValue(serverResp.JSON200.Server.Networks.SourceAndDestinationCheck)
	state.SeparatePrivateNetworkInterface = types.BoolPointerValue(serverResp.JSON200.Server.Networks.SeparatePrivateNetworkInterface)
	state.Memory = types.Int32Value(serverResp.JSON200.Server.Memory)
	state.Disk = types.Int32Value(serverResp.JSON200.Server.Disk)
	var readDiags diag.Diagnostics
	state.Disks, readDiags = serverDisksToList(ctx, serverResp.JSON200.Server.Disks, state.Disks)
	diags.Append(readDiags...)
	state.Backups = types.BoolValue(serverResp.JSON200.Server.NextBackupWindow != nil)
	state.Ipv6 = types.BoolValue(len(serverResp.JSON200.Server.Networks.V6) > 0)

	if serverResp.JSON200.Server.VpcId == nil {
		state.VpcIpv4Address = types.StringNull()
	} else if len(serverResp.JSON200.Server.Networks.V4) > 0 {
		for _, v4address := range serverResp.JSON200.Server.Networks.V4 {
			// Skip addresses in 172.21.0.0/16, these are BL internal addresses that are not part of the user's VPC
			if v4address.Type == "private" && !strings.HasPrefix(v4address.IpAddress, "172.21.") {
				state.VpcIpv4Address = types.StringValue(v4address.IpAddress)
				break
			}
		}
	}

	advFeat := serverResp.JSON200.Server.AdvancedFeatures.EnabledAdvancedFeatures
	state.AdvancedFeatures, diags = resources.NewAdvancedFeaturesValue(
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
		state.AdvancedFeatures = resources.NewAdvancedFeaturesValueUnknown()
	}

	publicIpv4Addresses := []string{}
	privateIpv4Addresses := []string{}

	for _, v4address := range serverResp.JSON200.Server.Networks.V4 {
		switch v4address.Type {
		case "public":
			publicIpv4Addresses = append(publicIpv4Addresses, v4address.IpAddress)
		case "private":
			privateIpv4Addresses = append(privateIpv4Addresses, v4address.IpAddress)
		}
	}

	tfPublicIpv4Addresses, diag := types.ListValueFrom(ctx, types.StringType, publicIpv4Addresses)
	diags.Append(diag...)
	if diag.HasError() {
		state.PublicIpv4Addresses = types.ListUnknown(state.PublicIpv4Addresses.ElementType(ctx))
		state.PublicIpv4Count = types.Int32Unknown()
	} else {
		state.PublicIpv4Addresses = tfPublicIpv4Addresses
		state.PublicIpv4Count = types.Int32Value(int32(len(publicIpv4Addresses)))
	}

	tfPrivateIpv4Addresses, diag := types.ListValueFrom(ctx, types.StringType, privateIpv4Addresses)
	diags.Append(diag...)
	if diag.HasError() {
		state.PrivateIPv4Addresses = types.ListUnknown(state.PrivateIPv4Addresses.ElementType(ctx))
	} else {
		state.PrivateIPv4Addresses = tfPrivateIpv4Addresses
	}

	publicIpv6Addresses := []string{}
	privateIpv6Addresses := []string{}

	for _, v6address := range serverResp.JSON200.Server.Networks.V6 {
		switch v6address.Type {
		case "public":
			publicIpv6Addresses = append(publicIpv6Addresses, v6address.IpAddress)
		case "private":
			privateIpv6Addresses = append(privateIpv6Addresses, v6address.IpAddress)
		}
	}

	tfPublicIpv6Addresses, diag := types.ListValueFrom(ctx, types.StringType, publicIpv6Addresses)
	diags.Append(diag...)
	if diag.HasError() || tfPublicIpv6Addresses.IsNull() {
		state.PublicIpv6Addresses = types.ListUnknown(state.PublicIpv6Addresses.ElementType(ctx))
	} else {
		state.PublicIpv6Addresses = tfPublicIpv6Addresses
	}

	tfPrivateIpv6Addresses, diag := types.ListValueFrom(ctx, types.StringType, privateIpv6Addresses)
	diags.Append(diag...)
	if diag.HasError() || tfPrivateIpv6Addresses.IsNull() {
		state.PrivateIpv6Addresses = types.ListUnknown(state.PrivateIpv6Addresses.ElementType(ctx))
	} else {
		state.PrivateIpv6Addresses = tfPrivateIpv6Addresses
	}

	return diags
}

func isAdvFeatChanged(config *resources.AdvancedFeaturesValue, data *resources.AdvancedFeaturesValue) bool {
	// Check if any of the writable advanced features have been modified by the user
	return (!config.EmulatedHyperv.IsNull() && !config.EmulatedHyperv.Equal(data.EmulatedHyperv)) ||
		(!config.EmulatedDevices.IsNull() && !config.EmulatedDevices.Equal(data.EmulatedDevices)) ||
		(!config.EmulatedTpm.IsNull() && !config.EmulatedTpm.Equal(data.EmulatedTpm)) ||
		(!config.NestedVirt.IsNull() && !config.NestedVirt.Equal(data.NestedVirt)) ||
		(!config.DriverDisk.IsNull() && !config.DriverDisk.Equal(data.DriverDisk)) ||
		(!config.UnsetUuid.IsNull() && !config.UnsetUuid.Equal(data.UnsetUuid)) ||
		(!config.LocalRtc.IsNull() && !config.LocalRtc.Equal(data.LocalRtc))
}

func (r *serverResource) updateAdvancedFeatures(
	ctx context.Context,
	serverId int64,
	config *resources.AdvancedFeaturesValue,
	data *resources.AdvancedFeaturesValue,
) error {
	// If none of the writable advanced features have been modified by the user, we can skip the update
	if !isAdvFeatChanged(config, data) {
		return nil
	}

	tflog.Info(ctx, fmt.Sprintf("Changing advanced features for server: server_id=%d", serverId))

	enabledAdvancedFeatures := []string{}

	// The enabled advanced features are flags that the user has configured, or are configured by default
	emulatedHyperV := config.EmulatedHyperv.ValueBool() || config.EmulatedHyperv.IsNull() && data.EmulatedDevices.ValueBool()
	if emulatedHyperV {
		enabledAdvancedFeatures = append(enabledAdvancedFeatures, "emulated-hyperv")
	}
	emulatedDevices := config.EmulatedDevices.ValueBool() || config.EmulatedDevices.IsNull() && data.EmulatedDevices.ValueBool()
	if emulatedDevices {
		enabledAdvancedFeatures = append(enabledAdvancedFeatures, "emulated-devices")
	}
	emulatedTPM := config.EmulatedTpm.ValueBool() || config.EmulatedTpm.IsNull() && data.EmulatedTpm.ValueBool()
	if emulatedTPM {
		enabledAdvancedFeatures = append(enabledAdvancedFeatures, "emulated-tpm")
	}
	nestedVirt := config.NestedVirt.ValueBool() || config.NestedVirt.IsNull() && data.NestedVirt.ValueBool()
	if nestedVirt {
		enabledAdvancedFeatures = append(enabledAdvancedFeatures, "nested-virt")
	}
	driverDisk := config.DriverDisk.ValueBool() || config.DriverDisk.IsNull() && data.DriverDisk.ValueBool()
	if driverDisk {
		enabledAdvancedFeatures = append(enabledAdvancedFeatures, "driver-disk")
	}
	unsetUUID := config.UnsetUuid.ValueBool() || config.UnsetUuid.IsNull() && data.UnsetUuid.ValueBool()
	if unsetUUID {
		enabledAdvancedFeatures = append(enabledAdvancedFeatures, "unset-uuid")
	}
	localRTC := config.LocalRtc.ValueBool() || config.LocalRtc.IsNull() && data.LocalRtc.ValueBool()
	if localRTC {
		enabledAdvancedFeatures = append(enabledAdvancedFeatures, "local-rtc")
	}

	// Include the current value of read-only advanced features in the payload to the server
	cloudInit := data.CloudInit.ValueBool()
	if cloudInit {
		enabledAdvancedFeatures = append(enabledAdvancedFeatures, "cloud-init")
	}
	qemuGuestAgent := data.QemuGuestAgent.ValueBool()
	if qemuGuestAgent {
		enabledAdvancedFeatures = append(enabledAdvancedFeatures, "qemu-guest-agent")
	}
	uefiBoot := data.UefiBoot.ValueBool()
	if uefiBoot {
		enabledAdvancedFeatures = append(enabledAdvancedFeatures, "uefi-boot")
	}

	resp, err := r.bc.client.PostServersServerIdActionsChangeAdvancedFeaturesWithResponse(
		ctx,
		serverId,
		binarylane.PostServersServerIdActionsChangeAdvancedFeaturesJSONRequestBody{
			Type:                    "change_advanced_features",
			EnabledAdvancedFeatures: &enabledAdvancedFeatures,
		},
	)
	if err != nil {
		return fmt.Errorf("error changing advanced features for server: server_id=%d, error: %w", serverId, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status code updating advanced features for server: server_id=%d, details: %s", serverId, resp.Body)
	}

	err = r.waitForServerAction(ctx, serverId, resp.JSON200.Action.Id)
	if err != nil {
		return fmt.Errorf("failed to confirm advanced features for server was successful: %w", err)
	}

	data.EmulatedHyperv = types.BoolValue(emulatedHyperV)
	data.EmulatedDevices = types.BoolValue(emulatedDevices)
	data.EmulatedTpm = types.BoolValue(emulatedTPM)
	data.NestedVirt = types.BoolValue(nestedVirt)
	data.DriverDisk = types.BoolValue(driverDisk)
	data.UnsetUuid = types.BoolValue(unsetUUID)
	data.LocalRtc = types.BoolValue(localRTC)
	data.CloudInit = types.BoolValue(cloudInit)
	data.QemuGuestAgent = types.BoolValue(qemuGuestAgent)
	data.UefiBoot = types.BoolValue(uefiBoot)

	return nil
}

// serverDisksToList converts the server's disks into the `disks` attribute value, ordered to
// match `order` — the value Terraform already holds, either planned or from prior state. A
// list's elements must stay where they were planned, so a configuration can list disks in any
// order as long as this keeps to it.
func serverDisksToList(ctx context.Context, apiDisks []binarylane.Disk, order types.List) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	var ordered []resources.DisksValue
	if !order.IsNull() && !order.IsUnknown() {
		diags.Append(order.ElementsAs(ctx, &ordered, true)...)
	}
	orderKeys := resources.DiskKeyer{}
	rank := make(map[resources.DiskKey]int, len(ordered))
	for i, d := range ordered {
		rank[orderKeys.Of(d.Primary.ValueBool(), d.Description.ValueString())] = i
	}

	primaryFirst := func(d binarylane.Disk) int {
		if d.Primary {
			return 0
		}
		return 1
	}

	// Sort into a stable order first, to key the disks in and to fall back on.
	disks := slices.Clone(apiDisks)
	slices.SortStableFunc(disks, func(a, b binarylane.Disk) int {
		return cmp.Or(cmp.Compare(primaryFirst(a), primaryFirst(b)), cmp.Compare(a.Id, b.Id))
	})

	// Disks the order doesn't mention keep that order, after those it does.
	diskKeys := resources.DiskKeyer{}
	rankById := make(map[int64]int, len(disks))
	for _, d := range disks {
		r, ranked := rank[diskKeys.Of(d.Primary, diskDescription(d))]
		if !ranked {
			r = len(rank)
		}
		rankById[d.Id] = r
	}
	slices.SortStableFunc(disks, func(a, b binarylane.Disk) int {
		return cmp.Compare(rankById[a.Id], rankById[b.Id])
	})

	list, listDiags := types.ListValueFrom(ctx, resources.DisksValue{}.Type(ctx), disks)
	diags.Append(listDiags...)
	return list, diags
}

// splitDisks divides a `disks` list into its primary entry and the additional disks, as the API's
// own type: binarylane.Disk carries tfsdk tags for exactly the attributes `disks` has, so the
// values reflect straight into it. The attributes Binary Lane assigns are null in a configuration
// and come back as their zero value. primary is nil when the list is null or unknown, or has no
// entry with `primary = true`.
func splitDisks(ctx context.Context, list types.List) (primary *binarylane.Disk, additional []binarylane.Disk, diags diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil, nil
	}

	var disks []binarylane.Disk
	diags = list.ElementsAs(ctx, &disks, true)
	if diags.HasError() {
		return nil, nil, diags
	}

	for i, d := range disks {
		if d.Primary && primary == nil {
			primary = &disks[i]
			continue
		}
		additional = append(additional, d)
	}
	return primary, additional, diags
}

// reconcileServerDisks brings the server's disks in line with desiredPrimary and
// desiredAdditional, and returns them as they stand afterwards, re-read if anything changed —
// including after a failure, since partially applied disk changes are not rolled back. current
// is the disks as last read, or nil to read them here.
//
// Deleting and shrinking only free space, so they always run. Growing the primary, growing a
// disk and adding one all claim space the server's total may not have yet, so they wait for
// growAndAdd: Update calls this before its resize action with growAndAdd false and after it
// with true, while Create has no resize of its own and always passes true.
func (r *serverResource) reconcileServerDisks(
	ctx context.Context,
	serverId int64,
	current []binarylane.Disk,
	desiredPrimary float64,
	desiredAdditional []binarylane.Disk,
	growAndAdd bool,
) ([]binarylane.Disk, error) {
	if current == nil {
		var err error
		if current, err = r.fetchServerDisks(ctx, serverId); err != nil {
			return nil, err
		}
	}

	mutated, err := r.applyDiskChanges(ctx, serverId, current, desiredPrimary, desiredAdditional, growAndAdd)
	if !mutated {
		return current, err
	}
	disks, fetchErr := r.fetchServerDisks(ctx, serverId)
	if fetchErr != nil {
		return current, errors.Join(err, fetchErr)
	}
	return disks, err
}

// applyDiskChanges makes the changes reconcileServerDisks needs, reporting whether it attempted
// any.
func (r *serverResource) applyDiskChanges(
	ctx context.Context,
	serverId int64,
	current []binarylane.Disk,
	desiredPrimary float64,
	desiredAdditional []binarylane.Disk,
	growAndAdd bool,
) (mutated bool, err error) {
	desiredKeys := make([]resources.DiskKey, len(desiredAdditional))
	desiredSizes := make(map[resources.DiskKey]float64, len(desiredAdditional))
	keyer := resources.DiskKeyer{}
	for i, d := range desiredAdditional {
		desiredKeys[i] = keyer.Of(false, diskDescription(d))
		desiredSizes[desiredKeys[i]] = d.SizeGigabytes
	}

	// Keyed by id order, which is the order the desired disks were planned against.
	byId := slices.Clone(current)
	slices.SortStableFunc(byId, func(a, b binarylane.Disk) int { return cmp.Compare(a.Id, b.Id) })

	var primary *binarylane.Disk
	currentKeys := make(map[int64]resources.DiskKey, len(current))
	currentByKey := make(map[resources.DiskKey]binarylane.Disk, len(current))
	keyer = resources.DiskKeyer{}
	for i, d := range byId {
		if d.Primary {
			primary = &byId[i]
			continue
		}
		key := keyer.Of(false, diskDescription(d))
		currentKeys[d.Id] = key
		currentByKey[key] = d
	}

	// Delete disks that are no longer wanted and shrink those that are getting smaller.
	for _, d := range current {
		if d.Primary {
			continue
		}
		desired, wanted := desiredSizes[currentKeys[d.Id]]
		switch {
		case !wanted:
			mutated = true
			if err := r.deleteServerDisk(ctx, serverId, d.Id); err != nil {
				return true, err
			}
		case desired < d.SizeGigabytes:
			mutated = true
			if err := r.resizeServerDisk(ctx, serverId, d.Id, desired); err != nil {
				return true, err
			}
		}
	}
	// The primary shrinks before a server-level resize and grows after it, never both.
	if primary != nil && (desiredPrimary < primary.SizeGigabytes ||
		(growAndAdd && desiredPrimary > primary.SizeGigabytes)) {
		mutated = true
		if err := r.resizeServerDisk(ctx, serverId, primary.Id, desiredPrimary); err != nil {
			return true, err
		}
	}

	if !growAndAdd {
		return mutated, nil
	}

	for i, d := range desiredAdditional {
		if existing, ok := currentByKey[desiredKeys[i]]; ok {
			if d.SizeGigabytes > existing.SizeGigabytes {
				mutated = true
				if err := r.resizeServerDisk(ctx, serverId, existing.Id, d.SizeGigabytes); err != nil {
					return true, err
				}
			}
			continue
		}
		mutated = true
		if err := r.addServerDisk(ctx, serverId, diskDescription(d), d.SizeGigabytes); err != nil {
			return true, err
		}
	}
	return mutated, nil
}

// diskDescription returns a disk's description, which is unset for a disk added without one.
func diskDescription(d binarylane.Disk) string {
	if d.Description == nil {
		return ""
	}
	return *d.Description
}

func (r *serverResource) fetchServerDisks(ctx context.Context, serverId int64) ([]binarylane.Disk, error) {
	resp, err := r.bc.client.GetServersServerIdWithResponse(ctx, serverId)
	if err != nil {
		return nil, fmt.Errorf("error fetching server disks: server_id=%d, error: %w", serverId, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status code fetching server disks: server_id=%d, status=%s, body: %s", serverId, resp.Status(), resp.Body)
	}
	return resp.JSON200.Server.Disks, nil
}

func (r *serverResource) resizeServerDisk(ctx context.Context, serverId int64, diskId int64, sizeGigabytes float64) error {
	tflog.Info(ctx, fmt.Sprintf("Resizing disk: server_id=%d, disk_id=%d, size_gigabytes=%g", serverId, diskId, sizeGigabytes))

	resizeResp, err := r.bc.client.PostServersServerIdActionsResizeDiskWithResponse(ctx, serverId, binarylane.ResizeDisk{
		Type:          binarylane.ResizeDiskTypeResizeDisk,
		DiskId:        diskId,
		SizeGigabytes: int32(sizeGigabytes),
	})
	if err != nil {
		return fmt.Errorf("error resizing disk: server_id=%d, disk_id=%d, error: %w", serverId, diskId, err)
	}
	if resizeResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status code resizing disk: server_id=%d, disk_id=%d, details: %s", serverId, diskId, resizeResp.Body)
	}

	err = r.waitForServerAction(ctx, serverId, resizeResp.JSON200.Action.Id)
	if err != nil {
		return fmt.Errorf("error resizing disk: %w", err)
	}

	return nil
}

func (r *serverResource) deleteServerDisk(ctx context.Context, serverId int64, diskId int64) error {
	tflog.Info(ctx, fmt.Sprintf("Deleting disk: server_id=%d, disk_id=%d", serverId, diskId))

	deleteResp, err := r.bc.client.PostServersServerIdActionsDeleteDiskWithResponse(ctx, serverId, binarylane.DeleteDisk{
		Type:   binarylane.DeleteDiskTypeDeleteDisk,
		DiskId: diskId,
	})
	if err != nil {
		return fmt.Errorf("error deleting disk: server_id=%d, disk_id=%d, error: %w", serverId, diskId, err)
	}
	if deleteResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status code deleting disk: server_id=%d, disk_id=%d, details: %s", serverId, diskId, deleteResp.Body)
	}

	err = r.waitForServerAction(ctx, serverId, deleteResp.JSON200.Action.Id)
	if err != nil {
		return fmt.Errorf("error deleting disk: %w", err)
	}

	return nil
}

func (r *serverResource) addServerDisk(ctx context.Context, serverId int64, description string, sizeGigabytes float64) error {
	tflog.Info(ctx, fmt.Sprintf("Adding disk: server_id=%d, description=%s, size_gigabytes=%g", serverId, description, sizeGigabytes))

	addDisk := binarylane.AddDisk{
		Type:          binarylane.AddDiskTypeAddDisk,
		SizeGigabytes: int32(sizeGigabytes),
	}
	if description != "" {
		addDisk.Description = &description
	}

	addResp, err := r.bc.client.PostServersServerIdActionsAddDiskWithResponse(ctx, serverId, addDisk)
	if err != nil {
		return fmt.Errorf("error adding disk: server_id=%d, description=%s, error: %w", serverId, description, err)
	}
	if addResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status code adding disk: server_id=%d, description=%s, details: %s", serverId, description, addResp.Body)
	}

	err = r.waitForServerAction(ctx, serverId, addResp.JSON200.Action.Id)
	if err != nil {
		return fmt.Errorf("error adding disk: %w", err)
	}

	return nil
}
