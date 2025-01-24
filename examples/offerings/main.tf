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

data "binarylane_images" "example" {
}

output "images" {
  value = data.binarylane_images.example.images[*].slug
}
