# WIP - examples/k8s

This WIP example shows how to create a Kubernetes cluster on Binary Lane.

## Prior art

- [MartinHodges/create_k8s](https://github.com/MartinHodges/create_k8s) repo and the [medium article](https://medium.com/@martin.hodges/creating-a-kubernetes-cluster-from-scratch-in-1-hour-using-automation-a25e387be547) that goes with it.
- https://github.com/inscapist/terraform-k3s-private-cloud
- https://github.com/schnerring/schnerring.github.io/blob/07d06fb40e3d01f2483a739f516ac4d711a5742c/content/blog/use-terraform-to-deploy-an-azure-kubernetes-service-aks-cluster-traefik-2-cert-manager-and-lets-encrypt-certificates/index.md

## Remaining TODOs

- [ ] Consider rewriting module to use `microk8s` instead of `k3s`
- [ ] Use `kustomize` instead of `helm`
- [ ] Use the `binarylane_load_balancer` resource instead of `kubectl port-forward` (or use the load balancer in another example)
