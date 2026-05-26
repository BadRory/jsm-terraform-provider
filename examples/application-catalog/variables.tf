variable "jsm_api_token" {
  description = "Atlassian API token. Generate at https://id.atlassian.com/manage/api-tokens"
  type        = string
  sensitive   = true
}

variable "okta_api_token" {
  description = "Okta API token"
  type        = string
  sensitive   = true
}

variable "okta_org_name" {
  description = "Okta organisation name (e.g. mycompany)"
  type        = string
}

variable "okta_base_url" {
  description = "Okta base URL (e.g. okta.com)"
  type        = string
  default     = "okta.com"
}

variable "jsm_site_url" {
  description = "Atlassian site URL (e.g. https://mycompany.atlassian.net)"
  type        = string
}

variable "jsm_email" {
  description = "Atlassian account email for authentication"
  type        = string
}

variable "jsm_approver_account_id" {
  description = "Jira account ID of the default application approver"
  type        = string
}
