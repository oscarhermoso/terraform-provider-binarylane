package provider

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"terraform-provider-binarylane/internal/binarylane"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestServerResource(t *testing.T) {
	// Must assign a password to the server or Binary Lane will send emails
	password1 := GenerateTestPassword(t)
	password2 := GenerateTestPassword(t)

	sshPublicKeyInitial := GenerateTestPublicKey(t)
	sshPublicKeyUpdated := GenerateTestPublicKey(t)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `

resource "binarylane_vpc" "test" {
  name     = "tf-test-server-resource"
  ip_range = "10.240.0.0/16"
}

resource "binarylane_ssh_key" "initial" {
  name       = "tf-test-server-resource-initial"
  public_key = "` + sshPublicKeyInitial + `"
  default    = true
}

resource "binarylane_ssh_key" "updated" {
  name       = "tf-test-server-resource-updated"
  public_key = "` + sshPublicKeyUpdated + `"
  default    = true
}

resource "binarylane_server" "test" {
  name              = "tf-test-server-resource"
  region            = "per"
  image             = "debian-11"
  size              = "std-min"
	memory            = 1152
  password          = "` + password1 + `"
  vpc_id            = binarylane_vpc.test.id
  public_ipv4_count = 1
  ssh_keys          = [binarylane_ssh_key.initial.id]
	source_and_destination_check = false
	separate_private_network_interface = true
	backups						= true
	port_blocking			= false
  user_data         = <<EOT
#cloud-config
echo "Hello World" > /var/tmp/output.txt
EOT
}

data "binarylane_server" "test" {
  depends_on = [binarylane_server.test]

  id = binarylane_server.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify resource values
					resource.TestCheckResourceAttrSet("binarylane_server.test", "id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "name", "tf-test-server-resource"),
					resource.TestCheckResourceAttr("binarylane_server.test", "region", "per"),
					resource.TestCheckResourceAttr("binarylane_server.test", "image", "debian-11"),
					resource.TestCheckResourceAttr("binarylane_server.test", "size", "std-min"),
					resource.TestCheckResourceAttr("binarylane_server.test", "memory", "1152"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disk", "20"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "vpc_id"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "vpc_ipv4_address"),
					resource.TestCheckResourceAttr("binarylane_server.test", "public_ipv4_count", "1"),
					resource.TestCheckResourceAttr("binarylane_server.test", "password", password1),
					resource.TestCheckResourceAttr("binarylane_server.test", "user_data", `#cloud-config
echo "Hello World" > /var/tmp/output.txt
`),
					resource.TestCheckResourceAttr("binarylane_server.test", "public_ipv4_addresses.#", "1"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "private_ipv4_addresses.0"),
					resource.TestCheckResourceAttr("binarylane_server.test", "port_blocking", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "ssh_keys.#", "1"),
					resource.TestCheckResourceAttrPair("binarylane_server.test", "ssh_keys.0", "binarylane_ssh_key.initial", "id"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "permalink"),
					resource.TestCheckResourceAttr("binarylane_server.test", "source_and_destination_check", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "separate_private_network_interface", "true"),
					resource.TestCheckResourceAttr("binarylane_server.test", "backups", "true"),
					resource.TestCheckResourceAttr("binarylane_server.test", "advanced_features.emulated_hyperv", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "advanced_features.emulated_devices", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "advanced_features.nested_virt", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "advanced_features.driver_disk", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "advanced_features.unset_uuid", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "advanced_features.local_rtc", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "advanced_features.emulated_tpm", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "advanced_features.cloud_init", "true"),
					resource.TestCheckResourceAttr("binarylane_server.test", "advanced_features.qemu_guest_agent", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "advanced_features.uefi_boot", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "ipv6", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "public_ipv6_addresses.#", "0"),

					// Verify data source values
					resource.TestCheckResourceAttrPair("data.binarylane_server.test", "id", "binarylane_server.test", "id"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "name", "tf-test-server-resource"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "region", "per"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "image", "debian-11"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "size", "std-min"),
					resource.TestCheckResourceAttrSet("data.binarylane_server.test", "vpc_id"),
					resource.TestCheckResourceAttrSet("data.binarylane_server.test", "vpc_ipv4_address"),
					resource.TestCheckResourceAttrPair("data.binarylane_server.test", "permalink", "binarylane_server.test", "permalink"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "user_data", `#cloud-config
echo "Hello World" > /var/tmp/output.txt
`),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "memory", "1152"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "disk", "20"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "backups", "true"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "port_blocking", "false"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "source_and_destination_check", "false"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "separate_private_network_interface", "true"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "advanced_features.emulated_hyperv", "false"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "advanced_features.emulated_devices", "false"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "advanced_features.nested_virt", "false"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "advanced_features.driver_disk", "false"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "advanced_features.unset_uuid", "false"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "advanced_features.local_rtc", "false"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "advanced_features.emulated_tpm", "false"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "advanced_features.cloud_init", "true"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "advanced_features.qemu_guest_agent", "false"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "advanced_features.uefi_boot", "false"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "ipv6", "false"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "public_ipv6_addresses.#", "0"),
				),
			},
			// Test import by ID
			{
				ResourceName:            "binarylane_server.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "ssh_keys", "timeouts"},
			},
			// Test import by name
			{
				ResourceName:            "binarylane_server.test",
				ImportState:             true,
				ImportStateId:           "tf-test-server-resource",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "ssh_keys", "timeouts"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "binarylane_vpc" "test" {
  name     = "tf-test-server-resource"
  ip_range = "10.240.0.0/16"
}

resource "binarylane_ssh_key" "initial" {
  name       = "tf-test-server-resource-initial"
  public_key = "` + sshPublicKeyInitial + `"
  default    = true
}

resource "binarylane_ssh_key" "updated" {
  name       = "tf-test-server-resource-updated"
  public_key = "` + sshPublicKeyUpdated + `"
  default    = true
}

resource "binarylane_server" "test" {
  name              = "tf-test-server-resource-2"
  region            = "per"
  image             = "debian-12"
  size              = "std-1vcpu"
  disk              = 40
  disks             = [
    { name = "zeta", size_gigabytes = 5 },
    { name = "alpha", size_gigabytes = 5 },
    { name = "beta", size_gigabytes = 5 },
  ]
  password          = "` + password1 + `"
  vpc_id            = null
  public_ipv4_count = 0
  ssh_keys          = [binarylane_ssh_key.updated.id]
	advanced_features = {
	  emulated_hyperv = true
	}
	ipv6							= true

	# source_and_destination_check =  null  # defaults to null
	# backups				  = false  # defaults to false
	# port_blocking		= true   # defaults to true

  user_data         = <<EOT
#cloud-config
echo "Hello Whitespace" > /var/tmp/output.txt


EOT
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("binarylane_server.test", "name", "tf-test-server-resource-2"),
					resource.TestCheckResourceAttr("binarylane_server.test", "size", "std-1vcpu"),
					resource.TestCheckResourceAttr("binarylane_server.test", "memory", "2048"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disk", "40"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.#", "3"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.name", "zeta"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.size_gigabytes", "5"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.0.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.name", "alpha"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.size_gigabytes", "5"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.1.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.2.name", "beta"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.2.size_gigabytes", "5"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.2.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "public_ipv4_count", "0"),
					resource.TestCheckResourceAttr("binarylane_server.test", "public_ipv4_addresses.#", "0"),
					resource.TestCheckResourceAttr("binarylane_server.test", "image", "debian-12"),
					resource.TestCheckNoResourceAttr("binarylane_server.test", "vpc_id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "ssh_keys.#", "1"),
					resource.TestCheckResourceAttrPair("binarylane_server.test", "ssh_keys.0", "binarylane_ssh_key.updated", "id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "advanced_features.emulated_hyperv", "true"),
					resource.TestCheckNoResourceAttr("binarylane_server.test", "source_and_destination_check"),
					resource.TestCheckNoResourceAttr("binarylane_server.test", "separate_private_network_interface"),
					resource.TestCheckResourceAttr("binarylane_server.test", "backups", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "port_blocking", "true"),
					resource.TestCheckResourceAttr("binarylane_server.test", "user_data", // test extra whitespace
						`#cloud-config
echo "Hello Whitespace" > /var/tmp/output.txt


`),
					resource.TestCheckResourceAttr("binarylane_server.test", "ipv6", "true"),
					resource.TestCheckResourceAttr("binarylane_server.test", "public_ipv6_addresses.#", "1"),
					resource.TestCheckResourceAttr("binarylane_server.test", "private_ipv6_addresses.#", "1"),
				),
			},
			// Change password testing (Cannot run at same time as Rebuild operation, so it has it's own test)
			{
				Config: providerConfig + `
resource "binarylane_vpc" "test" {
  name     = "tf-test-server-resource"
  ip_range = "10.240.0.0/16"
}

resource "binarylane_ssh_key" "initial" {
  name       = "tf-test-server-resource-initial"
  public_key = "` + sshPublicKeyInitial + `"
  default    = true
}

resource "binarylane_ssh_key" "updated" {
  name       = "tf-test-server-resource-updated"
  public_key = "` + sshPublicKeyUpdated + `"
  default    = true
}

resource "binarylane_server" "test" {
  name              = "tf-test-server-resource-2"
  region            = "per"
  image             = "debian-12"
  size              = "std-1vcpu"
  disk              = 35
  disks             = [
    { name = "alpha", size_gigabytes = 10 },
    { name = "beta", size_gigabytes = 5 },
    { name = "delta", size_gigabytes = 5 },
  ]
  password          = "` + password2 + `"
  vpc_id            = null
  public_ipv4_count = 0
  ssh_keys          = [binarylane_ssh_key.updated.id]

	# source_and_destination_check =  null  # defaults to null
	# backups				  = false  # defaults to false
	# port_blocking		= true   # defaults to true

  user_data         = <<EOT
#cloud-config
echo "Hello Whitespace" > /var/tmp/output.txt


EOT
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("binarylane_server.test", "name", "tf-test-server-resource-2"),
					resource.TestCheckResourceAttr("binarylane_server.test", "password", password2),
					resource.TestCheckResourceAttr("binarylane_server.test", "disk", "35"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.#", "3"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.name", "alpha"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.size_gigabytes", "10"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.0.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.name", "beta"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.size_gigabytes", "5"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.1.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.2.name", "delta"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.2.size_gigabytes", "5"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.2.id"),
				),
			},
		},
	})
}

