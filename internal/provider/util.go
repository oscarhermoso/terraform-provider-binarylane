package provider

import (
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	d_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	r_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AttributeConfig struct {
	RequiredAttributes *[]string
	OptionalAttributes *[]string
	ExcludedAttributes *[]string
}

func convertResourceSchemaToDataSourceSchema(rs r_schema.Schema, cfg AttributeConfig) (*d_schema.Schema, error) {
	ds := d_schema.Schema{
		Attributes:          make(map[string]d_schema.Attribute, len(rs.Attributes)),
		Blocks:              make(map[string]d_schema.Block),
		Description:         rs.GetDescription(),
		MarkdownDescription: rs.GetMarkdownDescription(),
		DeprecationMessage:  rs.GetDeprecationMessage(),
	}
	for name, attribute := range rs.Attributes {
		if cfg.ExcludedAttributes != nil && slices.Contains(*cfg.ExcludedAttributes, name) {
			continue
		}

		required := cfg.RequiredAttributes != nil && slices.Contains(*cfg.RequiredAttributes, name)
		optional := cfg.OptionalAttributes != nil && slices.Contains(*cfg.OptionalAttributes, name)
		switch attribute.GetType() {
		case types.BoolType:
			ds.Attributes[name] = d_schema.BoolAttribute{
				Description:         attribute.GetDescription(),
				Required:            required,
				Optional:            optional,
				Computed:            !required,
				Sensitive:           attribute.IsSensitive(),
				MarkdownDescription: attribute.GetMarkdownDescription(),
				DeprecationMessage:  attribute.GetDeprecationMessage(),
			}
		case types.DynamicType:
			ds.Attributes[name] = d_schema.DynamicAttribute{
				Description:         attribute.GetDescription(),
				Required:            required,
				Optional:            optional,
				Computed:            !required,
				Sensitive:           attribute.IsSensitive(),
				MarkdownDescription: attribute.GetMarkdownDescription(),
				DeprecationMessage:  attribute.GetDeprecationMessage(),
			}
		case types.Float32Type:
			ds.Attributes[name] = d_schema.Float32Attribute{
				Description:         attribute.GetDescription(),
				Required:            required,
				Optional:            optional,
				Computed:            !required,
				Sensitive:           attribute.IsSensitive(),
				MarkdownDescription: attribute.GetMarkdownDescription(),
				DeprecationMessage:  attribute.GetDeprecationMessage(),
			}
		case types.Float64Type:
			ds.Attributes[name] = d_schema.Float64Attribute{
				Description:         attribute.GetDescription(),
				Required:            required,
				Optional:            optional,
				Computed:            !required,
				Sensitive:           attribute.IsSensitive(),
				MarkdownDescription: attribute.GetMarkdownDescription(),
				DeprecationMessage:  attribute.GetDeprecationMessage(),
			}
		case types.Int32Type:
			ds.Attributes[name] = d_schema.Int32Attribute{
				Description:         attribute.GetDescription(),
				Required:            required,
				Optional:            optional,
				Computed:            !required,
				Sensitive:           attribute.IsSensitive(),
				MarkdownDescription: attribute.GetMarkdownDescription(),
				DeprecationMessage:  attribute.GetDeprecationMessage(),
			}
		case types.Int64Type:
			ds.Attributes[name] = d_schema.Int64Attribute{
				Description:         attribute.GetDescription(),
				Required:            required,
				Optional:            optional,
				Computed:            !required,
				Sensitive:           attribute.IsSensitive(),
				MarkdownDescription: attribute.GetMarkdownDescription(),
				DeprecationMessage:  attribute.GetDeprecationMessage(),
			}
		case types.NumberType:
			ds.Attributes[name] = d_schema.NumberAttribute{
				Description:         attribute.GetDescription(),
				Required:            required,
				Optional:            optional,
				Computed:            !required,
				Sensitive:           attribute.IsSensitive(),
				MarkdownDescription: attribute.GetMarkdownDescription(),
				DeprecationMessage:  attribute.GetDeprecationMessage(),
			}
		case types.StringType:
			ds.Attributes[name] = d_schema.StringAttribute{
				Description:         attribute.GetDescription(),
				Required:            required,
				Optional:            optional,
				Computed:            !required,
				Sensitive:           attribute.IsSensitive(),
				MarkdownDescription: attribute.GetMarkdownDescription(),
				DeprecationMessage:  attribute.GetDeprecationMessage(),
			}
		default:
			if listType, isList := attribute.GetType().(types.ListType); isList {
				ds.Attributes[name] = d_schema.ListAttribute{
					ElementType:         listType.ElemType,
					Description:         attribute.GetDescription(),
					Required:            required,
					Optional:            optional,
					Computed:            !required,
					Sensitive:           attribute.IsSensitive(),
					MarkdownDescription: attribute.GetMarkdownDescription(),
					DeprecationMessage:  attribute.GetDeprecationMessage(),
				}
				continue
			}

			if objType, isObject := attribute.(r_schema.SingleNestedAttribute); isObject {
				attributeTypes := make(map[string]attr.Type, len(objType.GetAttributes()))
				for name, attribute := range objType.GetAttributes() {
					attributeTypes[name] = attribute.GetType()
				}

				ds.Attributes[name] = d_schema.ObjectAttribute{
					CustomType:          objType.CustomType,
					Description:         attribute.GetDescription(),
					Required:            required,
					Optional:            optional,
					Computed:            !required,
					Sensitive:           attribute.IsSensitive(),
					MarkdownDescription: attribute.GetMarkdownDescription(),
					DeprecationMessage:  attribute.GetDeprecationMessage(),
					AttributeTypes:      attributeTypes,
				}
				continue
			}

			return nil, fmt.Errorf("failed to convert resource schema attribute to data source schema attribute: name=%s, type=%s", name, attribute.GetType())
		}
	}

	return &ds, nil
}
