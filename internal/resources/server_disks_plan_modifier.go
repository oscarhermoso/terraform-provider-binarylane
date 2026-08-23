package resources

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DiskKey identifies a disk across plans: the primary by being the primary, whose description
// Binary Lane assigns, and every other disk by its description. Binary Lane allows two disks to
// share a description, so they are also numbered within it: the second "data" disk configured is
// the second "data" disk on the server.
type DiskKey struct {
	primary     bool
	description string
	ordinal     int
}

// DiskKeyer numbers disks as it keys them, so both sides of a match must be keyed in the same
// order: id order for the server's own disks.
type DiskKeyer map[string]int

func (k DiskKeyer) Of(primary bool, description string) DiskKey {
	if primary {
		return DiskKey{primary: true}
	}
	ordinal := k[description]
	k[description]++
	return DiskKey{description: description, ordinal: ordinal}
}

// DiskShrinkWarning warns that shrinking a disk may corrupt the filesystem on it.
func DiskShrinkWarning(attrPath path.Path, label string, from int64, to int64) diag.Diagnostic {
	return diag.NewAttributeWarningDiagnostic(
		attrPath,
		fmt.Sprintf("%s shrink may corrupt the filesystem", label),
		fmt.Sprintf(
			"%s is being reduced from %d GB to %d GB. Binary Lane resizes the block device in place and does not touch the "+
				"filesystem on it, so the filesystem(s) (and any partitions) on it must first be shrunk to fit within %d GB. "+
				"Back up the server before applying, and make sure there's enough free space to shrink the filesystem to that size.",
			label, from, to, to,
		),
	)
}

// disksPlanModifier is referenced by name from the "disks" attribute in
// scripts/data/server_disks.json, the same way disksValidator is.
//
// Terraform builds the proposed plan for a list by merging prior state into the configuration
// element by element, so removing a disk hands the disks after it the id of their predecessor.
// This plans the list from the configuration instead, taking each disk's server-assigned values
// from the disk it matches in prior state rather than from whatever sits at the same index.
type disksPlanModifier struct{}

func (m disksPlanModifier) Description(ctx context.Context) string {
	return "Matches each disk with the one it corresponds to in prior state, rather than with whatever is at the same index."
}

func (m disksPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m disksPlanModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// Nothing to plan when the resource is being destroyed.
	if req.Plan.Raw.IsNull() {
		return
	}

	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		// An unset `disks` is whatever Binary Lane gives the server, which isn't known until the
		// server exists. Once it does, the resource's ModifyPlan plans it instead: what happens
		// to the disks then depends on `disk`, the server's total, which it only resolves after
		// attribute plan modifiers like this one have run.
		if req.State.Raw.IsNull() {
			resp.PlanValue = types.ListUnknown(DisksValue{}.Type(ctx))
		}
		return
	}

	var configured, current []DisksValue
	resp.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &configured, true)...)
	if !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
		resp.Diagnostics.Append(req.StateValue.ElementsAs(ctx, &current, true)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Key prior state in id order, the order the server's own disks are keyed in wherever they
	// are matched. serverDisksToList already leaves state in an order that agrees with it, but
	// this does not have to rely on that: get it wrong and disks that share a description are
	// paired with the wrong disk, and the plan promises a resize of one while the apply resizes
	// another.
	slices.SortStableFunc(current, func(a, b DisksValue) int {
		return cmp.Compare(a.Id.ValueInt64(), b.Id.ValueInt64())
	})

	stateKeys := DiskKeyer{}
	priorByKey := make(map[DiskKey]DisksValue, len(current))
	for _, d := range current {
		priorByKey[stateKeys.Of(d.Primary.ValueBool(), d.Description.ValueString())] = d
	}

	configKeys := DiskKeyer{}
	elements := make([]attr.Value, 0, len(configured))
	for _, cd := range configured {
		// Additional disks may leave `primary` unset, which means false.
		isPrimary := cd.Primary.ValueBool()
		prior, matched := priorByKey[configKeys.Of(isPrimary, cd.Description.ValueString())]

		// `id` and the primary's `description` are assigned by Binary Lane, so both are unknown
		// until there is a disk in state to take them from.
		id, description := types.Int64Unknown(), cd.Description
		if isPrimary {
			description = types.StringUnknown()
		}
		if matched {
			if !prior.Id.IsNull() && !prior.Id.IsUnknown() {
				id = prior.Id
			}
			if isPrimary && !prior.Description.IsNull() && !prior.Description.IsUnknown() {
				description = prior.Description
			}
			if !cd.SizeGigabytes.IsUnknown() && cd.SizeGigabytes.ValueInt64() < prior.SizeGigabytes.ValueInt64() {
				label := fmt.Sprintf("Disk %q", cd.Description.ValueString())
				if isPrimary {
					label = "The primary disk"
				}
				resp.Diagnostics.Append(DiskShrinkWarning(
					req.Path, label, prior.SizeGigabytes.ValueInt64(), cd.SizeGigabytes.ValueInt64()))
			}
		}

		disk, diskDiags := NewDisksValue(
			DisksValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"id":             id,
				"description":    description,
				"primary":        types.BoolValue(isPrimary),
				"size_gigabytes": cd.SizeGigabytes,
			},
		)
		resp.Diagnostics.Append(diskDiags...)
		elements = append(elements, disk)
	}

	planned, diags := types.ListValue(DisksValue{}.Type(ctx), elements)
	resp.Diagnostics.Append(diags...)
	if !diags.HasError() {
		resp.PlanValue = planned
	}
}
