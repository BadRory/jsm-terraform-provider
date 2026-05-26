###############################################################################
# Example: Okta + JSM Assets — full application access-request lifecycle
#
# This example shows how to:
#   1. Create an Okta OIDC application and its associated group.
#   2. Build an "Application Catalog" in JSM Assets that records the app with
#      its Okta IDs, assigned approvers, and available roles.
#   3. Combine both so that Jira Service Management automation can route access
#      requests to the right people and then add approved users to Okta groups.
###############################################################################

terraform {
  required_providers {
    jsm = {
      source  = "registry.terraform.io/badrory/jsm"
      version = "~> 0.1"
    }
    okta = {
      source  = "okta/okta"
      version = "~> 4.0"
    }
  }
}

# ---------------------------------------------------------------------------
# Provider configuration
# ---------------------------------------------------------------------------

provider "jsm" {
  site_url  = var.jsm_site_url
  email     = var.jsm_email
  api_token = var.jsm_api_token
}

provider "okta" {
  org_name  = var.okta_org_name
  base_url  = var.okta_base_url
  api_token = var.okta_api_token
}

# ---------------------------------------------------------------------------
# Okta — create the application and its access group
# ---------------------------------------------------------------------------

resource "okta_app_oauth" "github" {
  label          = "GitHub Enterprise"
  type           = "web"
  grant_types    = ["authorization_code", "refresh_token"]
  redirect_uris  = ["https://github.com/login/oauth2/callback"]
  response_types = ["code"]

  lifecycle {
    ignore_changes = [users]
  }
}

resource "okta_group" "github_users" {
  name        = "github-enterprise-users"
  description = "Users with access to GitHub Enterprise via Okta SSO"
}

resource "okta_app_group_assignment" "github_users" {
  app_id   = okta_app_oauth.github.id
  group_id = okta_group.github_users.id
}

# ---------------------------------------------------------------------------
# JSM Assets — define the Application Catalog schema
# ---------------------------------------------------------------------------

resource "jsm_assets_object_schema" "app_catalog" {
  name              = "Application Catalog"
  object_schema_key = "APPS"
  description       = "Catalog of all managed SaaS/enterprise applications used for access request workflows."
}

# ---------------------------------------------------------------------------
# JSM Assets — Application object type
# ---------------------------------------------------------------------------

resource "jsm_assets_object_type" "application" {
  name             = "Application"
  object_schema_id = jsm_assets_object_schema.app_catalog.id
  description      = "A managed application that users can request access to via JSM."
  icon_id          = "1" # Generic blue box — update to match your JSM icon set
}

# ---------------------------------------------------------------------------
# JSM Assets — Attributes on the Application object type
#
# The built-in "Name" attribute (system) is always present; we add custom ones.
# ---------------------------------------------------------------------------

# The OktaAppID links back to the SAML/OIDC app in Okta.
resource "jsm_assets_object_type_attribute" "app_okta_app_id" {
  object_type_id = jsm_assets_object_type.application.id
  name           = "OktaAppID"
  description    = "The Okta application client ID. Used by automation to provision access."
  type           = 0 # Default/scalar
  default_type_id = 0 # Text
  unique         = true
  indexed        = true
}

# The OktaGroupID links to the Okta group that grants app access.
resource "jsm_assets_object_type_attribute" "app_okta_group_id" {
  object_type_id = jsm_assets_object_type.application.id
  name           = "OktaGroupID"
  description    = "The Okta group ID whose membership grants access to this application."
  type           = 0
  default_type_id = 0 # Text
  unique         = true
  indexed        = true
}

# Approvers — multi-valued User attribute (up to 5 approvers per app).
resource "jsm_assets_object_type_attribute" "app_approvers" {
  object_type_id      = jsm_assets_object_type.application.id
  name                = "Approvers"
  description         = "Jira users who approve access requests for this application."
  type                = 2 # User
  minimum_cardinality = 1
  maximum_cardinality = 5
}

# Available roles — multi-valued text (e.g. "read", "write", "admin").
resource "jsm_assets_object_type_attribute" "app_roles" {
  object_type_id      = jsm_assets_object_type.application.id
  name                = "Roles"
  description         = "The access roles/levels available for this application."
  type                = 0
  default_type_id     = 10 # Select (dropdown)
  maximum_cardinality = 10
}

# Owner — single User attribute.
resource "jsm_assets_object_type_attribute" "app_owner" {
  object_type_id = jsm_assets_object_type.application.id
  name           = "Owner"
  description    = "The business owner responsible for this application."
  type           = 2 # User
}

# Team — free-text team name.
resource "jsm_assets_object_type_attribute" "app_team" {
  object_type_id  = jsm_assets_object_type.application.id
  name            = "Team"
  description     = "The team that owns this application."
  type            = 0
  default_type_id = 0 # Text
  indexed         = true
}

# Description — textarea.
resource "jsm_assets_object_type_attribute" "app_description" {
  object_type_id  = jsm_assets_object_type.application.id
  name            = "Description"
  description     = "A short description of what this application is used for."
  type            = 0
  default_type_id = 9 # Textarea
}

# ---------------------------------------------------------------------------
# JSM Assets — the GitHub application record
#
# We look up the Name attribute's system ID by finding the object type
# attributes after creation. For simplicity, the Name attribute for JSM Assets
# objects is typically attribute ID "1" within the schema — but since IDs vary
# per workspace we use data sources in real usage (see the locals block below).
#
# The attribute_id values below reference the Terraform-managed attributes we
# created above. Terraform will substitute the correct IDs at apply time.
# ---------------------------------------------------------------------------

resource "jsm_assets_object" "github" {
  object_type_id = jsm_assets_object_type.application.id

  attribute {
    attribute_id = jsm_assets_object_type_attribute.app_okta_app_id.id
    value        = okta_app_oauth.github.client_id
  }

  attribute {
    attribute_id = jsm_assets_object_type_attribute.app_okta_group_id.id
    value        = okta_group.github_users.id
  }

  attribute {
    attribute_id = jsm_assets_object_type_attribute.app_approvers.id
    value        = var.jsm_approver_account_id
  }

  attribute {
    attribute_id = jsm_assets_object_type_attribute.app_team.id
    value        = "Platform Engineering"
  }

  attribute {
    attribute_id = jsm_assets_object_type_attribute.app_description.id
    value        = "GitHub Enterprise — source code management for the engineering org."
  }

  depends_on = [
    jsm_assets_object_type_attribute.app_okta_app_id,
    jsm_assets_object_type_attribute.app_okta_group_id,
    jsm_assets_object_type_attribute.app_approvers,
    jsm_assets_object_type_attribute.app_team,
    jsm_assets_object_type_attribute.app_description,
  ]
}

# ---------------------------------------------------------------------------
# Outputs — useful for referencing from other modules or CI pipelines
# ---------------------------------------------------------------------------

output "okta_app_id" {
  description = "The Okta application client ID"
  value       = okta_app_oauth.github.client_id
}

output "okta_group_id" {
  description = "The Okta group ID for GitHub users"
  value       = okta_group.github_users.id
}

output "jsm_schema_id" {
  description = "The JSM Assets object schema ID"
  value       = jsm_assets_object_schema.app_catalog.id
}

output "jsm_application_object_type_id" {
  description = "The JSM Assets Application object type ID"
  value       = jsm_assets_object_type.application.id
}

output "jsm_github_object_id" {
  description = "The JSM Assets object key for the GitHub application record"
  value       = jsm_assets_object.github.id
}
