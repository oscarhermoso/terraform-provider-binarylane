package provider

import (
	"context"
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

		attr, err := convertResourceAttrToDataSourceAttr(name, attribute, required, optional)
		if err != nil {
			return nil, fmt.Errorf("failed to convert resource schema attribute to data source schema attribute: %w", err)
		}
		ds.Attributes[name] = attr
	}

	return &ds, nil
}

func convertResourceAttrToDataSourceAttr(name string, attribute r_schema.Attribute, required, optional bool) (d_schema.Attribute, error) {
	switch attribute.GetType() {
	case types.BoolType:
		return d_schema.BoolAttribute{
			Description:         attribute.GetDescription(),
			Required:            required,
			Optional:            optional,
			Computed:            !required,
			Sensitive:           attribute.IsSensitive(),
			MarkdownDescription: attribute.GetMarkdownDescription(),
			DeprecationMessage:  attribute.GetDeprecationMessage(),
		}, nil
	case types.DynamicType:
		return d_schema.DynamicAttribute{
			Description:         attribute.GetDescription(),
			Required:            required,
			Optional:            optional,
			Computed:            !required,
			Sensitive:           attribute.IsSensitive(),
			MarkdownDescription: attribute.GetMarkdownDescription(),
			DeprecationMessage:  attribute.GetDeprecationMessage(),
		}, nil
	case types.Float32Type:
		return d_schema.Float32Attribute{
			Description:         attribute.GetDescription(),
			Required:            required,
			Optional:            optional,
			Computed:            !required,
			Sensitive:           attribute.IsSensitive(),
			MarkdownDescription: attribute.GetMarkdownDescription(),
			DeprecationMessage:  attribute.GetDeprecationMessage(),
		}, nil
	case types.Float64Type:
		return d_schema.Float64Attribute{
			Description:         attribute.GetDescription(),
			Required:            required,
			Optional:            optional,
			Computed:            !required,
			Sensitive:           attribute.IsSensitive(),
			MarkdownDescription: attribute.GetMarkdownDescription(),
			DeprecationMessage:  attribute.GetDeprecationMessage(),
		}, nil
	case types.Int32Type:
		return d_schema.Int32Attribute{
			Description:         attribute.GetDescription(),
			Required:            required,
			Optional:            optional,
			Computed:            !required,
			Sensitive:           attribute.IsSensitive(),
			MarkdownDescription: attribute.GetMarkdownDescription(),
			DeprecationMessage:  attribute.GetDeprecationMessage(),
		}, nil
	case types.Int64Type:
		return d_schema.Int64Attribute{
			Description:         attribute.GetDescription(),
			Required:            required,
			Optional:            optional,
			Computed:            !required,
			Sensitive:           attribute.IsSensitive(),
			MarkdownDescription: attribute.GetMarkdownDescription(),
			DeprecationMessage:  attribute.GetDeprecationMessage(),
		}, nil
	case types.NumberType:
		return d_schema.NumberAttribute{
			Description:         attribute.GetDescription(),
			Required:            required,
			Optional:            optional,
			Computed:            !required,
			Sensitive:           attribute.IsSensitive(),
			MarkdownDescription: attribute.GetMarkdownDescription(),
			DeprecationMessage:  attribute.GetDeprecationMessage(),
		}, nil
	case types.StringType:
		return d_schema.StringAttribute{
			Description:         attribute.GetDescription(),
			Required:            required,
			Optional:            optional,
			Computed:            !required,
			Sensitive:           attribute.IsSensitive(),
			MarkdownDescription: attribute.GetMarkdownDescription(),
			DeprecationMessage:  attribute.GetDeprecationMessage(),
		}, nil

	default:
		if t, isList := attribute.(r_schema.ListNestedAttribute); isList {
			nestedObjectAttrs := make(map[string]d_schema.Attribute)
			for name, attribute := range t.NestedObject.Attributes {
				nestedAttribute, err := convertResourceAttrToDataSourceAttr(name, attribute, required, optional)
				if err != nil {
					return nil, err
				}
				nestedObjectAttrs[name] = nestedAttribute
			}

			return d_schema.ListNestedAttribute{
				NestedObject: d_schema.NestedAttributeObject{
					Attributes: nestedObjectAttrs,
					CustomType: t.NestedObject.CustomType,
					Validators: t.NestedObject.Validators,
				},
				Description:         attribute.GetDescription(),
				Required:            required,
				Optional:            optional,
				Computed:            !required,
				Sensitive:           attribute.IsSensitive(),
				MarkdownDescription: attribute.GetMarkdownDescription(),
				DeprecationMessage:  attribute.GetDeprecationMessage(),
			}, nil
		}

		if t, isList := attribute.GetType().(types.ListType); isList {
			return d_schema.ListAttribute{
				ElementType:         t.ElemType,
				Description:         attribute.GetDescription(),
				Required:            required,
				Optional:            optional,
				Computed:            !required,
				Sensitive:           attribute.IsSensitive(),
				MarkdownDescription: attribute.GetMarkdownDescription(),
				DeprecationMessage:  attribute.GetDeprecationMessage(),
			}, nil
		}
		if t, isObject := attribute.(r_schema.SingleNestedAttribute); isObject {
			attributeTypes := make(map[string]attr.Type, len(t.GetAttributes()))
			for name, attribute := range t.GetAttributes() {
				attributeTypes[name] = attribute.GetType()
			}

			return d_schema.ObjectAttribute{
				CustomType:          t.CustomType,
				Description:         attribute.GetDescription(),
				Required:            required,
				Optional:            optional,
				Computed:            !required,
				Sensitive:           attribute.IsSensitive(),
				MarkdownDescription: attribute.GetMarkdownDescription(),
				DeprecationMessage:  attribute.GetDeprecationMessage(),
				AttributeTypes:      attributeTypes,
			}, nil
		}

		if t, isObject := attribute.GetType().(types.ObjectType); isObject {
			return d_schema.ObjectAttribute{
				AttributeTypes:      t.AttrTypes,
				Description:         attribute.GetDescription(),
				Required:            required,
				Optional:            optional,
				Computed:            !required,
				Sensitive:           attribute.IsSensitive(),
				MarkdownDescription: attribute.GetMarkdownDescription(),
				DeprecationMessage:  attribute.GetDeprecationMessage(),
			}, nil
		}

		return nil, fmt.Errorf("conversion of attribute type is not implemented: name=%s, type=%s", name, attribute.GetType())
	}
}

func listContainsUnknown(ctx context.Context, list types.List) bool {
	// If the whole list is unknown, return true
	if list.IsUnknown() {
		return true
	}

	// Get elements as generic attr.Value to check individual unknown status
	var elements []attr.Value
	diags := list.ElementsAs(ctx, &elements, false)
	if diags.HasError() {
		return true // Assume unknown in case of errors
	}

	// Check if any element is unknown
	for _, elem := range elements {
		if elem.IsUnknown() {
			return true
		}
	}

	return false
}

func Pointer[T any](d T) *T {
	return &d
}
