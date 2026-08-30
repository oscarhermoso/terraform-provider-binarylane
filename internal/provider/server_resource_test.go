package provider

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"terraform-provider-binarylane/internal/binarylane"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestServerResource(t *testing.T) {
	// Must assign a password to the server or Binary Lane will send emails
	password1 := GenerateTestPassword(t)
	password2 := GenerateTestPassword(t)

	sshPublicKeyInitial := GenerateTestPublicKey(t)
	sshPublicKeyUpdated := GenerateTestPublicKey(t)

	// Captured while "zeta" still exists, then checked once it is gone and the disks after it
	// have shifted down a position.
	var alphaDiskId, betaDiskId string

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
  region            = "` + testRegion + `"
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
					resource.TestCheckResourceAttr("binarylane_server.test", "region", testRegion),
					resource.TestCheckResourceAttr("binarylane_server.test", "image", "debian-11"),
					resource.TestCheckResourceAttr("binarylane_server.test", "size", "std-min"),
					resource.TestCheckResourceAttr("binarylane_server.test", "memory", "1152"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disk", "20"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.#", "1"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.0.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.primary", "true"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.description", "SYSTEM"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.size_gigabytes", "20"),
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
					resource.TestCheckResourceAttr("data.binarylane_server.test", "region", testRegion),
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
					resource.TestCheckResourceAttr("data.binarylane_server.test", "disks.#", "1"),
					resource.TestCheckResourceAttrSet("data.binarylane_server.test", "disks.0.id"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "disks.0.primary", "true"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "disks.0.description", "SYSTEM"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "disks.0.size_gigabytes", "20"),
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
  region            = "` + testRegion + `"
  image             = "debian-12"
  size              = "std-min"
  disk              = 40
  disks             = [
    { primary = true, size_gigabytes = 25 },
    { description = "zeta", size_gigabytes = 5 },
    { description = "alpha", size_gigabytes = 5 },
    { description = "beta", size_gigabytes = 5 },
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
					resource.TestCheckResourceAttr("binarylane_server.test", "size", "std-min"),
					// `memory` is unset here and `size` is unchanged, so it carries forward
					// the 1152 set in the previous step.
					resource.TestCheckResourceAttr("binarylane_server.test", "memory", "1152"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disk", "40"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.#", "4"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.0.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.primary", "true"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.description", "SYSTEM"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.size_gigabytes", "25"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.1.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.primary", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.description", "zeta"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.size_gigabytes", "5"),
					resource.TestCheckResourceAttrWith("binarylane_server.test", "disks.2.id", func(value string) error {
						alphaDiskId = value
						return nil
					}),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.2.primary", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.2.description", "alpha"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.2.size_gigabytes", "5"),
					resource.TestCheckResourceAttrWith("binarylane_server.test", "disks.3.id", func(value string) error {
						betaDiskId = value
						return nil
					}),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.3.primary", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.3.description", "beta"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.3.size_gigabytes", "5"),
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
  region            = "` + testRegion + `"
  image             = "debian-12"
  size              = "std-min"
  disk              = 35
  disks             = [
    { primary = true, size_gigabytes = 15 },
    { description = "alpha", size_gigabytes = 10 },
    { description = "beta", size_gigabytes = 5 },
    { description = "delta", size_gigabytes = 5 },
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
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.#", "4"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.0.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.primary", "true"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.description", "SYSTEM"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.size_gigabytes", "15"),
					// Removing "zeta" shifts these down a position, but each keeps the disk it
					// had: they are matched to prior state by description, not by index.
					resource.TestCheckResourceAttrWith("binarylane_server.test", "disks.1.id", func(value string) error {
						if value != alphaDiskId {
							return fmt.Errorf("disk \"alpha\" has id %s, want %s", value, alphaDiskId)
						}
						return nil
					}),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.primary", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.description", "alpha"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.size_gigabytes", "10"),
					resource.TestCheckResourceAttrWith("binarylane_server.test", "disks.2.id", func(value string) error {
						if value != betaDiskId {
							return fmt.Errorf("disk \"beta\" has id %s, want %s", value, betaDiskId)
						}
						return nil
					}),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.2.primary", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.2.description", "beta"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.2.size_gigabytes", "5"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.3.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.3.primary", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.3.description", "delta"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.3.size_gigabytes", "5"),
				),
			},
			// Dropping `disks` leaves the server's disks alone: growing the total here only
			// adds unallocated space, since the primary is not the only disk.
			{
				Config: providerConfig + `
resource "binarylane_ssh_key" "updated" {
  name       = "tf-test-server-resource-updated"
  public_key = "` + sshPublicKeyUpdated + `"
  default    = true
}

resource "binarylane_server" "test" {
  name              = "tf-test-server-resource-2"
  region            = "` + testRegion + `"
  image             = "debian-12"
  size              = "std-min"
  disk              = 50
  password          = "` + password2 + `"
  vpc_id            = null
  public_ipv4_count = 0
  ssh_keys          = [binarylane_ssh_key.updated.id]

  user_data         = <<EOT
#cloud-config
echo "Hello Whitespace" > /var/tmp/output.txt


EOT
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("binarylane_server.test", "disk", "50"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.#", "4"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.0.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.primary", "true"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.description", "SYSTEM"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.size_gigabytes", "15"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.1.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.primary", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.description", "alpha"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.size_gigabytes", "10"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.2.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.2.primary", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.2.description", "beta"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.2.size_gigabytes", "5"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.3.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.3.primary", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.3.description", "delta"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.3.size_gigabytes", "5"),
				),
			},
		},
	})
}

// TestServerResourceDisksValidation exercises the plan-time `disks` validators. Every case only
// plans a fresh server, so no infrastructure is created.
func TestServerResourceDisksValidation(t *testing.T) {
	tests := []struct {
		name string
		// attributes are added to an otherwise valid server resource.
		attributes string
		// expectErr is the error the plan must fail with, or empty if it must succeed.
		expectErr string
	}{
		{
			name:       "no primary disk",
			attributes: `disks = [{ description = "data1", size_gigabytes = 10 }]`,
			expectErr:  "Missing primary disk",
		},
		{
			name:       "empty disks list",
			attributes: `disks = []`,
			expectErr:  "Missing primary disk",
		},
		{
			name: "two primary disks",
			attributes: `disks = [
    { primary = true, size_gigabytes = 20 },
    { primary = true, size_gigabytes = 10 },
  ]`,
			expectErr: "Multiple primary disks",
		},
		{
			// The primary disk's description is assigned by Binary Lane, not by the user.
			name:       "primary disk with a description",
			attributes: `disks = [{ primary = true, description = "SYSTEM", size_gigabytes = 20 }]`,
			expectErr:  "Primary disk description is server-assigned",
		},
		{
			name: "disks exceeding the total",
			attributes: `disk = 45
  disks = [
    { primary = true, size_gigabytes = 30 },
    { description = "data1", size_gigabytes = 20 },
  ]`,
			expectErr: "Disks exceed total disk allocation",
		},
		{
			name:       "disk below the minimum",
			attributes: `disk = 10`,
			expectErr:  "Attribute disk",
		},
		{
			// Binary Lane allows duplicate descriptions, and so does this provider.
			name: "duplicate descriptions",
			attributes: `disks = [
    { primary = true, size_gigabytes = 20 },
    { description = "data", size_gigabytes = 5 },
    { description = "data", size_gigabytes = 10 },
  ]`,
		},
		{
			// Also valid with room to spare, leaving the difference unallocated.
			name: "disks filling the total exactly",
			attributes: `disk = 40
  disks = [
    { primary = true, size_gigabytes = 30 },
    { description = "data1", size_gigabytes = 10 },
  ]`,
		},
		{
			name: "disks with the total left unset",
			attributes: `disks = [
    { primary = true, size_gigabytes = 30 },
    { description = "data1", size_gigabytes = 10 },
  ]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := resource.TestStep{
				Config: planOnlyProviderConfig + fmt.Sprintf(`
resource "binarylane_server" "test" {
  name              = "tf-test-disks-validation"
  region            = %q
  image             = "debian-12"
  size              = "std-min"
  public_ipv4_count = 0
  %s
}
`, testRegion, tt.attributes),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			}
			if tt.expectErr != "" {
				step.ExpectError = regexp.MustCompile(tt.expectErr)
			}
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    []resource.TestStep{step},
			})
		})
	}
}

