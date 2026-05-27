package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/badrory/jsm-terraform-provider/internal/client"
)

var _ datasource.DataSource = &ObjectTypeDataSource{}

// ObjectTypeDataSource reads an existing JSM Assets object type by name within
// a given schema.
type ObjectTypeDataSource struct {
	client *client.Client
}

// ObjectTypeDataModel is the Terraform state model for the data source.
type ObjectTypeDataModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	IconID             types.String `tfsdk:"icon_id"`
	ObjectSchemaID     types.String `tfsdk:"object_schema_id"`
	ParentObjectTypeID types.String `tfsdk:"parent_object_type_id"`
}

// NewObjectTypeDataSource returns a new datasource.DataSource factory.
func NewObjectTypeDataSource() datasource.DataSource {
	return &ObjectTypeDataSource{}
}

func (d *ObjectTypeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assets_object_type"
}

func (d *ObjectTypeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up an existing JSM Assets object type by name within a schema.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The numeric ID of the object type.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The exact name of the object type to look up.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The object type description.",
			},
			"icon_id": schema.StringAttribute{
				Computed:    true,
				Description: "The icon ID of the object type.",
			},
			"object_schema_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the schema that contains the object type.",
			},
			"parent_object_type_id": schema.StringAttribute{
				Computed:    true,
				Description: "The parent object type ID (if any).",
			},
		},
	}
}

func (d *ObjectTypeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ObjectTypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ObjectTypeDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	objectTypes, err := d.client.GetObjectTypesBySchema(ctx, config.ObjectSchemaID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing object types", err.Error())
		return
	}

	name := config.Name.ValueString()
	for _, ot := range objectTypes {
		if ot.Name == name {
			config.ID = types.StringValue(ot.ID)
			config.Description = types.StringValue(ot.Description)
			config.IconID = types.StringValue(ot.Icon.ID)
			config.ObjectSchemaID = types.StringValue(ot.ObjectSchemaID)
			config.ParentObjectTypeID = types.StringValue(ot.ParentObjectTypeID)
			resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"Object type not found",
		fmt.Sprintf("No object type named %q found in schema %s.", name, config.ObjectSchemaID.ValueString()),
	)
}