func TestServerResourceDisksRequirePrimaryDisk(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "binarylane_server" "test" {
  name              = "tf-test-disks-no-primary"
  region            = "per"
  image             = "debian-12"
  size              = "std-1vcpu"
  public_ipv4_count = 0
  disks = [
    { name = "data1", size_gigabytes = 10 },
  ]
}
`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Missing primary disk size`),
			},
			{
				Config: providerConfig + `
resource "binarylane_server" "test" {
  name              = "tf-test-disks-empty"
  region            = "per"
  image             = "debian-12"
  size              = "std-1vcpu"
  public_ipv4_count = 0
  disks             = []
}
`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: providerConfig + `
resource "binarylane_server" "test" {
  name              = "tf-test-disks-duplicate"
  region            = "per"
  image             = "debian-12"
  size              = "std-1vcpu"
  public_ipv4_count = 0
  disk              = 45
  disks = [
    { name = "data", size_gigabytes = 5 },
    { name = "data", size_gigabytes = 10 },
  ]
}
`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Duplicate disk name`),
			},
			{
				Config: providerConfig + `
resource "binarylane_server" "test" {
  name              = "tf-test-disks-bad-total"
  region            = "per"
  image             = "debian-12"
  size              = "std-1vcpu"
  public_ipv4_count = 0
  disk              = 45
  disks = [
    { name = "data1", size_gigabytes = 1 },
  ]
}
`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Invalid total disk allocation`),
			},
			{
				Config: providerConfig + `
resource "binarylane_server" "test" {
  name              = "tf-test-disks-bad-total-tier"
  region            = "per"
  image             = "debian-12"
  size              = "std-1vcpu"
  public_ipv4_count = 0
  disk              = 50
  disks = [
    { name = "data1", size_gigabytes = 15 },
  ]
}
`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Invalid total disk allocation`),
			},
			{
				Config: providerConfig + `
resource "binarylane_server" "test" {
  name              = "tf-test-disk-only-below-min"
  region            = "per"
  image             = "debian-12"
  size              = "std-1vcpu"
  public_ipv4_count = 0
  disk              = 10
}
`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Total disk allocation too small`),
			},
			{
				Config: providerConfig + `
resource "binarylane_server" "test" {
  name              = "tf-test-total-below-min"
  region            = "per"
  image             = "debian-12"
  size              = "std-1vcpu"
  public_ipv4_count = 0
  disk              = 10
  disks = [
    { name = "data1", size_gigabytes = 5 },
  ]
}
`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Total disk allocation too small`),
			},
			{
				Config: providerConfig + `
resource "binarylane_server" "test" {
  name              = "tf-test-small-primary-compensated"
  region            = "per"
  image             = "debian-12"
  size              = "std-1vcpu"
  public_ipv4_count = 0
  disk              = 10
  disks = [
    { name = "data1", size_gigabytes = 10 },
  ]
}
`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestReadServerDisksOrdering(t *testing.T) {
	ctx := context.Background()
	desc := func(s string) *string { return &s }

	mustPriorList := func(names ...string) types.List {
		elements := make([]attr.Value, 0, len(names))
		for i, n := range names {
			obj, diags := types.ObjectValue(serverDiskAttrTypes(), map[string]attr.Value{
				"id":             types.Int64Value(int64(i + 1)),
				"name":           types.StringValue(n),
				"size_gigabytes": types.Int32Value(10),
			})
			if diags.HasError() {
				t.Fatalf("building prior element: %s", diags)
			}
			elements = append(elements, obj)
		}
		list, diags := types.ListValue(serverDiskObjectType(), elements)
		if diags.HasError() {
			t.Fatalf("building prior list: %s", diags)
		}
		return list
	}

	extractNames := func(t *testing.T, list types.List) []string {
		t.Helper()
		var out []serverDiskModel
		if diags := list.ElementsAs(ctx, &out, false); diags.HasError() {
			t.Fatalf("extracting names: %s", diags)
		}
		names := make([]string, 0, len(out))
		for _, d := range out {
			names = append(names, d.Name.ValueString())
		}
		return names
	}

	t.Run("no additional disks and null prior returns ListNull", func(t *testing.T) {
		var diags diag.Diagnostics
		primary, list := readServerDisks(ctx,
			[]binarylane.Disk{{Id: 1, Primary: true, SizeGigabytes: 30}},
			types.ListNull(serverDiskObjectType()), &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %s", diags)
		}
		if primary.ValueInt32() != 30 {
			t.Fatalf("primary = %d, want 30", primary.ValueInt32())
		}
		if !list.IsNull() {
			t.Fatalf("list = %v, want null", list)
		}
	})

	t.Run("no additional disks and non-null prior returns empty list", func(t *testing.T) {
		var diags diag.Diagnostics
		_, list := readServerDisks(ctx,
			[]binarylane.Disk{{Id: 1, Primary: true, SizeGigabytes: 30}},
			types.ListValueMust(serverDiskObjectType(), []attr.Value{}), &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %s", diags)
		}
		if list.IsNull() {
			t.Fatalf("list is null, want empty list")
		}
		if len(list.Elements()) != 0 {
			t.Fatalf("len = %d, want 0", len(list.Elements()))
		}
	})

	t.Run("null prior falls back to id-sorted order", func(t *testing.T) {
		var diags diag.Diagnostics
		_, list := readServerDisks(ctx,
			[]binarylane.Disk{
				{Id: 1, Primary: true, SizeGigabytes: 30},
				{Id: 3, Description: desc("c"), SizeGigabytes: 10},
				{Id: 2, Description: desc("b"), SizeGigabytes: 10},
				{Id: 4, Description: desc("a"), SizeGigabytes: 10},
			},
			types.ListNull(serverDiskObjectType()), &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %s", diags)
		}
		got := extractNames(t, list)
		want := []string{"b", "c", "a"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("order = %v, want %v (id-sorted)", got, want)
		}
	})

	t.Run("non-null prior preserves prior order even when ids disagree", func(t *testing.T) {
		var diags diag.Diagnostics
		_, list := readServerDisks(ctx,
			[]binarylane.Disk{
				{Id: 1, Primary: true, SizeGigabytes: 30},
				{Id: 10, Description: desc("a"), SizeGigabytes: 10},
				{Id: 11, Description: desc("b"), SizeGigabytes: 10},
				{Id: 12, Description: desc("c"), SizeGigabytes: 10},
			},
			mustPriorList("b", "c", "a"), &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %s", diags)
		}
		got := extractNames(t, list)
		want := []string{"b", "c", "a"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("order = %v, want %v (prior-order)", got, want)
		}
	})

	t.Run("disks new since prior state are appended in id order", func(t *testing.T) {
		var diags diag.Diagnostics
		_, list := readServerDisks(ctx,
			[]binarylane.Disk{
				{Id: 1, Primary: true, SizeGigabytes: 30},
				{Id: 10, Description: desc("a"), SizeGigabytes: 10},
				{Id: 11, Description: desc("b"), SizeGigabytes: 10},
				{Id: 13, Description: desc("d"), SizeGigabytes: 10},
				{Id: 12, Description: desc("c"), SizeGigabytes: 10},
			},
			mustPriorList("b", "a"), &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %s", diags)
		}
		got := extractNames(t, list)
		want := []string{"b", "a", "c", "d"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("order = %v, want %v (prior names then new in id order)", got, want)
		}
	})

	t.Run("disks dropped from prior are excluded without errors", func(t *testing.T) {
		var diags diag.Diagnostics
		_, list := readServerDisks(ctx,
			[]binarylane.Disk{
				{Id: 1, Primary: true, SizeGigabytes: 30},
				{Id: 10, Description: desc("a"), SizeGigabytes: 10},
			},
			mustPriorList("a", "deleted-elsewhere"), &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %s", diags)
		}
		got := extractNames(t, list)
		want := []string{"a"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("order = %v, want %v", got, want)
		}
	})
}

func TestServerResourceRename(t *testing.T) {
	// Must assign a password to the server or Binary Lane will send emails
	password := GenerateTestPassword(t)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Setup
			{
				Config: providerConfig + `

resource "binarylane_server" "test" {
	name              = "tf-test-server-rename-1"
	region            = "per"
	image             = "debian-11"
	size              = "std-min"
	public_ipv4_count = 0
	password          = "` + password + `"
}
`,
			},
			// Rename
			{
				Config: providerConfig + `
resource "binarylane_server" "test" {
	name              = "tf-test-server-rename-2"
	region            = "per"
	image             = "debian-11"
	size              = "std-min"
	public_ipv4_count = 0
	password          = "` + password + `"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("binarylane_server.test", "name", "tf-test-server-rename-2"),
					resource.TestCheckResourceAttr("binarylane_server.test", "region", "per"),
					resource.TestCheckResourceAttr("binarylane_server.test", "image", "debian-11"),
					resource.TestCheckResourceAttr("binarylane_server.test", "size", "std-min"),
					resource.TestCheckResourceAttr("binarylane_server.test", "password", password),
				),
			},
		},
	})
}

