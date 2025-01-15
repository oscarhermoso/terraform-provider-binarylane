resource "binarylane_vpc" "example" {
  name     = "tf-example-vpc"
  ip_range = "10.240.0.0/16"
}

resource "binarylane_server" "web" {
  # ...
  vpc_id = binarylane_vpc.example.id
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
