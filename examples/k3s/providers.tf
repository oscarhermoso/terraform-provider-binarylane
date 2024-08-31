terraform {
  required_providers {
    binarylane = {
      source = "oscarhermoso/binarylane"
    }
  }
}

provider "binarylane" {}

output "permalink" {
  value = "${binarylane_server.server.0.permalink}:6443"
}

provider "helm" {
  kubernetes {
    host = "${binarylane_server.server.0.permalink}:6443"
    # token    = random_password.cluster_token.result # TODO: This token doesn't work, use client certificate instead
    insecure = true # TODO
  }
}

resource "helm_release" "nginx_ingress" {
  name = "nginx-ingress-controller"

  repository = "https://charts.bitnami.com/bitnami"
  chart      = "nginx-ingress-controller"

  set {
    name  = "service.type"
    value = binarylane_server.server.0.private_ipv4_addresses.0
  }
}
