resource "binarylane_server" "example" {
  name              = "tf-example"
  region            = "per"
  image             = "ubuntu-24.04"
  size              = "std-min"
  password          = random_password.example.result
  ssh_keys          = [binarylane_ssh_key.example.id]
  public_ipv4_count = 1
}

resource "binarylane_ssh_key" "example" {
  name       = "tf-example"
  public_key = file("~/.ssh/id_ed25519.pub") # Generate with: ssh-keygen -t ed25519 -C "name@example.com"
}

resource "random_password" "example" {
  length = 32
}
