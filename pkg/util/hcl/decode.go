//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package hcl

import (
	"encoding"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/defiweb/go-anymapper"
	"github.com/defiweb/go-eth/types"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type stringize interface {
	String() string
}

// Marshaler marshals cty.Value.
type Marshaler interface {
	MarshalHCL() (cty.Value, error)
}

// Unmarshaler unmarshals a value from cty.Value.
type Unmarshaler interface {
	UnmarshalHCL(cty.Value) error
}

// PreDecodeAttribute is called before an attribute is decoded.
type PreDecodeAttribute interface {
	PreDecodeAttribute(*hcl.EvalContext, *hcl.Attribute) hcl.Diagnostics
}

// PostDecodeAttribute is called after an attribute is decoded.
type PostDecodeAttribute interface {
	PostDecodeAttribute(*hcl.EvalContext, *hcl.Attribute) hcl.Diagnostics
}

// PreDecodeBlock is called before a block is decoded.
type PreDecodeBlock interface {
	PreDecodeBlock(*hcl.EvalContext, *hcl.BodySchema, *hcl.Block, *hcl.BodyContent) hcl.Diagnostics
}

// PostDecodeBlock is called after a block is decoded.
type PostDecodeBlock interface {
	PostDecodeBlock(*hcl.EvalContext, *hcl.BodySchema, *hcl.Block, *hcl.BodyContent) hcl.Diagnostics
}

// Decode decodes the given HCL body into the given value.
// The value must be a pointer to a struct.
//
// It works similar to the Decode function from the gohcl package, but it has
// improved support for decoding values into maps and slices. It also supports
// encoding.TextUnmarshaler interface and additional Unmarshaler interface that
// allows to decode values into custom types directly from cty.Value.
//
// Only fields with the "hcl" tag will be decoded. The tag contains the name of
// the attribute or block and additional options separated by comma.
//
// Supported options are
//   - attr - the field is an attribute, it will be decoded from the HCL body
//     attributes. It is the default tag and can be omitted.
//   - label - a label of the parent block. Multiple labels can be defined.
//   - optional - the field is optional, if it is not present in the HCL body,
//     the field will be left as zero value.
//   - ignore - the field will be ignored but still be a part of the schema.
//   - block - the field is a block, it will be decoded from the HCL blocks.
//     The field must be a struct, slice of structs or a map of structs.
//   - remain - the field is populated with the remaining HCL body. The field
//     must be hcl.Body.
//   - body - the field is populated with the HCL body. The field must
//     be hcl.BodyContent.
//   - content - the field is populated with the HCL body content. The field
//     must be hcl.BodyContent.
//   - schema - the field is populated with the HCL body schema. The field must
//     be hcl.BodySchema.
//   - range - the block range. The field must be hcl.Range.
//
// If name is omitted, the field name will be used.
func Decode(ctx *hcl.EvalContext, body hcl.Body, val any) hcl.Diagnostics {
	return decodeSingleBlock(ctx, &hcl.Block{Body: body}, reflect.ValueOf(val))
}

// DecodeBlock decodes the given HCL block into the given value.
func DecodeBlock(ctx *hcl.EvalContext, block *hcl.Block, val any) hcl.Diagnostics {
	return decodeSingleBlock(ctx, block, reflect.ValueOf(val))
}

// DecodeExpression decodes the given HCL expression into the given value.
func DecodeExpression(ctx *hcl.EvalContext, expr hcl.Expression, val any) hcl.Diagnostics {
	ctyVal, diags := expr.Value(ctx)
	if diags.HasErrors() {
		return diags
	}
	if err := mapper.Map(ctyVal, val); err != nil {
		return hcl.Diagnostics{{
			Severity: hcl.DiagError,
			Summary:  "Decode error",
			Detail:   err.Error(),
		}}
	}
	return nil
}

// decodeSingleBlock decodes a single block into the given value. A value must
// be a pointer to a struct.
//
//nolint:funlen,gocyclo
func decodeSingleBlock(ctx *hcl.EvalContext, block *hcl.Block, ptrVal reflect.Value) hcl.Diagnostics {
	// Dereference the pointer to get the struct value.
	val := derefValue(ptrVal)

	if !val.CanSet() || val.Kind() != reflect.Struct {
		return hcl.Diagnostics{{
			Severity: hcl.DiagError,
			Summary:  "Decode error",
			Detail:   "Value must be a pointer to a struct",
			Subject:  &block.DefRange,
		}}
	}

	// Build the schema for the given struct.
	meta, diags := getStructMeta(val.Type())
	if diags.HasErrors() {
		return diags
	}

	// Decode the body.
	var (
		content *hcl.BodyContent
		remain  hcl.Body
	)
	if meta.Remain != nil {
		// Decode the body using the schema, and get the remain body.
		content, remain, diags = block.Body.PartialContent(meta.BodySchema)
		if diags.HasErrors() {
			return diags
		}

		// Set remain field.
		if remain != nil {
			val.FieldByIndex(meta.Remain.Reflect.Index).Set(reflect.ValueOf(remain))
		}
	} else {
		// Decode the body using the schema. If there are remain parts, it
		// will return an error.
		content, diags = block.Body.Content(meta.BodySchema)
		if diags.HasErrors() {
			return diags
		}
	}

	// Set "body" field.
	if meta.Body != nil {
		val.FieldByIndex(meta.Body.Reflect.Index).Set(reflect.ValueOf(block.Body))
	}

	// Set "content" field.
	if meta.Content != nil {
		val.FieldByIndex(meta.Content.Reflect.Index).Set(reflect.ValueOf(content).Elem())
	}

	// Set "schema" field.
	if meta.Schema != nil {
		val.FieldByIndex(meta.Schema.Reflect.Index).Set(reflect.ValueOf(meta.BodySchema).Elem())
	}

	// Set "range" field.
	if meta.Range != nil {
		val.FieldByIndex(meta.Range.Reflect.Index).Set(reflect.ValueOf(block.DefRange))
	}

	// Pre decode hook.
	if n, ok := ptrVal.Interface().(PreDecodeBlock); ok {
		diags := n.PreDecodeBlock(ctx, meta.BodySchema, block, content)
		if diags.HasErrors() {
			return diags
		}
	}

	// Check for missing or extraneous blocks.
	for _, field := range meta.Blocks {
		if field.Ignore || field.Multiple {
			continue
		}
		blocksOfType := content.Blocks.OfType(field.Name)
		if !field.Optional && len(blocksOfType) == 0 {
			return hcl.Diagnostics{{
				Severity: hcl.DiagError,
				Summary:  "Decode error",
				Detail:   fmt.Sprintf("Missing block %q", field.Name),
				Subject:  &block.DefRange,
			}}
		}
		if len(blocksOfType) > 1 {
			var diags hcl.Diagnostics
			for _, block := range blocksOfType {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Decode error",
					Detail:   fmt.Sprintf("Extraneous block %q, only one is allowed", field.Name),
					Subject:  &block.DefRange,
				})
			}
			return diags
		}
	}

	// Decode labels.
	for i, label := range block.Labels {
		fieldRef := val.FieldByIndex(meta.Labels[i].Reflect.Index)
		labelRef := reflect.ValueOf(cty.StringVal(label))
		if err := mapper.MapRefl(labelRef, fieldRef); err != nil {
			return hcl.Diagnostics{{
				Severity: hcl.DiagError,
				Summary:  "Decode error",
				Detail:   fmt.Sprintf("Cannot decode label %q: %s", label, err),
				Subject:  &block.DefRange,
			}}
		}
	}

	// Decode blocks.
	for _, block := range content.Blocks {
		field, ok := meta.Blocks.get(block.Type)
		if !ok {
			continue
		}
		if field.Ignore {
			continue
		}
		diags := decodeBlock(ctx, block, val.FieldByIndex(field.Reflect.Index))
		if diags.HasErrors() {
			return diags
		}
	}

	// Decode attributes.
	for _, attr := range content.Attributes {
		field, ok := meta.Attrs.get(attr.Name)
		if !ok {
			continue
		}
		if field.Ignore {
			continue
		}
		diags := decodeAttribute(ctx, attr, val.FieldByIndex(field.Reflect.Index))
		if diags.HasErrors() {
			return diags
		}
	}

	// Post decode hook.
	if n, ok := ptrVal.Interface().(PostDecodeBlock); ok {
		diags := n.PostDecodeBlock(ctx, meta.BodySchema, block, content)
		if diags.HasErrors() {
			return diags
		}
	}

	return nil
}

