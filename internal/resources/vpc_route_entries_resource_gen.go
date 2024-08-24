// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resources

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func VpcRouteEntriesResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"route_entries": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Description:         "An optional description for the route.",
							MarkdownDescription: "An optional description for the route.",
							Validators: []validator.String{
								stringvalidator.LengthAtMost(250),
							},
						},
						"destination": schema.StringAttribute{
							Required:            true,
							Description:         "The destination address for this route entry. This may be in CIDR format.",
							MarkdownDescription: "The destination address for this route entry. This may be in CIDR format.",
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"router": schema.StringAttribute{
							Required:            true,
							Description:         "The server that will receive traffic sent to the destination property in this VPC.",
							MarkdownDescription: "The server that will receive traffic sent to the destination property in this VPC.",
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
					},
					CustomType: RouteEntriesType{
						ObjectType: types.ObjectType{
							AttrTypes: RouteEntriesValue{}.AttributeTypes(ctx),
						},
					},
				},
				Optional:            true,
				Computed:            true,
				Description:         "Submit null to leave unaltered, submit an empty list to clear all route entries. It is not possible to PATCH individual route entries, to alter a route entry submit the entire list of route entries you wish to save.",
				MarkdownDescription: "Submit null to leave unaltered, submit an empty list to clear all route entries. It is not possible to PATCH individual route entries, to alter a route entry submit the entire list of route entries you wish to save.",
			},
			"vpc_id": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Description:         "The target vpc id.",
				MarkdownDescription: "The target vpc id.",
			},
		},
	}
}

type VpcRouteEntriesModel struct {
	RouteEntries types.List  `tfsdk:"route_entries"`
	VpcId        types.Int64 `tfsdk:"vpc_id"`
}

var _ basetypes.ObjectTypable = RouteEntriesType{}

type RouteEntriesType struct {
	basetypes.ObjectType
}

func (t RouteEntriesType) Equal(o attr.Type) bool {
	other, ok := o.(RouteEntriesType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t RouteEntriesType) String() string {
	return "RouteEntriesType"
}

func (t RouteEntriesType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	descriptionAttribute, ok := attributes["description"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`description is missing from object`)

		return nil, diags
	}

	descriptionVal, ok := descriptionAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`description expected to be basetypes.StringValue, was: %T`, descriptionAttribute))
	}

	destinationAttribute, ok := attributes["destination"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`destination is missing from object`)

		return nil, diags
	}

	destinationVal, ok := destinationAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`destination expected to be basetypes.StringValue, was: %T`, destinationAttribute))
	}

	routerAttribute, ok := attributes["router"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`router is missing from object`)

		return nil, diags
	}

	routerVal, ok := routerAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`router expected to be basetypes.StringValue, was: %T`, routerAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return RouteEntriesValue{
		Description: descriptionVal,
		Destination: destinationVal,
		Router:      routerVal,
		state:       attr.ValueStateKnown,
	}, diags
}

func NewRouteEntriesValueNull() RouteEntriesValue {
	return RouteEntriesValue{
		state: attr.ValueStateNull,
	}
}

func NewRouteEntriesValueUnknown() RouteEntriesValue {
	return RouteEntriesValue{
		state: attr.ValueStateUnknown,
	}
}

func NewRouteEntriesValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (RouteEntriesValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing RouteEntriesValue Attribute Value",
				"While creating a RouteEntriesValue value, a missing attribute value was detected. "+
					"A RouteEntriesValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("RouteEntriesValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid RouteEntriesValue Attribute Type",
				"While creating a RouteEntriesValue value, an invalid attribute value was detected. "+
					"A RouteEntriesValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("RouteEntriesValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("RouteEntriesValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra RouteEntriesValue Attribute Value",
				"While creating a RouteEntriesValue value, an extra attribute value was detected. "+
					"A RouteEntriesValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra RouteEntriesValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewRouteEntriesValueUnknown(), diags
	}

	descriptionAttribute, ok := attributes["description"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`description is missing from object`)

		return NewRouteEntriesValueUnknown(), diags
	}

	descriptionVal, ok := descriptionAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`description expected to be basetypes.StringValue, was: %T`, descriptionAttribute))
	}

	destinationAttribute, ok := attributes["destination"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`destination is missing from object`)

		return NewRouteEntriesValueUnknown(), diags
	}

	destinationVal, ok := destinationAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`destination expected to be basetypes.StringValue, was: %T`, destinationAttribute))
	}

	routerAttribute, ok := attributes["router"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`router is missing from object`)

		return NewRouteEntriesValueUnknown(), diags
	}

	routerVal, ok := routerAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`router expected to be basetypes.StringValue, was: %T`, routerAttribute))
	}

	if diags.HasError() {
		return NewRouteEntriesValueUnknown(), diags
	}

	return RouteEntriesValue{
		Description: descriptionVal,
		Destination: destinationVal,
		Router:      routerVal,
		state:       attr.ValueStateKnown,
	}, diags
}

func NewRouteEntriesValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) RouteEntriesValue {
	object, diags := NewRouteEntriesValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewRouteEntriesValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t RouteEntriesType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewRouteEntriesValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewRouteEntriesValueUnknown(), nil
	}

	if in.IsNull() {
		return NewRouteEntriesValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewRouteEntriesValueMust(RouteEntriesValue{}.AttributeTypes(ctx), attributes), nil
}

func (t RouteEntriesType) ValueType(ctx context.Context) attr.Value {
	return RouteEntriesValue{}
}

var _ basetypes.ObjectValuable = RouteEntriesValue{}

type RouteEntriesValue struct {
	Description basetypes.StringValue `tfsdk:"description"`
	Destination basetypes.StringValue `tfsdk:"destination"`
	Router      basetypes.StringValue `tfsdk:"router"`
	state       attr.ValueState
}

func (v RouteEntriesValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 3)

	var val tftypes.Value
	var err error

	attrTypes["description"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["destination"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["router"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 3)

		val, err = v.Description.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["description"] = val

		val, err = v.Destination.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["destination"] = val

		val, err = v.Router.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["router"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v RouteEntriesValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v RouteEntriesValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v RouteEntriesValue) String() string {
	return "RouteEntriesValue"
}

func (v RouteEntriesValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := map[string]attr.Type{
		"description": basetypes.StringType{},
		"destination": basetypes.StringType{},
		"router":      basetypes.StringType{},
	}

	if v.IsNull() {
		return types.ObjectNull(attributeTypes), diags
	}

	if v.IsUnknown() {
		return types.ObjectUnknown(attributeTypes), diags
	}

	objVal, diags := types.ObjectValue(
		attributeTypes,
		map[string]attr.Value{
			"description": v.Description,
			"destination": v.Destination,
			"router":      v.Router,
		})

	return objVal, diags
}

func (v RouteEntriesValue) Equal(o attr.Value) bool {
	other, ok := o.(RouteEntriesValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Description.Equal(other.Description) {
		return false
	}

	if !v.Destination.Equal(other.Destination) {
		return false
	}

	if !v.Router.Equal(other.Router) {
		return false
	}

	return true
}

func (v RouteEntriesValue) Type(ctx context.Context) attr.Type {
	return RouteEntriesType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v RouteEntriesValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"description": basetypes.StringType{},
		"destination": basetypes.StringType{},
		"router":      basetypes.StringType{},
	}
}