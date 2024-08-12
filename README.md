# binary-lane-terraform-provider

See the examples in the [examples directory](./examples/basic/main.tf).

```terraform
resource "binarylane_server" "example" {
  name   = "example"
  region = "per"
  image  = "ubuntu-24.04"
  size   = "std-min"
}
```

## WIP

If somehow you use this in production I would be pretty impressed.

- [x] Create/delete a server when runing locally
- [ ] Publish to Terraform Registry (and maybe OpenTofu?)
- [ ] Deploy the rest of the owl

## Local development


Based on [this example from the terraform docs](https://developer.hashicorp.com/terraform/plugin/code-generation/workflow-example),

1. `go mod tidy`
2. Run `go generate` to fetch/transform the OpenAPI spec in `internal/binarylane/openapi.json`
3. Make any changes to `provider_gen_config.yml` (see https://developer.hashicorp.com/terraform/plugin/code-generation/openapi-generator#generator-config), and run `go generate` again
4. Scaffold any new resources and data sources

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

5. Create or modify `~/.terraformrc` in your home directory

```hcl
provider_installation {

  dev_overrides {
    # Example GOBIN path, will need to be replaced with your own GOBIN path. Default is $GOPATH/bin
    "hashicorp.com/oscarhermoso/binarylane" = "/home/oscarhermoso/Git/terraform-provider-binarylane/bin"
  }

  direct {}
}
```

6. Build and test the provider

```sh
go build -o bin/terraform-provider-binarylane
go install
cd examples/resources/binarylane_image
terraform plan
terraform apply
```

## Notes

Images:

```sh
curl -X GET "https://api.binarylane.com.au/v2/images?type=distribution&&page=1&per_page=200" \
  -H "Authorization: Bearer **********" > tmp/images.json

jq '[ .images[] | .slug ] | sort' tmp/images.json
```

```json
[
  "alma-8",
  "alma-9",
  "byo-os",
  "byo-os-virtio-disabled",
  "cpanel-plus-whm",
  "debian-11",
  "debian-12",
  "rocky-8",
  "rocky-9",
  "ubuntu-20.04-neon-desktop",
  "ubuntu-20.04.6",
  "ubuntu-22.04",
  "ubuntu-22.04-desktop",
  "ubuntu-24.04",
  "windows-2012-r2",
  "windows-2016",
  "windows-2016-sql-2016-web",
  "windows-2019",
  "windows-2019-sql-2017-std",
  "windows-2019-sql-2017-web",
  "windows-2022",
  "windows-2022-sql-2019-std",
  "windows-2022-sql-2019-web"
]
```

Regions:

```sh
curl -X GET "https://api.binarylane.com.au/v2/regions" \
  -H "Authorization: Bearer **********"" > tmp/regions.json

jq '[ .regions[] | .slug ] | sort' tmp/regions.json
```

```json
[
  "bne",
  "mel",
  "per",
  "sin",
  "syd"
]
```

Sizes (not all sizes are available in all regions):

```sh
curl -X GET "https://api.binarylane.com.au/v2/sizes" \
  -H "Authorization: Bearer **********"" > tmp/sizes.json

jq '[ .sizes[] | .slug ] | sort' tmp/sizes.json
```

```json
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
