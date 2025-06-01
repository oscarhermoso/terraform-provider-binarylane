package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestServerFirewallRulesResource(t *testing.T) {
	// Must assign a password to the server or Binary Lane will send emails
	password := GenerateTestPassword(t)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `

resource "binarylane_server" "test" {
  name              = "tf-test-server-fw-rules"
  region            = "per"
  image             = "debian-12"
  size              = "std-min"
  password          = "` + password + `"
  public_ipv4_count = 0
}

resource "binarylane_server_firewall_rules" "test" {
  server_id = binarylane_server.test.id
  firewall_rules = [
    {
      description           = "Allow SSH"
      protocol              = "tcp"
      source_addresses      = ["0.0.0.0/0"]
      destination_addresses = [binarylane_server.test.private_ipv4_addresses.0]
      destination_ports     = ["22"]
      action                = "accept"
    }
  ]
}

data "binarylane_server_firewall_rules" "test" {
  server_id = binarylane_server_firewall_rules.test.server_id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify resource values
					resource.TestCheckResourceAttrSet("binarylane_server_firewall_rules.test", "server_id"),
					resource.TestCheckResourceAttr("binarylane_server_firewall_rules.test", "firewall_rules.#", "1"),
					resource.TestCheckResourceAttr("binarylane_server_firewall_rules.test", "firewall_rules.0.description", "Allow SSH"),
					resource.TestCheckResourceAttr("binarylane_server_firewall_rules.test", "firewall_rules.0.protocol", "tcp"),
					resource.TestCheckResourceAttr("binarylane_server_firewall_rules.test", "firewall_rules.0.source_addresses.#", "1"),
					resource.TestCheckResourceAttr("binarylane_server_firewall_rules.test", "firewall_rules.0.source_addresses.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("binarylane_server_firewall_rules.test", "firewall_rules.0.destination_addresses.#", "1"),
					resource.TestCheckResourceAttrSet("binarylane_server_firewall_rules.test", "firewall_rules.0.destination_addresses.0"),
					resource.TestCheckResourceAttr("binarylane_server_firewall_rules.test", "firewall_rules.0.destination_ports.#", "1"),
					resource.TestCheckResourceAttr("binarylane_server_firewall_rules.test", "firewall_rules.0.destination_ports.0", "22"),
					resource.TestCheckResourceAttr("binarylane_server_firewall_rules.test", "firewall_rules.0.action", "accept"),

					// Verify data source values
					resource.TestCheckResourceAttrSet("data.binarylane_server_firewall_rules.test", "server_id"),
					resource.TestCheckResourceAttr("data.binarylane_server_firewall_rules.test", "firewall_rules.#", "1"),
					resource.TestCheckResourceAttr("data.binarylane_server_firewall_rules.test", "firewall_rules.0.description", "Allow SSH"),
					resource.TestCheckResourceAttr("data.binarylane_server_firewall_rules.test", "firewall_rules.0.protocol", "tcp"),
					resource.TestCheckResourceAttr("data.binarylane_server_firewall_rules.test", "firewall_rules.0.source_addresses.#", "1"),
					resource.TestCheckResourceAttr("data.binarylane_server_firewall_rules.test", "firewall_rules.0.source_addresses.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("binarylane_server_firewall_rules.test", "firewall_rules.0.destination_addresses.#", "1"),
					resource.TestCheckResourceAttrSet("binarylane_server_firewall_rules.test", "firewall_rules.0.destination_addresses.0"),
					resource.TestCheckResourceAttr("data.binarylane_server_firewall_rules.test", "firewall_rules.0.destination_ports.#", "1"),
					resource.TestCheckResourceAttr("data.binarylane_server_firewall_rules.test", "firewall_rules.0.destination_ports.0", "22"),
					resource.TestCheckResourceAttr("data.binarylane_server_firewall_rules.test", "firewall_rules.0.action", "accept"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "binarylane_server_firewall_rules.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "server_id",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					resourceState := s.RootModule().Resources["binarylane_server_firewall_rules.test"]
					return resourceState.Primary.Attributes["server_id"], nil
				},
			},
			// Update and Read testing
			{
				Config: providerConfig + `

resource "binarylane_server" "test" {
  name              = "tf-test-server-fw-rules"
  region            = "per"
  image             = "debian-12"
  size              = "std-min"
  password          = "` + password + `"
  public_ipv4_count = 0
}

resource "binarylane_server_firewall_rules" "test" {
  server_id = binarylane_server.test.id
  firewall_rules = [
    {
      description           = "Allow HTTP" # Updated description
      protocol              = "tcp"
      source_addresses      = ["0.0.0.0/0"]
      destination_addresses = [binarylane_server.test.private_ipv4_addresses.0]
      destination_ports     = ["80"] # Updated port
      action                = "accept"
    },
  ]
}
	`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify resource values updated
					resource.TestCheckResourceAttrSet("binarylane_server_firewall_rules.test", "server_id"),
					resource.TestCheckResourceAttr("binarylane_server_firewall_rules.test", "firewall_rules.#", "1"),
					resource.TestCheckResourceAttr("binarylane_server_firewall_rules.test", "firewall_rules.0.description", "Allow HTTP"),
					resource.TestCheckResourceAttr("binarylane_server_firewall_rules.test", "firewall_rules.0.protocol", "tcp"),
					resource.TestCheckResourceAttr("binarylane_server_firewall_rules.test", "firewall_rules.0.source_addresses.#", "1"),
					resource.TestCheckResourceAttr("binarylane_server_firewall_rules.test", "firewall_rules.0.source_addresses.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("binarylane_server_firewall_rules.test", "firewall_rules.0.destination_addresses.#", "1"),
					resource.TestCheckResourceAttrSet("binarylane_server_firewall_rules.test", "firewall_rules.0.destination_addresses.0"),
					resource.TestCheckResourceAttr("binarylane_server_firewall_rules.test", "firewall_rules.0.destination_ports.#", "1"),
					resource.TestCheckResourceAttr("binarylane_server_firewall_rules.test", "firewall_rules.0.destination_ports.0", "80"),
					resource.TestCheckResourceAttr("binarylane_server_firewall_rules.test", "firewall_rules.0.action", "accept"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
