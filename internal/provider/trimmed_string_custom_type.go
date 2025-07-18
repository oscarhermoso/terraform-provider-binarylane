package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// TrimmedStringType

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.StringTypable = TrimmedStringType{}

type TrimmedStringType struct {
	basetypes.StringType
}

func (t TrimmedStringType) Equal(o attr.Type) bool {
	other, ok := o.(TrimmedStringType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t TrimmedStringType) String() string {
	return "TrimmedStringType"
}

func (t TrimmedStringType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	value := TrimmedStringValue{
		StringValue: in,
	}

	return value, nil
}

func (t TrimmedStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t TrimmedStringType) ValueType(ctx context.Context) attr.Value {
	return TrimmedStringValue{}
}

// TrimmedStringValue

var _ basetypes.StringValuable = TrimmedStringValue{}
var _ basetypes.StringValuableWithSemanticEquals = TrimmedStringValue{}

type TrimmedStringValue struct {
	basetypes.StringValue
}

func (v TrimmedStringValue) Equal(o attr.Value) bool {
	other, ok := o.(TrimmedStringValue)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v TrimmedStringValue) Type(ctx context.Context) attr.Type {
	return TrimmedStringType{}
}

func (v TrimmedStringValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// The framework should always pass the correct value type, but always check
	newValue, ok := newValuable.(TrimmedStringValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}

	priorString := strings.TrimSpace(v.StringValue.ValueString())
	newString := strings.TrimSpace(newValue.ValueString())

	// If the strings are equivalent, keep the prior value
	return priorString == newString, diags
}

// PlanModifier
var _ planmodifier.String = RequiresReplaceForTrimmedStringModifier{}

type RequiresReplaceForTrimmedStringModifier struct{}

func (m RequiresReplaceForTrimmedStringModifier) Description(_ context.Context) string {
	return "If the value of this attribute changes, Terraform will destroy and recreate the resource."
}

func (m RequiresReplaceForTrimmedStringModifier) MarkdownDescription(_ context.Context) string {
	return "If the value of this attribute changes, Terraform will destroy and recreate the resource."
}

func (m RequiresReplaceForTrimmedStringModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do not replace on resource creation.
	if req.State.Raw.IsNull() {
		return
	}
	// Do not replace on resource destroy.
	if req.Plan.Raw.IsNull() {
		return
	}
	// Do not replace if the plan and state values are equal.
	if req.PlanValue.Equal(req.StateValue) {
		return
	}

	// Create Normalized values for the state and plan.
	stateNormalized := TrimmedStringValue{
		StringValue: req.StateValue,
	}
	planNormalized := TrimmedStringValue{
		StringValue: req.PlanValue,
	}

	// Perform semantic equality check.
	semanticallyEqual, diags := stateNormalized.StringSemanticEquals(ctx, planNormalized)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if semanticallyEqual {
		tflog.Info(ctx, "Plan and state values are semantically equal. Suppressing differences.")
		resp.PlanValue = req.StateValue
	} else {
		resp.RequiresReplace = true
	}
}
