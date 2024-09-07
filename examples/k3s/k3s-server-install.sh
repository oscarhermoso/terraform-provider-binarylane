#!/bin/bash -x
exec >/tmp/k3s-server-install-debug.log 2>&1

export INSTALL_K3S_NAME="${cluster_id}"
export K3S_TOKEN="${cluster_token}"
export K3S_KUBECONFIG_MODE="644"

curl -sfL https://get.k3s.io | sh -s - server \
  --node-name="$(hostname -f)"
# --disable-cloud-controller \
# --disable servicelb \
# --disable local-storage \
# --disable traefik

unset INSTALL_K3S_NAME
unset K3S_TOKEN
unset K3S_KUBECONFIG_MODE

echo "Installing Helm"
curl -sfL https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash

# Disable firewall recommended
ufw disable

echo "K3s Setup Completed"
