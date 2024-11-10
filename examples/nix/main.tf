terraform {
  required_providers {
    binarylane = {
      source = "oscarhermoso/binarylane"
    }
  }
}

provider "binarylane" {}

resource "binarylane_server" "example" {
  name              = "tf-nix-example"
  region            = "per"
  image             = "debian-12"
  size              = "std-min"
  user_data         = file("./cloud-config.yml")
  public_ipv4_count = 1
}
