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
	name     = "tf_test_vpc"
	ip_range = "10.240.0.0/16"
}

data "binarylane_vpc" "test" {
  depends_on = [binarylane_vpc.test]

	id = binarylane_vpc.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify resource values
					resource.TestCheckResourceAttr("binarylane_vpc.test", "name", "tf_test_vpc"),
					resource.TestCheckResourceAttr("binarylane_vpc.test", "ip_range", "10.240.0.0/16"),
					resource.TestCheckResourceAttrSet("binarylane_vpc.test", "id"),

					// Verify data source values
					resource.TestCheckResourceAttr("data.binarylane_vpc.test", "name", "tf_test_vpc"),
					resource.TestCheckResourceAttr("data.binarylane_vpc.test", "ip_range", "10.240.0.0/16"),
					resource.TestCheckResourceAttrSet("data.binarylane_vpc.test", "id"),
				),
			},
		},
	})
}
