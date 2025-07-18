locals {
  cluster_id    = "tf-example-k3s"
  custom_domain = "tf-example-k3s.internal"
}

resource "random_password" "binarylane" {
  length = 18
}

# SSH Key
# -------
resource "tls_private_key" "ed25519_provisioning" {
  algorithm = "ED25519"
}

resource "local_sensitive_file" "ssh_private_key" {
  content         = tls_private_key.ed25519_provisioning.private_key_openssh
  filename        = ".id_ed25519"
  file_permission = "0600"
}

resource "local_file" "ssh_public_key" {
  content         = tls_private_key.ed25519_provisioning.public_key_openssh
  filename        = ".id_ed25519.pub"
  file_permission = "0644"
}

resource "binarylane_ssh_key" "example" {
  name       = "tf-example-k3s"
  public_key = tls_private_key.ed25519_provisioning.public_key_openssh
}

# Virtual Private Cloud
# ---------------------
resource "binarylane_vpc" "example" {
  name     = local.cluster_id
  ip_range = "10.0.0.0/8"
}

# k3s Servers
# -----------
resource "binarylane_server" "servers" {
  count = 1

  name              = "${local.cluster_id}-server-${count.index + 1}"
  region            = "per"
  image             = "ubuntu-24.04"
  size              = "std-1vcpu" # 2GB memory, 1 vCPU
  password          = random_password.binarylane.result
  ssh_keys          = [binarylane_ssh_key.example.id]
  vpc_id            = binarylane_vpc.example.id
  public_ipv4_count = 1
}

# k3s Agents
# ----------
resource "binarylane_server" "agents" {
  count = 2

  name              = "${local.cluster_id}-agent-${count.index + 1}"
  region            = "per"
  image             = "ubuntu-24.04"
  size              = "std-min"
  password          = random_password.binarylane.result
  ssh_keys          = [binarylane_ssh_key.example.id]
  vpc_id            = binarylane_vpc.example.id
  public_ipv4_count = 1
}

# Virtual Private Cloud Routing
# -----------------------------
resource "binarylane_vpc_route_entries" "example" { # TODO
  vpc_id = binarylane_vpc.example.id
  route_entries = [
    # {
    #   description = "NAT"
    #   destination = "0.0.0.0/0"
    #   router      = binarylane_server.gateway.private_ipv4_addresses.0
    # }
  ]
}

locals {
  agent_ips  = flatten([for _, a in binarylane_server.agents : a.private_ipv4_addresses])
  server_ips = flatten([for _, s in binarylane_server.servers : s.private_ipv4_addresses])
}

resource "binarylane_server_firewall_rules" "example" {
  for_each = { for server in concat(binarylane_server.servers, binarylane_server.agents) : server.name => server }

  server_id = each.value.id

  firewall_rules = [
    {
      description           = "K3s supervisor and Kubernetes API Server"
      protocol              = "tcp"
      source_addresses      = ["0.0.0.0/0"]
      destination_addresses = local.server_ips
      destination_ports     = ["6443"]
      action                = "accept"
    },
    {
      description           = "Flannel VXLAN"
      protocol              = "udp"
      source_addresses      = [binarylane_vpc.example.ip_range]
      destination_addresses = [binarylane_vpc.example.ip_range]
      destination_ports     = ["8472"]
      action                = "accept"
    },
    {
      description           = "Kubelet metrics"
      protocol              = "tcp"
      source_addresses      = [binarylane_vpc.example.ip_range]
      destination_addresses = [binarylane_vpc.example.ip_range]
      destination_ports     = ["10250"]
      action                = "accept"
    },
    {
      description           = "SSH"
      protocol              = "udp"
      source_addresses      = ["0.0.0.0/0"]
      destination_addresses = [binarylane_vpc.example.ip_range]
      destination_ports     = ["22"]
      action                = "accept"
    },
    {
      description           = "HTTP"
      protocol              = "tcp"
      source_addresses      = ["0.0.0.0/0"]
      destination_addresses = [binarylane_vpc.example.ip_range]
      destination_ports     = ["80"]
      action                = "accept"
    },
  ]
}