func TestServerResourceCreateWithDisks(t *testing.T) {
	// Must assign a password to the server or Binary Lane will send emails
	password := GenerateTestPassword(t)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "binarylane_server" "test" {
	name              = "tf-test-server-create-disks"
	region            = "` + testRegion + `"
	image             = "debian-12"
	size              = "std-min"
	public_ipv4_count = 0
	password          = "` + password + `"
	disk              = 40
	disks = [
		{ primary = true, size_gigabytes = 20 },
		{ description = "data", size_gigabytes = 10 },
		{ description = "data", size_gigabytes = 5 },
	]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// 5 GB of the total is left unallocated.
					resource.TestCheckResourceAttr("binarylane_server.test", "disk", "40"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.#", "3"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.0.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.primary", "true"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.description", "SYSTEM"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.size_gigabytes", "20"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.1.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.primary", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.description", "data"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.size_gigabytes", "10"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.2.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.2.primary", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.2.description", "data"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.2.size_gigabytes", "5"),
				),
			},
			// The same disks in a different order, which is not a change.
			{
				Config: providerConfig + `
resource "binarylane_server" "test" {
	name              = "tf-test-server-create-disks"
	region            = "` + testRegion + `"
	image             = "debian-12"
	size              = "std-min"
	public_ipv4_count = 0
	password          = "` + password + `"
	disk              = 40
	disks = [
		{ description = "data", size_gigabytes = 10 },
		{ primary = true, size_gigabytes = 20 },
		{ description = "data", size_gigabytes = 5 },
	]
}

data "binarylane_server" "test" {
	depends_on = [binarylane_server.test]

	id = binarylane_server.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// The resource keeps the order the configuration lists the disks in
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.#", "3"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.0.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.primary", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.description", "data"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.size_gigabytes", "10"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.1.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.primary", "true"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.description", "SYSTEM"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.size_gigabytes", "20"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.2.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.2.primary", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.2.description", "data"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.2.size_gigabytes", "5"),

					// The data source has no order of its own, so it lists the primary disk
					// first and the rest by id
					resource.TestCheckResourceAttr("data.binarylane_server.test", "disks.#", "3"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "disks.0.primary", "true"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "disks.0.description", "SYSTEM"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "disks.0.size_gigabytes", "20"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "disks.1.primary", "false"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "disks.1.description", "data"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "disks.1.size_gigabytes", "10"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "disks.2.primary", "false"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "disks.2.description", "data"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "disks.2.size_gigabytes", "5"),
				),
			},
		},
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
	region            = "` + testRegion + `"
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
	region            = "` + testRegion + `"
	image             = "debian-11"
	size              = "std-min"
	public_ipv4_count = 0
	password          = "` + password + `"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("binarylane_server.test", "name", "tf-test-server-rename-2"),
					resource.TestCheckResourceAttr("binarylane_server.test", "region", testRegion),
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
	region            = "` + testRegion + `"
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
	region            = "` + testRegion + `"
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
	region            = "` + testRegion + `"
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
	region            = "` + testRegion + `"
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
