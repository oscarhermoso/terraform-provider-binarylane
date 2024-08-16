# Contributing

## Local development

### First time setup

Based on [this example from the terraform docs](https://developer.hashicorp.com/terraform/plugin/code-generation/workflow-example),

1. `go mod tidy`
2. Run `go generate` to fetch/transform the OpenAPI spec in `internal/binarylane/openapi.json`
3. Create or modify `~/.terraformrc` in your home directory

```hcl
provider_installation {

  dev_overrides {
    # Example GOBIN path, will need to be replaced with your own GOBIN path. Default is $GOPATH/bin
    "oscarhermoso/binarylane" = "/home/oscarhermoso/Git/terraform-provider-binarylane/bin"
  }

  direct {}
}
```

4. Build and test the provider

```sh
go build -o bin/terraform-provider-binarylane
go install
cd examples/basic
terraform plan
terraform apply
```

### Adding resources/data sources to the provider

1. Make any changes to `provider_gen_config.yml` (see https://developer.hashicorp.com/terraform/plugin/code-generation/openapi-generator#generator-config)
2. Update generated files

```sh
go generate
```

3. Scaffold any new resources and data sources

```sh
tfplugingen-framework scaffold resource \
    --output-dir ./internal/provider \
    --name REPLACE_ME
```

```sh
tfplugingen-framework scaffold data-source \
    --output-dir ./internal/provider \
    --name REPLACE_ME
```

4. Populate the template scaffolding:

Define the interfaces that the resource/data source should implement:

```diff
- var _ resource.Resource = (*replaceMeResource)(nil)
+ // Ensure the implementation satisfies the expected interfaces.
+ var (
+ 	_ resource.Resource              = &replaceMeResource{}
+ 	_ resource.ResourceWithConfigure = &replaceMeResource{}
+ 	// _ resource.ResourceWithImportState = &replaceMeResource{}
+ )
```

Pass the API client to the resource:

```diff
- type replaceMeResource struct{}
+ type replaceMeResource struct {
+ 	bc *BinarylaneClient
+ }
```

Add a `Configure` method to the resource:

```diff
+ func (d *replaceMeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
+ 	if req.ProviderData == nil {
+ 		return
+ 	}

+ 	bc, ok := req.ProviderData.(BinarylaneClient)
+ 	if !ok {
+ 		resp.Diagnostics.AddError(
+ 			"Unexpected Data Source Configure Type",
+ 			fmt.Sprintf("Expected *BinarylaneClient, got: %T.", req.ProviderData),
+ 		)

+ 		return
+ 	}

+ 	d.bc = &bc
+ }
```

Delete the local `*ResourceModel` struct, and import the generated `resources.*Model` instead.

```diff
- type replaceMeResourceModel struct {
- 	Id types.String `tfsdk:"id"`
- }
```

```diff
func (r *replaceMeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
-	var data replaceMeResourceModel
+	var data resources.ReplaceMeModel
```

Add the new data source/resource to `provider.go`:

```diff
  return []func() resource.Resource{
    # ...
+   NewReplaceMeResource,
  }
```
