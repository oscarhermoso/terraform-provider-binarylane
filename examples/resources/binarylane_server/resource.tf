resource "binarylane_server" "example" {
  name              = "tf-example"
  region            = "per" # or "syd", "mel", "bne", "sin"
  image             = "ubuntu-24.04"
  size              = "std-min" # 1 VPCU, 1 GB Memory, 20 GB NVME Storage, 1000 GB Data Transfer
  public_ipv4_count = 1

  # To manage multiple disks, every disk must be listed, including the primary
  # disk = 40
  # disks = [
  #   { primary = true, size_gigabytes = 20 },
  #   { description = "data1", size_gigabytes = 10 },
  #   { description = "data2", size_gigabytes = 10 },
  # ]

  # Accepts a cloud-init script or cloud-config YAML file to configure the server
  #   See more: https://cloudinit.readthedocs.io/en/latest/explanation/format.html#user-data-script
  user_data = file("./init.sh")
}
