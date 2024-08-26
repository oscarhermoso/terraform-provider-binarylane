terraform {
  required_providers {
    binarylane = {
      source = "oscarhermoso/binarylane"
    }
  }
}

provider "binarylane" {}

resource "binarylane_server" "example" {
  name            = "tf-nix-example"
  region          = "per"
  image           = "debian-12"
  size            = "std-min"
  user_data       = file("./cloud-config.yml")
  wait_for_create = 300 # 5 mins
}
