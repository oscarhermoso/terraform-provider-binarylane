resource "binarylane_ssh_key" "example" {
  name = "tf-example"

  # Generate with: ssh-keygen -t ed25519 -C "name@example.com"
  public_key = file("~/.ssh/id_ed25519.pub")
}

resource "binarylane_server" "example" {
  # ...
  ssh_keys = [binarylane_ssh_key.example.id]
}
