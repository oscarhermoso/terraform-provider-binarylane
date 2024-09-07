#!/bin/bash -x
exec >/tmp/k3s-agent-join-debug.log 2>&1

export K3S_TOKEN="${cluster_token}"
export K3S_URL="https://${cluster_server}:6443"

curl -sfL https://get.k3s.io | sh -s - agent \
  --node-name="$(hostname -f)"

unset K3S_TOKEN
unset K3S_URL

# Disable firewall recommended
ufw disable

echo "K3s Node Join Completed"
