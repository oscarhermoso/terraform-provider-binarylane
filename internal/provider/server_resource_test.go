package provider

import (
	"crypto/rand"
	"encoding/base64"
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

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "binarylane_vpc" "test" {
	name     = "tf-test-server-resource"
	ip_range = "10.240.0.0/16"
}

resource "binarylane_ssh_key" "test" {
  name       = "tf-test-server-resource"
  public_key = "` + GeneratePublicKey(t) + `"
  default    = true
}

resource "binarylane_ssh_key" "unused" {
  name       = "tf-test-server-resource-unused"
  public_key = "` + GeneratePublicKey(t) + `"
  default    = true
}

resource "binarylane_server" "test" {
	name   = "tf-test-server-resource"
  region = "per"
  image  = "debian-12"
  size   = "std-min"
	password = "` + password + `"
	vpc_id = binarylane_vpc.test.id
	public_ipv4_count = 0
	user_data = <<EOT
#cloud-config
echo "Hello World" > /var/tmp/output.txt
EOT
	ssh_keys = [binarylane_ssh_key.test.id]
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
					resource.TestCheckResourceAttr("binarylane_server.test", "image", "debian-12"),
					resource.TestCheckResourceAttr("binarylane_server.test", "size", "std-min"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "vpc_id"),
					resource.TestCheckResourceAttr("binarylane_server.test", "public_ipv4_count", "0"),
					resource.TestCheckResourceAttr("binarylane_server.test", "password", password),
					resource.TestCheckResourceAttr("binarylane_server.test", "user_data", `#cloud-config
echo "Hello World" > /var/tmp/output.txt
`),
					resource.TestCheckResourceAttr("binarylane_server.test", "public_ipv4_addresses.#", "0"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "private_ipv4_addresses.0"),
					resource.TestCheckResourceAttr("binarylane_server.test", "port_blocking", "true"),
					resource.TestCheckResourceAttr("binarylane_server.test", "ssh_keys.#", "1"), // Only the test SSH key should be registered
					resource.TestCheckResourceAttrPair("binarylane_server.test", "ssh_keys.0", "binarylane_ssh_key.test", "id"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "permalink"),

					// Verify data source values
					resource.TestCheckResourceAttrPair("data.binarylane_server.test", "id", "binarylane_server.test", "id"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "name", "tf-test-server-resource"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "region", "per"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "image", "debian-12"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "size", "std-min"),
					resource.TestCheckResourceAttrPair("data.binarylane_server.test", "permalink", "binarylane_server.test", "permalink"),
				),
			},
			// Test import by ID
			{
				ResourceName:            "binarylane_server.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "ssh_keys", "user_data", "wait_for_create", "public_ipv4_count"},
			},
			// Test import by name
			{
				ResourceName:            "binarylane_server.test",
				ImportState:             true,
				ImportStateId:           "tf-test-server-resource",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "ssh_keys", "user_data", "wait_for_create", "public_ipv4_count"},
			},
		},
	})
}
