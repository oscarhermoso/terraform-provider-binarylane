package provider

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"terraform-provider-binarylane/internal/binarylane"
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
}
`,
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

func init() {
	resource.AddTestSweepers("vpc", &resource.Sweeper{
		Name: "vpc",
		F: func(_ string) error {
			client, err := binarylane.NewClientWithDefaultConfig()

			if err != nil {
				return fmt.Errorf("Error creating Binary Lane API client: %w", err)
			}

			ctx := context.Background()

			var page int32 = 1
			perPage := int32(200)
			nextPage := true

			for nextPage {
				params := binarylane.GetVpcsParams{
					Page:    &page,
					PerPage: &perPage,
				}

				listResp, err := client.GetVpcsWithResponse(ctx, &params)
				if err != nil {
					return fmt.Errorf("Error getting VPCs for test sweep: %w", err)
				}

				if listResp.StatusCode() != http.StatusOK {
					return fmt.Errorf("Unexpected status code getting VPCs for test sweep: %s", listResp.Body)
				}

				vpcs := listResp.JSON200.Vpcs
				for _, vpc := range *vpcs {
					if strings.HasPrefix(*vpc.Name, "tf-test-") {

						deleteResp, err := client.DeleteVpcsVpcIdWithResponse(ctx, *vpc.Id)
						if err != nil {
							return fmt.Errorf("Error deleting VPC %d in test sweep: %w", *vpc.Id, err)
						}
						if deleteResp.StatusCode() != http.StatusNoContent {
							return fmt.Errorf("Unexpected status %d deleting VPC %d for test sweep: %s", deleteResp.StatusCode(), *vpc.Id, deleteResp.Body)
						}
						log.Println("Deleted VPC for test sweep:", *vpc.Id)
					}
				}
				if listResp.JSON200.Links == nil || listResp.JSON200.Links.Pages == nil || listResp.JSON200.Links.Pages.Next == nil {
					nextPage = false
					break
				}

				page++
			}

			return nil
		},
	})
}
