package schematic

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

var resolveTests = []struct {
	Schema *Schema
}{
	{
		Schema: &Schema{
			Title: "Selfreferencing",
			Type:  "object",
			Definitions: map[string]*Schema{
				"blog": {
					Title: "Blog",
					Type:  "object",
					Links: []*Link{
						{
							Title:  "Create",
							Rel:    "create",
							HRef:   NewHRef("/blogs"),
							Method: "POST",
							Schema: &Schema{
								Type: "object",
								Properties: map[string]*Schema{
									"name": {
										Ref: NewReference("#/definitions/blog/definitions/name"),
									},
								},
							},
							TargetSchema: &Schema{
								Ref: NewReference("#/definitions/blog"),
							},
						},
						{
							Title:  "Get",
							Rel:    "self",
							HRef:   NewHRef("/blogs/{(%23%2Fdefinitions%2Fblog%2Fdefinitions%2Fid)}"),
							Method: "GET",
							TargetSchema: &Schema{
								Ref: NewReference("#/definitions/blog"),
							},
						},
					},
					Definitions: map[string]*Schema{
						"id": {
							Type: "string",
						},
						"name": {
							Type: "string",
						},
					},
					Properties: map[string]*Schema{
						"id": {
							Ref: NewReference("#/definitions/blog/definitions/id"),
						},
						"name": {
							Ref: NewReference("#/definitions/blog/definitions/name"),
						},
					},
				},
			},
			Properties: map[string]*Schema{
				"blog": {
					Ref: NewReference("#/definitions/blog"),
				},
			},
		},
	},
}

func TestResolve(t *testing.T) {
	for i, rt := range resolveTests {
		t.Run(fmt.Sprintf("resolveTests[%d]", i), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatal(r)
				}
			}()

			rt.Schema.Resolve(nil, ResolvedSet{})
		})
	}
}

var typeTests = []struct {
	Schema *Schema
	Type   string
}{
	{
		Schema: &Schema{
			Type: "boolean",
		},
		Type: "bool",
	},
	{
		Schema: &Schema{
			Type: "number",
		},
		Type: "float64",
	},
	{
		Schema: &Schema{
			Type: "integer",
		},
		Type: "int",
	},
	{
		Schema: &Schema{
			Type: "any",
		},
		Type: "interface{}",
	},
	{
		Schema: &Schema{
			Type: "string",
		},
		Type: "string",
	},
	{
		Schema: &Schema{
			Type:   "string",
			Format: "date-time",
		},
		Type: "time.Time",
	},
	{
		Schema: &Schema{
			Type: []interface{}{"null", "string"},
		},
		Type: "*string",
	},
	{
		Schema: &Schema{
			Type: "array",
			Items: &Schema{
				Type: "string",
			},
		},
		Type: "[]string",
	},
	{
		Schema: &Schema{
			Type: "array",
		},
		Type: "[]interface{}",
	},
	{
		Schema: &Schema{
			Type:                 "object",
			AdditionalProperties: false,
			PatternProperties: map[string]*Schema{
				"^\\w+$": {
					Type: "string",
				},
			},
		},
		Type: "map[string]string",
	},
	{
		Schema: &Schema{
			Type:                 "object",
			AdditionalProperties: false,
			PatternProperties: map[string]*Schema{
				"^\\w+$": {
					Type: []interface{}{"string", "null"},
				},
			},
		},
		Type: "map[string]*string",
	},
	{
		Schema: &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"counter": {
					Type: "integer",
				},
			},
			Required: []string{"counter"},
		},
		Type: "Counter int",
	},
	{
		Schema: &Schema{
			Type:   []interface{}{"null", "string"},
			Format: "date-time",
		},
		Type: "*time.Time",
	},
}

func TestSchemaType(t *testing.T) {
	for i, tt := range typeTests {
		kind := tt.Schema.GoType()
		if !strings.Contains(kind, tt.Type) {
			t.Errorf("%d: wants %v, got %v", i, tt.Type, kind)
		}
	}
}

var linkTests = []struct {
	Link *Link
	Type string
}{
	{
		Link: &Link{
			Schema: &Schema{
				Properties: map[string]*Schema{
					"string": {
						Type: "string",
					},
				},
				Type:     "object",
				Required: []string{"string"},
			},
		},
		Type: "String string",
	},
	{
		Link: &Link{
			Schema: &Schema{
				Properties: map[string]*Schema{
					"int": {
						Type: "integer",
					},
				},
				Type: "object",
			},
		},
		Type: "Int *int",
	},
}

func TestLinkType(t *testing.T) {
	for i, lt := range linkTests {
		kind, _ := lt.Link.GoType()
		if !strings.Contains(kind, lt.Type) {
			t.Errorf("%d: wants %v, got %v", i, lt.Type, kind)
		}
	}
}

