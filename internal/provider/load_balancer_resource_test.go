package provider

import (
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestLoadBalancerResource(t *testing.T) {
	// Must assign a password to the server or Binary Lane will send emails
	pw_bytes := make([]byte, 12)
	_, err := rand.Read(pw_bytes)
	if err != nil {
		t.Errorf("Failed to generate password: %s", err)
		return
	}
	password := base64.URLEncoding.EncodeToString(pw_bytes)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "binarylane_server" "test" {
	count  = 2
	name   = "tf-test-lb-server-${count.index}"
  region = "per"
  image  = "debian-12"
  size   = "std-min"
	password = "` + password + `"
	wait_for_create = 60
}

resource "binarylane_load_balancer" "test" {
	name   = "tf-test-lb"
	server_ids = [binarylane_server.test.0.id, binarylane_server.test.1.id]
	forwarding_rules = [{ entry_protocol = "http" }]
}

data "binarylane_load_balancer" "test" {
  depends_on = [binarylane_load_balancer.test]

	id = binarylane_load_balancer.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify resource values
					resource.TestCheckResourceAttrSet("binarylane_load_balancer.test", "id"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "name", "tf-test-lb"),
					resource.TestCheckNoResourceAttr("binarylane_load_balancer.test", "region"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "server_ids.#", "2"),
					resource.TestCheckResourceAttrPair("binarylane_load_balancer.test", "server_ids.0", "binarylane_server.test.0", "id"),
					resource.TestCheckResourceAttrPair("binarylane_load_balancer.test", "server_ids.1", "binarylane_server.test.1", "id"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "forwarding_rules.#", "1"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "forwarding_rules.0.entry_protocol", "http"),

					// Verify data source values
					resource.TestCheckResourceAttrPair("data.binarylane_load_balancer.test", "id", "binarylane_load_balancer.test", "id"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "name", "tf-test-lb"),
					resource.TestCheckNoResourceAttr("data.binarylane_load_balancer.test", "region"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "server_ids.#", "2"),
					resource.TestCheckResourceAttrPair("data.binarylane_load_balancer.test", "server_ids.0", "binarylane_server.test.0", "id"),
					resource.TestCheckResourceAttrPair("data.binarylane_load_balancer.test", "server_ids.1", "binarylane_server.test.1", "id"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "forwarding_rules.#", "1"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "forwarding_rules.0.entry_protocol", "http"),
				),
			},
			// Test import by ID
			{
				ResourceName:      "binarylane_load_balancer.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Test import by name
			{
				ResourceName:      "binarylane_load_balancer.test",
				ImportState:       true,
				ImportStateId:     "tf-test-lb",
				ImportStateVerify: true,
			},
		},
	})
}