// decodeBlock decodes a block into the given value.
//   - If a value is a slice, it will append a new element to the slice.
//   - If a block is a map, it will append a new element to the map and label
//     will be used as a key. Block must have only one label.
//   - If a value is a pointer or interface, it will dereference the value.
//   - If a value is a nil pointer, it will be allocated.
func decodeBlock(ctx *hcl.EvalContext, block *hcl.Block, val reflect.Value) hcl.Diagnostics {
	switch val.Kind() {
	case reflect.Struct:
		return decodeSingleBlock(ctx, block, val.Addr())
	case reflect.Ptr:
		if val.IsNil() {
			val.Set(reflect.New(val.Type().Elem()))
		}
		return decodeBlock(ctx, block, val.Elem())
	case reflect.Interface:
		return decodeBlock(ctx, block, val.Elem())
	case reflect.Slice:
		if val.IsNil() {
			val.Set(reflect.MakeSlice(val.Type(), 0, 1))
		}
		elem := reflect.New(val.Type().Elem())
		if diags := decodeBlock(ctx, block, elem); diags.HasErrors() {
			return diags
		}
		val.Set(reflect.Append(val, elem.Elem()))
		return nil
	case reflect.Map:
		if len(block.Labels) != 1 {
			return hcl.Diagnostics{{
				Severity: hcl.DiagError,
				Summary:  "Decode error",
				Detail: fmt.Sprintf(
					"Cannot decode block %q into map: block must have only one label",
					block.Type,
				),
				Subject: block.DefRange.Ptr(),
			}}
		}
		if val.IsNil() {
			val.Set(reflect.MakeMap(val.Type()))
		}
		key := reflect.ValueOf(block.Labels[0])
		if val.MapIndex(key).IsValid() {
			return hcl.Diagnostics{{
				Severity: hcl.DiagError,
				Summary:  "Decode error",
				Detail: fmt.Sprintf(
					"Cannot decode block %q into map: duplicate label %q",
					block.Type, block.Labels[0],
				),
				Subject: block.DefRange.Ptr(),
			}}
		}
		elem := reflect.New(val.Type().Elem())
		if diags := decodeBlock(ctx, block, elem); diags.HasErrors() {
			return diags
		}
		val.SetMapIndex(key, elem.Elem())
		return nil
	}
	return hcl.Diagnostics{{
		Severity: hcl.DiagError,
		Summary:  "Decode error",
		Detail:   fmt.Sprintf("Cannot decode block %q into %s", block.Type, val.Type()),
		Subject:  block.DefRange.Ptr(),
	}}
}

