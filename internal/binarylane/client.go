package binarylane

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewClientWithAuth(endpoint string, token string) (*ClientWithResponses, error) {
	if token == "" {
		return nil, errors.New("missing or empty value for the Binary Lane API " +
			"token. Set the `api_token` value in the configuration or use the " +
			"BINARYLANE_API_TOKEN environment variable. If either is already set, " +
			"ensure the value is not empty")
	}

	auth, err := securityprovider.NewSecurityProviderBearerToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client with supplied API token: %w", err)
	}

	if endpoint == "" {
		endpoint = "https://api.binarylane.com.au/v2"
	}

	client, err := NewClientWithResponses(
		endpoint,
		WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			dump, err := httputil.DumpRequestOut(req, true)
			if err != nil {
				return err
			}
			tflog.Debug(ctx, fmt.Sprintf("%q\n", dump))
			return nil
		}),
		WithRequestEditorFn(auth.Intercept), // include auth AFTER the request logger
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create Binary Lane API client: %w", err)
	}

	return client, nil
}
