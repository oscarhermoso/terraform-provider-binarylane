resource "binarylane_server" "example" {
  name              = "tf-example"
  region            = "per" # or "syd", "mel", "bne", "sin"
  image             = "ubuntu-24.04"
  size              = "std-min" # 1 VPCU, 1 GB Memory, 20 GB NVME Storage, 1000 GB Data Transfer
  public_ipv4_count = 1

  # When `disks` is set, `disk` is the primary disk size and the server's total allocated
  # storage is `disk + sum(disks.size_gigabytes)`. Each additional disk is matched by name
  # across plans; renaming a disk will destroy and re-create it (data loss).
  # disk = 20
  # disks = [
  #   { name = "data1", size_gigabytes = 10 },
  #   { name = "data2", size_gigabytes = 20 },
  # ]

  # Accepts a cloud-init script or cloud-config YAML file to configure the server
  #   See more: https://cloudinit.readthedocs.io/en/latest/explanation/format.html#user-data-script
  user_data = file("./init.sh")
}