// decodeAttribute decodes a single attribute into the given value.
func decodeAttribute(ctx *hcl.EvalContext, attr *hcl.Attribute, val reflect.Value) hcl.Diagnostics {
	// Pre decode hook.
	if n, ok := val.Interface().(PreDecodeAttribute); ok {
		diags := n.PreDecodeAttribute(ctx, attr)
		if diags.HasErrors() {
			return diags
		}
	}

	// Evaluate the expression.
	ctyVal, diags := attr.Expr.Value(ctx)
	if diags.HasErrors() {
		return diags
	}

	// Map the value.
	if err := mapper.MapRefl(reflect.ValueOf(ctyVal), val); err != nil {
		return hcl.Diagnostics{{
			Severity: hcl.DiagError,
			Summary:  "Decode error",
			Detail:   err.Error(),
			Subject:  &attr.Range,
		}}
	}

	// Post decode hook.
	if n, ok := val.Interface().(PostDecodeAttribute); ok {
		diags := n.PostDecodeAttribute(ctx, attr)
		if diags.HasErrors() {
			return diags
		}
	}

	return nil
}

// getStructMeta parses the tags of a struct and returns a structMeta.
//
//nolint:funlen,gocyclo
func getStructMeta(s reflect.Type) (*structMeta, hcl.Diagnostics) {
	meta := &structMeta{BodySchema: &hcl.BodySchema{}}
	for _, fieldRef := range reflect.VisibleFields(s) {
		fieldMeta, diags := getStructFieldMeta(fieldRef)
		if diags.HasErrors() {
			return nil, diags
		}
		if !fieldMeta.Tagged {
			continue
		}
		switch fieldMeta.Type {
		case fieldAttr:
			if !meta.Attrs.add(fieldMeta) {
				return nil, hcl.Diagnostics{{
					Severity: hcl.DiagError,
					Summary:  "Schema error",
					Detail: fmt.Sprintf(
						"Duplicate attribute name %q in struct %s",
						fieldMeta.Name,
						s,
					),
				}}
			}
			meta.BodySchema.Attributes = append(meta.BodySchema.Attributes, hcl.AttributeSchema{
				Name:     fieldMeta.Name,
				Required: !fieldMeta.Optional,
			})
		case fieldLabel:
			if !meta.Labels.add(fieldMeta) {
				return nil, hcl.Diagnostics{{
					Severity: hcl.DiagError,
					Summary:  "Schema error",
					Detail: fmt.Sprintf(
						"Duplicate label name %q in struct %s",
						fieldMeta.Name,
						s,
					),
				}}
			}
		case fieldBlock:
			// Extract the labels from the struct.
			var labels []string
			for _, subFieldRef := range reflect.VisibleFields(fieldMeta.StructReflect) {
				subFieldMeta, diags := getStructFieldMeta(subFieldRef)
				if diags.HasErrors() {
					return nil, diags
				}
				if !subFieldMeta.Tagged {
					continue
				}
				if subFieldMeta.Type != fieldLabel {
					continue
				}
				labels = append(labels, subFieldMeta.Name)
			}

			// Add the block to the schema.
			if !meta.Blocks.add(fieldMeta) {
				return nil, hcl.Diagnostics{{
					Severity: hcl.DiagError,
					Summary:  "Schema error",
					Detail: fmt.Sprintf(
						"Duplicate block name %q in struct %s",
						fieldMeta.Name,
						s,
					),
				}}
			}
			meta.BodySchema.Blocks = append(meta.BodySchema.Blocks, hcl.BlockHeaderSchema{
				Type:       fieldMeta.Name,
				LabelNames: labels,
			})
		case fieldRemain:
			meta.Remain = &fieldMeta
		case fieldBody:
			meta.Body = &fieldMeta
		case fieldContent:
			meta.Content = &fieldMeta
		case fieldSchema:
			meta.Schema = &fieldMeta
		case fieldRange:
			meta.Range = &fieldMeta
		default:
			// Should never happen.
			return nil, hcl.Diagnostics{{
				Severity: hcl.DiagError,
				Summary:  "Schema error",
				Detail:   fmt.Sprintf("Unsupported field type %q", fieldMeta.Type),
			}}
		}
	}
	return meta, nil
}