var paramsTests = []struct {
	Schema     *Schema
	Link       *Link
	Order      []string
	Parameters map[string]string
}{
	{
		Schema: &Schema{},
		Link: &Link{
			HRef: NewHRef("/destroy/"),
			Rel:  "destroy",
		},
		Parameters: map[string]string{},
	},
	{
		Schema: &Schema{},
		Link: &Link{
			HRef:   NewHRef("/instances/"),
			Rel:    "instances",
			Method: "get",
		},
		Order:      []string{"lr"},
		Parameters: map[string]string{"lr": "*ListRange"},
	},
	{
		Schema: &Schema{},
		Link: &Link{
			Rel:  "update",
			HRef: NewHRef("/update/"),
			Schema: &Schema{
				Type: "string",
			},
		},
		Order:      []string{"o"},
		Parameters: map[string]string{"o": "string"},
	},
	{
		Schema: &Schema{
			Definitions: map[string]*Schema{
				"struct": {
					Definitions: map[string]*Schema{
						"uuid": {
							Type: "string",
						},
					},
				},
			},
		},
		Link: &Link{
			HRef: NewHRef("/results/{(%23%2Fdefinitions%2Fstruct%2Fdefinitions%2Fuuid)}"),
		},
		Order:      []string{"structUUID"},
		Parameters: map[string]string{"structUUID": "string"},
	},
	{
		Schema: &Schema{},
		Link: &Link{
			Title: "update",
			Rel:   "update",
			HRef:  NewHRef("/update/"),
			Schema: &Schema{
				Properties: map[string]*Schema{
					"id": {
						Type: "string",
					},
				},
				Type: "object",
			},
		},
		Order:      []string{"o"},
		Parameters: map[string]string{"o": "LinkUpdateOpts"},
	},
	{
		Schema: &Schema{},
		Link: &Link{
			Title: "update",
			Rel:   "update",
			HRef:  NewHRef("/update/"),
			Schema: &Schema{
				Properties: map[string]*Schema{
					"id": {
						Type: "string",
					},
				},
				Type: []interface{}{"object", "null"},
			},
		},
		Order:      []string{"o"},
		Parameters: map[string]string{"o": "*LinkUpdateOpts"},
	},
	{
		Schema: &Schema{},
		Link: &Link{
			Title:  "list",
			Rel:    "instances",
			Method: "get",
			HRef:   NewHRef("/list/"),
			Schema: &Schema{
				PatternProperties: map[string]*Schema{
					"^\\w+$": {
						Type: "string",
					},
				},
				Type: "object",
			},
		},
		Order:      []string{"o", "lr"},
		Parameters: map[string]string{"o": "map[string]string", "lr": "*ListRange"},
	},
	{
		Schema: &Schema{
			Definitions: map[string]*Schema{
				"struct": {
					Definitions: map[string]*Schema{
						"uuid": {
							Type: "string",
						},
					},
				},
			},
		},
		Link: &Link{
			Title: "check update",
			HRef:  NewHRef("/results/{(%23%2Fdefinitions%2Fstruct%2Fdefinitions%2Fuuid)}"),
			Schema: &Schema{
				Properties: map[string]*Schema{
					"id": {
						Type: "string",
					},
				},
				Type: "object",
			},
		},
		Order:      []string{"structUUID", "o"},
		Parameters: map[string]string{"structUUID": "string", "o": "LinkCheckUpdateOpts"},
	},
}

func TestParameters(t *testing.T) {
	for i, pt := range paramsTests {
		pt.Link.Resolve(pt.Schema, ResolvedSet{})
		order, params := pt.Link.Parameters("link")
		if !reflect.DeepEqual(order, pt.Order) {
			t.Errorf("%d: wants %v, got %v", i, pt.Order, order)
		}
		if !reflect.DeepEqual(params, pt.Parameters) {
			t.Errorf("%d: wants %v, got %v", i, pt.Parameters, params)
		}

	}
}

