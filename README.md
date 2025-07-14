# terraform-provider-binarylane

See the [documentation on the Terraform Registry](https://registry.terraform.io/providers/oscarhermoso/binarylane/latest), or see [examples in the `examples` directory](./examples/README.md).

```terraform
resource "binarylane_server" "example" {
  region            = "per"
  image             = "ubuntu-24.04"
  size              = "std-min"
  public_ipv4_count = 1
}
```

## Progress

Planned features:

- [x] Servers
  - [x] Firewall Rules
  - [x] Cloud Init
  - [x] Updating server properties
  - [x] "Advanced Features"
  - [ ] Customised Backups
  - [ ] Software
  - [ ] IPV6
  - [ ] Alerts
- [x] SSH Keys
- [x] Virtual Private Cloud
- [x] Load Balancers
- [ ] Images
- [ ] DNS
- [x] Docs
  - [x] Generated docs
  - [ ] Review & fill out remaining descriptions
  - [ ] Include examples in docs
- [x] Examples
  - [x] Basic
  - [x] Cloud Init
  - [x] K3s
  - [ ] NixOS
  - [ ] Virtual Private Cloud

## Server parameters

### Regions

<details>
<summary>curl for regions</summary>

```sh
curl -X GET "https://api.binarylane.com.au/v2/regions" \
  -H "Authorization: Bearer $BINARYLANE_API_TOKEN" > tmp/regions.json

jq '[ .regions[] | .slug ] | sort' tmp/regions.json
```
</details>

```json
[
  "bne",
  "mel",
  "per",
  "sin",
  "syd"
]
```

### Images

<details>
<summary>curl for images</summary>

```sh
curl -X GET "https://api.binarylane.com.au/v2/images?type=distribution&&page=1&per_page=200" \
  -H "Authorization: Bearer $BINARYLANE_API_TOKEN" > tmp/images.json

jq '[ .images[] | .slug ] | sort' tmp/images.json
```
</details>

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

### Sizes

<details>
<summary>curl for sizes</summary>

```sh
curl -X GET "https://api.binarylane.com.au/v2/sizes" \
  -H "Authorization: Bearer $BINARYLANE_API_TOKEN" > tmp/sizes.json

jq '[ .sizes[] | .slug ] | sort' tmp/sizes.json
```
</details>

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
