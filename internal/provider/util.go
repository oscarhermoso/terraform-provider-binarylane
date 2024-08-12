package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func assignStr(source *string, target *basetypes.StringValue) {
	if source != nil {
		*target = types.StringValue(*source)
	}
}

// func assignInt(source *int, target *basetypes.Int64Value) {
// 	if source != nil {
// 		*target = types.Int64Value(int64(*source))
// 	}
// }

func assignInt64(source *int64, target *basetypes.Int64Value) {
	if source != nil {
		*target = types.Int64Value(int64(*source))
	}
}

// func assignFloat(source *float32, target *basetypes.Float64Value) {
// 	if source != nil {
// 		*target = types.Float64Value(float64(*source))
// 	}
// }

// func assignBool(source *bool, target *basetypes.BoolValue) {
// 	if source != nil {
// 		*target = types.BoolValue(bool(*source))
// 	}
// }
