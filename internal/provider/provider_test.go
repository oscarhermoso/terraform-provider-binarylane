package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	providerConfig = `
provider "binarylane" {
  # api_token = "test123"
	# api_endpoint = "https://test.binarylane.internal.au/v2"
}

# End of provider config

`

	// planOnlyProviderConfig configures the provider with a placeholder token, for tests that
	// only ever plan and so never reach the API. providerConfig leaves the token to
	// BINARYLANE_API_TOKEN, which those tests would otherwise need set to run at all.
	planOnlyProviderConfig = `
provider "binarylane" {
  api_token = "placeholder-never-used"
}

# End of provider config

`

	// testRegion is the BinaryLane region acceptance tests create servers in.
	// TEMPORARY: set to "mel" because "per" is intermittently returning "Unable to find a
	// suitable host for the requested server configuration". Revert to "per" once that's
	// resolved.
	testRegion = "mel"
)

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"binarylane": providerserver.NewProtocol6WithError(New("test")()),
	}
)
