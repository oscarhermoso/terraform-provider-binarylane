package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestVpcResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
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
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify resource values
					resource.TestCheckResourceAttr("binarylane_vpc.test", "name", "tf-test-vpc"),
					resource.TestCheckResourceAttr("binarylane_vpc.test", "ip_range", "10.240.0.0/16"),
					resource.TestCheckResourceAttrSet("binarylane_vpc.test", "id"),
					// Verify data source values
					resource.TestCheckResourceAttr("data.binarylane_vpc.test", "name", "tf-test-vpc"),
					resource.TestCheckResourceAttr("data.binarylane_vpc.test", "ip_range", "10.240.0.0/16"),
					resource.TestCheckResourceAttrSet("data.binarylane_vpc.test", "id"),

					// Verify resource values
					resource.TestCheckResourceAttrSet("binarylane_vpc_route_entries.test", "vpc_id"),
					resource.TestCheckResourceAttr("binarylane_vpc_route_entries.test", "route_entries.#", "1"),
					resource.TestCheckResourceAttr("binarylane_vpc_route_entries.test", "route_entries.0.destination", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("binarylane_vpc_route_entries.test", "route_entries.0.router", "10.240.0.1"),
					resource.TestCheckResourceAttr("binarylane_vpc_route_entries.test", "route_entries.0.description", "test"),
				),
			},
		},
	})
}
