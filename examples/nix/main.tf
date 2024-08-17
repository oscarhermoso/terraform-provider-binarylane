terraform {
  required_providers {
    binarylane = {
      source = "oscarhermoso/binarylane"
    }
  }
}

provider "binarylane" {
  # api_url   = "" # Defaults to "https://api.binarylane.com"
  # api_token = "" # Recommend setting with environment variable BINARYLANE_API_TOKEN
}

resource "binarylane_server" "example" {
  name      = "tf-nix-example"
  region    = "per"
  image     = "debian-12"
  size      = "std-min"
  user_data = file("./cloud-config.yml")
}
