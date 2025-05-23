resource "binarylane_server" "example" {
  # ...
}

resource "binarylane_server_firewall_rules" "example" {
  server_id = binarylane_server.example.id
  firewall_rules = [
    {
      description           = "Allow SSH"
      protocol              = "tcp"
      source_addresses      = ["0.0.0.0/0"]
      destination_addresses = [binarylane_server.example.private_ipv4_addresses.0]
      destination_ports     = ["22"]
      action                = "accept"
    }
  ]
}
