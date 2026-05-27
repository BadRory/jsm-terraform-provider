// Package client provides a Go HTTP client for the JSM Assets REST API.
//
// Authentication uses Atlassian basic auth (email + API token).
// The workspace ID is auto-discovered from the site URL if not provided.
package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// assetsBaseFormat is the format string for the JSM Assets API base URL.
	// The workspace ID is substituted at runtime.
	assetsBaseFormat = "https://api.atlassian.com/jsm/assets/workspace/%s/v1"

	// workspaceDiscoveryPath is appended to the site URL to discover workspace IDs.
	workspaceDiscoveryPath = "/rest/servicedeskapi/assets/workspace"
)

// NotFoundError is returned when the API responds with a 404.
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID %q not found", e.Resource, e.ID)
}

// IsNotFound returns true if the error is a NotFoundError.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*NotFoundError)
	return ok
}

// Client is a configured API client for JSM Assets.
type Client struct {
	httpClient  *http.Client
	siteURL     string
	email       string
	apiToken    string
	WorkspaceID string
	assetsBase  string
}

// NewClient creates and configures a Client. If workspaceID is empty it is
// auto-discovered by calling the workspace discovery endpoint.
func NewClient(siteURL, email, apiToken, workspaceID string) (*Client, error) {
	// Normalise site URL (strip trailing slash)
	siteURL = strings.TrimRight(siteURL, "/")

	c := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		siteURL:    siteURL,
		email:      email,
		apiToken:   apiToken,
	}

	if workspaceID == "" {
		var err error
		workspaceID, err = c.discoverWorkspaceID(context.Background())
		if err != nil {
			return nil, fmt.Errorf("auto-discovering JSM Assets workspace ID: %w", err)
		}
	}

	c.WorkspaceID = workspaceID
	c.assetsBase = fmt.Sprintf(assetsBaseFormat, workspaceID)
	return c, nil
}

// ---------------------------------------------------------------------------
// internal helpers
// ---------------------------------------------------------------------------

func (c *Client) authHeader() string {
	raw := fmt.Sprintf("%s:%s", c.email, c.apiToken)
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(raw))
}

// request builds and executes an HTTP request, returning the raw response.
func (c *Client) request(ctx context.Context, method, url string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshalling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request %s %s: %w", method, url, err)
	}

	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request %s %s: %w", method, url, err)
	}
	return resp, nil
}

// doJSON executes a request and decodes the JSON response into result.
// result may be nil when no response body is expected (e.g. DELETE).
func (c *Client) doJSON(ctx context.Context, method, url string, body interface{}, result interface{}) error {
	resp, err := c.request(ctx, method, url, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return &NotFoundError{Resource: "resource", ID: url}
	case http.StatusNoContent:
		return nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, string(respBytes))
	}

	if result != nil && len(respBytes) > 0 {
		if err := json.Unmarshal(respBytes, result); err != nil {
			return fmt.Errorf("decoding response from %s %s: %w — body: %s", method, url, err, string(respBytes))
		}
	}
	return nil
}

// discoverWorkspaceID queries the site for its Assets workspace ID.
func (c *Client) discoverWorkspaceID(ctx context.Context) (string, error) {
	url := c.siteURL + workspaceDiscoveryPath
	var resp WorkspaceResponse
	if err := c.doJSON(ctx, http.MethodGet, url, nil, &resp); err != nil {
		return "", fmt.Errorf("fetching workspace list from %s: %w", url, err)
	}
	if len(resp.Values) == 0 {
		return "", fmt.Errorf("no JSM Assets workspaces found at %s — ensure JSM Assets is enabled", url)
	}
	return resp.Values[0].WorkspaceID, nil
}

// ---------------------------------------------------------------------------
// Object Schema
// ---------------------------------------------------------------------------

// ListObjectSchemas returns all object schemas in the workspace.
func (c *Client) ListObjectSchemas(ctx context.Context) ([]ObjectSchema, error) {
	var resp ObjectSchemaListResponse
	if err := c.doJSON(ctx, http.MethodGet, c.assetsBase+"/objectschema/list", nil, &resp); err != nil {
		return nil, fmt.Errorf("listing object schemas: %w", err)
	}
	return resp.ObjectschemaList, nil
}

// CreateObjectSchema creates a new object schema.
func (c *Client) CreateObjectSchema(ctx context.Context, req CreateObjectSchemaRequest) (*ObjectSchema, error) {
	var result ObjectSchema
	if err := c.doJSON(ctx, http.MethodPost, c.assetsBase+"/objectschema/create", req, &result); err != nil {
		return nil, fmt.Errorf("creating object schema %q: %w", req.Name, err)
	}
	return &result, nil
}

