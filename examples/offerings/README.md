# examples/offerings

You can follow this example to list the available offerings for servers.

Feel free to customize this example further to filter by specific regions, operating systems, or server sizes to suit your needs.


```sh
terraform init
terraform apply -refresh-only
terraform output -json | jq -r 'with_entries(.value |= .value)'
```

Example output:

```json
{
  "images": [
    "alma-8",
    "alma-9",
    "byo-os",
    "byo-os-virtio-disabled",
    "debian-12",
    "ubuntu-20.04-neon-desktop",
    "rocky-8",
    "rocky-9",
    "ubuntu-22.04",
    "ubuntu-24.04",
    "cpanel-plus-whm",
    "debian-11",
    "ubuntu-20.04.6",
    "ubuntu-22.04-desktop",
    "windows-2022",
    "windows-2022-sql-2019-web",
    "windows-2022-sql-2019-std",
    "windows-2016",
    "windows-2019",
    "windows-2016-sql-2016-web",
    "windows-2019-sql-2017-web",
    "windows-2019-sql-2017-std"
  ],
  "regions": [
    "syd",
    "mel",
    "bne",
    "per",
    "sin"
  ],
  "sizes": [
    "std-min",
    "std-1vcpu",
    "std-2vcpu",
    "std-4vcpu",
    "std-6vcpu",
    "std-8vcpu",
    "cpu-2thr",
    "cpu-4thr",
    "cpu-6thr",
    "cpu-8thr",
    "hdd-500gb",
    "hdd-1000gb",
    "hdd-2000gb",
    "ded-e2136-400gb",
    "ded-e2136-800gb",
    "ded-e2288g-800gb",
    "ded-3900x-1600gb"
  ]
}
```
