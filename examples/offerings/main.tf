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

data "binarylane_images" "example" {}

data "binarylane_sizes" "example" {}

data "binarylane_regions" "example" {}

output "images" {
  value = data.binarylane_images.example.images[*].slug
}

output "sizes" {
  value = data.binarylane_sizes.example.sizes[*].slug
}

output "regions" {
  value = data.binarylane_regions.example.regions[*].slug
}
