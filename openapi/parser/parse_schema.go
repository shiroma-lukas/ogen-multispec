package parser

import (
	"github.com/go-faster/errors"

	"github.com/shiroma-lukas/ogen-multispec"
	"github.com/shiroma-lukas/ogen-multispec/jsonpointer"
	"github.com/shiroma-lukas/ogen-multispec/jsonschema"
)

func (p *parser) parseSchema(schema *ogen.Schema, ctx *jsonpointer.ResolveCtx) (*jsonschema.Schema, error) {
	s, err := p.schemaParser.Parse(schema.ToJSONSchema(), ctx)
	if err != nil {
		return nil, errors.Wrap(err, "parse schema")
	}
	return s, nil
}
