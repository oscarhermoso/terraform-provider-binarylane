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
)

func TestLoadBalancerResource(t *testing.T) {
	// Must assign a password to the server or Binary Lane will send emails
	password := GenerateTestPassword(t)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "binarylane_server" "test" {
  count             = 2
  name              = "tf-test-lb-server-${count.index}"
  region            = "per"
  image             = "debian-12"
  size              = "std-min"
  password          = "` + password + `"
  public_ipv4_count = 1
}

resource "binarylane_load_balancer" "test" {
  name             = "tf-test-lb"
  server_ids       = [binarylane_server.test.0.id, binarylane_server.test.1.id]
	# initially, skip defining forwarding_rules, health_check, to test default values
}

data "binarylane_load_balancer" "test" {
  depends_on = [binarylane_load_balancer.test]

  id = binarylane_load_balancer.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify resource values
					resource.TestCheckResourceAttrSet("binarylane_load_balancer.test", "id"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "name", "tf-test-lb"),
					resource.TestCheckNoResourceAttr("binarylane_load_balancer.test", "region"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "server_ids.#", "2"),
					resource.TestCheckResourceAttrPair("binarylane_load_balancer.test", "server_ids.0", "binarylane_server.test.0", "id"),
					resource.TestCheckResourceAttrPair("binarylane_load_balancer.test", "server_ids.1", "binarylane_server.test.1", "id"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "forwarding_rules.#", "1"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "forwarding_rules.0.entry_protocol", "http"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "health_check.path", "/"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "health_check.protocol", "http"),

					// Verify data source values
					resource.TestCheckResourceAttrPair("data.binarylane_load_balancer.test", "id", "binarylane_load_balancer.test", "id"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "name", "tf-test-lb"),
					resource.TestCheckNoResourceAttr("data.binarylane_load_balancer.test", "region"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "server_ids.#", "2"),
					resource.TestCheckResourceAttrPair("data.binarylane_load_balancer.test", "server_ids.0", "binarylane_server.test.0", "id"),
					resource.TestCheckResourceAttrPair("data.binarylane_load_balancer.test", "server_ids.1", "binarylane_server.test.1", "id"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "forwarding_rules.#", "1"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "forwarding_rules.0.entry_protocol", "http"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "health_check.path", "/"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "health_check.protocol", "http"),
				),
			},
			// Test import by ID
			{
				ResourceName:      "binarylane_load_balancer.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Test import by name
			{
				ResourceName:      "binarylane_load_balancer.test",
				ImportState:       true,
				ImportStateId:     "tf-test-lb",
				ImportStateVerify: true,
			},
			{
				Config: providerConfig + `
resource "binarylane_server" "test" {
  count             = 2
  name              = "tf-test-lb-server-${count.index}"
  region            = "per"
  image             = "debian-12"
  size              = "std-min"
  password          = "` + password + `"
  public_ipv4_count = 1
}

resource "binarylane_load_balancer" "test" {
  name             = "tf-test-lb"
  server_ids       = [binarylane_server.test.0.id, binarylane_server.test.1.id]
	# updated
  forwarding_rules = [{ entry_protocol = "https" }]
	health_check     = {
		path 		 = "/test-health"
		protocol = "https"
	}
}

data "binarylane_load_balancer" "test" {
  depends_on = [binarylane_load_balancer.test]

  id = binarylane_load_balancer.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify resource values
					resource.TestCheckResourceAttrSet("binarylane_load_balancer.test", "id"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "name", "tf-test-lb"),
					resource.TestCheckNoResourceAttr("binarylane_load_balancer.test", "region"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "server_ids.#", "2"),
					resource.TestCheckResourceAttrPair("binarylane_load_balancer.test", "server_ids.0", "binarylane_server.test.0", "id"),
					resource.TestCheckResourceAttrPair("binarylane_load_balancer.test", "server_ids.1", "binarylane_server.test.1", "id"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "forwarding_rules.#", "1"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "forwarding_rules.0.entry_protocol", "https"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "health_check.path", "/test-health"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "health_check.protocol", "https"),

					// Verify data source values
					resource.TestCheckResourceAttrPair("data.binarylane_load_balancer.test", "id", "binarylane_load_balancer.test", "id"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "name", "tf-test-lb"),
					resource.TestCheckNoResourceAttr("data.binarylane_load_balancer.test", "region"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "server_ids.#", "2"),
					resource.TestCheckResourceAttrPair("data.binarylane_load_balancer.test", "server_ids.0", "binarylane_server.test.0", "id"),
					resource.TestCheckResourceAttrPair("data.binarylane_load_balancer.test", "server_ids.1", "binarylane_server.test.1", "id"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "forwarding_rules.#", "1"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "forwarding_rules.0.entry_protocol", "https"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "health_check.path", "/test-health"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "health_check.protocol", "https"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("load_balancer", &resource.Sweeper{
		Name:         "load_balancer",
		Dependencies: []string{"server"},
		F: func(_ string) error {
			client, err := binarylane.NewClientWithDefaultConfig()

			if err != nil {
				return fmt.Errorf("Error creating Binary Lane API client: %w", err)
			}

			ctx := context.Background()

			var page int32 = 1
			perPage := int32(200)
			var nextPage bool = true

			for nextPage {
				params := binarylane.GetLoadBalancersParams{
					Page:    &page,
					PerPage: &perPage,
				}

				listResp, err := client.GetLoadBalancersWithResponse(ctx, &params)
				if err != nil {
					return fmt.Errorf("Error getting load balancers for test sweep: %w", err)
				}

				if listResp.StatusCode() != http.StatusOK {
					return fmt.Errorf("Unexpected status code getting load balancers for test sweep: %s", listResp.Body)
				}

				loadBalancers := *listResp.JSON200.LoadBalancers
				for _, lb := range loadBalancers {
					if strings.HasPrefix(*lb.Name, "tf-test-") {
						deleteResp, err := client.DeleteLoadBalancersLoadBalancerIdWithResponse(ctx, *lb.Id)
						if err != nil {
							return fmt.Errorf("Error deleting load balancer %d for test sweep: %w", *lb.Id, err)
						}
						if deleteResp.StatusCode() != http.StatusNoContent {
							return fmt.Errorf("Unexpected status %d deleting load balancer %d in test sweep: %s", deleteResp.StatusCode(), *lb.Id, deleteResp.Body)
						}
						log.Println("Deleted load balancer during test sweep:", *lb.Id)
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
