package resources

import (
	"context"
	"fmt"

	"github.com/badrory/jsm-terraform-provider/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ObjectSchemaResource{}

// ObjectSchemaResource manages a JSM Assets object schema.
type ObjectSchemaResource struct {
	client *client.Client
}

// ObjectSchemaModel is the Terraform state model.
type ObjectSchemaModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	ObjectSchemaKey types.String `tfsdk:"object_schema_key"`
	Description     types.String `tfsdk:"description"`
	Status          types.String `tfsdk:"status"`
	WorkspaceID     types.String `tfsdk:"workspace_id"`
}

// NewObjectSchemaResource returns a new resource.Resource factory.
func NewObjectSchemaResource() resource.Resource {
	return &ObjectSchemaResource{}
}

func (r *ObjectSchemaResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assets_object_schema"
}

func (r *ObjectSchemaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
Manages a **JSM Assets object schema**.

An object schema is the top-level container for all object types and asset records.
Think of it as a database: it owns the type definitions (tables) and the data (rows).

**Important:** The ` + "`object_schema_key`" + ` is immutable after creation. Changing it
forces a new schema to be created (and the old one destroyed).
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The numeric ID assigned by JSM Assets.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable name for the schema (e.g. `Application Catalog`).",
			},
			"object_schema_key": schema.StringAttribute{
				Required: true,
				Description: "Short, uppercase identifier for the schema (e.g. `APP`). " +
					"Must be unique within the workspace. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Free-text description of the schema.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current lifecycle status of the schema as reported by the API.",
			},
			"workspace_id": schema.StringAttribute{
				Computed:    true,
				Description: "The JSM Assets workspace ID that owns this schema.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ObjectSchemaResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource configure type",
			fmt.Sprintf("Expected *client.Client, got %T.", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *ObjectSchemaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ObjectSchemaModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.CreateObjectSchema(ctx, client.CreateObjectSchemaRequest{
		Name:            plan.Name.ValueString(),
		ObjectSchemaKey: plan.ObjectSchemaKey.ValueString(),
		Description:     plan.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating object schema", err.Error())
		return
	}

	schemaToModel(s, &plan, r.client.WorkspaceID)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ObjectSchemaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ObjectSchemaModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.GetObjectSchema(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading object schema", err.Error())
		return
	}

	schemaToModel(s, &state, r.client.WorkspaceID)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ObjectSchemaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ObjectSchemaModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateObjectSchema(ctx, state.ID.ValueString(), client.UpdateObjectSchemaRequest{
		Name:            plan.Name.ValueString(),
		ObjectSchemaKey: plan.ObjectSchemaKey.ValueString(),
		Description:     plan.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating object schema", err.Error())
		return
	}

	schemaToModel(s, &plan, r.client.WorkspaceID)
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ObjectSchemaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ObjectSchemaModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteObjectSchema(ctx, state.ID.ValueString()); err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError("Error deleting object schema", err.Error())
		}
	}
}

// schemaToModel maps an API ObjectSchema to the Terraform state model.
func schemaToModel(s *client.ObjectSchema, m *ObjectSchemaModel, workspaceID string) {
	m.ID = types.StringValue(s.ID)
	m.Name = types.StringValue(s.Name)
	m.ObjectSchemaKey = types.StringValue(s.ObjectSchemaKey)
	m.Description = types.StringValue(s.Description)
	m.Status = types.StringValue(s.Status)
	m.WorkspaceID = types.StringValue(workspaceID)
}