// getStructFieldMeta parses the hcl tag of a struct field and returns a
// structFieldMeta.
//
//nolint:funlen,gocyclo
func getStructFieldMeta(field reflect.StructField) (structFieldMeta, hcl.Diagnostics) {
	var (
		tag string
		sfm = structFieldMeta{Reflect: field}
	)

	// Parse the tag.
	tag, sfm.Tagged = field.Tag.Lookup("hcl")
	if !sfm.Tagged {
		return sfm, nil
	}

	// Tagged fields must be exported.
	if !field.IsExported() {
		return sfm, hcl.Diagnostics{{
			Severity: hcl.DiagError,
			Summary:  "Schema error",
			Detail: fmt.Sprintf(
				"Field %q is not exported but has an hcl tag",
				field.Name,
			),
		}}
	}

	parts := strings.Split(tag, ",")
	sfm.Name = parts[0]
	if len(sfm.Name) == 0 {
		sfm.Name = field.Name
	}
	for _, part := range parts[1:] {
		switch part {
		case "attr":
			sfm.Type = fieldAttr
		case "label":
			sfm.Type = fieldLabel
		case "block":
			sfm.Type = fieldBlock
		case "remain":
			sfm.Type = fieldRemain
		case "body":
			sfm.Type = fieldBody
		case "content":
			sfm.Type = fieldContent
		case "schema":
			sfm.Type = fieldSchema
		case "range":
			sfm.Type = fieldRange
		case "optional":
			sfm.Optional = true
		case "ignore":
			sfm.Ignore = true
		default:
			return sfm, hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid tag",
				Detail:   fmt.Sprintf("Invalid tag %q", part),
			}}
		}
	}

	// Find a struct type for this block.
	// A field may be also a slice or map of structs, in which case the struct
	// type must be extracted.
	if sfm.Type == fieldBlock {
		typ := derefType(sfm.Reflect.Type)
		if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Map {
			typ = derefType(typ.Elem())
			// If it is a slice or map, the block can be repeated.
			sfm.Multiple = true
		}
		sfm.StructReflect = typ
		if typ.Kind() != reflect.Struct {
			return sfm, hcl.Diagnostics{{
				Severity: hcl.DiagError,
				Summary:  "Schema error",
				Detail: fmt.Sprintf(
					"Cannot use block tag on field %q of type %s, only structs, slices of structs, and maps of structs are supported",
					sfm.Name,
					sfm.Reflect.Type,
				),
			}}
		}
	}

	// Validate the tag.
	if sfm.Type != fieldAttr && sfm.Type != fieldBlock && sfm.Optional {
		return sfm, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid tag",
			Detail:   "A optional tag can only be used with attributes and blocks",
		}}
	}
	if sfm.Type == fieldBlock && sfm.Multiple && sfm.Optional {
		return sfm, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid tag",
			Detail:   "A optional tag cannot be used with a block that can be repeated",
		}}
	}
	if sfm.Type == fieldRemain && field.Type != bodyTy {
		return sfm, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid tag",
			Detail:   "A remain tag must be used with a field of type hcl.Body",
		}}
	}
	if sfm.Type == fieldBody && field.Type != bodyTy {
		return sfm, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid tag",
			Detail:   "A body tag must be used with a field of type hcl.Body",
		}}
	}
	if sfm.Type == fieldContent && field.Type != bodyContentTy {
		return sfm, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid tag",
			Detail:   "A body tag must be used with a field of type hcl.BodyContent",
		}}
	}
	if sfm.Type == fieldSchema && field.Type != bodySchemaTy {
		return sfm, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid tag",
			Detail:   "A body tag must be used with a field of type hcl.BodySchema",
		}}
	}
	if sfm.Type == fieldRange && field.Type != rangeTy {
		return sfm, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid tag",
			Detail:   "A range tag must be used with a field of type hcl.Range",
		}}
	}

	return sfm, nil
}

