package client

// ---------------------------------------------------------------------------
// Workspace
// ---------------------------------------------------------------------------

// WorkspaceResponse is the response from the JSM Assets workspace discovery
// endpoint: GET {site}/rest/servicedeskapi/assets/workspace
type WorkspaceResponse struct {
	Values []Workspace `json:"values"`
}

// Workspace represents a single JSM Assets workspace.
type Workspace struct {
	WorkspaceID string `json:"workspaceId"`
}

// ---------------------------------------------------------------------------
// Object Schema
// ---------------------------------------------------------------------------

// ObjectSchema represents a JSM Assets object schema (top-level container).
type ObjectSchema struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	ObjectSchemaKey string `json:"objectSchemaKey"`
	Description     string `json:"description,omitempty"`
	Status          string `json:"status,omitempty"`
	Created         string `json:"created,omitempty"`
	Updated         string `json:"updated,omitempty"`
	WorkspaceID     string `json:"workspaceId,omitempty"`
}

// ObjectSchemaListResponse is the paginated list of object schemas.
type ObjectSchemaListResponse struct {
	ObjectschemaList []ObjectSchema `json:"objectschemaList"`
	StartIndex       int            `json:"startIndex"`
	Size             int            `json:"size"`
	IsLast           bool           `json:"isLast"`
}

// CreateObjectSchemaRequest is the payload for POST /objectschema/create.
type CreateObjectSchemaRequest struct {
	Name            string `json:"name"`
	ObjectSchemaKey string `json:"objectSchemaKey"`
	Description     string `json:"description,omitempty"`
}

// UpdateObjectSchemaRequest is the payload for PUT /objectschema/{id}.
// The key cannot be changed after creation; the API ignores changes to it.
type UpdateObjectSchemaRequest struct {
	Name            string `json:"name"`
	ObjectSchemaKey string `json:"objectSchemaKey"`
	Description     string `json:"description,omitempty"`
}

// ---------------------------------------------------------------------------
// Object Type
// ---------------------------------------------------------------------------

// ObjectType represents a JSM Assets object type (a class of assets).
type ObjectType struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	Icon               Icon   `json:"icon"`
	ObjectSchemaID     string `json:"objectSchemaId,omitempty"`
	ParentObjectTypeID string `json:"parentObjectTypeId,omitempty"`
	Inherited          bool   `json:"inherited,omitempty"`
	AbstractObjectType bool   `json:"abstractObjectType,omitempty"`
}

// Icon is the visual icon associated with an object type.
type Icon struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	URL16 string `json:"url16,omitempty"`
	URL48 string `json:"url48,omitempty"`
}

// CreateObjectTypeRequest is the payload for POST /objecttype/create.
type CreateObjectTypeRequest struct {
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	IconID             string `json:"iconId"`
	ObjectSchemaID     string `json:"objectSchemaId"`
	ParentObjectTypeID string `json:"parentObjectTypeId,omitempty"`
}

// UpdateObjectTypeRequest is the payload for PUT /objecttype/{id}.
type UpdateObjectTypeRequest struct {
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	IconID             string `json:"iconId"`
	ObjectSchemaID     string `json:"objectSchemaId"`
	ParentObjectTypeID string `json:"parentObjectTypeId,omitempty"`
}

// ---------------------------------------------------------------------------
// Object Type Attribute
// ---------------------------------------------------------------------------

// AttributeType enumerates the high-level attribute kind.
// 0 = Default (basic data types), 1 = Object reference, 2 = User, 6 = Status.
type AttributeType int

const (
	AttributeTypeDefault AttributeType = 0
	AttributeTypeObject  AttributeType = 1
	AttributeTypeUser    AttributeType = 2
	AttributeTypeStatus  AttributeType = 6
)

// DefaultType maps the integer ID to a human-readable data type for
// AttributeType=0 (Default) attributes.
//
//	0 = Text, 1 = Integer, 2 = Boolean, 3 = Double, 4 = Date,
//	5 = Time, 6 = DateTime, 7 = URL, 8 = Email, 9 = Textarea,
//	10 = Select, 11 = IP Address
type DefaultType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ObjectTypeAttribute is the definition of a single attribute on an object type.
type ObjectTypeAttribute struct {
	ID                 string      `json:"id"`
	Name               string      `json:"name"`
	Description        string      `json:"description,omitempty"`
	Type               int         `json:"type"`
	DefaultType        DefaultType `json:"defaultType"`
	TypeValue          string      `json:"typeValue,omitempty"`
	AdditionalValue    string      `json:"additionalValue,omitempty"`
	MinimumCardinality int         `json:"minimumCardinality"`
	MaximumCardinality int         `json:"maximumCardinality"`
	System             bool        `json:"system,omitempty"`
	Removable          bool        `json:"removable,omitempty"`
	Indexed            bool        `json:"indexed,omitempty"`
	Unique             bool        `json:"unique,omitempty"`
	Summable           bool        `json:"summable,omitempty"`
	RegexValidation    string      `json:"regexValidation,omitempty"`
	ObjectType         ObjectType  `json:"objectType"`
}

