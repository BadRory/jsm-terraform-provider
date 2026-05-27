# JSM Terraform Provider

[![CI](https://github.com/badrory/jsm-terraform-provider/actions/workflows/ci.yml/badge.svg)](https://github.com/badrory/jsm-terraform-provider/actions/workflows/ci.yml)
[![Release](https://github.com/badrory/jsm-terraform-provider/actions/workflows/release.yml/badge.svg)](https://github.com/badrory/jsm-terraform-provider/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/badrory/jsm-terraform-provider)](https://goreportcard.com/report/github.com/badrory/jsm-terraform-provider)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A Terraform provider for managing [Jira Service Management (JSM) Assets](https://developer.atlassian.com/cloud/assets/rest/intro/) — the asset and configuration management database (CMDB) built into Jira Service Management Cloud.

Use this provider to define your asset schema, object types, attributes, and individual asset records as code, keeping your CMDB schema consistent across environments and under version control.

---

## Table of Contents

- [Overview](#overview)
- [Requirements](#requirements)
- [Installation](#installation)
- [Authentication](#authentication)
- [Provider Configuration](#provider-configuration)
- [Resources](#resources)
  - [jsm\_assets\_object\_schema](#jsm_assets_object_schema)
  - [jsm\_assets\_object\_type](#jsm_assets_object_type)
  - [jsm\_assets\_object\_type\_attribute](#jsm_assets_object_type_attribute)
  - [jsm\_assets\_object](#jsm_assets_object)
- [Data Sources](#data-sources)
  - [data.jsm\_assets\_object\_schema](#datajsm_assets_object_schema)
  - [data.jsm\_assets\_object\_type](#datajsm_assets_object_type)
  - [data.jsm\_assets\_object](#datajsm_assets_object)
- [AQL Reference](#aql-reference)
- [Examples](#examples)
  - [Minimal — single object](#minimal--single-object)
  - [Full — Application Catalog with Okta](#full--application-catalog-with-okta)
- [Data Model Hierarchy](#data-model-hierarchy)
- [Key Design Decisions](#key-design-decisions)
- [Development](#development)
- [Contributing](#contributing)

---

## Overview

JSM Assets is Atlassian's cloud CMDB. It models your infrastructure and services as a graph of typed objects. The provider maps the four levels of that model directly to Terraform resources:

| JSM Concept | Terraform Resource | Analogy |
|---|---|---|
| Schema | `jsm_assets_object_schema` | Database |
| Object Type | `jsm_assets_object_type` | Table / class |
| Object Type Attribute | `jsm_assets_object_type_attribute` | Column / field definition |
| Object | `jsm_assets_object` | Row / instance |

---

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) ≥ 1.0
- Go 1.25+ (only for building from source)
- A **Jira Service Management Cloud** instance with the Assets feature enabled
- An [Atlassian API token](https://id.atlassian.com/manage/api-tokens)

---

## Installation

Add the provider to your `terraform` block:

```hcl
terraform {
  required_providers {
    jsm = {
      source  = "badrory/jsm"
      version = "~> 0.1"
    }
  }
}
```

Then run:

```bash
terraform init
```

### Building from source

```bash
git clone https://github.com/badrory/jsm-terraform-provider.git
cd jsm-terraform-provider
make install   # installs the binary to $GOPATH/bin
```

To use a locally built binary, create or update `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "badrory/jsm" = "/path/to/your/GOPATH/bin"
  }
  direct {}
}
```

---

## Authentication

The provider authenticates to Atlassian APIs using HTTP Basic auth with your **email address** and an **API token** (not your account password).

Generate a token at: <https://id.atlassian.com/manage/api-tokens>

Credentials can be supplied via provider arguments or environment variables. Environment variables are recommended so secrets never appear in Terraform state or plan output.

| Environment Variable | Description |
|---|---|
| `JSM_SITE_URL` | Your Atlassian site URL, e.g. `https://mycompany.atlassian.net` |
| `JSM_EMAIL` | The email address for the API token owner |
| `JSM_API_TOKEN` | The Atlassian API token |
| `JSM_WORKSPACE_ID` | *(Optional)* JSM workspace ID — auto-discovered if omitted |

---

## Provider Configuration

```hcl
provider "jsm" {
  site_url     = "https://mycompany.atlassian.net"  # or JSM_SITE_URL
  email        = "admin@mycompany.com"              # or JSM_EMAIL
  api_token    = var.jsm_api_token                  # or JSM_API_TOKEN
  workspace_id = "abc123"                           # or JSM_WORKSPACE_ID (optional)
}
```

### Arguments

| Argument | Type | Required | Description |
|---|---|---|---|
| `site_url` | string | **yes** | Base URL of your Atlassian Cloud site |
| `email` | string | **yes** | Email address associated with the API token |
| `api_token` | string | **yes** | Atlassian API token (treat as a secret) |
| `workspace_id` | string | no | JSM workspace ID. If omitted the provider calls `{site_url}/rest/servicedeskapi/assets/workspace` and uses the first workspace returned |

> **Tip:** Store `api_token` in a secrets manager and pass it via a variable. Never hard-code it in `.tf` files.

---

## Resources

### jsm_assets_object_schema

Creates and manages a top-level **Object Schema** — the container for all object types and objects in a logical domain (e.g. "Application Catalog", "Infrastructure").

```hcl
resource "jsm_assets_object_schema" "app_catalog" {
  name               = "Application Catalog"
  object_schema_key  = "APPS"
  description        = "All production applications and their metadata"
}
```

#### Arguments

| Argument | Type | Required | Force Replace | Description |
|---|---|---|---|---|
| `name` | string | **yes** | no | Human-readable schema name |
| `object_schema_key` | string | **yes** | **yes** | Short uppercase identifier for keys (e.g. `APPS`). Changing this destroys and recreates the schema |
| `description` | string | no | no | Free-text description |

#### Attributes (read-only)

| Attribute | Type | Description |
|---|---|---|
| `id` | string | Numeric schema ID assigned by JSM |
| `status` | string | Lifecycle status (e.g. `Ok`) |
| `workspace_id` | string | The workspace this schema belongs to |

---

### jsm_assets_object_type

Defines a **class of assets** within a schema. Object types are analogous to database tables; each one represents a category of thing you want to track (e.g. "Application", "Server", "Team").

```hcl
resource "jsm_assets_object_type" "application" {
  name             = "Application"
  description      = "A production software application"
  object_schema_id = jsm_assets_object_schema.app_catalog.id
  icon_id          = "1"
}
```

#### Arguments

| Argument | Type | Required | Force Replace | Description |
|---|---|---|---|---|
| `name` | string | **yes** | no | Object type display name |
| `object_schema_id` | string | **yes** | **yes** | ID of the parent schema. Changing this destroys and recreates the type |
| `description` | string | no | no | Free-text description. Defaults to `""` |
| `icon_id` | string | no | no | Icon to display in the JSM UI. Defaults to `"1"` (generic blue box). Find IDs in the JSM Assets icon picker |
| `parent_object_type_id` | string | no | no | ID of a parent object type, enabling a type hierarchy. Defaults to `""` |
| `abstract_type` | bool | no | no | If `true`, marks this type as abstract (used as a base type only). Defaults to `false` |

#### Attributes (read-only)

| Attribute | Type | Description |
|---|---|---|
| `id` | string | Numeric object type ID |

---

### jsm_assets_object_type_attribute

Defines a **field** on an object type. Attributes are the building blocks of your schema — they determine what data each asset can hold.

```hcl
resource "jsm_assets_object_type_attribute" "okta_app_id" {
  object_type_id  = jsm_assets_object_type.application.id
  name            = "OktaAppID"
  description     = "The Okta Application ID"
  type            = 0   # Default / scalar
  default_type_id = 0   # Text
  indexed         = true
  unique          = true
}
```

#### Arguments

| Argument | Type | Required | Force Replace | Description |
|---|---|---|---|---|
| `object_type_id` | string | **yes** | **yes** | ID of the parent object type. Changing this destroys and recreates the attribute |
| `name` | string | **yes** | no | Attribute name |
| `description` | string | no | no | Free-text description. Defaults to `""` |
| `type` | number | no | no | Attribute kind. See [Attribute Types](#attribute-types) below. Defaults to `0` |
| `default_type_id` | number | no | no | Scalar data type when `type = 0`. See [Scalar Types](#scalar-types) below. Defaults to `0` (Text) |
| `type_value` | string | no | no | For `type = 1`: the target object type ID. For `type = 6`: the status type ID. Defaults to `""` |
| `additional_value` | string | no | no | Extra configuration for the attribute (e.g. select options). Defaults to `""` |
| `minimum_cardinality` | number | no | no | Minimum number of values required (0 = optional). Defaults to `0` |
| `maximum_cardinality` | number | no | no | Maximum number of values allowed (1 = single-value, -1 = unlimited). Defaults to `1` |
| `indexed` | bool | no | no | Enable AQL indexing for faster queries. Defaults to `false` |
| `unique` | bool | no | no | Enforce uniqueness across all objects of this type. Defaults to `false` |
| `summable` | bool | no | no | Allow numeric values to be summed in the UI. Defaults to `false` |
| `regex_validation` | string | no | no | Regex pattern for input validation. Defaults to `""` |

#### Attributes (read-only)

| Attribute | Type | Description |
|---|---|---|
| `id` | string | Numeric attribute ID |
| `system` | bool | `true` if this is a system-managed attribute (e.g. the built-in `Name` field). System attributes cannot be deleted via Terraform — a warning is emitted and the API call is skipped |

#### Attribute Types

| `type` | Kind | Notes |
|---|---|---|
| `0` | **Default** (scalar) | Data type is governed by `default_type_id` |
| `1` | **Object reference** | Set `type_value` to the target object type ID |
| `2` | **User** | Jira user picker; supply account IDs as values |
| `6` | **Status** | Set `type_value` to the status type ID |

#### Scalar Types (`default_type_id` when `type = 0`)

| `default_type_id` | Data Type |
|---|---|
| `0` | Text |
| `1` | Integer |
| `2` | Boolean |
| `3` | Double |
| `4` | Date |
| `5` | Time |
| `6` | DateTime |
| `7` | URL |
| `8` | Email |
| `9` | Textarea |
| `10` | Select (dropdown) |
| `11` | IP Address |

---

### jsm_assets_object

Creates an individual **asset record** (an instance of an object type). Objects hold the actual data — a specific server, application, team, etc.

```hcl
resource "jsm_assets_object" "github" {
  object_type_id = jsm_assets_object_type.application.id

  attribute {
    attribute_id = jsm_assets_object_type_attribute.name.id
    value        = "GitHub"
  }

  attribute {
    attribute_id = jsm_assets_object_type_attribute.okta_app_id.id
    value        = okta_app_oauth.github.id
  }
}
```

#### Arguments

| Argument | Type | Required | Force Replace | Description |
|---|---|---|---|---|
| `object_type_id` | string | **yes** | **yes** | ID of the object type this record belongs to. Changing this destroys and recreates the object |
| `attribute` | block | no | no | One block per attribute value to manage. Attributes not listed are ignored by Terraform |

Each `attribute` block accepts:

| Field | Type | Required | Description |
|---|---|---|---|
| `attribute_id` | string | **yes** | ID of the object type attribute definition |
| `value` | string | **yes** | String value. For User attributes supply the Jira account ID. For Object references supply the target object's key (e.g. `"APPS-5"`) |

> **Important:** Terraform only manages attributes explicitly declared in `attribute {}` blocks. System-managed fields (e.g. `Created`, `Updated`) and any undeclared attributes are left untouched, preventing noisy perpetual diffs.

#### Attributes (read-only)

| Attribute | Type | Description |
|---|---|---|
| `id` | string | Object key assigned by JSM (e.g. `APPS-1`) |
| `label` | string | Display label derived by JSM from the object's `Name` attribute |

---

## Data Sources

Data sources let you look up existing JSM assets that were created outside Terraform, or reference objects managed in a separate Terraform workspace.

### data.jsm_assets_object_schema

Look up an existing schema by name.

```hcl
data "jsm_assets_object_schema" "existing" {
  name = "Application Catalog"
}

output "schema_id" {
  value = data.jsm_assets_object_schema.existing.id
}
```

#### Arguments

| Argument | Type | Required | Description |
|---|---|---|---|
| `name` | string | **yes** | Exact schema name to search for |

#### Attributes (read-only)

| Attribute | Type | Description |
|---|---|---|
| `id` | string | Numeric schema ID |
| `object_schema_key` | string | Short uppercase key (e.g. `APPS`) |
| `description` | string | Schema description |
| `status` | string | Lifecycle status |
| `workspace_id` | string | Workspace ID |

---

### data.jsm_assets_object_type

Look up an existing object type by name within a schema.

```hcl
data "jsm_assets_object_type" "application" {
  name             = "Application"
  object_schema_id = data.jsm_assets_object_schema.existing.id
}
```

#### Arguments

| Argument | Type | Required | Description |
|---|---|---|---|
| `name` | string | **yes** | Exact object type name |
| `object_schema_id` | string | **yes** | ID of the schema to search within |

#### Attributes (read-only)

| Attribute | Type | Description |
|---|---|---|
| `id` | string | Numeric object type ID |
| `description` | string | Object type description |
| `icon_id` | string | Icon ID |
| `parent_object_type_id` | string | ID of the parent object type, if any |

---

### data.jsm_assets_object

Search for an asset using an **AQL** (Assets Query Language) query. Returns the first matching object.

```hcl
data "jsm_assets_object" "github_app" {
  aql = "objectType = \"Application\" AND Name = \"GitHub\""
}

output "github_object_key" {
  value = data.jsm_assets_object.github_app.id   # e.g. "APPS-3"
}
```

#### Arguments

| Argument | Type | Required | Description |
|---|---|---|---|
| `aql` | string | **yes** | AQL query string. The first result is returned |

#### Attributes (read-only)

| Attribute | Type | Description |
|---|---|---|
| `id` | string | Object key (e.g. `APPS-3`) |
| `label` | string | Display label |
| `object_type_id` | string | Object type ID |
| `attributes` | map(string) | Map of attribute definition IDs → first string value |

---

## AQL Reference

AQL (Assets Query Language) is used in the `data.jsm_assets_object` data source and in the JSM Assets UI to query objects.

```
# Match by object type and attribute value
objectType = "Application" AND Name = "GitHub"

# Match within a specific schema
objectSchemaId = 5 AND objectType = "Server"

# Match on a custom attribute
"OktaGroupID" = "00g1abc123"

# Combined filter
objectType = "Application" AND "Owner" = "62ab1c3d4e5f6a7b8c9d0e1f"
```

Full AQL documentation: <https://developer.atlassian.com/cloud/assets/aql/>

---

## Examples

### Minimal — single object

Create a schema, one object type with a single custom attribute, and one asset record:

```hcl
terraform {
  required_providers {
    jsm = {
      source  = "badrory/jsm"
      version = "~> 0.1"
    }
  }
}

provider "jsm" {
  # Credentials are read from environment variables:
  # JSM_SITE_URL, JSM_EMAIL, JSM_API_TOKEN
}

# ── Schema ────────────────────────────────────────────────────────────────────
resource "jsm_assets_object_schema" "services" {
  name              = "Services"
  object_schema_key = "SVC"
  description       = "Internal services and APIs"
}

# ── Object Type ───────────────────────────────────────────────────────────────
resource "jsm_assets_object_type" "service" {
  name             = "Service"
  object_schema_id = jsm_assets_object_schema.services.id
}

# ── Attributes ────────────────────────────────────────────────────────────────
resource "jsm_assets_object_type_attribute" "repo_url" {
  object_type_id  = jsm_assets_object_type.service.id
  name            = "RepositoryURL"
  description     = "Link to the source code repository"
  type            = 0
  default_type_id = 7   # URL
}

resource "jsm_assets_object_type_attribute" "owner" {
  object_type_id      = jsm_assets_object_type.service.id
  name                = "Owner"
  type                = 2   # User
  minimum_cardinality = 1
}

# ── Object (asset record) ─────────────────────────────────────────────────────
resource "jsm_assets_object" "payments_api" {
  object_type_id = jsm_assets_object_type.service.id

  attribute {
    attribute_id = jsm_assets_object_type_attribute.repo_url.id
    value        = "https://github.com/mycompany/payments-api"
  }

  attribute {
    attribute_id = jsm_assets_object_type_attribute.owner.id
    value        = "62ab1c3d4e5f6a7b8c9d0e1f"   # Jira account ID
  }
}
```

---

### Full — Application Catalog with Okta

The [`examples/application-catalog/`](examples/application-catalog/) directory contains a production-ready example that:

1. Creates an Okta OAuth application and access group
2. Builds a JSM "Application Catalog" schema with the key `APPS`
3. Defines an `Application` object type with seven custom attributes:

| Attribute | Type | Notes |
|---|---|---|
| `OktaAppID` | Text (unique, indexed) | Linked from Okta |
| `OktaGroupID` | Text (unique, indexed) | Linked from Okta |
| `Approvers` | User (1–5 values) | Multi-valued user picker |
| `Roles` | Select dropdown | Multi-valued |
| `Owner` | User | Single-value |
| `Team` | Text (indexed) | Team name |
| `Description` | Textarea | Long-form description |

4. Creates a `GitHub` asset record that joins the Okta app data with JSM

#### Running the example

```bash
cd examples/application-catalog

# Provide required variables
export TF_VAR_jsm_api_token="<your-token>"
export TF_VAR_okta_api_token="<your-okta-token>"
export JSM_SITE_URL="https://mycompany.atlassian.net"
export JSM_EMAIL="admin@mycompany.com"

terraform init
terraform plan
terraform apply
```

---

## Data Model Hierarchy

```
Workspace (auto-discovered)
└── Object Schema  (jsm_assets_object_schema)
    └── Object Type  (jsm_assets_object_type)
        ├── Object Type Attribute  (jsm_assets_object_type_attribute)
        │   ├── Scalar attributes  (text, integer, boolean, URL, email, …)
        │   ├── Object references  (links to other object types)
        │   ├── User attributes    (Jira user pickers)
        │   └── Status attributes
        └── Object  (jsm_assets_object)
            └── Attribute values  (attribute {} blocks)
```

---

## Key Design Decisions

### Workspace auto-discovery

If `workspace_id` is not supplied the provider queries `{site_url}/rest/servicedeskapi/assets/workspace` at startup and uses the first workspace returned. Explicitly setting `workspace_id` is recommended for production to avoid unexpected workspace changes.

### Immutable fields cause replacement

The following arguments force resource destruction and recreation when changed:

| Resource | Immutable argument |
|---|---|
| `jsm_assets_object_schema` | `object_schema_key` |
| `jsm_assets_object_type` | `object_schema_id` |
| `jsm_assets_object_type_attribute` | `object_type_id` |
| `jsm_assets_object` | `object_type_id` |

### Not-found handling and drift detection

All resources call `resp.State.RemoveResource(ctx)` on a 404 response instead of returning an error. This means that if an asset is deleted outside Terraform, `terraform plan` will show it needs to be recreated rather than failing with an error.

### Attribute filtering on objects

`jsm_assets_object` only reconciles attributes explicitly listed in `attribute {}` blocks. System-managed attributes such as `Created` and `Updated` are populated by JSM and are intentionally excluded from Terraform's view to prevent perpetual diffs on every plan.

### System attribute protection

`jsm_assets_object_type_attribute` emits a **warning** (not an error) when asked to delete a system attribute and skips the DELETE API call. System attributes are identified by `system = true` on the attribute resource.

---

## Development

### Prerequisites

- Go 1.25+
- [Terraform](https://developer.hashicorp.com/terraform/downloads) (for acceptance tests)
- [golangci-lint](https://golangci-lint.run/usage/install/) (for linting)
- [GoReleaser](https://goreleaser.com/install/) (for releases)

### Commands

```bash
# Build the provider binary
make build

# Install to $GOPATH/bin (for local dev_overrides use)
make install

# Run unit tests
make test

# Run acceptance tests (requires live JSM credentials via env vars)
make testacc

# Lint
golangci-lint run
```

### Environment variables for acceptance tests

```bash
export JSM_SITE_URL="https://mycompany.atlassian.net"
export JSM_EMAIL="admin@mycompany.com"
export JSM_API_TOKEN="<token>"
export JSM_WORKSPACE_ID="<workspace-id>"   # optional
```

### Project layout

```
main.go                               Provider entry-point
internal/
  provider/provider.go                Provider schema, Configure, Resources, DataSources
  client/
    client.go                         HTTP client: auth, doJSON, all API methods
    models.go                         Go structs mirroring the JSM Assets REST API
  resources/
    object_schema.go                  jsm_assets_object_schema resource
    object_type.go                    jsm_assets_object_type resource
    object_type_attribute.go          jsm_assets_object_type_attribute resource
    object.go                         jsm_assets_object resource
  datasources/
    object_schema.go                  data.jsm_assets_object_schema
    object_type.go                    data.jsm_assets_object_type
    object.go                         data.jsm_assets_object (AQL search)
examples/
  provider/provider.tf                Minimal provider configuration
  application-catalog/                Full Okta + JSM working example
docs/index.md                         Provider documentation
.github/
  workflows/ci.yml                    Lint, build, and test on every PR
  workflows/release.yml               GoReleaser publish on v* tags
  dependabot.yml                      Automated dependency updates
.goreleaser.yml                       Multi-platform release configuration
.golangci.yml                         Linter configuration
```

### Adding a new resource

1. Add API methods to `internal/client/client.go` and the corresponding structs to `models.go`
2. Create `internal/resources/<name>.go` implementing `resource.Resource`
3. Register the factory in `internal/provider/provider.go` under `Resources()`
4. Add a data source under `internal/datasources/` if read-only lookup is needed
5. Add an example under `examples/` and update `docs/index.md`

### Releasing

Create and push a semver tag — the release workflow handles everything else:

```bash
git tag v0.2.0
git push origin v0.2.0
```

GoReleaser builds binaries for all supported platforms, packages them as `.zip` archives, generates `SHA256SUMS`, and publishes everything as a GitHub Release. If `GPG_PRIVATE_KEY` and `PASSPHRASE` repository secrets are configured, the checksum file is signed automatically (required for publishing to the [Terraform Registry](https://registry.terraform.io/)).

---

## Contributing

Contributions are welcome! Please:

1. Fork the repository and create a feature branch
2. Make your changes with tests where appropriate
3. Run `golangci-lint run` and `go test ./...` before opening a PR
4. Open a pull request against `main` with a clear description of the change

---

## API Reference

- [JSM Assets REST API](https://developer.atlassian.com/cloud/assets/rest/intro/)
- [AQL documentation](https://developer.atlassian.com/cloud/assets/aql/)
- [Atlassian API tokens](https://id.atlassian.com/manage/api-tokens)
- [terraform-plugin-framework](https://developer.hashicorp.com/terraform/plugin/framework)
