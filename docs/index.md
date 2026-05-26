# JSM Provider

The **JSM** provider lets you manage [Jira Service Management (JSM) Assets](https://www.atlassian.com/software/jira/service-management/features/asset-management) — the built-in CMDB — entirely in Terraform.

Use it alongside the [Okta provider](https://registry.terraform.io/providers/okta/okta/latest) to codify your full application access-request lifecycle:

1. **Okta** — create the OIDC/SAML app, the user group, and the group assignment.
2. **JSM Assets** — create the Application Catalog schema, define the `Application` object type with custom attributes (OktaAppID, OktaGroupID, Approvers, Roles …), and insert a record for each app.
3. **JSM Automation** — reads the asset record to find who the approvers are, routes the JSM service-request to them, and — on approval — adds the user to the Okta group.

---

## Authentication

Generate an Atlassian API token at <https://id.atlassian.com/manage/api-tokens>.

```hcl
provider "jsm" {
  site_url  = "https://mycompany.atlassian.net"
  email     = "admin@mycompany.com"
  api_token = var.jsm_api_token   # sensitive
}
```

All three values can also be supplied via environment variables:

| Variable | Description |
|---|---|
| `JSM_SITE_URL` | Your Atlassian site URL |
| `JSM_EMAIL` | Atlassian account email |
| `JSM_API_TOKEN` | Atlassian API token |
| `JSM_WORKSPACE_ID` | *(optional)* Assets workspace ID — auto-discovered if omitted |

---

## Resources

| Resource | Description |
|---|---|
| [`jsm_assets_object_schema`](resources/jsm_assets_object_schema.md) | Top-level schema container (the "database") |
| [`jsm_assets_object_type`](resources/jsm_assets_object_type.md) | A class of assets (the "table") |
| [`jsm_assets_object_type_attribute`](resources/jsm_assets_object_type_attribute.md) | A field definition on an object type (the "column") |
| [`jsm_assets_object`](resources/jsm_assets_object.md) | An individual asset record (the "row") |

## Data Sources

| Data Source | Description |
|---|---|
| [`data.jsm_assets_object_schema`](data-sources/jsm_assets_object_schema.md) | Look up an existing schema by name |
| [`data.jsm_assets_object_type`](data-sources/jsm_assets_object_type.md) | Look up an existing object type by name + schema |
| [`data.jsm_assets_object`](data-sources/jsm_assets_object.md) | Find an object via an AQL query |

---

## Attribute type reference

### `type` values for `jsm_assets_object_type_attribute`

| Value | Kind |
|-------|------|
| `0` | Default scalar — governed by `default_type_id` |
| `1` | Object reference (set `type_value` to the target object type ID) |
| `2` | User (Jira user picker) |
| `6` | Status |

### `default_type_id` values (when `type = 0`)

| Value | Data type |
|-------|-----------|
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

## Full example

See [`examples/application-catalog/`](../examples/application-catalog/) for a complete
Okta + JSM Assets integration.

```hcl
# 1 — Schema
resource "jsm_assets_object_schema" "apps" {
  name              = "Application Catalog"
  object_schema_key = "APPS"
}

# 2 — Object type
resource "jsm_assets_object_type" "application" {
  name             = "Application"
  object_schema_id = jsm_assets_object_schema.apps.id
}

# 3 — Custom attributes
resource "jsm_assets_object_type_attribute" "okta_group" {
  object_type_id  = jsm_assets_object_type.application.id
  name            = "OktaGroupID"
  type            = 0
  default_type_id = 0  # Text
  unique          = true
  indexed         = true
}

resource "jsm_assets_object_type_attribute" "approvers" {
  object_type_id      = jsm_assets_object_type.application.id
  name                = "Approvers"
  type                = 2  # User
  minimum_cardinality = 1
  maximum_cardinality = 5
}

# 4 — Asset record (cross-references Okta output)
resource "jsm_assets_object" "github" {
  object_type_id = jsm_assets_object_type.application.id

  attribute {
    attribute_id = jsm_assets_object_type_attribute.okta_group.id
    value        = okta_group.github_users.id
  }

  attribute {
    attribute_id = jsm_assets_object_type_attribute.approvers.id
    value        = "jira-account-id-of-approver"
  }
}
```
