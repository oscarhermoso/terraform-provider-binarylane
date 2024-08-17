package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestServerResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "binarylane_server" "test" {
	name   = "tf-server-resource-test"
  region = "per"
  image  = "ubuntu-24.04"
  size   = "std-min"
	user_data = <<EOT
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
					resource.TestCheckResourceAttr("binarylane_server.test", "name", "tf-server-resource-test"),
					resource.TestCheckResourceAttr("binarylane_server.test", "region", "per"),
					resource.TestCheckResourceAttr("binarylane_server.test", "image", "ubuntu-24.04"),
					resource.TestCheckResourceAttr("binarylane_server.test", "size", "std-min"),
					resource.TestCheckResourceAttr("binarylane_server.test", "user_data", `#cloud-config
echo "Hello World" > /var/tmp/output.txt
`),
					resource.TestCheckResourceAttrSet("binarylane_server.test", "id"),

					// Verify data source values
					resource.TestCheckResourceAttr("data.binarylane_server.test", "name", "tf-server-resource-test"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "region", "per"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "image", "ubuntu-24.04"),
					resource.TestCheckResourceAttr("data.binarylane_server.test", "size", "std-min"),
					resource.TestCheckResourceAttrSet("data.binarylane_server.test", "id"),
					// 					resource.TestCheckResourceAttr("data.binarylane_server.test", "user_data", `
					// #cloud-config
					// echo "Hello World" > /var/tmp/output.txt
					// 					`),
				),
			},
			// ImportState testing
			// {
			// 	ResourceName:            "binarylane_server.test",
			// 	ImportState:             true,
			// 	ImportStateVerify:       true,
			// 	ImportStateVerifyIgnore: []string{}, // nothing to ignore
			// },
			// TODO: Update and Read testing
			// 			{
			// 				Config: providerConfig + `
			// resource "binarylane_server" "test" {
			//   items = [
			//     {
			//       coffee = {
			//         id = 2
			//       }
			//       quantity = 2
			//     },
			//   ]
			// }
			// `,
			// 				Check: resource.ComposeAggregateTestCheckFunc(
			// 					// Verify first order item updated
			// 					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.quantity", "2"),
			// 					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.coffee.id", "2"),
			// 					// Verify first coffee item has Computed attributes updated.
			// 					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.coffee.description", ""),
			// 					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.coffee.image", "/packer.png"),
			// 					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.coffee.name", "Packer Spiced Latte"),
			// 					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.coffee.price", "350"),
			// 					resource.TestCheckResourceAttr("binarylane_server.test", "items.0.coffee.teaser", "Packed with goodness to spice up your images"),
			// 				),
			// 			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
