resource "binarylane_server" "example" {
  count = 2
  # ...
}

resource "binarylane_load_balancer" "example" {
  name             = "tf-example-lb"
  forwarding_rules = [{ entry_protocol = "https" }]
  server_ids = [
    binarylane_server.example.0.id,
    binarylane_server.example.1.id
  ]
}
