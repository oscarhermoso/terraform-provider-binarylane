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

### Testing


```sh
cp .env.example .env
# Add your API token to .env
eval export $(cat .env)
go test -v ./internal/provider/...
```

### Update modules

```sh
go get -u && go mod tidy && go generate
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

#### Resource

Define the interfaces that the resource should implement:

```diff
- var _ resource.Resource = (*exampleResource)(nil)
+ var (
+ 	_ resource.Resource                = &exampleResource{}
+ 	_ resource.ResourceWithConfigure   = &exampleResource{}
+ 	_ resource.ResourceWithImportState = &exampleResource{}
+ )
```

Add the API client to the resource:

```diff
- type exampleResource struct{}
+ type exampleResource struct {
+ 	bc *BinarylaneClient
+ }
```

Extend the resource model from the generated `resources.*Model`.

```diff
  type exampleResourceModel struct {
- 	Id types.String `tfsdk:"id"`
+ 	resources.ExampleModel
+   // Add any additional fields here
  }
```

Add a `Configure` method to the resource:

```diff
+ func (d *exampleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
+ 	if req.ProviderData == nil {
+ 		return
+ 	}

+ 	bc, ok := req.ProviderData.(BinarylaneClient)
+ 	if !ok {
+ 		resp.Diagnostics.AddError(
+ 			"Unexpected Resource Configure Type",
+ 			fmt.Sprintf("Expected *BinarylaneClient, got: %T.", req.ProviderData),
+ 		)

+ 		return
+ 	}

+ 	d.bc = &bc
+ }
```

```diff
+ func (r *exampleResource) ImportState(
+   ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse,
+ ) {
+ 	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
+ }
```

Add the new resource to `provider.go`:

```diff
  return []func() resource.Resource{
    # ...
+   NewExampleResource,
  }
```

#### Data Source

Define the interfaces that the data source should implement:

```diff
- var _ resource.Resource = (*exampleDataSource)(nil)
+ var (
+ 	_ datasource.DataSource              = &exampleDataSource{}
+ 	_ datasource.DataSourceWithConfigure = &exampleDataSource{}
+ )

Add the API client to the data source:

```diff
- type exampleResource struct{}
+ type exampleResource struct {
+ 	bc *BinarylaneClient
+ }
```

Derive the data source model from the generated `resources.*Model`.

```diff
  type exampleDataSourceModel struct {
- 	Id types.String `tfsdk:"id"`
+ 	resources.ExampleModel
  }
```

Add a `Configure` method to the data source.

```diff
+ func (d *exampleDataSource) Configure(
+   _ context.Context,
+   req datasource.ConfigureRequest,
+   resp *datasource.ConfigureResponse,
+ ) {
+ 	if req.ProviderData == nil {
+ 		return
+ 	}
+
+ 	bc, ok := req.ProviderData.(BinarylaneClient)
+ 	if !ok {
+ 		resp.Diagnostics.AddError(
+ 			"Unexpected Data Source Configure Type",
+ 			fmt.Sprintf("Expected *BinarylaneClient, got: %T.", req.ProviderData))
+ 		return
+ 	}
+
+ 	d.bc = &bc
+ }
```

Use the generated schema to define the data source schema.

```diff
  func (d *exampleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
- 	resp.Schema = schema.Schema{
- 		Attributes: map[string]schema.Attribute{
- 			"id": schema.StringAttribute{
- 				Computed: true,
- 			},
- 		},
- 	}
+   ds, err := convertResourceSchemaToDataSourceSchema(
+     resources.ExampleResourceSchema(ctx)
+ 		AttributeConfig{
+ 			RequiredAttributes: &[]string{"id"},
+ 		},
+ 	)
+ 	if err != nil {
+ 		resp.Diagnostics.AddError("Failed to convert resource schema to data source schema", err.Error())
+ 		return
+ 	}
+ 	resp.Schema = *ds
+ 	// resp.Schema.Description = "TODO"
  }
```

Add the data source to `provider.go`:

```diff
  return []func() datasource.DataSource{
    # ...
+   NewExampleDataSource,
  }
