package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrderResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "binarylane_server" "example" {
  name     = "example"
  region   = "per"
  image    = "ubuntu-24.04"
  size     = "std-min"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of items
					resource.TestCheckResourceAttr("binarylane_server.test", "items.#", "1"),
					// Verify first order item
					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.quantity", "2"),
					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.coffee.id", "1"),
					// Verify first coffee item has Computed attributes filled.
					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.coffee.description", ""),
					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.coffee.image", "/hashicorp.png"),
					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.coffee.name", "HCP Aeropress"),
					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.coffee.price", "200"),
					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.coffee.teaser", "Automation in a cup"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("binarylane_server.test", "id"),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "last_updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "binarylane_server.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the HashiCups
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "binarylane_server" "test" {
  items = [
    {
      coffee = {
        id = 2
      }
      quantity = 2
    },
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify first order item updated
					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.quantity", "2"),
					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.coffee.id", "2"),
					// Verify first coffee item has Computed attributes updated.
					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.coffee.description", ""),
					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.coffee.image", "/packer.png"),
					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.coffee.name", "Packer Spiced Latte"),
					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.coffee.price", "350"),
					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.coffee.teaser", "Packed with goodness to spice up your images"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
