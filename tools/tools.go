//go:build tools
// +build tools

package tools

import (
	_ "github.com/hashicorp/terraform-plugin-codegen-framework/cmd/tfplugingen-framework"
	_ "github.com/hashicorp/terraform-plugin-codegen-openapi/cmd/tfplugingen-openapi"
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"

	// Documentation generation
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
