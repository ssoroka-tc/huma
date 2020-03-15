package huma

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Schema represents a JSON Schema which can be generated from Go structs
type Schema struct {
	Type        string             `json:"type,omitempty"`
	Description string             `json:"description,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Format      string             `json:"format,omitempty"`
	Enum        []interface{}      `json:"enum,omitempty"`
	Default     interface{}        `json:"default,omitempty"`
	Example     interface{}        `json:"example,omitempty"`
	Minimum     *int               `json:"minimum,omitempty"`
	Maximum     *int               `json:"maximum,omitempty"`
}

// GenerateSchema creates a JSON schema for a Go type. Struct field tags
// can be used to provide additional metadata such as descriptions and
// validation.
func GenerateSchema(t reflect.Type) (*Schema, error) {
	schema := &Schema{}

	switch t.Kind() {
	case reflect.Struct:
		// TODO: support time and URI types
		properties := make(map[string]*Schema)
		required := make([]string, 0)
		schema.Type = "object"

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)

			jsonTags := strings.Split(f.Tag.Get("json"), ",")

			name := f.Name
			if len(jsonTags) > 0 {
				name = jsonTags[0]
			}

			s, err := GenerateSchema(f.Type)
			if err != nil {
				return nil, err
			}
			properties[name] = s

			if d, ok := f.Tag.Lookup("description"); ok {
				s.Description = d
			}

			if e, ok := f.Tag.Lookup("enum"); ok {
				s.Enum = []interface{}{}
				for _, v := range strings.Split(e, ",") {
					// TODO: convert to correct type
					s.Enum = append(s.Enum, v)
				}
			}

			if d, ok := f.Tag.Lookup("minimum"); ok {
				min, err := strconv.Atoi(d)
				if err != nil {
					return nil, err
				}
				s.Minimum = &min
			}

			if d, ok := f.Tag.Lookup("maximum"); ok {
				max, err := strconv.Atoi(d)
				if err != nil {
					return nil, err
				}
				s.Maximum = &max
			}

			if e, ok := f.Tag.Lookup("example"); ok {
				if s.Type == "string" {
					s.Example = e
				} else {
					var v interface{}
					if err := json.Unmarshal([]byte(e), &v); err != nil {
						return nil, err
					}
					s.Example = v
				}
			}

			optional := false
			for _, tag := range jsonTags[1:] {
				if tag == "omitempty" {
					optional = true
				}
			}
			if !optional {
				required = append(required, name)
			}
		}

		if len(properties) > 0 {
			schema.Properties = properties
		}

		if len(required) > 0 {
			schema.Required = required
		}

	case reflect.Map:
		// pass
	case reflect.Slice, reflect.Array:
		schema.Type = "array"
		s, err := GenerateSchema(t.Elem())
		if err != nil {
			return nil, err
		}
		schema.Items = s
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &Schema{
			Type: "integer",
		}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// Unsigned integers can't be negative.
		min := 0
		return &Schema{
			Type:    "integer",
			Minimum: &min,
		}, nil
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}, nil
	case reflect.Bool:
		return &Schema{Type: "boolean"}, nil
	case reflect.String:
		return &Schema{Type: "string"}, nil
	case reflect.Ptr:
		return GenerateSchema(t.Elem())
	default:
		return nil, fmt.Errorf("unsupported type %s from %s", t.Kind(), t)
	}

	return schema, nil
}
