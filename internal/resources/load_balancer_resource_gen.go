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
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func LoadBalancerResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"forwarding_rules": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"entry_protocol": schema.StringAttribute{
							Required:            true,
							Description:         "The protocol that traffic must match for the load balancer to forward it. Valid values are \"http\" and \"https\".",
							MarkdownDescription: "The protocol that traffic must match for the load balancer to forward it. Valid values are \"http\" and \"https\".",
							Validators: []validator.String{
								stringvalidator.OneOf(
									"http",
									"https",
								),
							},
						},
					},
					CustomType: ForwardingRulesType{
						ObjectType: types.ObjectType{
							AttrTypes: ForwardingRulesValue{}.AttributeTypes(ctx),
						},
					},
				},
				Optional:            true,
				Computed:            true,
				Description:         "The rules that control which traffic the load balancer will forward to servers in the pool. Leave null to accept a default \"HTTP\" only forwarding rule.",
				MarkdownDescription: "The rules that control which traffic the load balancer will forward to servers in the pool. Leave null to accept a default \"HTTP\" only forwarding rule.",
			},
			"health_check": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"path": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "Leave null to accept the default '/' path.",
						MarkdownDescription: "Leave null to accept the default '/' path.",
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexp.MustCompile("/[A-Za-z0-9/.?=&+%_-]*"), ""),
						},
					},
					"protocol": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "Leave null to accept the default HTTP protocol.\n\n| Value | Description |\n| ----- | ----------- |\n| http | The health check will be performed via HTTP. |\n| https | The health check will be performed via HTTPS. |\n| both | The health check will be performed via both HTTP and HTTPS. Failing a health check on one protocol will remove the server from the pool of servers only for that protocol. |\n\n",
						MarkdownDescription: "Leave null to accept the default HTTP protocol.\n\n| Value | Description |\n| ----- | ----------- |\n| http | The health check will be performed via HTTP. |\n| https | The health check will be performed via HTTPS. |\n| both | The health check will be performed via both HTTP and HTTPS. Failing a health check on one protocol will remove the server from the pool of servers only for that protocol. |\n\n",
					},
				},
				CustomType: HealthCheckType{
					ObjectType: types.ObjectType{
						AttrTypes: HealthCheckValue{}.AttributeTypes(ctx),
					},
				},
				Optional:            true,
				Computed:            true,
				Description:         "The rules that determine which servers are considered 'healthy' and in the server pool for the load balancer. Leave this null to accept appropriate defaults based on the forwarding_rules.",
				MarkdownDescription: "The rules that determine which servers are considered 'healthy' and in the server pool for the load balancer. Leave this null to accept appropriate defaults based on the forwarding_rules.",
			},
			"id": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Description:         "The ID of the load balancer to fetch.",
				MarkdownDescription: "The ID of the load balancer to fetch.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The hostname of the load balancer.",
				MarkdownDescription: "The hostname of the load balancer.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Leave null to create an anycast load balancer.",
				MarkdownDescription: "Leave null to create an anycast load balancer.",
			},
			"server_ids": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Optional:            true,
				Computed:            true,
				Description:         "A list of server IDs to assign to this load balancer.",
				MarkdownDescription: "A list of server IDs to assign to this load balancer.",
			},
		},
	}
}

type LoadBalancerModel struct {
	ForwardingRules types.List       `tfsdk:"forwarding_rules"`
	HealthCheck     HealthCheckValue `tfsdk:"health_check"`
	Id              types.Int64      `tfsdk:"id"`
	Name            types.String     `tfsdk:"name"`
	Region          types.String     `tfsdk:"region"`
	ServerIds       types.List       `tfsdk:"server_ids"`
}

var _ basetypes.ObjectTypable = ForwardingRulesType{}

type ForwardingRulesType struct {
	basetypes.ObjectType
}

