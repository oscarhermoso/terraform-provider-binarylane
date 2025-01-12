resource "binarylane_server" "example" {
  count = 2

  name              = "tf-example-server-${count.index}"
  region            = "per"
  image             = "debian-12"
  size              = "std-min"
  public_ipv4_count = 1
}

resource "binarylane_load_balancer" "example" {
  name             = "tf-example-lb"
  forwarding_rules = [{ entry_protocol = "http" }]
  server_ids = [
    binarylane_server.example.0.id,
    binarylane_server.example.1.id
  ]
}
