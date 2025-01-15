resource "binarylane_vpc" "example" {
  name     = "tf-example-vpc"
  ip_range = "10.240.0.0/16"
}

resource "binarylane_server" "example" {
  # ...
  vpc_id = binarylane_vpc.example.id
}