// GetObjectSchema retrieves an object schema by ID.
func (c *Client) GetObjectSchema(ctx context.Context, id string) (*ObjectSchema, error) {
	var result ObjectSchema
	url := fmt.Sprintf("%s/objectschema/%s", c.assetsBase, id)
	if err := c.doJSON(ctx, http.MethodGet, url, nil, &result); err != nil {
		return nil, fmt.Errorf("getting object schema %s: %w", id, err)
	}
	return &result, nil
}

// UpdateObjectSchema updates an existing object schema.
func (c *Client) UpdateObjectSchema(ctx context.Context, id string, req UpdateObjectSchemaRequest) (*ObjectSchema, error) {
	var result ObjectSchema
	url := fmt.Sprintf("%s/objectschema/%s", c.assetsBase, id)
	if err := c.doJSON(ctx, http.MethodPut, url, req, &result); err != nil {
		return nil, fmt.Errorf("updating object schema %s: %w", id, err)
	}
	return &result, nil
}

// DeleteObjectSchema deletes an object schema by ID.
func (c *Client) DeleteObjectSchema(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/objectschema/%s", c.assetsBase, id)
	if err := c.doJSON(ctx, http.MethodDelete, url, nil, nil); err != nil {
		return fmt.Errorf("deleting object schema %s: %w", id, err)
	}
	return nil
}

