# binary-lane-terraform-provider

## Following the example from the terraform docs

https://developer.hashicorp.com/terraform/plugin/code-generation/workflow-example

Initial setup

```sh
mkdir terraform-provider-binarylane
cd terraform-provider-binarylane
go mod init terraform-provider-binarylane
touch main.go
touch provider.go
```

1. Run `go genereate` to generate the OpenAPI spec in `internal/binarylane/openapi.yml`
2. Make any changes to `provider_gen_config.yml` (see https://developer.hashicorp.com/terraform/plugin/code-generation/openapi-generator#generator-config)
3. Generate JSON spec for provider code

```sh
go install github.com/hashicorp/terraform-plugin-codegen-openapi/cmd/tfplugingen-openapi@latest

tfplugingen-openapi generate \
  --config ./provider_gen_config.yml \
  ./openapi.yml \
  --output ./provider_code_spec.json
```

4. Generate code for provider, resources and data sources

```sh
go install github.com/hashicorp/terraform-plugin-codegen-framework/cmd/tfplugingen-framework@latest

tfplugingen-framework generate all \
    --input provider_code_spec.json \
    --force \
    --output internal
```

5. Scaffold provider, resources and data sources

```sh
mkdir -p internal/provider

tfplugingen-framework scaffold provider \
  --name binarylane \
  --output-dir ./internal/provider
```

```sh
tfplugingen-framework scaffold data-source \
    --name REPLACE_ME \
    --output-dir ./internal/provider
```

```sh
tfplugingen-framework scaffold data-source \
    --name REPLACE_ME \
    --output-dir ./internal/provider
```

6. Add this to your `~/.terraformrc` (your home directory!)

```hcl
provider_installation {

  dev_overrides {
    # Example GOBIN path, will need to be replaced with your own GOBIN path. Default is $GOPATH/bin
    "hashicorp.com/edu/binarylane" = "/home/oscarhermoso/Git/terraform-provider-binarylane/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

## Notes

Getting a list of images

```sh
curl -X GET "https://api.binarylane.com.au/v2/images?type=distribution&&page=1&per_page=200" \
  -H "Authorization: Bearer **********" > images.json

jq '[ .images[] | .slug ] | sort' images.json
```

Getting a list of regions

```sh
curl -X GET "https://api.binarylane.com.au/v2/regions" \
  -H "Authorization: Bearer **********"" > regions.json

jq '[ .regions[] | .slug ] | sort' regions.json

[
  "bne",
  "mel",
  "per",
  "sin",
  "syd"
]
```

Getting a list of sizes

```sh
curl -X GET "https://api.binarylane.com.au/v2/sizes" \
  -H "Authorization: Bearer **********"" > sizes.json

jq '[ .sizes[] | .slug ] | sort' sizes.json

[
  "cpu-2thr",
  "cpu-4thr",
  "cpu-6thr",
  "cpu-8thr",
  "ded-3900x-1600gb",
  "ded-e2136-400gb",
  "ded-e2136-800gb",
  "ded-e2288g-400gb",
  "ded-e2288g-800gb",
  "hdd-1000gb",
  "hdd-2000gb",
  "hdd-500gb",
  "std-1vcpu",
  "std-2vcpu",
  "std-4vcpu",
  "std-6vcpu",
  "std-8vcpu",
  "std-min"
]
```
