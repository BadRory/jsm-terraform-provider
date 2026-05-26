package resources

import (
	"context"
	"fmt"

	"github.com/badrory/jsm-terraform-provider/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ObjectTypeAttributeResource{}

// ObjectTypeAttributeResource manages a JSM Assets object type attribute.
type ObjectTypeAttributeResource struct {
	client *client.Client
}

// ObjectTypeAttributeModel is the Terraform state model.
type ObjectTypeAttributeModel struct {
	ID                 types.String `tfsdk:"id"`
	ObjectTypeID       types.String `tfsdk:"object_type_id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Type               types.Int64  `tfsdk:"type"`
	DefaultTypeID      types.Int64  `tfsdk:"default_type_id"`
	TypeValue          types.String `tfsdk:"type_value"`
	AdditionalValue    types.String `tfsdk:"additional_value"`
	MinimumCardinality types.Int64  `tfsdk:"minimum_cardinality"`
	MaximumCardinality types.Int64  `tfsdk:"maximum_cardinality"`
	Indexed            types.Bool   `tfsdk:"indexed"`
	Unique             types.Bool   `tfsdk:"unique"`
	Summable           types.Bool   `tfsdk:"summable"`
	RegexValidation    types.String `tfsdk:"regex_validation"`
	System             types.Bool   `tfsdk:"system"`
}

// NewObjectTypeAttributeResource returns a new resource.Resource factory.
func NewObjectTypeAttributeResource() resource.Resource {
	return &ObjectTypeAttributeResource{}
}

func (r *ObjectTypeAttributeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assets_object_type_attribute"
}

func (r *ObjectTypeAttributeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
Manages a **JSM Assets object type attribute** (a field definition on an object type).

### ` + "`type`" + ` values

| Value | Meaning |
|-------|---------|
| ` + "`0`" + ` | Default — basic scalar type governed by ` + "`default_type_id`" + ` |
| ` + "`1`" + ` | Object reference (link to another asset) — set ` + "`type_value`" + ` to the target object type ID |
| ` + "`2`" + ` | User (Jira user picker) |
| ` + "`6`" + ` | Status (set ` + "`type_value`" + ` to the target status type ID) |

### ` + "`default_type_id`" + ` values (only when ` + "`type = 0`" + `)

| Value | Meaning |
|-------|---------|
| ` + "`0`" + ` | Text |
| ` + "`1`" + ` | Integer |
| ` + "`2`" + ` | Boolean |
| ` + "`3`" + ` | Double |
| ` + "`4`" + ` | Date |
| ` + "`5`" + ` | Time |
| ` + "`6`" + ` | DateTime |
| ` + "`7`" + ` | URL |
| ` + "`8`" + ` | Email |
| ` + "`9`" + ` | Textarea |
| ` + "`10`" + ` | Select (dropdown) |
| ` + "`11`" + ` | IP Address |
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The numeric ID assigned by JSM Assets.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"object_type_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the parent object type. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The attribute name (e.g. `OktaGroupID`, `Approvers`).",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Free-text description of the attribute.",
			},
			"type": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Description: "High-level attribute kind. See the table above. Default: `0` (Default/scalar).",
			},
			"default_type_id": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Description: "Scalar data type when `type = 0`. See the table above. Default: `0` (Text).",
			},
			"type_value": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "For type=1 (Object reference): the target object type ID. For type=6 (Status): the status type ID.",
			},
			"additional_value": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Extra configuration value; usage depends on the attribute type.",
			},
			"minimum_cardinality": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Description: "Minimum number of values allowed (0 = not required).",
			},
			"maximum_cardinality": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Description: "Maximum number of values allowed (1 = single-value attribute).",
			},
			"indexed": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether to index this attribute for faster searches.",
			},
			"unique": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Enforce uniqueness across all objects of this type.",
			},
			"summable": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the attribute values can be summed (useful for numeric attributes).",
			},
			"regex_validation": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Optional regular expression for input validation.",
			},
			"system": schema.BoolAttribute{
				Computed:    true,
				Description: "True when this is a system-managed attribute that cannot be deleted.",
			},
		},
	}
}

func (r *ObjectTypeAttributeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ObjectTypeAttributeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ObjectTypeAttributeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attr, err := r.client.CreateObjectTypeAttribute(ctx, plan.ObjectTypeID.ValueString(), modelToCreateAttrReq(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error creating object type attribute", err.Error())
		return
	}

	attributeToModel(attr, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ObjectTypeAttributeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ObjectTypeAttributeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attr, err := r.client.GetObjectTypeAttributeByID(ctx, state.ObjectTypeID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading object type attribute", err.Error())
		return
	}

	attributeToModel(attr, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ObjectTypeAttributeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ObjectTypeAttributeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attr, err := r.client.UpdateObjectTypeAttribute(
		ctx,
		state.ObjectTypeID.ValueString(),
		state.ID.ValueString(),
		modelToCreateAttrReq(plan),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating object type attribute", err.Error())
		return
	}

	attributeToModel(attr, &plan)
	plan.ID = state.ID
	plan.ObjectTypeID = state.ObjectTypeID
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ObjectTypeAttributeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ObjectTypeAttributeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.System.ValueBool() {
		resp.Diagnostics.AddWarning(
			"Cannot delete system attribute",
			fmt.Sprintf("Attribute %q (ID %s) is a system-managed attribute and cannot be deleted. It has been removed from Terraform state only.", state.Name.ValueString(), state.ID.ValueString()),
		)
		return
	}

	if err := r.client.DeleteObjectTypeAttribute(ctx, state.ID.ValueString()); err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError("Error deleting object type attribute", err.Error())
		}
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func modelToCreateAttrReq(m ObjectTypeAttributeModel) client.CreateObjectTypeAttributeRequest {
	return client.CreateObjectTypeAttributeRequest{
		Name:               m.Name.ValueString(),
		Description:        m.Description.ValueString(),
		Type:               int(m.Type.ValueInt64()),
		DefaultTypeID:      int(m.DefaultTypeID.ValueInt64()),
		TypeValue:          m.TypeValue.ValueString(),
		AdditionalValue:    m.AdditionalValue.ValueString(),
		MinimumCardinality: int(m.MinimumCardinality.ValueInt64()),
		MaximumCardinality: int(m.MaximumCardinality.ValueInt64()),
		Indexed:            m.Indexed.ValueBool(),
		Unique:             m.Unique.ValueBool(),
		Summable:           m.Summable.ValueBool(),
		RegexValidation:    m.RegexValidation.ValueString(),
	}
}

func attributeToModel(attr *client.ObjectTypeAttribute, m *ObjectTypeAttributeModel) {
	m.ID = types.StringValue(attr.ID)
	m.Name = types.StringValue(attr.Name)
	m.Description = types.StringValue(attr.Description)
	m.Type = types.Int64Value(int64(attr.Type))
	m.DefaultTypeID = types.Int64Value(int64(attr.DefaultType.ID))
	m.TypeValue = types.StringValue(attr.TypeValue)
	m.AdditionalValue = types.StringValue(attr.AdditionalValue)
	m.MinimumCardinality = types.Int64Value(int64(attr.MinimumCardinality))
	m.MaximumCardinality = types.Int64Value(int64(attr.MaximumCardinality))
	m.Indexed = types.BoolValue(attr.Indexed)
	m.Unique = types.BoolValue(attr.Unique)
	m.Summable = types.BoolValue(attr.Summable)
	m.RegexValidation = types.StringValue(attr.RegexValidation)
	m.System = types.BoolValue(attr.System)
	m.ObjectTypeID = types.StringValue(attr.ObjectType.ID)
}