// GetObjectTypesBySchema returns all object types (flat list) for a schema.
func (c *Client) GetObjectTypesBySchema(ctx context.Context, schemaID string) ([]ObjectType, error) {
	url := fmt.Sprintf("%s/objectschema/%s/objecttypes/flat", c.assetsBase, schemaID)
	var result []ObjectType
	if err := c.doJSON(ctx, http.MethodGet, url, nil, &result); err != nil {
		return nil, fmt.Errorf("listing object types for schema %s: %w", schemaID, err)
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Object Type
// ---------------------------------------------------------------------------

// CreateObjectType creates a new object type.
func (c *Client) CreateObjectType(ctx context.Context, req CreateObjectTypeRequest) (*ObjectType, error) {
	var result ObjectType
	if err := c.doJSON(ctx, http.MethodPost, c.assetsBase+"/objecttype/create", req, &result); err != nil {
		return nil, fmt.Errorf("creating object type %q: %w", req.Name, err)
	}
	return &result, nil
}

// GetObjectType retrieves an object type by ID.
func (c *Client) GetObjectType(ctx context.Context, id string) (*ObjectType, error) {
	var result ObjectType
	url := fmt.Sprintf("%s/objecttype/%s", c.assetsBase, id)
	if err := c.doJSON(ctx, http.MethodGet, url, nil, &result); err != nil {
		return nil, fmt.Errorf("getting object type %s: %w", id, err)
	}
	return &result, nil
}

// UpdateObjectType updates an existing object type.
func (c *Client) UpdateObjectType(ctx context.Context, id string, req UpdateObjectTypeRequest) (*ObjectType, error) {
	var result ObjectType
	url := fmt.Sprintf("%s/objecttype/%s", c.assetsBase, id)
	if err := c.doJSON(ctx, http.MethodPut, url, req, &result); err != nil {
		return nil, fmt.Errorf("updating object type %s: %w", id, err)
	}
	return &result, nil
}

// DeleteObjectType deletes an object type by ID.
func (c *Client) DeleteObjectType(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/objecttype/%s", c.assetsBase, id)
	if err := c.doJSON(ctx, http.MethodDelete, url, nil, nil); err != nil {
		return fmt.Errorf("deleting object type %s: %w", id, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Object Type Attribute
// ---------------------------------------------------------------------------

// GetObjectTypeAttributes lists all attributes for an object type.
func (c *Client) GetObjectTypeAttributes(ctx context.Context, objectTypeID string) ([]ObjectTypeAttribute, error) {
	url := fmt.Sprintf("%s/objecttype/%s/attributes", c.assetsBase, objectTypeID)
	var result []ObjectTypeAttribute
	if err := c.doJSON(ctx, http.MethodGet, url, nil, &result); err != nil {
		return nil, fmt.Errorf("listing attributes for object type %s: %w", objectTypeID, err)
	}
	return result, nil
}

// GetObjectTypeAttributeByID fetches a single attribute by listing the parent
// type's attributes and matching on ID (the API has no direct GET for one).
func (c *Client) GetObjectTypeAttributeByID(ctx context.Context, objectTypeID, attributeID string) (*ObjectTypeAttribute, error) {
	attrs, err := c.GetObjectTypeAttributes(ctx, objectTypeID)
	if err != nil {
		return nil, err
	}
	for i := range attrs {
		if attrs[i].ID == attributeID {
			return &attrs[i], nil
		}
	}
	return nil, &NotFoundError{Resource: "object type attribute", ID: attributeID}
}

// CreateObjectTypeAttribute creates a new attribute on an object type.
func (c *Client) CreateObjectTypeAttribute(ctx context.Context, objectTypeID string, req CreateObjectTypeAttributeRequest) (*ObjectTypeAttribute, error) {
	var result ObjectTypeAttribute
	url := fmt.Sprintf("%s/objecttypeattribute/%s", c.assetsBase, objectTypeID)
	if err := c.doJSON(ctx, http.MethodPost, url, req, &result); err != nil {
		return nil, fmt.Errorf("creating attribute %q on object type %s: %w", req.Name, objectTypeID, err)
	}
	return &result, nil
}

// UpdateObjectTypeAttribute updates an existing attribute.
func (c *Client) UpdateObjectTypeAttribute(ctx context.Context, objectTypeID, attributeID string, req CreateObjectTypeAttributeRequest) (*ObjectTypeAttribute, error) {
	var result ObjectTypeAttribute
	url := fmt.Sprintf("%s/objecttypeattribute/%s/%s", c.assetsBase, objectTypeID, attributeID)
	if err := c.doJSON(ctx, http.MethodPut, url, req, &result); err != nil {
		return nil, fmt.Errorf("updating attribute %s: %w", attributeID, err)
	}
	return &result, nil
}

// DeleteObjectTypeAttribute deletes an attribute by ID.
func (c *Client) DeleteObjectTypeAttribute(ctx context.Context, attributeID string) error {
	url := fmt.Sprintf("%s/objecttypeattribute/%s", c.assetsBase, attributeID)
	if err := c.doJSON(ctx, http.MethodDelete, url, nil, nil); err != nil {
		return fmt.Errorf("deleting attribute %s: %w", attributeID, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Object (asset instance)
// ---------------------------------------------------------------------------

// CreateObject creates a new asset object.
func (c *Client) CreateObject(ctx context.Context, req CreateObjectRequest) (*Object, error) {
	var result Object
	if err := c.doJSON(ctx, http.MethodPost, c.assetsBase+"/object/create", req, &result); err != nil {
		return nil, fmt.Errorf("creating object of type %s: %w", req.ObjectTypeID, err)
	}
	return &result, nil
}

// GetObject retrieves an object and its attributes by ID.
func (c *Client) GetObject(ctx context.Context, id string) (*Object, error) {
	var result Object
	url := fmt.Sprintf("%s/object/%s", c.assetsBase, id)
	if err := c.doJSON(ctx, http.MethodGet, url, nil, &result); err != nil {
		return nil, fmt.Errorf("getting object %s: %w", id, err)
	}

	// Fetch attributes separately — the base GET may not include them.
	attrs, err := c.GetObjectAttributes(ctx, id)
	if err != nil {
		return nil, err
	}
	result.Attributes = attrs
	return &result, nil
}

// GetObjectAttributes returns all attributes for an object instance.
func (c *Client) GetObjectAttributes(ctx context.Context, objectID string) ([]ObjectAttribute, error) {
	url := fmt.Sprintf("%s/object/%s/attributes", c.assetsBase, objectID)
	var result []ObjectAttribute
	if err := c.doJSON(ctx, http.MethodGet, url, nil, &result); err != nil {
		return nil, fmt.Errorf("listing attributes for object %s: %w", objectID, err)
	}
	return result, nil
}

// UpdateObject updates an existing object.
func (c *Client) UpdateObject(ctx context.Context, id string, req CreateObjectRequest) (*Object, error) {
	var result Object
	url := fmt.Sprintf("%s/object/%s", c.assetsBase, id)
	if err := c.doJSON(ctx, http.MethodPut, url, req, &result); err != nil {
		return nil, fmt.Errorf("updating object %s: %w", id, err)
	}
	return &result, nil
}

// DeleteObject deletes an object by ID.
func (c *Client) DeleteObject(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/object/%s", c.assetsBase, id)
	if err := c.doJSON(ctx, http.MethodDelete, url, nil, nil); err != nil {
		return fmt.Errorf("deleting object %s: %w", id, err)
	}
	return nil
}

// SearchObjects executes an AQL query and returns matching objects.
// The results include attributes when includeAttributes is true.
func (c *Client) SearchObjects(ctx context.Context, aql string, includeAttrs bool) ([]Object, error) {
	req := AQLRequest{
		QLQuery:        aql,
		Page:           1,
		ResultsPerPage: 25,
		IncludeAttrs:   includeAttrs,
	}
	var resp AQLResponse
	if err := c.doJSON(ctx, http.MethodPost, c.assetsBase+"/object/aql", req, &resp); err != nil {
		return nil, fmt.Errorf("AQL search %q: %w", aql, err)
	}
	return resp.Values, nil
}
