terraform {
  required_providers {
    binarylane = {
      source = "oscarhermoso/binarylane"
    }
  }
}

provider "binarylane" {
  # api_token = "" # Recommend setting with environment variable BINARYLANE_API_TOKEN
}

resource "binarylane_server" "example" {
  name              = "tf-example-basic"
  region            = "per" # or "syd", "mel", "bne", "sin"
  image             = "ubuntu-24.04"
  size              = "std-min" # 1 VPCU, 1 GB Memory,  20 GB NVME Storage, 1000 GB Data Transfer
  public_ipv4_count = 1
  # Password will be generated automatically and mailed to the account holder
}
