resource "binarylane_ssh_key" "example" {
  name       = "tf-example"
  public_key = file("~/.ssh/id_ed25519.pub") # Generate with: ssh-keygen -t ed25519 -C "name@example.com"
}

resource "binarylane_server" "example" {
  # ...
  ssh_keys = [binarylane_ssh_key.example.id]
}