func (t ForwardingRulesType) Equal(o attr.Type) bool {
	other, ok := o.(ForwardingRulesType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t ForwardingRulesType) String() string {
	return "ForwardingRulesType"
}

func (t ForwardingRulesType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	entryProtocolAttribute, ok := attributes["entry_protocol"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`entry_protocol is missing from object`)

		return nil, diags
	}

	entryProtocolVal, ok := entryProtocolAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`entry_protocol expected to be basetypes.StringValue, was: %T`, entryProtocolAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return ForwardingRulesValue{
		EntryProtocol: entryProtocolVal,
		state:         attr.ValueStateKnown,
	}, diags
}

func NewForwardingRulesValueNull() ForwardingRulesValue {
	return ForwardingRulesValue{
		state: attr.ValueStateNull,
	}
}

func NewForwardingRulesValueUnknown() ForwardingRulesValue {
	return ForwardingRulesValue{
		state: attr.ValueStateUnknown,
	}
}

func NewForwardingRulesValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (ForwardingRulesValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing ForwardingRulesValue Attribute Value",
				"While creating a ForwardingRulesValue value, a missing attribute value was detected. "+
					"A ForwardingRulesValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("ForwardingRulesValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid ForwardingRulesValue Attribute Type",
				"While creating a ForwardingRulesValue value, an invalid attribute value was detected. "+
					"A ForwardingRulesValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("ForwardingRulesValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("ForwardingRulesValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra ForwardingRulesValue Attribute Value",
				"While creating a ForwardingRulesValue value, an extra attribute value was detected. "+
					"A ForwardingRulesValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra ForwardingRulesValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewForwardingRulesValueUnknown(), diags
	}

	entryProtocolAttribute, ok := attributes["entry_protocol"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`entry_protocol is missing from object`)

		return NewForwardingRulesValueUnknown(), diags
	}

	entryProtocolVal, ok := entryProtocolAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`entry_protocol expected to be basetypes.StringValue, was: %T`, entryProtocolAttribute))
	}

	if diags.HasError() {
		return NewForwardingRulesValueUnknown(), diags
	}

	return ForwardingRulesValue{
		EntryProtocol: entryProtocolVal,
		state:         attr.ValueStateKnown,
	}, diags
}

func NewForwardingRulesValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) ForwardingRulesValue {
	object, diags := NewForwardingRulesValue(attributeTypes, attributes)

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

		panic("NewForwardingRulesValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t ForwardingRulesType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewForwardingRulesValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewForwardingRulesValueUnknown(), nil
	}

	if in.IsNull() {
		return NewForwardingRulesValueNull(), nil
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

	return NewForwardingRulesValueMust(ForwardingRulesValue{}.AttributeTypes(ctx), attributes), nil
}

func (t ForwardingRulesType) ValueType(ctx context.Context) attr.Value {
	return ForwardingRulesValue{}
}

var _ basetypes.ObjectValuable = ForwardingRulesValue{}

type ForwardingRulesValue struct {
	EntryProtocol basetypes.StringValue `tfsdk:"entry_protocol"`
	state         attr.ValueState
}

func (v ForwardingRulesValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 1)

	var val tftypes.Value
	var err error

	attrTypes["entry_protocol"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 1)

		val, err = v.EntryProtocol.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["entry_protocol"] = val

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

func (v ForwardingRulesValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ForwardingRulesValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ForwardingRulesValue) String() string {
	return "ForwardingRulesValue"
}

func (v ForwardingRulesValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := map[string]attr.Type{
		"entry_protocol": basetypes.StringType{},
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
			"entry_protocol": v.EntryProtocol,
		})

	return objVal, diags
}

func (v ForwardingRulesValue) Equal(o attr.Value) bool {
	other, ok := o.(ForwardingRulesValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.EntryProtocol.Equal(other.EntryProtocol) {
		return false
	}

	return true
}

func (v ForwardingRulesValue) Type(ctx context.Context) attr.Type {
	return ForwardingRulesType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v ForwardingRulesValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"entry_protocol": basetypes.StringType{},
	}
}

var _ basetypes.ObjectTypable = HealthCheckType{}

type HealthCheckType struct {
	basetypes.ObjectType
}

