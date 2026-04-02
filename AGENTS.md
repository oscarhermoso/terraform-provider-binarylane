# AGENTS.md

## Before starting

Run `go generate` to pull in the latest API schema before making changes.

## Code style

- Follow patterns in existing resources, especially `server_resource.go` and `server_data_source.go`.
- New server attributes that map to a post-creation API action (e.g. `change_*`) follow the `source_and_destination_check` / `separate_private_network_interface` pattern: read the planned value before the first state save, call the action after creation if the value differs from the server default, and handle updates in the `Update` method.
- Run `go fmt ./...` before finishing — the CI will fail without it.
- Don't add `Required: false` when `Optional: true` is already set; it is redundant.

## Testing

- Tests live in `internal/provider/*_test.go`.
- A valid `BINARYLANE_API_TOKEN` is required; tests create real servers that cost money — be conscientious.
- **Always run the sweeper after tests**, regardless of success or failure:
  ```sh
  go test -v ./internal/provider/... -sweep=all
  ```
- Run a specific test with:
  ```sh
  go test -v -run TestServerResource ./internal/provider/...
  ```
- New attributes on the server resource should be exercised inside an existing test step where possible (e.g. the first step of `TestServerResource` already provisions a VPC server).

## Workflow

1. `go generate` — fetch latest OpenAPI spec and regenerate code + docs.
2. Make changes.
3. `go fmt ./...` — format code.
4. `go build ./...` — verify it compiles.
5. `go vet ./...` — catch common mistakes.
6. Run relevant tests, then sweep.
7. `go generate` again to regenerate docs if schema or descriptions changed.
