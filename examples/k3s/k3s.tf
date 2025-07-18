module "k3s" {
  # source = "xunleii/k3s/module"  # Uncomment after https://github.com/xunleii/terraform-module-k3s/pull/207 is merged
  source = "github.com/dvdmuckle/terraform-module-k3s?ref=62dca7731b78f1120a141a2703e3ed14470276c0"

  depends_on_    = binarylane_server.agents
  k3s_version    = "latest"
  cluster_domain = "cluster.local"
  managed_fields = ["label"]

  // These subnets are used by the Flannel network interface in K3s, and are NOT the same as the VPC address space
  cidr = {
    pods     = "10.42.0.0/16",
    services = "10.43.0.0/16",
  }

  servers = {
    for i in range(length(binarylane_server.servers)) :
    binarylane_server.servers[i].name => {
      ip   = binarylane_server.servers[i].private_ipv4_addresses[0]
      name = binarylane_server.servers[i].name
      connection = {
        host        = binarylane_server.servers[i].permalink
        private_key = trimspace(tls_private_key.ed25519_provisioning.private_key_pem)
      }
      labels = { "node.kubernetes.io/type" = "master" }
      flags = [
        "--tls-san ${binarylane_server.servers[0].private_ipv4_addresses[0]}",
      ]
    }
  }

  agents = {
    for i in range(length(binarylane_server.agents)) :
    binarylane_server.agents[i].name => {
      ip   = binarylane_server.agents[i].private_ipv4_addresses[0]
      name = binarylane_server.agents[i].name
      connection = {
        host        = binarylane_server.agents[i].permalink
        private_key = trimspace(tls_private_key.ed25519_provisioning.private_key_pem)
      }
      labels = { "node.kubernetes.io/pool" = "general" }
    }
  }
}

resource "local_sensitive_file" "kube_config" {
  content         = module.k3s.kube_config
  filename        = ".kube/config"
  file_permission = "0644"
}
