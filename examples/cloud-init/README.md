# examples/cloud-init

Binary lane servers initialize cloud-init configurations to the server upon creation.

If you want to deploy multiple parts of a cloud-init configuration, you can use the [cloudinit_config](https://registry.terraform.io/providers/hashicorp/cloudinit/latest/docs/data-sources/config) data source.

```terraform
terraform {
  required_providers {
    binarylane = {
      source = "oscarhermoso/binarylane"
    }
  }
}

provider "binarylane" {
  # api_url   = "" # Defaults to "https://api.binarylane.com.au/v2"
  # api_token = "" # Recommend setting with environment variable BINARYLANE_API_TOKEN
}

data "cloudinit_config" "example" {
  gzip          = false
  base64_encode = false

  part {
    filename     = "hello-script.sh"
    content_type = "text/x-shellscript"

    content = file("./hello-script.sh")
  }

  part {
    filename     = "cloud-config.yaml"
    content_type = "text/cloud-config"

    content = file("./cloud-config.yml")
  }
}

resource "binarylane_server" "example" {
  name      = "tf-nix-example"
  region    = "per"
  image     = "ubuntu-24.04"
  size      = "std-min"
  user_data = data.cloudinit_config.example.rendered
}
```