const (
	fieldAttr = iota
	fieldBlock
	fieldLabel
	fieldRemain
	fieldBody
	fieldContent
	fieldSchema
	fieldRange
)

// structMeta contains the information about fields of a struct to which
// HCL is decoded.
type structMeta struct {
	BodySchema *hcl.BodySchema  // BodySchema for the struct.
	Labels     structFieldsMeta // List of fields that are labels.
	Attrs      structFieldsMeta // List of fields that are attributes.
	Blocks     structFieldsMeta // List of fields that are blocks.
	Remain     *structFieldMeta // Field that is tagged with "remain".
	Body       *structFieldMeta // Field that is tagged with "body".
	Content    *structFieldMeta // Field that is tagged with "content".
	Schema     *structFieldMeta // Field that is tagged with "schema".
	Range      *structFieldMeta // Field that is tagged with "range".
}

// structFieldMeta contains the information about a struct field.
type structFieldMeta struct {
	Name          string              // Name of the field as defined in the hcl tag.
	Tagged        bool                // True if the field has a hcl tag.
	Type          int                 // Type of the field, one of the field* constants.
	Optional      bool                // True if the field is optional.
	Multiple      bool                // True if the field is a block and can be repeated.
	Ignore        bool                // True if the field is ignored.
	Reflect       reflect.StructField // The reflect.StructField of the field.
	StructReflect reflect.Type        // The reflect.Type of the struct to which block is decoded (if field is a block).
}

type structFieldsMeta []structFieldMeta

// add adds a struct field. It returns false if the field with the same name
// already exists.
func (s *structFieldsMeta) add(field structFieldMeta) bool {
	if s.has(field.Name) {
		return false
	}
	*s = append(*s, field)
	return true
}

// get returns the struct field with the given name. It returns false if the
// field does not exist.
func (s *structFieldsMeta) get(name string) (structFieldMeta, bool) {
	for _, f := range *s {
		if f.Name == name {
			return f, true
		}
	}
	return structFieldMeta{}, false
}

// has returns true if the struct field with the given name exists.
func (s *structFieldsMeta) has(name string) bool {
	for _, f := range *s {
		if f.Name == name {
			return true
		}
	}
	return false
}

var mapper *anymapper.Mapper

var (
	bodyTy        = reflect.TypeOf((*hcl.Body)(nil)).Elem()
	bodyContentTy = reflect.TypeOf((*hcl.BodyContent)(nil)).Elem()
	bodySchemaTy  = reflect.TypeOf((*hcl.BodySchema)(nil)).Elem()
	rangeTy       = reflect.TypeOf((*hcl.Range)(nil)).Elem()
	ctyValTy      = reflect.TypeOf((*cty.Value)(nil)).Elem()
	bigIntTy      = reflect.TypeOf((*big.Int)(nil)).Elem()
	bigIntPtrTy   = reflect.TypeOf((*big.Int)(nil))
	bigFloatTy    = reflect.TypeOf((*big.Float)(nil)).Elem()
	addressTy     = reflect.TypeOf((*types.Address)(nil)).Elem()
	stringTy      = reflect.TypeOf("")
	anyTy         = reflect.TypeOf((*any)(nil)).Elem()
)

// derefType dereferences the given type until it is not a pointer or an
// interface.
func derefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface {
		t = t.Elem()
	}
	return t
}

// derefValue dereferences the given value until it is not a pointer or an
// interface. If the value is a nil pointer, it is initialized.
func derefValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.Kind() == reflect.Ptr && v.IsNil() && v.CanSet() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

func ctyMapper(m *anymapper.Mapper, src, dst reflect.Type) anymapper.MapFunc {
	if src == ctyValTy {
		return fromCtyMapper(m, src, dst)
	}
	if dst == ctyValTy {
		return toCtyMapper(m, src, dst)
	}
	return nil
}

