locals {
  cluster_id    = "tf-example-k8s"
  custom_domain = "tf-example-k8s.internal"
}

resource "random_password" "binarylane" {
  length = 18
}

resource "random_password" "cluster_token" {
  length  = 64
  special = false
}

# Virtual Private Cloud
# ---------------------
resource "binarylane_vpc" "example" {
  name     = local.cluster_id
  ip_range = "10.240.0.0/16"
}

# k3s Servers
# -----------
data "cloudinit_config" "server" {
  gzip          = false
  base64_encode = false

  part {
    content_type = "text/x-shellscript"
    content = templatefile("${path.module}/k3s-server-install.sh", {
      cluster_id    = local.cluster_id,
      cluster_token = random_password.cluster_token.result,
    })
  }
}

resource "binarylane_server" "server" {
  count = 1

  name              = "${local.cluster_id}-server-${count.index + 1}"
  region            = "per"
  image             = "ubuntu-24.04"
  size              = "std-min"
  password          = random_password.binarylane.result
  vpc_id            = binarylane_vpc.example.id
  public_ipv4_count = 1
  user_data         = sensitive(data.cloudinit_config.server.rendered)
}

# k3s Agents
# ----------
data "cloudinit_config" "agent" {
  gzip          = false
  base64_encode = false

  part {
    content_type = "text/x-shellscript"
    content = templatefile("${path.module}/k3s-agent-join.sh", {
      cluster_token  = random_password.cluster_token.result,
      cluster_server = binarylane_server.server.0.private_ipv4_addresses.0
    })
  }
}

resource "binarylane_server" "agent" {
  count = 2

  name              = "${local.cluster_id}-agent-${count.index + 1}"
  region            = "per"
  image             = "ubuntu-24.04"
  size              = "std-min"
  password          = random_password.binarylane.result
  vpc_id            = binarylane_vpc.example.id
  public_ipv4_count = 1
  user_data         = sensitive(data.cloudinit_config.agent.rendered)
}

# Virtual Private Cloud Routing
# -----------------------------
resource "binarylane_vpc_route_entries" "example" {
  vpc_id = binarylane_vpc.example.id
  route_entries = [
    # {
    #   description = "NAT"
    #   destination = "0.0.0.0/0"
    #   router      = binarylane_server.gateway.private_ipv4_addresses.0
    # }
  ]
}

resource "binarylane_server_firewall_rules" "example" {
  for_each = { for server in flatten([binarylane_server.server, binarylane_server.agent]) : server.id => server }

  server_id = each.value.id

  firewall_rules = [
    {
      description           = "K3s supervisor and Kubernetes API Server"
      protocol              = "tcp"
      source_addresses      = flatten([for _, a in binarylane_server.agent : a.private_ipv4_addresses])
      destination_addresses = flatten([for _, s in binarylane_server.server : s.private_ipv4_addresses])
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
  ]
}
