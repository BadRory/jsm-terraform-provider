package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/badrory/jsm-terraform-provider/internal/client"
)

var _ resource.Resource = &ObjectResource{}

// ObjectResource manages a JSM Assets object (asset instance).
type ObjectResource struct {
	client *client.Client
}

// ObjectModel is the Terraform state model for an asset object.
type ObjectModel struct {
	ID           types.String `tfsdk:"id"`
	ObjectTypeID types.String `tfsdk:"object_type_id"`
	Label        types.String `tfsdk:"label"`
	Attributes   types.Set    `tfsdk:"attributes"`
}

// objectAttributeModel models a single attribute value on an object.
// `attribute_id` maps to the ObjectTypeAttribute.ID (the definition), not the
// ObjectAttribute.ID (the instance value row).
type objectAttributeModel struct {
	AttributeID types.String `tfsdk:"attribute_id"`
	Value       types.String `tfsdk:"value"`
}

// objectAttrAttrTypes is the framework attr.Type map for attribute set elements.
var objectAttrAttrTypes = map[string]attr.Type{
	"attribute_id": types.StringType,
	"value":        types.StringType,
}

// NewObjectResource returns a new resource.Resource factory.
func NewObjectResource() resource.Resource {
	return &ObjectResource{}
}

func (r *ObjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assets_object"
}

func (r *ObjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
Manages a **JSM Assets object** — a concrete asset record (an instance of an
object type).

### attributes block

Use one ` + "`attribute`" + ` block per field you want to manage. The ` + "`attribute_id`" + `
must match the ` + "`id`" + ` exported by a ` + "`jsm_assets_object_type_attribute`" + ` resource
(or be looked up via ` + "`data.jsm_assets_object_type`" + `).

System attributes (like the built-in ` + "`Name`" + `) are also set through this block —
look up their IDs using the JSM UI or via data sources.

Attributes **not** listed here are left at their API defaults and are **not**
tracked by Terraform, preventing noisy diffs from system-populated fields.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The key assigned by JSM Assets (e.g. `APP-1`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"object_type_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the object type this object is an instance of. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"label": schema.StringAttribute{
				Computed:    true,
				Description: "The display label of the object (derived from the `Name` attribute by JSM).",
			},
			"attributes": schema.SetNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Attribute values to set on this object. Only attributes listed here are managed by Terraform.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"attribute_id": schema.StringAttribute{
							Required:    true,
							Description: "ID of the object type attribute definition (`jsm_assets_object_type_attribute.*.id`).",
						},
						"value": schema.StringAttribute{
							Required:    true,
							Description: "String value for this attribute. For User attributes supply the Jira account ID; for Object references supply the target object key/ID.",
						},
					},
				},
			},
		},
	}
}

