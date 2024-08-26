terraform {
  required_providers {
    binarylane = {
      source = "oscarhermoso/binarylane"
    }
  }
}

provider "binarylane" {}

resource "binarylane_ssh_key" "example" {
  name       = "test-user"
  public_key = file("./id_ed25519.pub") # Likely "~/.ssh/id_ed25519.pub"
  default    = true                     # Currently, the server will only register SSH keys that are added as global defaults
}


resource "binarylane_server" "example" {
  name   = "tf-example"
  region = "per"
  image  = "ubuntu-24.04"
  size   = "std-min"

  depends_on = [binarylane_ssh_key.example] # Wait for SSH key to be created as a global default

  # TODO: Extend provider with "ssh_key_id" option for explicit registraton of keys
}
