package resources

import (
	"context"
	"fmt"

	"github.com/badrory/jsm-terraform-provider/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ObjectTypeResource{}

// ObjectTypeResource manages a JSM Assets object type.
type ObjectTypeResource struct {
	client *client.Client
}

// ObjectTypeModel is the Terraform state model.
type ObjectTypeModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	IconID             types.String `tfsdk:"icon_id"`
	ObjectSchemaID     types.String `tfsdk:"object_schema_id"`
	ParentObjectTypeID types.String `tfsdk:"parent_object_type_id"`
	AbstractType       types.Bool   `tfsdk:"abstract_type"`
}

// NewObjectTypeResource returns a new resource.Resource factory.
func NewObjectTypeResource() resource.Resource {
	return &ObjectTypeResource{}
}

func (r *ObjectTypeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assets_object_type"
}

func (r *ObjectTypeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
Manages a **JSM Assets object type**.

An object type is a class (template) for asset records within a schema.
Examples: ` + "`Application`" + `, ` + "`Server`" + `, ` + "`Database`" + `.

The ` + "`object_schema_id`" + ` is immutable — changing it forces a new resource.

### Icon IDs

Icons are managed inside JSM. To find a valid icon ID:
1. Open **Assets** in JSM and navigate to your schema.
2. Create or view an object type and inspect the icon picker URL — the numeric
   segment is the icon ID.
3. The default generic icon is typically ID ` + "`1`" + `.
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
				Description: "The name of the object type (e.g. `Application`).",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Free-text description of the object type.",
			},
			"icon_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("1"),
				Description: "ID of the icon to display for this object type. Defaults to `1` (generic blue box).",
			},
			"object_schema_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the parent object schema. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"parent_object_type_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Optional ID of a parent object type to create a hierarchy. Leave empty for top-level types.",
			},
			"abstract_type": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "When true the type acts as an abstract base — it cannot have object instances directly.",
			},
		},
	}
}

func (r *ObjectTypeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected resource configure type",
			fmt.Sprintf("Expected *client.Client, got %T.", req.ProviderData))
		return
	}
	r.client = c
}

func (r *ObjectTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ObjectTypeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ot, err := r.client.CreateObjectType(ctx, client.CreateObjectTypeRequest{
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		IconID:             plan.IconID.ValueString(),
		ObjectSchemaID:     plan.ObjectSchemaID.ValueString(),
		ParentObjectTypeID: plan.ParentObjectTypeID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating object type", err.Error())
		return
	}

	objectTypeToModel(ot, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ObjectTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ObjectTypeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ot, err := r.client.GetObjectType(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading object type", err.Error())
		return
	}

	objectTypeToModel(ot, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ObjectTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ObjectTypeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ot, err := r.client.UpdateObjectType(ctx, state.ID.ValueString(), client.UpdateObjectTypeRequest{
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		IconID:             plan.IconID.ValueString(),
		ObjectSchemaID:     state.ObjectSchemaID.ValueString(), // immutable
		ParentObjectTypeID: plan.ParentObjectTypeID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating object type", err.Error())
		return
	}

	objectTypeToModel(ot, &plan)
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ObjectTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ObjectTypeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteObjectType(ctx, state.ID.ValueString()); err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError("Error deleting object type", err.Error())
		}
	}
}

// objectTypeToModel maps an API ObjectType to the Terraform state model.
func objectTypeToModel(ot *client.ObjectType, m *ObjectTypeModel) {
	m.ID = types.StringValue(ot.ID)
	m.Name = types.StringValue(ot.Name)
	m.Description = types.StringValue(ot.Description)
	m.IconID = types.StringValue(ot.Icon.ID)
	m.ObjectSchemaID = types.StringValue(ot.ObjectSchemaID)
	m.ParentObjectTypeID = types.StringValue(ot.ParentObjectTypeID)
	m.AbstractType = types.BoolValue(ot.AbstractObjectType)
}
