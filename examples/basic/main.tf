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
  name   = "tf-example"
  region = "per" # or "syd", "mel", "bne", "sin"
  image  = "ubuntu-24.04"
  size   = "std-min"
}
