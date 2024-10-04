package provider

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
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
	name       = "tf_ssh_key_resource_test"
	public_key = "` + publicKey + `"
}

data "binarylane_ssh_key" "test" {
  depends_on = [binarylane_ssh_key.test]

	id = binarylane_ssh_key.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify resource values
					resource.TestCheckResourceAttr("binarylane_ssh_key.test", "name", "tf_ssh_key_resource_test"),
					resource.TestCheckResourceAttr("binarylane_ssh_key.test", "public_key", publicKey),
					resource.TestCheckResourceAttr("binarylane_ssh_key.test", "default", "false"),
					resource.TestCheckResourceAttrSet("binarylane_ssh_key.test", "id"),

					// Verify data source values
					resource.TestCheckResourceAttr("data.binarylane_ssh_key.test", "name", "tf_ssh_key_resource_test"),
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
			// 	name       = "tf_ssh_key_resource_test"
			// 	public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJCsuosklP0T4fJcQDgkeVh7dQu+eV+vev1CfwdUkj7h test@company.internal"
			// 	default    = true
			// }
			// 			`,
			// 				Check: resource.ComposeAggregateTestCheckFunc(
			// 					// Verify resource values
			// 					resource.TestCheckResourceAttr("binarylane_ssh_key.test", "name", "tf_ssh_key_resource_test"),
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
