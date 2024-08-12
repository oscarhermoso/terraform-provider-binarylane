package provider

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	providerConfig = `
provider "binarylane" {
}
`
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"binarylane": providerserver.NewProtocol6WithError(New("test")()),
	}
)

const TestAccNamePrefix = "tf-acc"

func TestAccPreCheck(t *testing.T) {
	if v := os.Getenv("BINARYLANE_API_TOKEN"); v == "" {
		t.Fatal("BINARYLANE_API_TOKEN must be set for acceptance tests")
	}

	if v := os.Getenv("BINARYLANE_ACC_TEST_PROJECT"); v == "" {
		t.Fatal("BINARYLANE_ACC_TEST_PROJECT must be set for acceptance tests")
	}

	if v := os.Getenv("BINARYLANE_ACC_TEST_LOCATION"); v == "" {
		t.Fatal("BINARYLANE_ACC_TEST_LOCATION must be set for acceptance tests")
	}
}

func GetRandomResourceName(resType string) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return fmt.Sprintf("%s-%s-%s", TestAccNamePrefix, resType, string(b))
}