func (r *ObjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ObjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ObjectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiAttrs, ds := planAttrsToAPIAttrs(ctx, plan.Attributes)
	resp.Diagnostics.Append(ds...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.client.CreateObject(ctx, client.CreateObjectRequest{
		ObjectTypeID: plan.ObjectTypeID.ValueString(),
		Attributes:   apiAttrs,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating object", err.Error())
		return
	}

	// Re-read to get the complete, server-side attribute list.
	obj, err = r.client.GetObject(ctx, obj.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading object after create", err.Error())
		return
	}

	filteredObj := filterObjectAttributes(obj, plan.Attributes)
	ds = objectToModel(ctx, filteredObj, &plan)
	resp.Diagnostics.Append(ds...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ObjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ObjectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.client.GetObject(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading object", err.Error())
		return
	}

	// Only reconcile attributes that are already tracked in state to avoid
	// picking up system attributes and causing perpetual diffs.
	filteredObj := filterObjectAttributes(obj, state.Attributes)
	ds := objectToModel(ctx, filteredObj, &state)
	resp.Diagnostics.Append(ds...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ObjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ObjectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiAttrs, ds := planAttrsToAPIAttrs(ctx, plan.Attributes)
	resp.Diagnostics.Append(ds...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.client.UpdateObject(ctx, state.ID.ValueString(), client.CreateObjectRequest{
		ObjectTypeID: state.ObjectTypeID.ValueString(),
		Attributes:   apiAttrs,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating object", err.Error())
		return
	}

	obj, err = r.client.GetObject(ctx, obj.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading object after update", err.Error())
		return
	}

	filteredObj := filterObjectAttributes(obj, plan.Attributes)
	ds = objectToModel(ctx, filteredObj, &plan)
	resp.Diagnostics.Append(ds...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ObjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ObjectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteObject(ctx, state.ID.ValueString()); err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError("Error deleting object", err.Error())
		}
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// planAttrsToAPIAttrs converts a types.Set of attribute blocks into the API
// request slice.
func planAttrsToAPIAttrs(ctx context.Context, attrsSet types.Set) ([]client.ObjectAttributeIn, diag.Diagnostics) {
	var ds diag.Diagnostics
	if attrsSet.IsNull() || attrsSet.IsUnknown() {
		return nil, ds
	}

	var models []objectAttributeModel
	ds.Append(attrsSet.ElementsAs(ctx, &models, false)...)
	if ds.HasError() {
		return nil, ds
	}

	out := make([]client.ObjectAttributeIn, 0, len(models))
	for _, m := range models {
		out = append(out, client.ObjectAttributeIn{
			ObjectTypeAttributeID: m.AttributeID.ValueString(),
			ObjectAttributeValues: []client.ObjectAttributeValueIn{
				{Value: m.Value.ValueString()},
			},
		})
	}
	return out, ds
}

// objectToModel writes API Object data into the Terraform state model.
func objectToModel(_ context.Context, obj *client.Object, m *ObjectModel) diag.Diagnostics {
	var ds diag.Diagnostics

	m.ID = types.StringValue(obj.ID)
	m.Label = types.StringValue(obj.Label)
	m.ObjectTypeID = types.StringValue(obj.ObjectType.ID)

	attrObjs := make([]attr.Value, 0, len(obj.Attributes))
	for _, a := range obj.Attributes {
		val := ""
		if len(a.ObjectAttributeValues) > 0 {
			val = a.ObjectAttributeValues[0].Value
		}

		// Prefer the nested ObjectTypeAttribute.ID; fall back to the top-level field.
		attrID := a.ObjectTypeAttribute.ID
		if attrID == "" {
			attrID = a.ObjectTypeAttributeID
		}

		attrObj, d := types.ObjectValue(objectAttrAttrTypes, map[string]attr.Value{
			"attribute_id": types.StringValue(attrID),
			"value":        types.StringValue(val),
		})
		ds.Append(d...)
		if ds.HasError() {
			return ds
		}
		attrObjs = append(attrObjs, attrObj)
	}

	set, d := types.SetValue(types.ObjectType{AttrTypes: objectAttrAttrTypes}, attrObjs)
	ds.Append(d...)
	if !ds.HasError() {
		m.Attributes = set
	}
	return ds
}

// filterObjectAttributes returns a shallow copy of obj with only the attributes
// whose IDs are tracked in the managedAttrs set, avoiding perpetual diffs from
// system-populated fields.
func filterObjectAttributes(obj *client.Object, managedAttrs types.Set) *client.Object {
	if managedAttrs.IsNull() || managedAttrs.IsUnknown() {
		return obj
	}

	// Build a lookup set of tracked attribute definition IDs.
	tracked := map[string]bool{}
	for _, elem := range managedAttrs.Elements() {
		if o, ok := elem.(types.Object); ok {
			if idVal, ok2 := o.Attributes()["attribute_id"]; ok2 {
				if sv, ok3 := idVal.(types.String); ok3 {
					tracked[sv.ValueString()] = true
				}
			}
		}
	}

	filtered := *obj // shallow copy
	filtered.Attributes = nil
	for _, a := range obj.Attributes {
		attrID := a.ObjectTypeAttribute.ID
		if attrID == "" {
			attrID = a.ObjectTypeAttributeID
		}
		if tracked[attrID] {
			filtered.Attributes = append(filtered.Attributes, a)
		}
	}
	return &filtered
}