// CreateObjectTypeAttributeRequest is the payload for
// POST /objecttypeattribute/{objectTypeId}.
type CreateObjectTypeAttributeRequest struct {
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	Type               int    `json:"type"`
	DefaultTypeID      int    `json:"defaultTypeId,omitempty"`
	TypeValue          string `json:"typeValue,omitempty"`
	AdditionalValue    string `json:"additionalValue,omitempty"`
	MinimumCardinality int    `json:"minimumCardinality,omitempty"`
	MaximumCardinality int    `json:"maximumCardinality,omitempty"`
	Indexed            bool   `json:"indexed,omitempty"`
	Unique             bool   `json:"unique,omitempty"`
	Summable           bool   `json:"summable,omitempty"`
	RegexValidation    string `json:"regexValidation,omitempty"`
}

// ---------------------------------------------------------------------------
// Object (asset instance)
// ---------------------------------------------------------------------------

// Object is an instance of an ObjectType — an individual asset record.
type Object struct {
	ID          string            `json:"id"`
	Label       string            `json:"label"`
	ObjectType  ObjectType        `json:"objectType"`
	Attributes  []ObjectAttribute `json:"attributes,omitempty"`
	HasAvatar   bool              `json:"hasAvatar,omitempty"`
	AvatarUUID  string            `json:"avatarUUID,omitempty"`
	WorkspaceID string            `json:"workspaceId,omitempty"`
	Created     string            `json:"created,omitempty"`
	Updated     string            `json:"updated,omitempty"`
}

// ObjectAttribute is the value of a single attribute on an object instance.
type ObjectAttribute struct {
	ID                    string                 `json:"id"`
	ObjectTypeAttribute   ObjectTypeAttribute    `json:"objectTypeAttribute"`
	ObjectTypeAttributeID string                 `json:"objectTypeAttributeId,omitempty"`
	ObjectAttributeValues []ObjectAttributeValue `json:"objectAttributeValues"`
}

// ObjectAttributeValue holds one value inside an ObjectAttribute.
type ObjectAttributeValue struct {
	Value        string `json:"value"`
	DisplayValue string `json:"displayValue,omitempty"`
	SearchValue  string `json:"searchValue,omitempty"`
}

// CreateObjectRequest is the payload for POST /object/create and
// PUT /object/{id}.
type CreateObjectRequest struct {
	ObjectTypeID string              `json:"objectTypeId"`
	Attributes   []ObjectAttributeIn `json:"attributes"`
	HasAvatar    bool                `json:"hasAvatar,omitempty"`
	AvatarUUID   string              `json:"avatarUUID,omitempty"`
}

// ObjectAttributeIn is a single attribute entry in a create/update request.
type ObjectAttributeIn struct {
	ObjectTypeAttributeID string                   `json:"objectTypeAttributeId"`
	ObjectAttributeValues []ObjectAttributeValueIn `json:"objectAttributeValues"`
}

// ObjectAttributeValueIn is a single value in an attribute entry.
type ObjectAttributeValueIn struct {
	Value string `json:"value"`
}

// AQLRequest is the payload for POST /object/aql.
type AQLRequest struct {
	ObjectTypeID   string `json:"objectTypeId,omitempty"`
	ObjectSchemaID string `json:"objectSchemaId,omitempty"`
	QLQuery        string `json:"qlQuery"`
	Page           int    `json:"page,omitempty"`
	ResultsPerPage int    `json:"resultsPerPage,omitempty"`
	Asc            int    `json:"asc,omitempty"`
	IncludeAttrs   bool   `json:"includeAttributes"`
}

// AQLResponse is the paginated response from POST /object/aql.
type AQLResponse struct {
	Values     []Object `json:"values"`
	StartAt    int      `json:"startAt"`
	MaxResults int      `json:"maxResults"`
	TotalCount int      `json:"totalCount"`
	IsLast     bool     `json:"isLast"`
}