func TestServerVpcIpv4Change(t *testing.T) {
	// Must assign a password to the server or Binary Lane will send emails
	password := GenerateTestPassword(t)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Plan should fail if vpc_ipv4_address is set without vpc_id
			{
				Config: providerConfig + `
resource "binarylane_server" "test" {
	name              = "tf-test-server-vpcipv4-1"
	region            = "per"
	image             = "debian-11"
	size              = "std-min"
	public_ipv4_count = 0
	password          = "` + password + `"
	vpc_ipv4_address  = "10.241.69.69"
}
`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("Attribute \"vpc_id\" must be specified when \"vpc_ipv4_address\" is specified"),
			},
			// Setup
			{
				Config: providerConfig + `
resource "binarylane_vpc" "test1" {
  name     = "tf-test-server-vpcipv4-1"
  ip_range = "10.240.0.0/16"
}

resource "binarylane_vpc" "test2" {
  name     = "tf-test-server-vpcipv4-2"
  ip_range = "10.241.0.0/16"
}

resource "binarylane_server" "test" {
	name              = "tf-test-server-vpcipv4-1"
	region            = "per"
	image             = "debian-11"
	size              = "std-min"
	public_ipv4_count = 0
	password          = "` + password + `"
	vpc_id            = binarylane_vpc.test1.id
	vpc_ipv4_address  = "10.240.0.69"
}

data "binarylane_server" "test" {
  depends_on = [binarylane_server.test]

  id = binarylane_server.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("binarylane_server.test", "vpc_id", "binarylane_vpc.test1", "id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "vpc_ipv4_address", "10.240.0.69"),
					resource.TestCheckResourceAttr("binarylane_server.test", "private_ipv4_addresses.#", "1"),
					resource.TestCheckResourceAttr("binarylane_server.test", "private_ipv4_addresses.0", "10.240.0.69"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "vpc_ipv4_address", "10.240.0.69"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "private_ipv4_addresses.#", "1"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "private_ipv4_addresses.0", "10.240.0.69"),
				),
			},
			// Change IP without changing VPC
			{
				Config: providerConfig + `
resource "binarylane_vpc" "test1" {
  name     = "tf-test-server-vpcipv4-1"
  ip_range = "10.240.0.0/16"
}

resource "binarylane_vpc" "test2" {
  name     = "tf-test-server-vpcipv4-2"
  ip_range = "10.241.0.0/16"
}

resource "binarylane_server" "test" {
	name              = "tf-test-server-vpcipv4-1"
	region            = "per"
	image             = "debian-11"
	size              = "std-min"
	public_ipv4_count = 0
	password          = "` + password + `"
	vpc_id            = binarylane_vpc.test1.id
	vpc_ipv4_address  = "10.240.4.20"
}

data "binarylane_server" "test" {
  depends_on = [binarylane_server.test]

  id = binarylane_server.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("binarylane_server.test", "vpc_id", "binarylane_vpc.test1", "id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "vpc_ipv4_address", "10.240.4.20"),
					resource.TestCheckResourceAttr("binarylane_server.test", "private_ipv4_addresses.#", "1"),
					resource.TestCheckResourceAttr("binarylane_server.test", "private_ipv4_addresses.0", "10.240.4.20"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "vpc_ipv4_address", "10.240.4.20"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "private_ipv4_addresses.#", "1"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "private_ipv4_addresses.0", "10.240.4.20"),
				),
			},
			// Change VPC (should also change IP)
			{
				Config: providerConfig + `
resource "binarylane_vpc" "test1" {
  name     = "tf-test-server-vpcipv4-1"
  ip_range = "10.240.0.0/16"
}

resource "binarylane_vpc" "test2" {
  name     = "tf-test-server-vpcipv4-2"
  ip_range = "10.241.0.0/16"
}

resource "binarylane_server" "test" {
	name              = "tf-test-server-vpcipv4-1"
	region            = "per"
	image             = "debian-11"
	size              = "std-min"
	public_ipv4_count = 0
	password          = "` + password + `"
	vpc_id						= binarylane_vpc.test2.id
	vpc_ipv4_address  = "10.241.69.69"
}

data "binarylane_server" "test" {
  depends_on = [binarylane_server.test]

  id = binarylane_server.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("binarylane_server.test", "vpc_id", "binarylane_vpc.test2", "id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "vpc_ipv4_address", "10.241.69.69"),
					resource.TestCheckResourceAttr("binarylane_server.test", "private_ipv4_addresses.#", "1"),
					resource.TestCheckResourceAttr("binarylane_server.test", "private_ipv4_addresses.0", "10.241.69.69"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "vpc_ipv4_address", "10.241.69.69"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "private_ipv4_addresses.#", "1"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "private_ipv4_addresses.0", "10.241.69.69"),
				),
			},
		},
	})
}

func GenerateTestPassword(t *testing.T) string {
	t.Helper()
	pwBytes := make([]byte, 12)
	_, err := rand.Read(pwBytes)
	if err != nil {
		t.Errorf("Failed to generate password: %s", err)
	}
	return base64.URLEncoding.EncodeToString(pwBytes)
}

func init() {
	resource.AddTestSweepers("server", &resource.Sweeper{
		Name: "server",
		F: func(_ string) error {
			client, err := binarylane.NewClientWithDefaultConfig()
			if err != nil {
				return fmt.Errorf("Error creating Binary Lane API client: %w", err)
			}

			ctx := context.Background()

			var page int32 = 1
			perPage := int32(200)
			nextPage := true

			for nextPage {
				params := binarylane.GetServersParams{
					Page:    &page,
					PerPage: &perPage,
				}

				listResp, err := client.GetServersWithResponse(ctx, &params)
				if err != nil {
					return fmt.Errorf("Error getting servers for test sweep: %w", err)
				}

				if listResp.StatusCode() != http.StatusOK {
					return fmt.Errorf("Unexpected status code getting servers in test sweep: %s", listResp.Body)
				}

				servers := listResp.JSON200.Servers
				for _, s := range servers {
					if strings.HasPrefix(s.Name, "tf-test-") {
						reason := "Terraform deletion"
						params := binarylane.DeleteServersServerIdParams{
							Reason: &reason,
						}

						deleteResp, err := client.DeleteServersServerIdWithResponse(ctx, s.Id, &params)
						if err != nil {
							return fmt.Errorf("Error deleting server %d during test sweep: %w", s.Id, err)
						}
						if deleteResp.StatusCode() != http.StatusNoContent {
							return fmt.Errorf("Unexpected status %d deleting server %d in test sweep: %s", deleteResp.StatusCode(), s.Id, deleteResp.Body)
						}
						log.Println("Deleted server during test sweep:", s.Id)
					}
				}
				if listResp.JSON200.Links == nil || listResp.JSON200.Links.Pages.Next == nil {
					nextPage = false
					break
				}

				page++
			}

			return nil
		},
	})
}
