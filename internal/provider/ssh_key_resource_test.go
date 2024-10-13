package provider

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"terraform-provider-binarylane/internal/binarylane"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestSshKeyResource(t *testing.T) {
	publicKey := GeneratePublicKey(t)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "binarylane_ssh_key" "test" {
	name       = "tf-test-key-resource-test"
	public_key = "` + publicKey + `"
}

data "binarylane_ssh_key" "test" {
  depends_on = [binarylane_ssh_key.test]

	id = binarylane_ssh_key.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify resource values
					resource.TestCheckResourceAttr("binarylane_ssh_key.test", "name", "tf-test-key-resource-test"),
					resource.TestCheckResourceAttr("binarylane_ssh_key.test", "public_key", publicKey),
					resource.TestCheckResourceAttr("binarylane_ssh_key.test", "default", "false"),
					resource.TestCheckResourceAttrSet("binarylane_ssh_key.test", "id"),

					// Verify data source values
					resource.TestCheckResourceAttr("data.binarylane_ssh_key.test", "name", "tf-test-key-resource-test"),
					resource.TestCheckResourceAttrSet("data.binarylane_ssh_key.test", "public_key"), // Ideally would check this is identical, but whitespace is not preserved
					resource.TestCheckResourceAttr("data.binarylane_ssh_key.test", "default", "false"),
					resource.TestCheckResourceAttrSet("data.binarylane_ssh_key.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "binarylane_ssh_key.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{}, // nothing to ignore
			},
			{
				ResourceName:            "binarylane_ssh_key.test",
				ImportStateIdFunc:       ImportByFingerprint,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{}, // nothing to ignore
			},
			// TODO: Update and Read testing
			// 			{
			// 				Config: providerConfig + `
			// resource "binarylane_ssh_key" "test" {
			// 	name       = "tf-test-key-resource-test"
			// 	public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJCsuosklP0T4fJcQDgkeVh7dQu+eV+vev1CfwdUkj7h test@company.internal"
			// 	default    = true
			// }
			// 			`,
			// 				Check: resource.ComposeAggregateTestCheckFunc(
			// 					// Verify resource values
			// 					resource.TestCheckResourceAttr("binarylane_ssh_key.test", "name", "tf-test-key-resource-test"),
			// 					resource.TestCheckResourceAttr("data.binarylane_ssh_key.test", "default", "true"),
			// 				),
			// 			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func GeneratePublicKey(t *testing.T) string {
	t.Helper()

	pub, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	encoded := base64.StdEncoding.EncodeToString(pub)
	return fmt.Sprintf("ssh-ed25519 %s test@company.internal", encoded)
}

func ImportByFingerprint(state *terraform.State) (fingerprint string, err error) {
	resourceName := "binarylane_ssh_key.test"
	var rawState map[string]string
	for _, m := range state.Modules {
		if len(m.Resources) > 0 {
			if v, ok := m.Resources[resourceName]; ok {
				rawState = v.Primary.Attributes
			}
		}
	}
	if rawState == nil {
		return "", fmt.Errorf("resource not found: %s", resourceName)
	}

	return rawState["fingerprint"], nil
}

func init() {
	resource.AddTestSweepers("ssh_key", &resource.Sweeper{
		Name: "ssh_key",
		F: func(_ string) error {
			endpoint := os.Getenv("BINARYLANE_API_ENDPOINT")
			if endpoint == "" {
				endpoint = "https://api.binarylane.com.au/v2"
			}
			token := os.Getenv("BINARYLANE_API_TOKEN")

			client, err := binarylane.NewClientWithAuth(
				endpoint,
				token,
			)

			if err != nil {
				return fmt.Errorf("Error creating Binary Lane API client: %w", err)
			}

			ctx := context.Background()

			var page int32 = 1
			perPage := int32(200)
			nextPage := true

			for nextPage {
				params := binarylane.GetAccountKeysParams{
					Page:    &page,
					PerPage: &perPage,
				}

				keyResp, err := client.GetAccountKeysWithResponse(ctx, &params)
				if err != nil {
					return fmt.Errorf("Error getting SSH keys for test sweep: %w", err)
				}

				if keyResp.StatusCode() != http.StatusOK {
					return fmt.Errorf("Unexpected status code getting SSH keys for test sweep: %s", keyResp.Body)
				}

				keys := keyResp.JSON200.SshKeys
				for _, k := range keys {
					if strings.HasPrefix(*k.Name, "tf-test-") {

						keyResp, err := client.DeleteAccountKeysKeyIdWithResponse(ctx, int(*k.Id))
						if err != nil {
							return fmt.Errorf("Error deleting SSH key %d for test sweep: %w", *k.Id, err)
						}
						if keyResp.StatusCode() != http.StatusNoContent {
							return fmt.Errorf("Unexpected status %d deleting SSH key %d for test sweep: %s", keyResp.StatusCode(), *k.Id, keyResp.Body)
						}
						log.Println("Deleted SSH key for test sweep:", *k.Id)
					}
				}
				if keyResp.JSON200.Links == nil || keyResp.JSON200.Links.Pages == nil || keyResp.JSON200.Links.Pages.Next == nil {
					nextPage = false
					break
				}

				page++
			}

			return nil
		},
	})
}
