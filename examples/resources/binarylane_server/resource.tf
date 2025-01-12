resource "binarylane_server" "example" {
  name              = "tf-example"
  region            = "per" # or "syd", "mel", "bne", "sin"
  image             = "ubuntu-24.04"
  size              = "std-min" # 1 VPCU, 1 GB Memory, 20 GB NVME Storage, 1000 GB Data Transfer
  public_ipv4_count = 1

  # Accepts a cloud-init script or cloud-config YAML file to configure the server
  #   See more: https://cloudinit.readthedocs.io/en/latest/explanation/format.html#user-data-script
  user_data = file("./init.sh")
}
