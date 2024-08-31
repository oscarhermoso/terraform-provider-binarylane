// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resources

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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

func ServerFirewallRulesResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"firewall_rules": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"action": schema.StringAttribute{
							Required:            true,
							Description:         "The action to take when there is a match on this rule.",
							MarkdownDescription: "The action to take when there is a match on this rule.",
							Validators: []validator.String{
								stringvalidator.OneOf(
									"drop",
									"accept",
								),
							},
						},
						"description": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Description:         "A description to assist in identifying this rule. Commonly used to record the reason for the rule or the intent behind it, e.g. \"Block access to RDP\" or \"Allow access from HQ\".",
							MarkdownDescription: "A description to assist in identifying this rule. Commonly used to record the reason for the rule or the intent behind it, e.g. \"Block access to RDP\" or \"Allow access from HQ\".",
							Validators: []validator.String{
								stringvalidator.LengthBetween(0, 250),
							},
						},
						"destination_addresses": schema.ListAttribute{
							ElementType:         types.StringType,
							Required:            true,
							Description:         "The destination addresses to match for this rule. Each address may be an individual IPv4 address or a range in IPv4 CIDR notation.",
							MarkdownDescription: "The destination addresses to match for this rule. Each address may be an individual IPv4 address or a range in IPv4 CIDR notation.",
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
							},
						},
						"destination_ports": schema.ListAttribute{
							ElementType:         types.StringType,
							Optional:            true,
							Computed:            true,
							Description:         "The destination ports to match for this rule. Leave null or empty to match on all ports.",
							MarkdownDescription: "The destination ports to match for this rule. Leave null or empty to match on all ports.",
						},
						"protocol": schema.StringAttribute{
							Required:            true,
							Description:         "The protocol to match for this rule.",
							MarkdownDescription: "The protocol to match for this rule.",
							Validators: []validator.String{
								stringvalidator.OneOf(
									"all",
									"icmp",
									"tcp",
									"udp",
								),
							},
						},
						"source_addresses": schema.ListAttribute{
							ElementType:         types.StringType,
							Required:            true,
							Description:         "The source addresses to match for this rule. Each address may be an individual IPv4 address or a range in IPv4 CIDR notation.",
							MarkdownDescription: "The source addresses to match for this rule. Each address may be an individual IPv4 address or a range in IPv4 CIDR notation.",
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
							},
						},
					},
					CustomType: FirewallRulesType{
						ObjectType: types.ObjectType{
							AttrTypes: FirewallRulesValue{}.AttributeTypes(ctx),
						},
					},
				},
				Required:            true,
				Description:         "A list of rules for the server. NB: that any existing rules that are not included will be removed. Submit an empty list to clear all rules.",
				MarkdownDescription: "A list of rules for the server. NB: that any existing rules that are not included will be removed. Submit an empty list to clear all rules.",
			},
			"server_id": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Description:         "The ID of the server for which firewall rules should be listed.",
				MarkdownDescription: "The ID of the server for which firewall rules should be listed.",
			},
		},
	}
}

type ServerFirewallRulesModel struct {
	FirewallRules types.List  `tfsdk:"firewall_rules"`
	ServerId      types.Int64 `tfsdk:"server_id"`
}

var _ basetypes.ObjectTypable = FirewallRulesType{}

type FirewallRulesType struct {
	basetypes.ObjectType
}

func (t FirewallRulesType) Equal(o attr.Type) bool {
	other, ok := o.(FirewallRulesType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t FirewallRulesType) String() string {
	return "FirewallRulesType"
}

func (t FirewallRulesType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	actionAttribute, ok := attributes["action"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`action is missing from object`)

		return nil, diags
	}

	actionVal, ok := actionAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`action expected to be basetypes.StringValue, was: %T`, actionAttribute))
	}

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

	destinationAddressesAttribute, ok := attributes["destination_addresses"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`destination_addresses is missing from object`)

		return nil, diags
	}

	destinationAddressesVal, ok := destinationAddressesAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`destination_addresses expected to be basetypes.ListValue, was: %T`, destinationAddressesAttribute))
	}

	destinationPortsAttribute, ok := attributes["destination_ports"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`destination_ports is missing from object`)

		return nil, diags
	}

	destinationPortsVal, ok := destinationPortsAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`destination_ports expected to be basetypes.ListValue, was: %T`, destinationPortsAttribute))
	}

	protocolAttribute, ok := attributes["protocol"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`protocol is missing from object`)

		return nil, diags
	}

	protocolVal, ok := protocolAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`protocol expected to be basetypes.StringValue, was: %T`, protocolAttribute))
	}

	sourceAddressesAttribute, ok := attributes["source_addresses"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`source_addresses is missing from object`)

		return nil, diags
	}

	sourceAddressesVal, ok := sourceAddressesAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`source_addresses expected to be basetypes.ListValue, was: %T`, sourceAddressesAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return FirewallRulesValue{
		Action:               actionVal,
		Description:          descriptionVal,
		DestinationAddresses: destinationAddressesVal,
		DestinationPorts:     destinationPortsVal,
		Protocol:             protocolVal,
		SourceAddresses:      sourceAddressesVal,
		state:                attr.ValueStateKnown,
	}, diags
}

