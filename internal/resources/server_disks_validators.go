package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// disksValidator is referenced by name from the "disks" attribute in
// scripts/data/server_disks.json, which extend-provider.sh splices into the generated schema. It
// lives here rather than in internal/provider, which internal/resources cannot import.
type disksValidator struct{}

func (v disksValidator) Description(ctx context.Context) string {
	return "Requires exactly one `disks` entry to have `primary = true` (with no `description` set) whenever `disks` is configured, and `sum(disks[*].size_gigabytes)` to not exceed `disk`."
}

func (v disksValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v disksValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var disks []DisksValue
	resp.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &disks, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	primaries, primaryUnknown := 0, false
	sum, sumKnown := int64(0), true
	for i, d := range disks {
		switch {
		case d.Primary.IsUnknown():
			primaryUnknown = true
		case d.Primary.ValueBool():
			primaries++
			if !d.Description.IsNull() && !d.Description.IsUnknown() {
				resp.Diagnostics.AddAttributeError(
					req.Path.AtListIndex(i).AtName("description"),
					"Primary disk description is server-assigned",
					"The primary disk's `description` is always assigned by Binary Lane (typically \"SYSTEM\"); leave it unset on the entry with `primary = true`.",
				)
			}
		}

		if d.SizeGigabytes.IsNull() || d.SizeGigabytes.IsUnknown() {
			sumKnown = false
			continue
		}
		sum += d.SizeGigabytes.ValueInt64()
	}
	if resp.Diagnostics.HasError() {
		return
	}

	if !primaryUnknown {
		if primaries == 0 {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Missing primary disk",
				"When `disks` is specified, it must fully describe every disk, including the primary — add an entry with `primary = true`.",
			)
			return
		}
		if primaries > 1 {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Multiple primary disks",
				"Only one entry in `disks` can have `primary = true`.",
			)
			return
		}
	}

	var disk types.Int32
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("disk"), &disk)...)
	if resp.Diagnostics.HasError() || !sumKnown || disk.IsNull() || disk.IsUnknown() {
		return
	}
	if sum > int64(disk.ValueInt32()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Disks exceed total disk allocation",
			fmt.Sprintf(
				"The sum of `disks[*].size_gigabytes` is %d GB, which exceeds `disk` (%d GB total). "+
					"`sum(disks[*].size_gigabytes)` must not exceed `disk`.",
				sum, disk.ValueInt32(),
			),
		)
	}
}
