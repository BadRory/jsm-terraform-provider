terraform {
  required_providers {
    jsm = {
      source  = "registry.terraform.io/badrory/jsm"
      version = "~> 0.1"
    }
  }
}

# Credentials via environment variables (recommended):
#   JSM_SITE_URL   = "https://mycompany.atlassian.net"
#   JSM_EMAIL      = "admin@mycompany.com"
#   JSM_API_TOKEN  = "..."
#
# Or inline (not recommended for secrets):
provider "jsm" {
  site_url  = "https://mycompany.atlassian.net"
  email     = "admin@mycompany.com"
  api_token = var.jsm_api_token

  # workspace_id is optional — auto-discovered from site_url.
  # workspace_id = "abc123"
}
