package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestImagesDataSource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
data "binarylane_images" "test" {
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source values
					resource.TestCheckResourceAttrWith("data.binarylane_images.test", "images.#", func(value string) error {
						count, err := strconv.Atoi(value)
						if err != nil {
							return err
						}
						if count < 1 {
							return fmt.Errorf("expected at least one image, got: %d", count)
						}
						return nil
					}),
					resource.TestCheckResourceAttrSet("data.binarylane_images.test", "images.0.id"),
				),
			},
		},
	})
}
