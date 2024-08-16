# examples/ssh

## Generate SSH key pair

https://docs.github.com/en/authentication/connecting-to-github-with-ssh/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent

## Testing

```sh
ssh-keygen -t ed25519 -C "test@company.internal" \
  -f ./id_ed25519 -N ""
```
