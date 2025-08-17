package provider

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
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
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.#", "1"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.0.id"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.0.description"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.primary", "true"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.size_gigabytes", "20"),

					// Verify data source values
					resource.TestCheckResourceAttrPair("data.binarylane_server.test", "id", "binarylane_server.test", "id"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "name", "tf-test-server-resource"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "region", "per"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "image", "debian-11"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "size", "std-min"),
					resource.TestCheckResourceAttrSet("data.binarylane_server.test", "vpc_id"),
					resource.TestCheckResourceAttrPair("data.binarylane_server.test", "permalink", "binarylane_server.test", "permalink"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "user_data", `#cloud-config
echo "Hello World" > /var/tmp/output.txt
`),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "memory", "1152"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "disk", "20"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "backups", "true"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "port_blocking", "false"),
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
					resource.TestCheckResourceAttr("data.binarylane_server.test", "disks.#", "1"),
					resource.TestCheckResourceAttrSet("data.binarylane_server.test", "disks.0.id"),
					resource.TestCheckResourceAttrSet("data.binarylane_server.test", "disks.0.description"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "disks.0.primary", "true"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "disks.0.size_gigabytes", "20"),
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
  disk              = "45"
  password          = "` + password1 + `"
  vpc_id            = null
  public_ipv4_count = 0
  ssh_keys          = [binarylane_ssh_key.updated.id]
	advanced_features = {
	  emulated_hyperv = true
	}

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
					resource.TestCheckResourceAttr("binarylane_server.test", "disk", "45"),
					resource.TestCheckResourceAttr("binarylane_server.test", "public_ipv4_count", "0"),
					resource.TestCheckResourceAttr("binarylane_server.test", "public_ipv4_addresses.#", "0"),
					resource.TestCheckResourceAttr("binarylane_server.test", "image", "debian-12"),
					resource.TestCheckNoResourceAttr("binarylane_server.test", "vpc_id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "ssh_keys.#", "1"),
					resource.TestCheckResourceAttrPair("binarylane_server.test", "ssh_keys.0", "binarylane_ssh_key.updated", "id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "advanced_features.emulated_hyperv", "true"),
					resource.TestCheckNoResourceAttr("binarylane_server.test", "source_and_destination_check"),
					resource.TestCheckResourceAttr("binarylane_server.test", "backups", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "port_blocking", "true"),
					resource.TestCheckResourceAttr("binarylane_server.test", "user_data", // test extra whitespace
						`#cloud-config
echo "Hello Whitespace" > /var/tmp/output.txt


`),
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
  disk              = "45"
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

func TestServerResourceDisks(t *testing.T) {
	// Must assign a password to the server or Binary Lane will send emails
	password := GenerateTestPassword(t)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "binarylane_server" "test" {
  name     = "tf-test-server-resource"
  region   = "per"
  image    = "debian-12"
  size     = "std-min"
  password = "` + password + `"
	public_ipv4_count = 0
	disks    = [
		{
			description = "Primary Disk"
			size_gigabytes = 9
		},
		{
			description = "Secondary Disk"
			size_gigabytes = 2
		},
		{
			description = "Tertiary Disk"
			size_gigabytes = 4
		}
	]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.#", "2"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.0.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.description", "Renamed Primary Disk"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.primary", "true"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.0.size_gigabytes", "10"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "disks.1.id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.description", "Secondary Disk"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.primary", "false"),
					resource.TestCheckResourceAttr("binarylane_server.test", "disks.1.size_gigabytes", "5"),
				),
			},
			// Test import by ID
			{
				ResourceName:            "binarylane_server.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "ssh_keys", "timeouts"},
			},
			// Test import by ID
			{
				Config: providerConfig + `
resource "binarylane_server" "test" {
  name     = "tf-test-server-resource"
  region   = "per"
  image    = "debian-12"
  size     = "std-min"
  password = "` + password + `"
	public_ipv4_count = 0
	disks    = [
		{
			description = "Renamed Primary Disk"
			size_gigabytes = 10
		},
		{
			description = "Secondary Disk"
			size_gigabytes = 10
		}
	]
}
`,
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

				servers := *listResp.JSON200.Servers
				for _, s := range servers {
					if strings.HasPrefix(*s.Name, "tf-test-") {
						reason := "Terraform deletion"
						params := binarylane.DeleteServersServerIdParams{
							Reason: &reason,
						}

						deleteResp, err := client.DeleteServersServerIdWithResponse(ctx, *s.Id, &params)
						if err != nil {
							return fmt.Errorf("Error deleting server %d during test sweep: %w", *s.Id, err)
						}
						if deleteResp.StatusCode() != http.StatusNoContent {
							return fmt.Errorf("Unexpected status %d deleting server %d in test sweep: %s", deleteResp.StatusCode(), *s.Id, deleteResp.Body)
						}
						log.Println("Deleted server during test sweep:", *s.Id)
					}
				}
				if listResp.JSON200.Links == nil || listResp.JSON200.Links.Pages == nil || listResp.JSON200.Links.Pages.Next == nil {
					nextPage = false
					break
				}

				page++
			}

			return nil
		},
	})
}
