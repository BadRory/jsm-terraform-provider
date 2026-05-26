// Package datasources implements Terraform data sources for JSM Assets.
package datasources

import (
	"context"
	"fmt"

	"github.com/badrory/jsm-terraform-provider/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ObjectSchemaDataSource{}

// ObjectSchemaDataSource reads an existing JSM Assets object schema by name.
type ObjectSchemaDataSource struct {
	client *client.Client
}

// ObjectSchemaDataModel is the Terraform state model for the data source.
type ObjectSchemaDataModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	ObjectSchemaKey types.String `tfsdk:"object_schema_key"`
	Description     types.String `tfsdk:"description"`
	Status          types.String `tfsdk:"status"`
	WorkspaceID     types.String `tfsdk:"workspace_id"`
}

// NewObjectSchemaDataSource returns a new datasource.DataSource factory.
func NewObjectSchemaDataSource() datasource.DataSource {
	return &ObjectSchemaDataSource{}
}

func (d *ObjectSchemaDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assets_object_schema"
}

func (d *ObjectSchemaDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up an existing JSM Assets object schema by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The numeric ID of the schema.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The exact name of the schema to look up.",
			},
			"object_schema_key": schema.StringAttribute{
				Computed:    true,
				Description: "The short uppercase key of the schema.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The schema description.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The current status of the schema.",
			},
			"workspace_id": schema.StringAttribute{
				Computed:    true,
				Description: "The JSM Assets workspace ID that owns this schema.",
			},
		},
	}
}

func (d *ObjectSchemaDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected data source configure type",
			fmt.Sprintf("Expected *client.Client, got %T.", req.ProviderData))
		return
	}
	d.client = c
}

func (d *ObjectSchemaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ObjectSchemaDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemas, err := d.client.ListObjectSchemas(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing object schemas", err.Error())
		return
	}

	name := config.Name.ValueString()
	for _, s := range schemas {
		if s.Name == name {
			config.ID = types.StringValue(s.ID)
			config.ObjectSchemaKey = types.StringValue(s.ObjectSchemaKey)
			config.Description = types.StringValue(s.Description)
			config.Status = types.StringValue(s.Status)
			config.WorkspaceID = types.StringValue(d.client.WorkspaceID)
			resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"Object schema not found",
		fmt.Sprintf("No object schema with name %q was found in the workspace.", name),
	)
}
