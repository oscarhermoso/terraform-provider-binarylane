package provider

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"terraform-provider-binarylane/internal/binarylane"
	"terraform-provider-binarylane/internal/resources"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
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
	WaitForCreateSeconds types.Int32  `tfsdk:"wait_for_create"`
	PublicIpv4Count      types.Int32  `tfsdk:"public_ipv4_count"`
	PublicIpv4Addresses  types.List   `tfsdk:"public_ipv4_addresses"`
	PrivateIPv4Addresses types.List   `tfsdk:"private_ipv4_addresses"`
	Permalink            types.String `tfsdk:"permalink"`
	Password             types.String `tfsdk:"password"`
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

	backups := resp.Schema.Attributes["backups"]
	resp.Schema.Attributes["backups"] = &schema.BoolAttribute{
		Description:         backups.GetDescription(),
		MarkdownDescription: backups.GetMarkdownDescription(),
		Optional:            backups.IsOptional(),
		Computed:            backups.IsComputed(),
		Default:             booldefault.StaticBool(false), // Add default to backups
	}

	user_data := resp.Schema.Attributes["user_data"]
	resp.Schema.Attributes["user_data"] = &schema.StringAttribute{
		Description:         user_data.GetDescription(),
		MarkdownDescription: user_data.GetMarkdownDescription(),
		Optional:            true,  // Optional as not all servers have an initialization script
		Computed:            false, // User defined
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(), // If user changes init script, assume they want to replace server
		},
	}

	vpcId := resp.Schema.Attributes["vpc_id"]
	resp.Schema.Attributes["vpc_id"] = &schema.Int64Attribute{
		Description:         vpcId.GetDescription(),
		MarkdownDescription: vpcId.GetMarkdownDescription(),
		Optional:            vpcId.IsOptional(),
		Computed:            false, // VPC ID is not computed, defined at creation
	}

	portBlocking := resp.Schema.Attributes["port_blocking"]
	resp.Schema.Attributes["port_blocking"] = &schema.BoolAttribute{
		Description:         portBlocking.GetDescription(),
		MarkdownDescription: portBlocking.GetMarkdownDescription(),
		Optional:            portBlocking.IsOptional(),
		Computed:            portBlocking.IsComputed(),
		Default:             booldefault.StaticBool(true), // Add default to port_blocking
	}

	sshKeys := resp.Schema.Attributes["ssh_keys"]
	resp.Schema.Attributes["ssh_keys"] = &schema.ListAttribute{
		ElementType:         types.Int64Type,
		Description:         sshKeys.GetMarkdownDescription(),
		MarkdownDescription: sshKeys.GetDescription(),
		Optional:            sshKeys.IsOptional(),
		Computed:            false,                                                   // SSH keys are not computed, defined at creation
		PlanModifiers:       []planmodifier.List{listplanmodifier.RequiresReplace()}, // Cannot update SSH keys with API
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
	}

	waitDescription := "The number of seconds to wait for the server to be created, after which, a timeout error will " +
		"be reported. If `wait_seconds` is left empty or set to 0, Terraform will succeed without waiting for the " +
		"server creation to complete."
	resp.Schema.Attributes["wait_for_create"] = &schema.Int32Attribute{
		Description:         waitDescription,
		MarkdownDescription: waitDescription,
		Optional:            true,
		Computed:            false,
	}

	publicIpv4CountDescription := "The number of public IPv4 addresses to assign to the server. If this is not provided, the " +
		"server will be created with the default number of public IPv4 addresses."
	resp.Schema.Attributes["public_ipv4_count"] = &schema.Int32Attribute{
		Description:         publicIpv4CountDescription,
		MarkdownDescription: publicIpv4CountDescription,
		Optional:            true,
		Computed:            true,
		Default:             int32default.StaticInt32(1),
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
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
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
	}
}

func (r *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data serverModel

	// Read Terraform plan data into the model
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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
			fmt.Sprintf("Received %s creating new server: name=%s. Details: %s", serverResp.Status(), data.Name.ValueString(), serverResp.Body))
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

	var publicIpv4Addresses []string
	var privateIpv4Addresses []string

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

	if data.WaitForCreateSeconds.ValueInt32() <= 0 {
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

	timeLimit := time.Now().Add(time.Duration(data.WaitForCreateSeconds.ValueInt32()) * time.Second)
	for {
		tflog.Info(ctx, "Waiting for server to be ready...")

		readyResp, err := r.bc.client.GetServersServerIdActionsActionIdWithResponse(ctx, data.Id.ValueInt64(), createActionId)
		if err != nil {
			resp.Diagnostics.AddError("Error waiting for server to be ready", err.Error())
			return
		}
		if readyResp.StatusCode() == http.StatusOK && readyResp.JSON200.Action.CompletedAt != nil {
			tflog.Info(ctx, "Server is ready")
			break
		}
		if time.Now().After(timeLimit) {
			resp.Diagnostics.AddError(
				"Timed out waiting for server to be ready",
				fmt.Sprintf(
					"Timed out waiting for server to be created, as `wait_for_create` was surpassed without "+
						"recieving a `completed_at` in response: name=%s, status=%s, details: %s",
					data.Name.ValueString(), readyResp.Status(), readyResp.Body,
				),
			)
			return
		}
		tflog.Debug(ctx, fmt.Sprintf("Waiting for server to be ready: name=%s, status=%s, details: %s", data.Name.ValueString(), readyResp.Status(), readyResp.Body))
		time.Sleep(time.Second * 5)
	}
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

	var publicIpv4Addresses []string
	var privateIpv4Addresses []string

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

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data serverModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	// TODO

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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
	id, err := strconv.ParseInt(req.ID, 10, 32)
	if err == nil {
		diags := resp.State.SetAttribute(ctx, path.Root("id"), int32(id))
		resp.Diagnostics.Append(diags...)
	} else {
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
}
