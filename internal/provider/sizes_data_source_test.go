package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestSizesDataSource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
data "binarylane_sizes" "test" {
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source values
					resource.TestCheckResourceAttrWith("data.binarylane_sizes.test", "sizes.#", func(value string) error {
						count, err := strconv.Atoi(value)
						if err != nil {
							return err
						}
						if count < 1 {
							return fmt.Errorf("expected at least one size, got: %d", count)
						}
						return nil
					}),
					resource.TestCheckResourceAttrSet("data.binarylane_sizes.test", "sizes.0.slug"),
				),
			},
		},
	})
}