// ctyMapper is a mapping function that maps cty.Value to other types.
//
//nolint:funlen,gocyclo
func fromCtyMapper(_ *anymapper.Mapper, src, dst reflect.Type) anymapper.MapFunc {
	if src != ctyValTy {
		return nil
	}

	// cty.Value -> any
	// To be able to reuse the existing mapping functions defined below, we
	// create an auxiliary variable based on the cty.Value type, and we use
	// that variable as the destination.
	if dst == anyTy {
		return func(m *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
			typ := src.Interface().(cty.Value).Type()
			switch {
			case typ == cty.String:
				var aux string
				if err := m.MapRefl(src, reflect.ValueOf(&aux)); err != nil {
					return err
				}
				dst.Set(reflect.ValueOf(aux))
			case typ == cty.Number:
				var aux float64
				if err := m.MapRefl(src, reflect.ValueOf(&aux)); err != nil {
					return err
				}
				dst.Set(reflect.ValueOf(aux))
			case typ == cty.Bool:
				var aux bool
				if err := m.MapRefl(src, reflect.ValueOf(&aux)); err != nil {
					return err
				}
				dst.Set(reflect.ValueOf(aux))
			case typ.IsListType() || typ.IsSetType() || typ.IsTupleType():
				var aux []any
				if err := m.MapRefl(src, reflect.ValueOf(&aux)); err != nil {
					return err
				}
				dst.Set(reflect.ValueOf(aux))
			case typ.IsMapType() || typ.IsObjectType():
				var aux map[string]any
				if err := m.MapRefl(src, reflect.ValueOf(&aux)); err != nil {
					return err
				}
				dst.Set(reflect.ValueOf(aux))
			case typ == cty.DynamicPseudoType:
				dst.Set(reflect.Zero(dst.Type()))
			default:
				dst.Set(src)
			}
			return nil
		}
	}

	// cty.Value -> cty.Value
	if dst == ctyValTy {
		return func(m *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
			dst.Set(src)
			return nil
		}
	}

	// cty.Value -> big.Int
	if dst == bigIntTy {
		return func(m *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
			val := src.Interface().(cty.Value)
			if val.Type() != cty.Number {
				return fmt.Errorf("cannot decode %s into big.Int", val.Type().FriendlyName())
			}
			if !val.AsBigFloat().IsInt() {
				return fmt.Errorf("cannot decode a float number into big.Int")
			}
			bi, acc := val.AsBigFloat().Int(nil)
			if acc != big.Exact {
				return fmt.Errorf("cannot decode a float number into big.Int")
			}
			dst.Set(reflect.ValueOf(bi).Elem())
			return nil
		}
	}

	// cty.Value -> big.Float
	if dst == bigFloatTy {
		return func(m *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
			val := src.Interface().(cty.Value)
			if val.Type() != cty.Number {
				return fmt.Errorf("cannot decode %s into big.Float", val.Type().FriendlyName())
			}
			dst.Set(reflect.ValueOf(val.AsBigFloat()).Elem())
			return nil
		}
	}

	// cty.Value -> Unmarshaler
	// cty.Value -> TextUnmarshaler
	// cty.Value -> string
	// cty.Value -> bool
	// cty.Value -> int*
	// cty.Value -> uint*
	// cty.Value -> float*
	// cty.Value -> slice
	// cty.Value -> map
	return func(m *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
		ctyVal := src.Interface().(cty.Value)

		// Try to use unmarshaler interfaces.
		if dst.CanAddr() {
			if u, ok := dst.Addr().Interface().(Unmarshaler); ok {
				return u.UnmarshalHCL(ctyVal)
			}
			if u, ok := dst.Addr().Interface().(encoding.TextUnmarshaler); ok && ctyVal.Type() == cty.String {
				return u.UnmarshalText([]byte(ctyVal.AsString()))
			}
		}

		// Try to map the cty.Value to the basic types.
		switch dst.Kind() {
		case reflect.String:
			if ctyVal.Type() != cty.String {
				return fmt.Errorf(
					"cannot decode %s type into a string",
					ctyVal.Type().FriendlyName(),
				)
			}
			dst.SetString(ctyVal.AsString())
		case reflect.Bool:
			if ctyVal.Type() != cty.Bool {
				return fmt.Errorf(
					"cannot decode %s type into a bool",
					ctyVal.Type().FriendlyName(),
				)
			}
			dst.SetBool(ctyVal.True())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if ctyVal.Type() != cty.Number {
				return fmt.Errorf(
					"cannot decode %s type into a %s type",
					ctyVal.Type().FriendlyName(), dst.Kind(),
				)
			}
			if !ctyVal.AsBigFloat().IsInt() {
				return fmt.Errorf(
					"cannot decode %s type into a %s type: not an integer",
					ctyVal.Type().FriendlyName(),
					dst.Kind(),
				)
			}
			i64, acc := ctyVal.AsBigFloat().Int64()
			if acc != big.Exact {
				return fmt.Errorf(
					"cannot decode %s type into a %s type: too large",
					ctyVal.Type().FriendlyName(),
					dst.Kind(),
				)
			}
			return m.MapReflContext(m.Context.WithStrictTypes(false), reflect.ValueOf(i64), dst)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if ctyVal.Type() != cty.Number {
				return fmt.Errorf(
					"cannot decode %s type into a %s type",
					ctyVal.Type().FriendlyName(),
					dst.Kind(),
				)
			}
			if !ctyVal.AsBigFloat().IsInt() {
				return fmt.Errorf(
					"cannot decode %s type into a %s type: not an integer",
					ctyVal.Type().FriendlyName(),
					dst.Kind(),
				)
			}
			u64, acc := ctyVal.AsBigFloat().Uint64()
			if acc != big.Exact {
				return fmt.Errorf(
					"cannot decode %s type into a %s type: too large",
					ctyVal.Type().FriendlyName(),
					dst.Kind(),
				)
			}
			return m.MapReflContext(m.Context.WithStrictTypes(false), reflect.ValueOf(u64), dst)
		case reflect.Float32, reflect.Float64:
			if ctyVal.Type() != cty.Number {
				return fmt.Errorf(
					"cannot decode %s type into a %s type",
					ctyVal.Type().FriendlyName(),
					dst.Kind(),
				)
			}
			return m.MapReflContext(m.Context.WithStrictTypes(false), reflect.ValueOf(ctyVal.AsBigFloat()), dst)
		case reflect.Slice:
			if !ctyVal.Type().IsListType() && !ctyVal.Type().IsSetType() && !ctyVal.Type().IsTupleType() {
				return fmt.Errorf(
					"cannot decode %s type into a slice",
					ctyVal.Type().FriendlyName(),
				)
			}
			dstSlice := reflect.MakeSlice(dst.Type(), 0, ctyVal.LengthInt())
			for it := ctyVal.ElementIterator(); it.Next(); {
				_, v := it.Element()
				elem := reflect.New(dst.Type().Elem())
				if err := m.MapRefl(reflect.ValueOf(v), elem); err != nil {
					return err
				}
				dstSlice = reflect.Append(dstSlice, elem.Elem())
			}
			dst.Set(dstSlice)
		case reflect.Map:
			if !ctyVal.Type().IsMapType() && !ctyVal.Type().IsObjectType() {
				return fmt.Errorf(
					"cannot decode %s type into a map",
					ctyVal.Type().FriendlyName(),
				)
			}
			dstMap := reflect.MakeMap(dst.Type())
			for it := ctyVal.ElementIterator(); it.Next(); {
				k, v := it.Element()
				key := reflect.New(dst.Type().Key())
				if err := m.MapRefl(reflect.ValueOf(k), key); err != nil {
					return err
				}
				val := reflect.New(dst.Type().Elem())
				if err := m.MapRefl(reflect.ValueOf(v), val); err != nil {
					return err
				}
				dstMap.SetMapIndex(key.Elem(), val.Elem())
			}
			dst.Set(dstMap)
		default:
			return fmt.Errorf("unsupported type %s", dst.Type())
		}
		return nil
	}
}

