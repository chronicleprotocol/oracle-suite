package hcl

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// OnEncodeBlock is called during encoding of a struct to HCL.
type OnEncodeBlock interface {
	OnEncodeBlock(block *Block) hcl.Diagnostics
}

// Encode encodes a struct to HCL.
// The value must be a struct.
func Encode(val any, block *Block) hcl.Diagnostics {
	return encodeSingleBlock(reflect.ValueOf(val), block, "", nil)
}

// EncodeBlock encodes a struct to HCL block and appends it to the given block.
// The value must be a struct.
//
// The typeName argument specifies the block type name. Labels are specified
// using the labels argument. If labels is nil, labels are read from the struct
// tags.
func EncodeBlock(val any, block *Block, typeName string, labels []string) hcl.Diagnostics {
	if typeName == "" {
		return hcl.Diagnostics{{
			Severity: hcl.DiagError,
			Summary:  "Encode error",
			Detail:   "EncodeBlock requires a non-empty typeName",
		}}
	}
	return encodeSingleBlock(reflect.ValueOf(val), block, typeName, labels)
}

// encodeSingleBlock encodes given struct value as HCL block.
//
//nolint:gocyclo
func encodeSingleBlock(val reflect.Value, body *Block, typeName string, labels []string) hcl.Diagnostics {
	val = derefValue(val)

	if val.Kind() != reflect.Struct {
		return hcl.Diagnostics{{
			Severity: hcl.DiagError,
			Summary:  "Encode error",
			Detail:   fmt.Sprintf("Unable to encode %s as block; value must be a struct", val.Kind()),
		}}
	}

	// Decode struct tags.
	meta, diags := getStructMeta(val.Type())
	if diags.HasErrors() {
		return diags
	}

	// If typeName is empty, we are encoding the root block, so there is no need
	// to create a new block.
	if typeName != "" {
		// Encode labels from fields with `label` tag, but only if labels are not
		// already set.
		if labels == nil {
			for _, label := range meta.Labels {
				var labelVal cty.Value
				fieldRef := val.FieldByIndex(label.Reflect.Index)
				labelRef := reflect.ValueOf(&labelVal)
				if err := mapper.MapRefl(fieldRef, labelRef); err != nil {
					return hcl.Diagnostics{{
						Severity: hcl.DiagError,
						Summary:  "Decode error",
						Detail:   fmt.Sprintf("Cannot encode label %q: %s", label.Name, err),
					}}
				}
				if labelVal.Type() != cty.String {
					return hcl.Diagnostics{{
						Severity: hcl.DiagError,
						Summary:  "Decode error",
						Detail:   fmt.Sprintf("Cannot encode label %q: value must be a string", label.Name),
					}}
				}
				labels = append(labels, labelVal.AsString())
			}
		}
		body = body.AppendBlock(typeName, labels)
	}

	// Encode attributes.
	for _, attr := range meta.Attrs {
		if attr.Ignore {
			continue
		}
		field := val.FieldByIndex(attr.Reflect.Index)
		if attr.Optional && canBeSkipped(field) {
			continue
		}
		body.AppendAttribute(attr.Name, field.Interface())
	}

	// Encode blocks.
	for _, block := range meta.Blocks {
		if block.Ignore {
			continue
		}
		field := val.FieldByIndex(block.Reflect.Index)
		if block.Optional && canBeSkipped(field) {
			continue
		}
		if block.Multiple {
			diags := encodeMultipleBlocks(body, field, block.Name)
			if diags.HasErrors() {
				return diags
			}
		} else {
			diags := encodeSingleBlock(field, body, block.Name, nil)
			if diags.HasErrors() {
				return diags
			}
		}
	}

	// Encode hook.
	if n, ok := val.Interface().(OnEncodeBlock); ok {
		diags := n.OnEncodeBlock(body)
		if diags.HasErrors() {
			return diags
		}
	}

	return nil
}

// encodeMultipleBlocks encodes a map or slice value as HCL blocks.
func encodeMultipleBlocks(body *Block, val reflect.Value, typeName string) hcl.Diagnostics {
	val = derefValue(val)

	switch val.Kind() {
	case reflect.Map:
		if val.Type().Key().Kind() != reflect.String {
			return hcl.Diagnostics{{
				Severity: hcl.DiagError,
				Summary:  "Encode error",
				Detail:   "Unable to encode value as HCL; map key is not a string",
			}}
		}
		// Keys are sorted to ensure consistent output.
		keys := val.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})
		for _, key := range keys {
			diags := encodeSingleBlock(val.MapIndex(key), body, typeName, []string{key.String()})
			if diags.HasErrors() {
				return diags
			}
		}
		return nil
	case reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			diags := encodeSingleBlock(val.Index(i), body, typeName, nil)
			if diags.HasErrors() {
				return diags
			}
		}
		return nil
	}
	return hcl.Diagnostics{{
		Severity: hcl.DiagError,
		Summary:  "Encode error",
		Detail:   fmt.Sprintf("Unable to encode %s as block", val.Kind()),
	}}
}

// canBeSkipped returns true if the given value can be skipped during encoding
// to HCL if the field has the `optional` tag set.
func canBeSkipped(v reflect.Value) bool {
	if !v.IsValid() {
		// Can this happen?
		return true
	}
	if v.Type().PkgPath() != "" {
		// Custom types may implement MarshalHCL or MarshalText, so we cannot
		// skip them.
		return false
	}
	return v.IsZero()
}
