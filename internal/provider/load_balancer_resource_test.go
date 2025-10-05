package provider

import (
	"context"
	"fmt"
	"log"
	"maps"
	"net/http"
	"slices"
	"strings"
	"terraform-provider-binarylane/internal/binarylane"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
					testCheckIfResourceAttrContainsAttr("binarylane_load_balancer.test", "server_ids", "binarylane_server.test.0", "id"),
					testCheckIfResourceAttrContainsAttr("binarylane_load_balancer.test", "server_ids", "binarylane_server.test.1", "id"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "forwarding_rules.#", "1"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "forwarding_rules.0.entry_protocol", "http"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "health_check.path", "/"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "health_check.protocol", "http"),
					resource.TestCheckResourceAttrSet("binarylane_load_balancer.test", "ip"),

					// Verify data source values
					resource.TestCheckResourceAttrPair("data.binarylane_load_balancer.test", "id", "binarylane_load_balancer.test", "id"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "name", "tf-test-lb"),
					resource.TestCheckNoResourceAttr("data.binarylane_load_balancer.test", "region"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "server_ids.#", "2"),
					testCheckIfResourceAttrContainsAttr("data.binarylane_load_balancer.test", "server_ids", "binarylane_server.test.0", "id"),
					testCheckIfResourceAttrContainsAttr("data.binarylane_load_balancer.test", "server_ids", "binarylane_server.test.1", "id"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "forwarding_rules.#", "1"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "forwarding_rules.0.entry_protocol", "http"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "health_check.path", "/"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "health_check.protocol", "http"),
					resource.TestCheckResourceAttrSet("data.binarylane_load_balancer.test", "ip"),
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
				Taint: []string{"binarylane_server.test[0]"},
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
					testCheckIfResourceAttrContainsAttr("binarylane_load_balancer.test", "server_ids", "binarylane_server.test.0", "id"),
					testCheckIfResourceAttrContainsAttr("binarylane_load_balancer.test", "server_ids", "binarylane_server.test.1", "id"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "forwarding_rules.#", "1"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "forwarding_rules.0.entry_protocol", "https"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "health_check.path", "/test-health"),
					resource.TestCheckResourceAttr("binarylane_load_balancer.test", "health_check.protocol", "https"),
					resource.TestCheckResourceAttrSet("binarylane_load_balancer.test", "ip"),

					// Verify data source values
					resource.TestCheckResourceAttrPair("data.binarylane_load_balancer.test", "id", "binarylane_load_balancer.test", "id"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "name", "tf-test-lb"),
					resource.TestCheckNoResourceAttr("data.binarylane_load_balancer.test", "region"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "server_ids.#", "2"),
					testCheckIfResourceAttrContainsAttr("data.binarylane_load_balancer.test", "server_ids", "binarylane_server.test.0", "id"),
					testCheckIfResourceAttrContainsAttr("data.binarylane_load_balancer.test", "server_ids", "binarylane_server.test.1", "id"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "forwarding_rules.#", "1"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "forwarding_rules.0.entry_protocol", "https"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "health_check.path", "/test-health"),
					resource.TestCheckResourceAttr("data.binarylane_load_balancer.test", "health_check.protocol", "https"),
					resource.TestCheckResourceAttrSet("data.binarylane_load_balancer.test", "ip"),
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
					if strings.HasPrefix(lb.Name, "tf-test-") {
						deleteResp, err := client.DeleteLoadBalancersLoadBalancerIdWithResponse(ctx, lb.Id)
						if err != nil {
							return fmt.Errorf("Error deleting load balancer %d for test sweep: %w", lb.Id, err)
						}
						if deleteResp.StatusCode() != http.StatusNoContent {
							return fmt.Errorf("Unexpected status %d deleting load balancer %d in test sweep: %s", deleteResp.StatusCode(), lb.Id, deleteResp.Body)
						}
						log.Println("Deleted load balancer during test sweep:", lb.Id)
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

func primaryInstanceState(s *terraform.State, name string) (*terraform.InstanceState, error) {
	ms := s.RootModule()

	rs, ok := ms.Resources[name]
	if !ok {
		return nil, fmt.Errorf("Not found: %s in %s, resources: %v", name, ms.Path, slices.Collect(maps.Keys(ms.Resources)))
	}
	is := rs.Primary
	if is == nil {
		return nil, fmt.Errorf("No primary instance: %s in %s", name, ms.Path)
	}
	return is, nil
}

func testCheckIfResourceAttrContainsAttr(collectionKey, collectionPath, otherResourceKey, otherAttrPath string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		collectionState, err := primaryInstanceState(s, collectionKey)
		if err != nil {
			return err
		}

		elementState, err := primaryInstanceState(s, otherResourceKey)
		if err != nil {
			return err
		}

		if collectionKey == otherResourceKey && collectionPath == otherAttrPath {
			return fmt.Errorf(
				"comparing self: resource %s attribute %s",
				collectionKey,
				collectionPath,
			)
		}

		vElement, okElement := elementState.Attributes[otherAttrPath]
		if !okElement {
			return fmt.Errorf("%s: Attribute %q not set, cannot be contained in %q from %s", otherResourceKey, otherAttrPath, collectionPath, collectionKey)
		}

		vCollection, okCollection := collectionState.Attributes[collectionPath+".#"]
		if !okCollection {
			return fmt.Errorf("%s: Attribute %q not a collection, cannot contain %q from %s", collectionKey, collectionPath, otherAttrPath, otherResourceKey)
		}

		for attrKey, attrValue := range collectionState.Attributes {
			if strings.HasPrefix(attrKey, collectionPath+".") &&
				!(strings.HasSuffix(attrKey, ".%")) &&
				!strings.HasSuffix(attrKey, ".#") &&
				attrValue == vElement {
				return nil
			}
		}

		return fmt.Errorf(
			"%s: Attribute '%s' expected to contain %#v, got %#v",
			collectionKey,
			collectionPath,
			vElement,
			vCollection)
	}
}
