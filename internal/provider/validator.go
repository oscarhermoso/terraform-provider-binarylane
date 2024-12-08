package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.Int32 = &MultipleOfValidator{}
)

type MultipleOfValidator struct {
	Multiple, RangeFrom, RangeTo int32
}

func (v MultipleOfValidator) Description(ctx context.Context) string {
	if v.RangeFrom == 0 {
		return fmt.Sprintf(`must be a multiple of %d`, v.Multiple)
	}
	return fmt.Sprintf("when greater than %d, must be a multiple of %d", v.RangeFrom, v.Multiple)
}

func (v MultipleOfValidator) MarkdownDescription(ctx context.Context) string {
	if v.RangeFrom == 0 {
		return fmt.Sprintf(`must be a multiple of %d`, v.Multiple)
	}
	return fmt.Sprintf("when greater than %d, must be a multiple of %d", v.RangeFrom, v.Multiple)
}

func (v MultipleOfValidator) ValidateInt32(ctx context.Context, req validator.Int32Request, resp *validator.Int32Response) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	memory := req.ConfigValue.ValueInt32()

	if memory%v.Multiple == 0 || v.RangeFrom != 0 && memory < v.RangeFrom || v.RangeTo != 0 && memory >= v.RangeTo {
		return
	}

	if v.RangeFrom != 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Not a Multiple",
			fmt.Sprintf("if greater than %d, value must be a multiple of %d", v.RangeFrom, v.Multiple),
		)
	} else {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Not a Multiple",
			fmt.Sprintf("value must be a multiple of %d", v.Multiple),
		)
	}
}
