package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/badrory/jsm-terraform-provider/internal/client"
)

var _ datasource.DataSource = &ObjectDataSource{}

// ObjectDataSource searches for a JSM Assets object via an AQL query and
// exposes the first result.
type ObjectDataSource struct {
	client *client.Client
}

// ObjectDataModel is the Terraform state model for the data source.
type ObjectDataModel struct {
	ID           types.String `tfsdk:"id"`
	AQL          types.String `tfsdk:"aql"`
	Label        types.String `tfsdk:"label"`
	ObjectTypeID types.String `tfsdk:"object_type_id"`
	Attributes   types.Map    `tfsdk:"attributes"`
}

// NewObjectDataSource returns a new datasource.DataSource factory.
func NewObjectDataSource() datasource.DataSource {
	return &ObjectDataSource{}
}

func (d *ObjectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assets_object"
}

func (d *ObjectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
Searches for a JSM Assets object using an **AQL (Assets Query Language)** expression
and returns the first result.

### AQL examples

` + "```" + `
# Find by object type and name
objectType = "Application" AND Name = "GitHub"

# Find within a schema
objectSchemaId = 5 AND objectType = "Application"

# Reference field match
"OktaGroupID" = "00g1abc123"
` + "```" + `

> **Note:** This data source returns only the **first** match. Ensure your AQL
> query is specific enough to identify a unique object.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The key of the matched object (e.g. `APP-1`).",
			},
			"aql": schema.StringAttribute{
				Required:    true,
				Description: "An AQL query string. The first matching object is returned.",
			},
			"label": schema.StringAttribute{
				Computed:    true,
				Description: "The display label of the matched object.",
			},
			"object_type_id": schema.StringAttribute{
				Computed:    true,
				Description: "The object type ID of the matched object.",
			},
			"attributes": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Map of attribute definition IDs to their first string value.",
			},
		},
	}
}

func (d *ObjectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ObjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ObjectDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	objects, err := d.client.SearchObjects(ctx, config.AQL.ValueString(), true)
	if err != nil {
		resp.Diagnostics.AddError("Error searching objects with AQL", err.Error())
		return
	}

	if len(objects) == 0 {
		resp.Diagnostics.AddError(
			"No objects found",
			fmt.Sprintf("AQL query %q returned no results.", config.AQL.ValueString()),
		)
		return
	}

	obj := objects[0]
	config.ID = types.StringValue(obj.ID)
	config.Label = types.StringValue(obj.Label)
	config.ObjectTypeID = types.StringValue(obj.ObjectType.ID)

	// Build a map of attribute definition ID → first value string.
	attrMap := map[string]attr.Value{}
	for _, a := range obj.Attributes {
		attrID := a.ObjectTypeAttribute.ID
		if attrID == "" {
			attrID = a.ObjectTypeAttributeID
		}
		val := ""
		if len(a.ObjectAttributeValues) > 0 {
			val = a.ObjectAttributeValues[0].Value
		}
		attrMap[attrID] = types.StringValue(val)
	}

	m, ds := types.MapValue(types.StringType, attrMap)
	resp.Diagnostics.Append(ds...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.Attributes = m

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