func NewFirewallRulesValueNull() FirewallRulesValue {
	return FirewallRulesValue{
		state: attr.ValueStateNull,
	}
}

func NewFirewallRulesValueUnknown() FirewallRulesValue {
	return FirewallRulesValue{
		state: attr.ValueStateUnknown,
	}
}

func NewFirewallRulesValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (FirewallRulesValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing FirewallRulesValue Attribute Value",
				"While creating a FirewallRulesValue value, a missing attribute value was detected. "+
					"A FirewallRulesValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("FirewallRulesValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid FirewallRulesValue Attribute Type",
				"While creating a FirewallRulesValue value, an invalid attribute value was detected. "+
					"A FirewallRulesValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("FirewallRulesValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("FirewallRulesValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra FirewallRulesValue Attribute Value",
				"While creating a FirewallRulesValue value, an extra attribute value was detected. "+
					"A FirewallRulesValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra FirewallRulesValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewFirewallRulesValueUnknown(), diags
	}

	actionAttribute, ok := attributes["action"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`action is missing from object`)

		return NewFirewallRulesValueUnknown(), diags
	}

	actionVal, ok := actionAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`action expected to be basetypes.StringValue, was: %T`, actionAttribute))
	}

	descriptionAttribute, ok := attributes["description"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`description is missing from object`)

		return NewFirewallRulesValueUnknown(), diags
	}

	descriptionVal, ok := descriptionAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`description expected to be basetypes.StringValue, was: %T`, descriptionAttribute))
	}

	destinationAddressesAttribute, ok := attributes["destination_addresses"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`destination_addresses is missing from object`)

		return NewFirewallRulesValueUnknown(), diags
	}

	destinationAddressesVal, ok := destinationAddressesAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`destination_addresses expected to be basetypes.ListValue, was: %T`, destinationAddressesAttribute))
	}

	destinationPortsAttribute, ok := attributes["destination_ports"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`destination_ports is missing from object`)

		return NewFirewallRulesValueUnknown(), diags
	}

	destinationPortsVal, ok := destinationPortsAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`destination_ports expected to be basetypes.ListValue, was: %T`, destinationPortsAttribute))
	}

	protocolAttribute, ok := attributes["protocol"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`protocol is missing from object`)

		return NewFirewallRulesValueUnknown(), diags
	}

	protocolVal, ok := protocolAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`protocol expected to be basetypes.StringValue, was: %T`, protocolAttribute))
	}

	sourceAddressesAttribute, ok := attributes["source_addresses"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`source_addresses is missing from object`)

		return NewFirewallRulesValueUnknown(), diags
	}

	sourceAddressesVal, ok := sourceAddressesAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`source_addresses expected to be basetypes.ListValue, was: %T`, sourceAddressesAttribute))
	}

	if diags.HasError() {
		return NewFirewallRulesValueUnknown(), diags
	}

	return FirewallRulesValue{
		Action:               actionVal,
		Description:          descriptionVal,
		DestinationAddresses: destinationAddressesVal,
		DestinationPorts:     destinationPortsVal,
		Protocol:             protocolVal,
		SourceAddresses:      sourceAddressesVal,
		state:                attr.ValueStateKnown,
	}, diags
}

func NewFirewallRulesValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) FirewallRulesValue {
	object, diags := NewFirewallRulesValue(attributeTypes, attributes)

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

		panic("NewFirewallRulesValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t FirewallRulesType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewFirewallRulesValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewFirewallRulesValueUnknown(), nil
	}

	if in.IsNull() {
		return NewFirewallRulesValueNull(), nil
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

	return NewFirewallRulesValueMust(FirewallRulesValue{}.AttributeTypes(ctx), attributes), nil
}

func (t FirewallRulesType) ValueType(ctx context.Context) attr.Value {
	return FirewallRulesValue{}
}

var _ basetypes.ObjectValuable = FirewallRulesValue{}

type FirewallRulesValue struct {
	Action               basetypes.StringValue `tfsdk:"action"`
	Description          basetypes.StringValue `tfsdk:"description"`
	DestinationAddresses basetypes.ListValue   `tfsdk:"destination_addresses"`
	DestinationPorts     basetypes.ListValue   `tfsdk:"destination_ports"`
	Protocol             basetypes.StringValue `tfsdk:"protocol"`
	SourceAddresses      basetypes.ListValue   `tfsdk:"source_addresses"`
	state                attr.ValueState
}

func (v FirewallRulesValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 6)

	var val tftypes.Value
	var err error

	attrTypes["action"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["description"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["destination_addresses"] = basetypes.ListType{
		ElemType: types.StringType,
	}.TerraformType(ctx)
	attrTypes["destination_ports"] = basetypes.ListType{
		ElemType: types.StringType,
	}.TerraformType(ctx)
	attrTypes["protocol"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["source_addresses"] = basetypes.ListType{
		ElemType: types.StringType,
	}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 6)

		val, err = v.Action.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["action"] = val

		val, err = v.Description.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["description"] = val

		val, err = v.DestinationAddresses.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["destination_addresses"] = val

		val, err = v.DestinationPorts.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["destination_ports"] = val

		val, err = v.Protocol.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["protocol"] = val

		val, err = v.SourceAddresses.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["source_addresses"] = val

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

func (v FirewallRulesValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v FirewallRulesValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v FirewallRulesValue) String() string {
	return "FirewallRulesValue"
}

func (v FirewallRulesValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	destinationAddressesVal, d := types.ListValue(types.StringType, v.DestinationAddresses.Elements())

	diags.Append(d...)

	if d.HasError() {
		return types.ObjectUnknown(map[string]attr.Type{
			"action":      basetypes.StringType{},
			"description": basetypes.StringType{},
			"destination_addresses": basetypes.ListType{
				ElemType: types.StringType,
			},
			"destination_ports": basetypes.ListType{
				ElemType: types.StringType,
			},
			"protocol": basetypes.StringType{},
			"source_addresses": basetypes.ListType{
				ElemType: types.StringType,
			},
		}), diags
	}

	destinationPortsVal, d := types.ListValue(types.StringType, v.DestinationPorts.Elements())

	diags.Append(d...)

	if d.HasError() {
		return types.ObjectUnknown(map[string]attr.Type{
			"action":      basetypes.StringType{},
			"description": basetypes.StringType{},
			"destination_addresses": basetypes.ListType{
				ElemType: types.StringType,
			},
			"destination_ports": basetypes.ListType{
				ElemType: types.StringType,
			},
			"protocol": basetypes.StringType{},
			"source_addresses": basetypes.ListType{
				ElemType: types.StringType,
			},
		}), diags
	}

	sourceAddressesVal, d := types.ListValue(types.StringType, v.SourceAddresses.Elements())

	diags.Append(d...)

	if d.HasError() {
		return types.ObjectUnknown(map[string]attr.Type{
			"action":      basetypes.StringType{},
			"description": basetypes.StringType{},
			"destination_addresses": basetypes.ListType{
				ElemType: types.StringType,
			},
			"destination_ports": basetypes.ListType{
				ElemType: types.StringType,
			},
			"protocol": basetypes.StringType{},
			"source_addresses": basetypes.ListType{
				ElemType: types.StringType,
			},
		}), diags
	}

	attributeTypes := map[string]attr.Type{
		"action":      basetypes.StringType{},
		"description": basetypes.StringType{},
		"destination_addresses": basetypes.ListType{
			ElemType: types.StringType,
		},
		"destination_ports": basetypes.ListType{
			ElemType: types.StringType,
		},
		"protocol": basetypes.StringType{},
		"source_addresses": basetypes.ListType{
			ElemType: types.StringType,
		},
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
			"action":                v.Action,
			"description":           v.Description,
			"destination_addresses": destinationAddressesVal,
			"destination_ports":     destinationPortsVal,
			"protocol":              v.Protocol,
			"source_addresses":      sourceAddressesVal,
		})

	return objVal, diags
}

func (v FirewallRulesValue) Equal(o attr.Value) bool {
	other, ok := o.(FirewallRulesValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Action.Equal(other.Action) {
		return false
	}

	if !v.Description.Equal(other.Description) {
		return false
	}

	if !v.DestinationAddresses.Equal(other.DestinationAddresses) {
		return false
	}

	if !v.DestinationPorts.Equal(other.DestinationPorts) {
		return false
	}

	if !v.Protocol.Equal(other.Protocol) {
		return false
	}

	if !v.SourceAddresses.Equal(other.SourceAddresses) {
		return false
	}

	return true
}

func (v FirewallRulesValue) Type(ctx context.Context) attr.Type {
	return FirewallRulesType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v FirewallRulesValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"action":      basetypes.StringType{},
		"description": basetypes.StringType{},
		"destination_addresses": basetypes.ListType{
			ElemType: types.StringType,
		},
		"destination_ports": basetypes.ListType{
			ElemType: types.StringType,
		},
		"protocol": basetypes.StringType{},
		"source_addresses": basetypes.ListType{
			ElemType: types.StringType,
		},
	}
}