var valuesTests = []struct {
	Schema *Schema
	Name   string
	Link   *Link
	Values []string
}{
	{
		Schema: &Schema{},
		Name:   "Result",
		Link: &Link{
			Rel: "destroy",
		},
		Values: []string{"error"},
	},
	{
		Schema: &Schema{
			Properties: map[string]*Schema{
				"value": {
					Type: "integer",
				},
			},
			Type: "object",
		},
		Name: "Result",
		Link: &Link{
			Rel: "instances",
		},
		Values: []string{"*Result", "error"},
	},
	{
		Schema: &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"value": {
					Type: "integer",
				},
			},
			Required: []string{"value"},
		},
		Name: "Result",
		Link: &Link{
			Rel: "self",
		},
		Values: []string{"*Result", "error"},
	},
	{
		Schema: &Schema{
			Type:                 "object",
			AdditionalProperties: false,
			PatternProperties: map[string]*Schema{
				"^\\w+$": {
					Type: "string",
				},
			},
		},
		Name: "ConfigVar",
		Link: &Link{
			Rel: "self",
		},
		Values: []string{"ConfigVar", "error"},
	},
	{
		Schema: &Schema{},
		Name:   "Result",
		Link: &Link{
			Rel:   "instances",
			Title: "List",
			TargetSchema: &Schema{
				Type: "object",
				Properties: map[string]*Schema{
					"value": {
						Type: "integer",
					},
				},
			},
		},
		Values: []string{"*ResultListResult", "error"},
	},
	{
		Schema: &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"value": {
					Type: "integer",
				},
			},
			Required: []string{"value"},
		},
		Name: "Result",
		Link: &Link{
			Rel:   "self",
			Title: "Info",
			TargetSchema: &Schema{
				Type: "object",
				Properties: map[string]*Schema{
					"value": {
						Type: "integer",
					},
				},
			},
		},
		Values: []string{"*ResultInfoResult", "error"},
	},
	{
		Schema: &Schema{},
		Name:   "ConfigVar",
		Link: &Link{
			Rel:   "self",
			Title: "Info",
			TargetSchema: &Schema{
				Type: "object",
				Properties: map[string]*Schema{
					"value": {
						Type: "integer",
					},
				},
			},
		},
		Values: []string{"*ConfigVarInfoResult", "error"},
	},
	{
		Schema: &Schema{
			Type:                 "object",
			AdditionalProperties: false,
			PatternProperties: map[string]*Schema{
				"^\\w+$": {
					Type: "string",
				},
			},
		},
		Name: "ConfigVar",
		Link: &Link{
			Rel:   "self",
			Title: "Info",
			TargetSchema: &Schema{
				Type: "object",
				Properties: map[string]*Schema{
					"value": {
						Type: "integer",
					},
				},
			},
		},
		Values: []string{"*ConfigVarInfoResult", "error"},
	},
	{
		Schema: &Schema{},
		Name:   "Result",
		Link: &Link{
			Rel:   "instances",
			Title: "List",
			TargetSchema: &Schema{
				Type: "string",
			},
		},
		Values: []string{"ResultListResult", "error"},
	},
	{
		Schema: &Schema{},
		Name:   "Result",
		Link: &Link{
			Rel:   "instances",
			Title: "List",
			TargetSchema: &Schema{
				Type: []interface{}{"string", "null"},
			},
		},
		Values: []string{"ResultListResult", "error"},
	},
	{
		Schema: &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"value": {
					Type: "integer",
				},
			},
			Required: []string{"value"},
		},
		Name: "Result",
		Link: &Link{
			Rel:   "self",
			Title: "Info",
			TargetSchema: &Schema{
				Type: "string",
			},
		},
		Values: []string{"ResultInfoResult", "error"},
	},
	{
		Schema: &Schema{
			Type:                 "object",
			AdditionalProperties: false,
			PatternProperties: map[string]*Schema{
				"^\\w+$": {
					Type: "string",
				},
			},
		},
		Name: "ConfigVar",
		Link: &Link{
			Rel:   "self",
			Title: "Info",
			TargetSchema: &Schema{
				Type: "string",
			},
		},
		Values: []string{"ConfigVarInfoResult", "error"},
	},
}

func TestValues(t *testing.T) {
	for i, vt := range valuesTests {
		values := vt.Schema.Values(vt.Name, vt.Link)
		if !reflect.DeepEqual(values, vt.Values) {
			t.Errorf("%d: wants %v, got %v", i, vt.Values, values)
		}
	}
}

var linkTitleTests = []struct {
	Schema   *Schema
	Expected bool
}{
	{
		Schema: &Schema{
			Title: "Selfreferencing",
			Type:  "object",
			Links: []*Link{
				{
					Title: "Create",
				},
				{
					Title: "Create",
				},
			},
		},
		Expected: true,
	},
	{
		Schema: &Schema{
			Title: "Selfreferencing",
			Type:  "object",
			Links: []*Link{
				{
					Title: "update",
				},
				{
					Title: "Update",
				},
			},
		},
		Expected: true,
	},
	{
		Schema: &Schema{
			Title: "Selfreferencing",
			Type:  "object",
			Links: []*Link{
				{
					Title: "Create",
				},
				{
					Title: "Delete",
				},
			},
		},
		Expected: false,
	},
}

func TestLinkTitles(t *testing.T) {
	for i, lt := range linkTitleTests {
		resp := lt.Schema.CheckForDuplicateTitles()
		if resp != lt.Expected {
			t.Errorf("%d: wants %v, got %v", i, lt.Expected, resp)
		}
	}
}
