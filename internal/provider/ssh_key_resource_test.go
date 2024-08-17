package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestSshKeyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "binarylane_ssh_key" "test" {
	name       = "tf_ssh_key_resource_test"
	public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJCsuosklP0T4fJcQDgkeVh7dQu+eV+vev1CfwdUkj7h test@company.internal"
}

data "binarylane_ssh_key" "test" {
  depends_on = [binarylane_ssh_key.test]

	id = binarylane_ssh_key.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify resource values
					resource.TestCheckResourceAttr("binarylane_ssh_key.test", "name", "tf_ssh_key_resource_test"),
					resource.TestCheckResourceAttr("binarylane_ssh_key.test", "public_key", "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJCsuosklP0T4fJcQDgkeVh7dQu+eV+vev1CfwdUkj7h test@company.internal"),
					resource.TestCheckResourceAttr("binarylane_ssh_key.test", "default", "false"),
					resource.TestCheckResourceAttrSet("binarylane_ssh_key.test", "id"),

					// Verify data source values
					resource.TestCheckResourceAttr("data.binarylane_ssh_key.test", "name", "tf_ssh_key_resource_test"),
					resource.TestCheckResourceAttr("data.binarylane_ssh_key.test", "public_key", "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJCsuosklP0T4fJcQDgkeVh7dQu+eV+vev1CfwdUkj7h test@company.internal"),
					resource.TestCheckResourceAttr("data.binarylane_ssh_key.test", "default", "false"),
					resource.TestCheckResourceAttrSet("data.binarylane_ssh_key.test", "id"),
				),
			},
			// ImportState testing
			// TODO
			// {
			// 	ResourceName:            "binarylane_ssh_key.test",
			// 	ImportState:             true,
			// 	ImportStateVerify:       true,
			// 	ImportStateVerifyIgnore: []string{}, // nothing to ignore
			// },
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
