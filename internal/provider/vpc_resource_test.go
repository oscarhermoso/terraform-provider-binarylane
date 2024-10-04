package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestVpcResource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "binarylane_vpc" "test" {
	name     = "tf-test-vpc"
	ip_range = "10.240.0.0/16"
}

data "binarylane_vpc" "test" {
  depends_on = [binarylane_vpc.test]

	id = binarylane_vpc.test.id
}

resource "binarylane_vpc_route_entries" "test" {
	vpc_id = binarylane_vpc.test.id
	route_entries = [
		{
			description = "test"
			destination = "0.0.0.0/0"
			router      = "10.240.0.1"
		}
	]
}

data "binarylane_vpc_route_entries" "test" {
	vpc_id = binarylane_vpc_route_entries.test.vpc_id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify binarylane_vpc resource values
					resource.TestCheckResourceAttr("binarylane_vpc.test", "name", "tf-test-vpc"),
					resource.TestCheckResourceAttr("binarylane_vpc.test", "ip_range", "10.240.0.0/16"),
					resource.TestCheckResourceAttrSet("binarylane_vpc.test", "id"),
					// Verify binarylane_vpc data source values
					resource.TestCheckResourceAttr("data.binarylane_vpc.test", "name", "tf-test-vpc"),
					resource.TestCheckResourceAttr("data.binarylane_vpc.test", "ip_range", "10.240.0.0/16"),
					resource.TestCheckResourceAttrSet("data.binarylane_vpc.test", "id"),
					// Verify binarylane_vpc_route_entries resource values
					resource.TestCheckResourceAttrSet("binarylane_vpc_route_entries.test", "vpc_id"),
					resource.TestCheckResourceAttr("binarylane_vpc_route_entries.test", "route_entries.#", "1"),
					resource.TestCheckResourceAttr("binarylane_vpc_route_entries.test", "route_entries.0.destination", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("binarylane_vpc_route_entries.test", "route_entries.0.router", "10.240.0.1"),
					resource.TestCheckResourceAttr("binarylane_vpc_route_entries.test", "route_entries.0.description", "test"),
					// Verify binarylane_vpc_route_entries data source values
					resource.TestCheckResourceAttrSet("data.binarylane_vpc_route_entries.test", "vpc_id"),
					resource.TestCheckResourceAttr("data.binarylane_vpc_route_entries.test", "route_entries.#", "1"),
					resource.TestCheckResourceAttr("data.binarylane_vpc_route_entries.test", "route_entries.0.destination", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("data.binarylane_vpc_route_entries.test", "route_entries.0.router", "10.240.0.1"),
					resource.TestCheckResourceAttr("data.binarylane_vpc_route_entries.test", "route_entries.0.description", "test"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "binarylane_vpc.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:                         "binarylane_vpc_route_entries.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "vpc_id",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					resourceState := s.RootModule().Resources["binarylane_vpc_route_entries.test"]
					return resourceState.Primary.Attributes["vpc_id"], nil
				},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "binarylane_vpc" "test" {
	name     = "tf-test-vpc-renamed"
	ip_range = "10.240.0.0/16"
}

resource "binarylane_vpc_route_entries" "test" {
	vpc_id = binarylane_vpc.test.id
	route_entries = [
		{
			description = "test-renamed"
			destination = "0.0.0.0/0"
			router      = "10.240.0.2"
		}
	]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify binarylane_vpc resource values updated
					resource.TestCheckResourceAttr("binarylane_vpc.test", "name", "tf-test-vpc-renamed"), // Updated name
					resource.TestCheckResourceAttr("binarylane_vpc.test", "ip_range", "10.240.0.0/16"),
					resource.TestCheckResourceAttrSet("binarylane_vpc.test", "id"),
					// Verify binarylane_vpc_route_entries resource values updated
					resource.TestCheckResourceAttrSet("binarylane_vpc_route_entries.test", "vpc_id"),
					resource.TestCheckResourceAttr("binarylane_vpc_route_entries.test", "route_entries.#", "1"),
					resource.TestCheckResourceAttr("binarylane_vpc_route_entries.test", "route_entries.0.destination", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("binarylane_vpc_route_entries.test", "route_entries.0.router", "10.240.0.2"),        // Updated router
					resource.TestCheckResourceAttr("binarylane_vpc_route_entries.test", "route_entries.0.description", "test-renamed"), // Updated description
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
