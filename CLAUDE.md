# JSM Terraform Provider

A Terraform provider for [Jira Service Management (JSM) Assets](https://developer.atlassian.com/cloud/assets/rest/intro/).

## Project layout

```
main.go                         — provider entry-point (providerserver.Serve)
internal/
  provider/provider.go          — provider schema, Configure, Resources, DataSources
  client/
    client.go                   — HTTP client: auth, doJSON, all API methods
    models.go                   — Go structs mirroring the JSM Assets REST API
  resources/
    object_schema.go            — jsm_assets_object_schema resource
    object_type.go              — jsm_assets_object_type resource
    object_type_attribute.go    — jsm_assets_object_type_attribute resource
    object.go                   — jsm_assets_object resource
  datasources/
    object_schema.go            — data.jsm_assets_object_schema
    object_type.go              — data.jsm_assets_object_type
    object.go                   — data.jsm_assets_object (AQL search)
examples/
  provider/provider.tf          — minimal provider config
  application-catalog/          — full Okta + JSM example
docs/index.md                   — provider documentation
```

## Build & install

```bash
make build      # go build ./...
make install    # go install ./...  (installs to $GOPATH/bin)
make test       # unit tests
make testacc    # acceptance tests (requires live JSM creds via env vars)
```

## Environment variables (auth)

```
JSM_SITE_URL=https://mycompany.atlassian.net
JSM_EMAIL=admin@mycompany.com
JSM_API_TOKEN=<atlassian-api-token>
JSM_WORKSPACE_ID=<optional, auto-discovered>
```

## API base URLs

- **Assets API**: `https://api.atlassian.com/jsm/assets/workspace/{workspaceId}/v1`
- **Workspace discovery**: `{site_url}/rest/servicedeskapi/assets/workspace`

## Key design decisions

- **Workspace auto-discovery**: The provider queries `{site_url}/rest/servicedeskapi/assets/workspace`
  on startup if `workspace_id` is not supplied; the first workspace is used.
- **Not-Found handling**: All resources call `resp.State.RemoveResource(ctx)` on 404
  to support drift detection without erroring.
- **Object attribute filtering**: `jsm_assets_object` only reconciles the attributes
  explicitly declared in the `attribute {}` blocks, avoiding diffs from JSM-managed
  system attributes (e.g. the internal `Created`, `Updated` fields).
- **System attribute protection**: `jsm_assets_object_type_attribute` emits a warning
  (not an error) when asked to delete a system attribute, and skips the API call.
- **terraform-plugin-framework v1**: Uses the modern framework (not SDKv2) for
  typed schemas, nested attributes, and `diag.Diagnostics`.

## Adding a new resource

1. Add API methods to `internal/client/client.go` and models to `models.go`.
2. Create `internal/resources/<name>.go` implementing `resource.Resource`.
3. Register the factory in `internal/provider/provider.go` under `Resources()`.
4. Add a data source if read-only lookup is needed (same pattern under `datasources/`).
5. Add an example under `examples/` and update `docs/index.md`.
