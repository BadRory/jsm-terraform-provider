// Package provider implements the Terraform provider for JSM Assets.
package provider

import (
	"context"
	"os"

	"github.com/badrory/jsm-terraform-provider/internal/client"
	"github.com/badrory/jsm-terraform-provider/internal/datasources"
	"github.com/badrory/jsm-terraform-provider/internal/resources"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the provider satisfies the framework interface.
var _ provider.Provider = &JSMProvider{}

// JSMProvider is the main Terraform provider implementation.
type JSMProvider struct {
	version string
}

// JSMProviderModel is the schema model for provider configuration.
type JSMProviderModel struct {
	SiteURL     types.String `tfsdk:"site_url"`
	Email       types.String `tfsdk:"email"`
	APIToken    types.String `tfsdk:"api_token"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
}

// New returns a factory function that creates a new JSMProvider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &JSMProvider{version: version}
	}
}

func (p *JSMProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "jsm"
	resp.Version = p.version
}

func (p *JSMProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
The **JSM** provider manages Jira Service Management (JSM) Assets — the CMDB layer
built into JSM Cloud. Use it to define object schemas, object types, attributes,
and individual asset records entirely in Terraform.

Pair it with the Okta provider to fully codify the lifecycle of an application:
create the Okta app and group, then record the application in JSM Assets complete
with its approvers, roles, and metadata so that users can request access via
JSM service catalog forms.

## Authentication

Generate an Atlassian API token at https://id.atlassian.com/manage/api-tokens
and pass it via the ` + "`api_token`" + ` argument or the ` + "`JSM_API_TOKEN`" + ` environment variable.
`,
		Attributes: map[string]schema.Attribute{
			"site_url": schema.StringAttribute{
				Optional:    true,
				Description: "The base URL of your Atlassian site, e.g. `https://mycompany.atlassian.net`. Can also be set via the `JSM_SITE_URL` environment variable.",
			},
			"email": schema.StringAttribute{
				Optional:    true,
				Description: "The email address of the Atlassian account used for authentication. Can also be set via the `JSM_EMAIL` environment variable.",
			},
			"api_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "An Atlassian API token. Generate one at https://id.atlassian.com/manage/api-tokens. Can also be set via the `JSM_API_TOKEN` environment variable.",
			},
			"workspace_id": schema.StringAttribute{
				Optional:    true,
				Description: "The JSM Assets workspace ID. If omitted the provider auto-discovers it from the site URL. Can also be set via the `JSM_WORKSPACE_ID` environment variable.",
			},
		},
	}
}

func (p *JSMProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config JSMProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve values: explicit config > environment variable.
	siteURL := resolveString(config.SiteURL, "JSM_SITE_URL")
	email := resolveString(config.Email, "JSM_EMAIL")
	apiToken := resolveString(config.APIToken, "JSM_API_TOKEN")
	workspaceID := resolveString(config.WorkspaceID, "JSM_WORKSPACE_ID")

	if siteURL == "" {
		resp.Diagnostics.AddError(
			"Missing site_url",
			"Set site_url in the provider block or via the JSM_SITE_URL environment variable.",
		)
	}
	if email == "" {
		resp.Diagnostics.AddError(
			"Missing email",
			"Set email in the provider block or via the JSM_EMAIL environment variable.",
		)
	}
	if apiToken == "" {
		resp.Diagnostics.AddError(
			"Missing api_token",
			"Set api_token in the provider block or via the JSM_API_TOKEN environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	c, err := client.NewClient(siteURL, email, apiToken, workspaceID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create JSM client", err.Error())
		return
	}

	// Share the configured client with resources and data sources.
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *JSMProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewObjectSchemaResource,
		resources.NewObjectTypeResource,
		resources.NewObjectTypeAttributeResource,
		resources.NewObjectResource,
	}
}

func (p *JSMProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewObjectSchemaDataSource,
		datasources.NewObjectTypeDataSource,
		datasources.NewObjectDataSource,
	}
}

// resolveString returns the non-null string from a types.String or falls back
// to reading the given environment variable.
func resolveString(v types.String, envVar string) string {
	if !v.IsNull() && !v.IsUnknown() {
		return v.ValueString()
	}
	return os.Getenv(envVar)
}
