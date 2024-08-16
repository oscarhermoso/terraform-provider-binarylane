package main

//go:generate ./internal/binarylane/fetch-openapi.sh
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=internal/binarylane/client.cfg.yml internal/binarylane/openapi.json
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=internal/binarylane/types.cfg.yml internal/binarylane/openapi.json
//go:generate go run github.com/hashicorp/terraform-plugin-codegen-openapi/cmd/tfplugingen-openapi generate --config=./provider_gen_config.yml internal/binarylane/openapi.json --output=./provider_code_spec.json
//go:generate ./provider-transform.sh
//go:generate mkdir -p internal/resources internal/data_sources
//go:generate go run github.com/hashicorp/terraform-plugin-codegen-framework/cmd/tfplugingen-framework generate resources --input=./provider_code_spec.json --output=./internal/resources --package=resources
//go:generate go run github.com/hashicorp/terraform-plugin-codegen-framework/cmd/tfplugingen-framework generate data-sources --input=./provider_code_spec.json --output=./internal/data_sources --package=data_sources
//go:generate tfplugindocs generate
