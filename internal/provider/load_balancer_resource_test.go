package provider

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"terraform-provider-binarylane/internal/binarylane"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestLoadBalancerResource(t *testing.T) {
	// Must assign a password to the server or Binary Lane will send emails
	pw_bytes := make([]byte, 12)
	_, err := rand.Read(pw_bytes)
	if err != nil {
		t.Errorf("Failed to generate password: %s", err)
		return
	}
	password := base64.URLEncoding.EncodeToString(pw_bytes)

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
  wait_for_create   = 60
  public_ipv4_count = 1
}

resource "binarylane_load_balancer" "test" {
  name             = "tf-test-lb"
  server_ids       = [binarylane_server.test.0.id, binarylane_server.test.1.id]
  forwarding_rules = [{ entry_protocol = "http" }]
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

					// Verify data source values
					resource.TestCheckResourceAttrPair("data.binarylane_load_balancer.test", "id", "binarylane_load_balancer.test", "id"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "name", "tf-test-lb"),
					resource.TestCheckNoResourceAttr("data.binarylane_load_balancer.test", "region"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "server_ids.#", "2"),
					resource.TestCheckResourceAttrPair("data.binarylane_load_balancer.test", "server_ids.0", "binarylane_server.test.0", "id"),
					resource.TestCheckResourceAttrPair("data.binarylane_load_balancer.test", "server_ids.1", "binarylane_server.test.1", "id"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "forwarding_rules.#", "1"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "forwarding_rules.0.entry_protocol", "http"),
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
		},
	})
}

func init() {
	resource.AddTestSweepers("load_balancer", &resource.Sweeper{
		Name:         "load_balancer",
		Dependencies: []string{"server"},
		F: func(_ string) error {
			endpoint := os.Getenv("BINARYLANE_API_ENDPOINT")
			if endpoint == "" {
				endpoint = "https://api.binarylane.com.au/v2"
			}
			token := os.Getenv("BINARYLANE_API_TOKEN")

			client, err := binarylane.NewClientWithAuth(
				endpoint,
				token,
			)

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

				lbResp, err := client.GetLoadBalancersWithResponse(ctx, &params)
				if err != nil {
					return fmt.Errorf("Error getting load balancers for test sweep: %w", err)
				}

				if lbResp.StatusCode() != http.StatusOK {
					return fmt.Errorf("Unexpected status code getting load balancers for test sweep: %s", lbResp.Body)
				}

				loadBalancers := *lbResp.JSON200.LoadBalancers
				for _, lb := range loadBalancers {
					if strings.HasPrefix(*lb.Name, "tf-test-") {
						lbResp, err := client.DeleteLoadBalancersLoadBalancerIdWithResponse(ctx, *lb.Id)
						if err != nil {
							return fmt.Errorf("Error deleting load balancer %d for test sweep: %w", *lb.Id, err)
						}
						if lbResp.StatusCode() != http.StatusNoContent {
							return fmt.Errorf("Unexpected status %d deleting load balancer %d in test sweep: %s", lbResp.StatusCode(), *lb.Id, lbResp.Body)
						}
						log.Println("Deleted load balancer during test sweep:", *lb.Id)
					}
				}
				if lbResp.JSON200.Links == nil || lbResp.JSON200.Links.Pages == nil || lbResp.JSON200.Links.Pages.Next == nil {
					nextPage = false
					break
				}

				page++
			}
			return nil
		},
	})
}