func toCtyMapper(_ *anymapper.Mapper, src, dst reflect.Type) anymapper.MapFunc { //nolint:gocyclo
	if dst != ctyValTy {
		return nil
	}

	if src == addressTy {
		return func(m *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
			ctyVal := cty.StringVal(src.String())
			dst.Set(reflect.ValueOf(ctyVal))
			return nil
		}
	}

	// cty.Value -> cty.Value
	if src == ctyValTy {
		return func(m *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
			dst.Set(src)
			return nil
		}
	}

	// big.Int -> cty.Value
	if src == bigIntTy {
		return func(_ *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
			val, ok := src.Interface().(big.Int)
			if !ok {
				return fmt.Errorf("cannot encode from big.Int %s", src.Type().Name())
			}
			dst.Set(reflect.ValueOf(cty.NumberUIntVal(val.Uint64())))
			return nil
		}
	}

	// *big.Int -> cty.Value
	if src == bigIntPtrTy {
		return func(_ *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
			val, ok := src.Interface().(*big.Int)
			if !ok {
				return fmt.Errorf("cannot encode from *big.Int %s", src.Type().Name())
			}
			dst.Set(reflect.ValueOf(cty.NumberUIntVal(val.Uint64())))
			return nil
		}
	}

	// big.Float -> cty.Value
	if src == bigFloatTy {
		return func(_ *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
			val, ok := src.Interface().(big.Float)
			if !ok {
				return fmt.Errorf("cannot encode from big.Float %s", src.Type().Name())
			}
			valFloat, _ := val.Float64()
			dst.Set(reflect.ValueOf(cty.NumberFloatVal(valFloat)))
			return nil
		}
	}

	// string -> cty.Value
	// bool -> cty.Value
	// int -> cty.Value
	// uint -> cty.Value
	// flaot -> cty.Value
	// slice -> cty.Value
	// map -> cty.Value
	return func(m *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
		// Try to use marshaler interfaces.
		if src.CanAddr() {
			if u, ok := src.Addr().Interface().(Marshaler); ok {
				ctyVal, err := u.MarshalHCL()
				if err != nil {
					return err
				}
				dst.Set(reflect.ValueOf(ctyVal))
				return nil
			}
			if u, ok := src.Addr().Interface().(encoding.TextMarshaler); ok && src.Type() == stringTy {
				text, err := u.MarshalText()
				if err != nil {
					return err
				}
				ctyVal := cty.StringVal(string(text))
				dst.Set(reflect.ValueOf(ctyVal))
				return nil
			}
		} else {
			if u, ok := src.Interface().(Marshaler); ok {
				ctyVal, err := u.MarshalHCL()
				if err != nil {
					return err
				}
				dst.Set(reflect.ValueOf(ctyVal))
				return nil
			}
			if u, ok := src.Interface().(encoding.TextMarshaler); ok && src.Type() == stringTy {
				text, err := u.MarshalText()
				if err != nil {
					return err
				}
				ctyVal := cty.StringVal(string(text))
				dst.Set(reflect.ValueOf(ctyVal))
				return nil
			}
		}

		switch src.Kind() {
		case reflect.String:
			ctyVal := cty.StringVal(src.String())
			dst.Set(reflect.ValueOf(ctyVal))
		case reflect.Bool:
			ctyVal := cty.BoolVal(src.Bool())
			dst.Set(reflect.ValueOf(ctyVal))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			ctyVal := cty.NumberIntVal(src.Int())
			dst.Set(reflect.ValueOf(ctyVal))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			ctyVal := cty.NumberUIntVal(src.Uint())
			dst.Set(reflect.ValueOf(ctyVal))
		case reflect.Float32, reflect.Float64:
			ctyVal := cty.NumberFloatVal(src.Float())
			dst.Set(reflect.ValueOf(ctyVal))
		case reflect.Slice:
			dstSlice := make([]cty.Value, src.Len())
			for i := 0; i < src.Len(); i++ {
				elem := reflect.New(dst.Type())
				if err := m.MapRefl(src.Index(i), elem); err != nil {
					return err
				}
				dstSlice[i] = *(elem.Interface().(*cty.Value))
			}
			if src.Len() > 0 {
				dst.Set(reflect.ValueOf(cty.ListVal(dstSlice)))
			}
		case reflect.Map:
			dstMap := make(map[string]cty.Value)
			for it := src.MapRange(); it.Next(); {
				keyRv := reflect.New(stringTy)
				if err := m.MapRefl(it.Value(), keyRv); err != nil {
					return err
				}
				key := keyRv.Interface().(string)
				if key == "" {
					continue
				}
				val := reflect.New(dst.Type())
				if err := m.MapRefl(it.Value(), val); err != nil {
					return err
				}
				dstMap[key] = *(val.Interface().(*cty.Value))
			}
			dst.Set(reflect.ValueOf(cty.MapVal(dstMap)))
		default:
			return fmt.Errorf("unsupported type %s to ctyValue", src.Type())
		}
		return nil
	}
}

func strMapper(_ *anymapper.Mapper, _, dst reflect.Type) anymapper.MapFunc {
	if dst != stringTy {
		return nil
	}

	return func(m *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
		val := src.Interface()
		str, ok := val.(string)
		if ok {
			dst.Set(reflect.ValueOf(str))
			return nil
		}
		if u, ok := val.(stringize); !ok {
			str = u.String()
			dst.Set(reflect.ValueOf(str))
			return nil
		}
		return fmt.Errorf("unsupported type %s to String", src.Type())
	}
}

func init() {
	mapper = anymapper.New()
	mapper.Context.StrictTypes = true
	mapper.Mappers[ctyValTy] = ctyMapper
	mapper.Mappers[stringTy] = strMapper
}
