# terraform-provider-binarylane

See the examples in the [examples directory](./examples/basic/main.tf).

```terraform
resource "binarylane_server" "example" {
  region = "per"
  image  = "ubuntu-24.04"
  size   = "std-min"
}
```

## WIP

If somehow you use this in production I would be pretty impressed.

- [x] Create/delete a server when runing locally
- [x] Publish
  - [x] [Terraform Registry](https://registry.terraform.io/providers/oscarhermoso/binarylane/latest/docs/resources/server)
  - [x] OpenTofu
- [ ] Deploy the rest of the owl

## Parameters

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
