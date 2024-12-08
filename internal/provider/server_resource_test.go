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
	pw_bytes := make([]byte, 12)
	_, err := rand.Read(pw_bytes)
	if err != nil {
		t.Errorf("Failed to generate password: %s", err)
		return
	}
	password := base64.URLEncoding.EncodeToString(pw_bytes)

	sshPublicKeyInitial := GeneratePublicKey(t)
	sshPublicKeyUpdated := GeneratePublicKey(t)

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
  password          = "` + password + `"
  vpc_id            = binarylane_vpc.test.id
  public_ipv4_count = 1
  ssh_keys          = [binarylane_ssh_key.initial.id]
	source_and_destination_check = false
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
					resource.TestCheckResourceAttr("binarylane_server.test", "password", password),
					resource.TestCheckResourceAttr("binarylane_server.test", "user_data", `#cloud-config
echo "Hello World" > /var/tmp/output.txt
`),
					resource.TestCheckResourceAttr("binarylane_server.test", "public_ipv4_addresses.#", "1"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "private_ipv4_addresses.0"),
					resource.TestCheckResourceAttr("binarylane_server.test", "port_blocking", "true"),
					resource.TestCheckResourceAttr("binarylane_server.test", "ssh_keys.#", "1"),
					resource.TestCheckResourceAttrPair("binarylane_server.test", "ssh_keys.0", "binarylane_ssh_key.initial", "id"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "permalink"),
					resource.TestCheckResourceAttr("binarylane_server.test", "source_and_destination_check", "false"),

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
  password          = "` + password + `"
  vpc_id            = null
  public_ipv4_count = 0
  ssh_keys          = [binarylane_ssh_key.updated.id]
	# source_and_destination_check =  null  # defaults to null when vpc_id is null
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
					resource.TestCheckNoResourceAttr("binarylane_server.test", "source_and_destination_check"),
					resource.TestCheckResourceAttr("binarylane_server.test", "user_data", // test extra whitespace
						`#cloud-config
echo "Hello Whitespace" > /var/tmp/output.txt


`),
				),
			},
		},
	})
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
