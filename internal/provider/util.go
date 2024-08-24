package provider

import (
	"fmt"
	"strings"

	d_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	r_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func convertResourceSchemaToDataSourceSchema(rs r_schema.Schema) (*d_schema.Schema, error) {
	ds := d_schema.Schema{
		Attributes:          make(map[string]d_schema.Attribute, len(rs.Attributes)),
		Blocks:              make(map[string]d_schema.Block),
		Description:         rs.GetDescription(),
		MarkdownDescription: rs.GetMarkdownDescription(),
		DeprecationMessage:  rs.GetDeprecationMessage(),
	}
	for name, attr := range rs.Attributes {
		switch attr.GetType() {
		case types.BoolType:
			ds.Attributes[name] = d_schema.BoolAttribute{
				Description:         attr.GetDescription(),
				Optional:            attr.IsOptional(),
				Computed:            true,
				Sensitive:           attr.IsSensitive(),
				MarkdownDescription: attr.GetMarkdownDescription(),
				DeprecationMessage:  attr.GetDeprecationMessage(),
			}
		case types.DynamicType:
			ds.Attributes[name] = d_schema.DynamicAttribute{
				Description:         attr.GetDescription(),
				Optional:            attr.IsOptional(),
				Computed:            true,
				Sensitive:           attr.IsSensitive(),
				MarkdownDescription: attr.GetMarkdownDescription(),
				DeprecationMessage:  attr.GetDeprecationMessage(),
			}
		case types.Float32Type:
			ds.Attributes[name] = d_schema.Float32Attribute{
				Description:         attr.GetDescription(),
				Optional:            attr.IsOptional(),
				Computed:            true,
				Sensitive:           attr.IsSensitive(),
				MarkdownDescription: attr.GetMarkdownDescription(),
				DeprecationMessage:  attr.GetDeprecationMessage(),
			}
		case types.Float64Type:
			ds.Attributes[name] = d_schema.Float64Attribute{
				Description:         attr.GetDescription(),
				Optional:            attr.IsOptional(),
				Computed:            true,
				Sensitive:           attr.IsSensitive(),
				MarkdownDescription: attr.GetMarkdownDescription(),
				DeprecationMessage:  attr.GetDeprecationMessage(),
			}
		case types.Int32Type:
			ds.Attributes[name] = d_schema.Int32Attribute{
				Description:         attr.GetDescription(),
				Optional:            attr.IsOptional(),
				Computed:            true,
				Sensitive:           attr.IsSensitive(),
				MarkdownDescription: attr.GetMarkdownDescription(),
				DeprecationMessage:  attr.GetDeprecationMessage(),
			}
		case types.Int64Type:
			ds.Attributes[name] = d_schema.Int64Attribute{
				Description:         attr.GetDescription(),
				Optional:            attr.IsOptional(),
				Computed:            true,
				Sensitive:           attr.IsSensitive(),
				MarkdownDescription: attr.GetMarkdownDescription(),
				DeprecationMessage:  attr.GetDeprecationMessage(),
			}
		case types.NumberType:
			ds.Attributes[name] = d_schema.NumberAttribute{
				Description:         attr.GetDescription(),
				Optional:            attr.IsOptional(),
				Computed:            true,
				Sensitive:           attr.IsSensitive(),
				MarkdownDescription: attr.GetMarkdownDescription(),
				DeprecationMessage:  attr.GetDeprecationMessage(),
			}
		case types.StringType:
			ds.Attributes[name] = d_schema.StringAttribute{
				Description:         attr.GetDescription(),
				Optional:            attr.IsOptional(),
				Computed:            true,
				Sensitive:           attr.IsSensitive(),
				MarkdownDescription: attr.GetMarkdownDescription(),
				DeprecationMessage:  attr.GetDeprecationMessage(),
			}
		default:
			// Feel free to to raise a PR and remove this hack
			if strings.HasPrefix(attr.GetType().String(), "types.ListType") {
				ds.Attributes[name] = d_schema.ListAttribute{
					ElementType:         attr.GetType().(types.ListType).ElemType,
					Description:         attr.GetDescription(),
					Optional:            attr.IsOptional(),
					Computed:            true,
					Sensitive:           attr.IsSensitive(),
					MarkdownDescription: attr.GetMarkdownDescription(),
					DeprecationMessage:  attr.GetDeprecationMessage(),
				}
			} else {
				return nil, fmt.Errorf("failed to convert resource schema attribute to data source schema attribute: name=%s, type=%s", name, attr.GetType())
			}
		}
	}

	return &ds, nil
}