func (t HealthCheckType) Equal(o attr.Type) bool {
	other, ok := o.(HealthCheckType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t HealthCheckType) String() string {
	return "HealthCheckType"
}

func (t HealthCheckType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	pathAttribute, ok := attributes["path"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`path is missing from object`)

		return nil, diags
	}

	pathVal, ok := pathAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`path expected to be basetypes.StringValue, was: %T`, pathAttribute))
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

	if diags.HasError() {
		return nil, diags
	}

	return HealthCheckValue{
		Path:     pathVal,
		Protocol: protocolVal,
		state:    attr.ValueStateKnown,
	}, diags
}

func NewHealthCheckValueNull() HealthCheckValue {
	return HealthCheckValue{
		state: attr.ValueStateNull,
	}
}

func NewHealthCheckValueUnknown() HealthCheckValue {
	return HealthCheckValue{
		state: attr.ValueStateUnknown,
	}
}

func NewHealthCheckValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (HealthCheckValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing HealthCheckValue Attribute Value",
				"While creating a HealthCheckValue value, a missing attribute value was detected. "+
					"A HealthCheckValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("HealthCheckValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid HealthCheckValue Attribute Type",
				"While creating a HealthCheckValue value, an invalid attribute value was detected. "+
					"A HealthCheckValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("HealthCheckValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("HealthCheckValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra HealthCheckValue Attribute Value",
				"While creating a HealthCheckValue value, an extra attribute value was detected. "+
					"A HealthCheckValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra HealthCheckValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewHealthCheckValueUnknown(), diags
	}

	pathAttribute, ok := attributes["path"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`path is missing from object`)

		return NewHealthCheckValueUnknown(), diags
	}

	pathVal, ok := pathAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`path expected to be basetypes.StringValue, was: %T`, pathAttribute))
	}

	protocolAttribute, ok := attributes["protocol"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`protocol is missing from object`)

		return NewHealthCheckValueUnknown(), diags
	}

	protocolVal, ok := protocolAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`protocol expected to be basetypes.StringValue, was: %T`, protocolAttribute))
	}

	if diags.HasError() {
		return NewHealthCheckValueUnknown(), diags
	}

	return HealthCheckValue{
		Path:     pathVal,
		Protocol: protocolVal,
		state:    attr.ValueStateKnown,
	}, diags
}

func NewHealthCheckValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) HealthCheckValue {
	object, diags := NewHealthCheckValue(attributeTypes, attributes)

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

		panic("NewHealthCheckValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t HealthCheckType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewHealthCheckValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewHealthCheckValueUnknown(), nil
	}

	if in.IsNull() {
		return NewHealthCheckValueNull(), nil
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

	return NewHealthCheckValueMust(HealthCheckValue{}.AttributeTypes(ctx), attributes), nil
}

func (t HealthCheckType) ValueType(ctx context.Context) attr.Value {
	return HealthCheckValue{}
}

var _ basetypes.ObjectValuable = HealthCheckValue{}

type HealthCheckValue struct {
	Path     basetypes.StringValue `tfsdk:"path"`
	Protocol basetypes.StringValue `tfsdk:"protocol"`
	state    attr.ValueState
}

func (v HealthCheckValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 2)

	var val tftypes.Value
	var err error

	attrTypes["path"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["protocol"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 2)

		val, err = v.Path.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["path"] = val

		val, err = v.Protocol.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["protocol"] = val

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

func (v HealthCheckValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v HealthCheckValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v HealthCheckValue) String() string {
	return "HealthCheckValue"
}

func (v HealthCheckValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := map[string]attr.Type{
		"path":     basetypes.StringType{},
		"protocol": basetypes.StringType{},
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
			"path":     v.Path,
			"protocol": v.Protocol,
		})

	return objVal, diags
}

func (v HealthCheckValue) Equal(o attr.Value) bool {
	other, ok := o.(HealthCheckValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Path.Equal(other.Path) {
		return false
	}

	if !v.Protocol.Equal(other.Protocol) {
		return false
	}

	return true
}

func (v HealthCheckValue) Type(ctx context.Context) attr.Type {
	return HealthCheckType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v HealthCheckValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"path":     basetypes.StringType{},
		"protocol": basetypes.StringType{},
	}
}
