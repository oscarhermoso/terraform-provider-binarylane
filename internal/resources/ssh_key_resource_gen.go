// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resources

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func SshKeyResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"default": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Optional: If true this will be added to all new server installations (if we support SSH Key injection for the server's operating system).",
				MarkdownDescription: "Optional: If true this will be added to all new server installations (if we support SSH Key injection for the server's operating system).",
			},
			"id": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Description:         "The ID or fingerprint of the SSH Key to fetch.",
				MarkdownDescription: "The ID or fingerprint of the SSH Key to fetch.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "A name to help you identify the key.",
				MarkdownDescription: "A name to help you identify the key.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
	}
}

type SshKeyModel struct {
	Default types.Bool   `tfsdk:"default"`
	Id      types.Int64  `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
}
