package provider

import (
	"context"
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

type serverModel struct {
	resources.ServerModel
	PublicIpv4Count           types.Int32    `tfsdk:"public_ipv4_count"`
	PublicIpv4Addresses       types.List     `tfsdk:"public_ipv4_addresses"`
	PrivateIPv4Addresses      types.List     `tfsdk:"private_ipv4_addresses"`
	SourceAndDestinationCheck types.Bool     `tfsdk:"source_and_destination_check"`
	Permalink                 types.String   `tfsdk:"permalink"`
	Password                  types.String   `tfsdk:"password"`
	PasswordChangeSupported   types.Bool     `tfsdk:"password_change_supported"`
	Timeouts                  timeouts.Value `tfsdk:"timeouts"`
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

func (r *serverResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resources.ServerResourceSchema(ctx)
	resp.Schema.Description = "Provides a Binary Lane Server resource. This can be used to create and delete servers."

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

	image := resp.Schema.Attributes["image"]
	resp.Schema.Attributes["image"] = &schema.StringAttribute{
		Description:         image.GetDescription(),
		MarkdownDescription: image.GetMarkdownDescription(),
		Required:            image.IsRequired(),
		Optional:            image.IsOptional(),
		Computed:            image.IsComputed(),
		Validators: []validator.String{
			stringvalidator.LengthAtLeast(1),
		},
	}

	backups := resp.Schema.Attributes["backups"]
	resp.Schema.Attributes["backups"] = &schema.BoolAttribute{
		Description:         backups.GetDescription(),
		MarkdownDescription: backups.GetMarkdownDescription(),
		Optional:            backups.IsOptional(),
		Computed:            backups.IsComputed(),
		Default:             booldefault.StaticBool(false),                           // Add default to backups
		PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()}, // TODO (#37)
	}

	user_data := resp.Schema.Attributes["user_data"]
	resp.Schema.Attributes["user_data"] = &schema.StringAttribute{
		Description:         user_data.GetDescription(),
		MarkdownDescription: user_data.GetMarkdownDescription(),
		Optional:            true,  // Optional as not all servers have an initialization script
		Computed:            false, // User defined
	}

	vpcId := resp.Schema.Attributes["vpc_id"]
	resp.Schema.Attributes["vpc_id"] = &schema.Int64Attribute{
		Description:         vpcId.GetDescription(),
		MarkdownDescription: vpcId.GetMarkdownDescription(),
		Optional:            vpcId.IsOptional(),
		Computed:            false, // VPC ID is not computed, defined at creation
		PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
	}

	portBlocking := resp.Schema.Attributes["port_blocking"]
	resp.Schema.Attributes["port_blocking"] = &schema.BoolAttribute{
		Description:         portBlocking.GetDescription(),
		MarkdownDescription: portBlocking.GetMarkdownDescription(),
		Optional:            portBlocking.IsOptional(),
		Computed:            portBlocking.IsComputed(),
		Default:             booldefault.StaticBool(true), // Add default to port_blocking
		PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
	}

	sshKeys := resp.Schema.Attributes["ssh_keys"]
	resp.Schema.Attributes["ssh_keys"] = &schema.ListAttribute{
		ElementType:         types.Int64Type,
		Description:         sshKeys.GetMarkdownDescription(),
		MarkdownDescription: sshKeys.GetDescription(),
		Optional:            sshKeys.IsOptional(),
		Computed:            false, // SSH keys are not computed, defined at creation
		Validators: []validator.List{
			listvalidator.ValueInt64sAre(int64validator.AtLeast(1)),
		},
	}

	// Additional attributes
	pwDescription :=
		"If this is provided the specified or default remote user's account password will be set to this value. " +
			"Only valid if the server supports password change actions. If omitted and the server supports password " +
			"change actions a random password will be generated and emailed to the account email address."
	resp.Schema.Attributes["password"] = &schema.StringAttribute{
		Description:         pwDescription,
		MarkdownDescription: pwDescription,
		Optional:            true,  // Password optional, if not set will be emailed to user
		Computed:            false, // Computed must be false to allow server to be created without password
		Sensitive:           true,  // Mark password as sensitive
		PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
	}

	publicIpv4CountDescription := "The number of public IPv4 addresses to assign to the server. If this is not provided, the " +
		"server will be created with the default number of public IPv4 addresses."
	resp.Schema.Attributes["public_ipv4_count"] = &schema.Int32Attribute{
		Description:         publicIpv4CountDescription,
		MarkdownDescription: publicIpv4CountDescription,
		Required:            true,
		Optional:            false,
		Computed:            false,
		// Default:             int32default.StaticInt32(0), // TODO: Uncomment with 1.0 release (see issue #30)
		Validators: []validator.Int32{
			int32validator.AtLeast(0),
		},
	}

	publicIpv4AddressesDescription := "The public IPv4 addresses assigned to the server."
	resp.Schema.Attributes["public_ipv4_addresses"] = &schema.ListAttribute{
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
	resp.Schema.Attributes["source_and_destination_check"] = &schema.BoolAttribute{
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

	privateIpv4AddressesDescription := "The private IPv4 addresses assigned to the server."
	resp.Schema.Attributes["private_ipv4_addresses"] = &schema.ListAttribute{
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

	resp.Schema.Attributes["permalink"] = &schema.StringAttribute{
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
	resp.Schema.Attributes["password_change_supported"] = &schema.BoolAttribute{
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

	resp.Schema.Attributes["timeouts"] =
		timeouts.Attributes(ctx, timeouts.Opts{
			Create: true,
			Update: true,
		})
}

func (r *serverResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan, state serverModel

	if req.Plan.Raw.IsNull() {
		// Destruction plan, no modification needed
		return
	}
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.SourceAndDestinationCheck.IsUnknown() {
		if plan.VpcId.IsNull() {
			plan.SourceAndDestinationCheck = types.BoolNull()
		} else {
			plan.SourceAndDestinationCheck = types.BoolPointerValue(Pointer(true))
		}
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
	}

	if req.State.Raw.IsNull() {
		// Creation plan, no further modification needed
		return
	}
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
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
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

func (r *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data serverModel

	// Read Terraform plan data into the model
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
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
	err = r.waitForServerAction(ctx, *serverResp.JSON200.Server.Id, createActionId)
	if err != nil {
		resp.Diagnostics.AddError("Error waiting for server to be created", err.Error())
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
	data.PasswordChangeSupported = types.BoolValue(*serverResp.JSON200.Server.PasswordChangeSupported)
	plannedSourceDestCheck := data.SourceAndDestinationCheck
	serverRespSourceDestCheck := types.BoolPointerValue(serverResp.JSON200.Server.Networks.SourceAndDestinationCheck)
	data.SourceAndDestinationCheck = serverRespSourceDestCheck

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

	// Update source_and_destination_check if needed
	if plannedSourceDestCheck.Equal(types.BoolPointerValue(Pointer(false))) {
		err := r.updateSourceDestCheck(ctx, data.Id.ValueInt64(), false)
		if err != nil {
			resp.Diagnostics.AddError("Error updating source and destination check", err.Error())
			return
		}
		data.SourceAndDestinationCheck = plannedSourceDestCheck
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data serverModel

	// Read Terraform prior state data into the model
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	tflog.Debug(ctx, fmt.Sprintf("Reading server: id=%s, name=%s", data.Id.String(), data.Name.ValueString()))

	serverResp, err := r.bc.client.GetServersServerIdWithResponse(ctx, data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading server: id=%s, name=%s", data.Id.String(), data.Name.ValueString()),
			err.Error(),
		)
		return
	}
	if serverResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unexpected HTTP status code %s reading server: name=%s, id=%s", serverResp.Status(), data.Name.ValueString(), data.Id.String()),
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
	data.PasswordChangeSupported = types.BoolValue(*serverResp.JSON200.Server.PasswordChangeSupported)
	data.SourceAndDestinationCheck = types.BoolPointerValue(serverResp.JSON200.Server.Networks.SourceAndDestinationCheck)

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
	data.PublicIpv4Count = types.Int32Value(int32(len(publicIpv4Addresses)))

	if resp.Diagnostics.HasError() {
		return
	}

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
	var plan, state serverModel

	// Read Terraform plan data into the model
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
		}
		if renameResp.StatusCode() != http.StatusOK {
			resp.Diagnostics.AddError(
				"Unexpected HTTP status code renaming server",
				fmt.Sprintf("Received %s renaming server: server_id=%s. Details: %s", renameResp.Status(), state.Id.String(), renameResp.Body))
			return
		}
		if *renameResp.JSON200.Action.Status == "errored" {
			resp.Diagnostics.AddError(
				"Unexpected response with \"errored\" status when renaming server",
				fmt.Sprintf("Received %s renaming server: server_id=%s. Details: %s", renameResp.Status(), state.Id.String(), renameResp.Body))
			return
		}

		// TODO - Currently, the API does not support polling for the rename to complete, because the action ID returns a 404 response (see #13)

		state.Name = types.StringValue(plan.Name.ValueString())
	}

	// Resize operation
	if !plan.Size.Equal(state.Size) || !plan.PublicIpv4Count.Equal(state.PublicIpv4Count) || !plan.Image.Equal(state.Image) {
		resizeReq := &binarylane.PostServersServerIdActionsResizeJSONRequestBody{
			Type: "resize",
		}
		if !plan.Size.Equal(state.Size) {
			resizeReq.Size = plan.Size.ValueStringPointer()
			state.Size = plan.Size
		}
		if !plan.PublicIpv4Count.Equal(state.PublicIpv4Count) {
			resizeReq.Options = &binarylane.ChangeSizeOptionsRequest{
				Ipv4Addresses: plan.PublicIpv4Count.ValueInt32Pointer(),
			}
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
		if !plan.Image.Equal(state.Image) {
			resizeReq.ChangeImage = &binarylane.ChangeImage{
				Image: plan.Image.ValueStringPointer(),
			}
			state.Image = plan.Image
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
		err = r.waitForServerAction(ctx, state.Id.ValueInt64(), *resizeResp.JSON200.Action.Id)
		if err != nil {
			resp.Diagnostics.AddError("Error waiting for server to be resized", err.Error())
			return
		}

		// Success

		if state.PublicIpv4Addresses.IsUnknown() || listContainsUnknown(ctx, state.PublicIpv4Addresses) {
			// New IPs may have been allocated, so we need to check the server again
			serverResp, err := r.bc.client.GetServersServerIdWithResponse(ctx, state.Id.ValueInt64())
			if err != nil {
				resp.Diagnostics.AddError("Error checking IPs for server after resize", err.Error())
				return
			}
			if serverResp.StatusCode() != http.StatusOK {
				resp.Diagnostics.AddError("Unexpected HTTP status code checking IPs for server after resize",
					fmt.Sprintf("Received %s checking IPs for server after resize: name=%s. Details: %s", serverResp.Status(), state.Name.ValueString(), serverResp.Body))
				return
			}
			publicIpv4Addresses := []string{}
			for _, v4address := range serverResp.JSON200.Server.Networks.V4 {
				if v4address.Type == "public" {
					publicIpv4Addresses = append(publicIpv4Addresses, v4address.IpAddress)
				}
			}
			state.PublicIpv4Addresses, diags = types.ListValueFrom(ctx, types.StringType, publicIpv4Addresses)
			resp.Diagnostics.Append(diags...)
		}

		// Save updated data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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
		err = r.waitForServerAction(ctx, state.Id.ValueInt64(), *rebuildResp.JSON200.Action.Id)
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
	}
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

	servers := *serverResp.JSON200.Servers
	idx := slices.IndexFunc(servers, func(s binarylane.Server) bool { return *s.Name == name })
	if idx == -1 {
		resp.Diagnostics.AddError(
			"Could not find server by hostname",
			fmt.Sprintf("Error finding server: hostname=%s", name),
		)
		return
	}
	server := servers[idx]

	diags := resp.State.SetAttribute(ctx, path.Root("id"), int32(*server.Id))
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
			if readyResp.StatusCode() == http.StatusOK && *readyResp.JSON200.Action.Status == binarylane.Errored {
				return fmt.Errorf("server action failed to with error: server_id=%d, action_id=%d, error: %s", serverId, actionId, *readyResp.JSON200.Action.ResultData)
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

func attrsRequiringRebuild(plan *serverModel, state *serverModel) []string {
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

	err = r.waitForServerAction(ctx, serverId, *sourceDestCheckResp.JSON200.Action.Id)
	if err != nil {
		return fmt.Errorf("error changing source and destination check: %w", err)
	}

	return nil
}
