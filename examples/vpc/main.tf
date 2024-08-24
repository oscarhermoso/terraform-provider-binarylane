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

resource "binarylane_vpc" "example" {
  name     = "tf-example"
  ip_range = "10.240.0.0/16"
}

# Web & NAT server
resource "binarylane_server" "web" {
  name              = "tf-vpc-example-web"
  region            = "per"
  image             = "ubuntu-24.04"
  size              = "std-min"
  vpc_id            = binarylane_vpc.example.id
  public_ipv4_count = 1
}

# Database server
resource "binarylane_server" "db" {
  name              = "tf-vpc-example-db"
  region            = "per"
  image             = "ubuntu-24.04"
  size              = "std-min"
  vpc_id            = binarylane_vpc.example.id
  public_ipv4_count = 0
}

# VPN server
resource "binarylane_server" "vpn" {
  name              = "tf-vpc-example-vpn"
  region            = "per"
  image             = "ubuntu-24.04"
  size              = "std-min"
  vpc_id            = binarylane_vpc.example.id
  public_ipv4_count = 1
}

resource "binarylane_vpc_route_entries" "example" {
  vpc_id = binarylane_vpc.example.id
  route_entries = [
    {
      description = "NAT"
      destination = "0.0.0.0/0"
      router      = binarylane_server.web.private_ipv4_addresses.0
    },
    {
      description = "VPN"
      destination = "192.168.1.0/24"
      router      = binarylane_server.vpn.private_ipv4_addresses.0
    }
  ]
}
