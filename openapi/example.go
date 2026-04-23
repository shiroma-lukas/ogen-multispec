package openapi

import (
	"github.com/shiroma-lukas/ogen-multispec/jsonschema"
	"github.com/shiroma-lukas/ogen-multispec/location"
)

// Example is an OpenAPI Example.
type Example struct {
	Ref Ref

	Summary       string
	Description   string
	Value         jsonschema.Example
	ExternalValue string

	location.Pointer `json:"-" yaml:"-"`
}
