# examples/k3s

This example shows how to create a K3s cluster on Binary Lane.

After setting up the cluster by running `terraform apply`, the kubeconfig file will be downloaded to the `.kube/config` file in the current directory. You can then use `kubectl` to interact with the cluster, for example:

The SSH key used to access the servers is stored in the `.id_ed25519` file. You can use it to SSH into the servers:

```sh
ssh -o "IdentitiesOnly=yes" -i ./.id_ed25519 root@neptune-stand.bnr.la
```

## Prior art

- [MartinHodges/create_k8s](https://github.com/MartinHodges/create_k8s) repo and the [medium article](https://medium.com/@martin.hodges/creating-a-kubernetes-cluster-from-scratch-in-1-hour-using-automation-a25e387be547) that goes with it.
- https://github.com/inscapist/terraform-k3s-private-cloud
- https://github.com/schnerring/schnerring.github.io/blob/07d06fb40e3d01f2483a739f516ac4d711a5742c/content/blog/use-terraform-to-deploy-an-azure-kubernetes-service-aks-cluster-traefik-2-cert-manager-and-lets-encrypt-certificates/index.md
- https://github.com/xunleii/terraform-module-k3s
