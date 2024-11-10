terraform {
  required_providers {
    binarylane = {
      source = "oscarhermoso/binarylane"
    }
  }
}

provider "binarylane" {}

/**
 * This example uses the public key in the git repository, but your SSH key will
 * be different. You can use the following command to generate a new key pair:
 *
 *   ssh-keygen -t ed25519 -C "your_email@example.com"
 *
 * Then, update the "public_key" attribute to point to the location of your
 * generated public key.
 */
resource "binarylane_ssh_key" "example" {
  name = "tf-example-ssh"

  public_key = file("./id_ed25519.pub") # Change to "~/.ssh/id_ed25519.pub"

  # You also have the option to register an SSH key as a global default, for any
  # new VM but it is not the recommended approach.
  # default    = false
}

resource "binarylane_server" "example" {
  name              = "tf-example-ssh"
  region            = "per"
  image             = "ubuntu-24.04"
  size              = "std-min"
  ssh_keys          = [binarylane_ssh_key.example.id]
  public_ipv4_count = 1

  # If you are using a global default SSH key, you need to explicitly wait for
  # it to be created before creating the server.
  # depends_on = [binarylane_ssh_key.example]
}
