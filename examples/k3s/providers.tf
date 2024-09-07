terraform {
  required_providers {
    binarylane = {
      source = "oscarhermoso/binarylane"
    }
    kubectl = {
      source  = "alekc/kubectl"
      version = ">= 2.0.0"
    }
  }
}

provider "binarylane" {}

provider "helm" {
  kubernetes {
    host        = "${binarylane_server.server.0.permalink}:6443"
    config_path = local_sensitive_file.kubeconfig.filename
    insecure    = true # TODO
  }
}

provider "kubectl" {
  host        = "${binarylane_server.server.0.permalink}:6443"
  config_path = local_sensitive_file.kubeconfig.filename
  insecure    = true # TODO
}
