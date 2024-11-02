terraform {
  required_providers {
    binarylane = {
      source = "oscarhermoso/binarylane"
    }
  }
}

provider "binarylane" {}

resource "binarylane_server" "example" {
  name              = "tf-cloud-init-example"
  region            = "per"
  image             = "ubuntu-24.04"
  size              = "std-min"
  user_data         = file("./cloud-config.yml")
  wait_for_create   = 300 # 5 mins
  public_ipv4_count = 1
}
