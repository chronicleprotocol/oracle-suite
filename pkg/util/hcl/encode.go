package hcl

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type PreEncodeBody interface {
	PreEncodeBody(block *hclwrite.Body, val interface{}) error
}

type PostEncodeBody interface {
	PostEncodeBody(block *hclwrite.Body, val interface{}) error
}

func Encode(val interface{}, body *hclwrite.Body) error {
	ptrVal := reflect.ValueOf(val)
	rv := derefValue(ptrVal)
	ty := rv.Type()
	if ty.Kind() != reflect.Struct {
		return fmt.Errorf("value is %s, not struct", ty.Kind())
	}
	return populateBody(rv, body)
}

func EncodeAsBlock(val interface{}, blockType string, body *hclwrite.Body) error {
	ptrVal := reflect.ValueOf(val)
	rv := derefValue(ptrVal)
	ty := rv.Type()
	if ty.Kind() != reflect.Struct {
		return fmt.Errorf("value is %s, not struct", ty.Kind())
	}

	meta, diags := getStructMeta(ty)
	if diags.HasErrors() {
		return fmt.Errorf(diags.Error())
	}

	if blockType == "ethereum" {
		fmt.Println("ethereum")
	}

	var labels []string
	for _, lf := range meta.Labels {
		fieldVal := rv.FieldByIndex(lf.Reflect.Index)

		var label string
		if err := mapper.Map(fieldVal.Interface(), &label); err != nil {
			return fmt.Errorf("cannot encode %T as HCL expression: %s", fieldVal.Interface(), err)
		}
		if !lf.Optional && label == "" {
			return fmt.Errorf("missing block label: %s", blockType)
		}
		labels = append(labels, label)
	}
	if len(labels) != len(meta.Labels) {
		return fmt.Errorf("missing block label")
	}

	newBlock := hclwrite.NewBlock(blockType, labels)
	err := populateBody(rv, newBlock.Body())
	if err != nil {
		return err
	}
	body.AppendBlock(newBlock)
	return nil
}

func populateBody(ptrVal reflect.Value, body *hclwrite.Body) error { //nolint:funlen,gocyclo
	rv := derefValue(ptrVal)
	ty := rv.Type()
	meta, diags := getStructMeta(ty)
	if diags.HasErrors() {
		return fmt.Errorf(diags.Error())
	}

	if n, ok := rv.Interface().(PreEncodeBody); ok {
		err := n.PreEncodeBody(body, rv.Interface())
		if err != nil {
			return err
		}
	}

	for _, block := range meta.Blocks {
		if block.Ignore {
			continue
		}

		fieldVal := rv.FieldByIndex(block.Reflect.Index)
		if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
			if !block.Optional {
				return fmt.Errorf("missing block: %s", block.Name)
			}
			continue
		}

		fieldRv := derefValue(fieldVal)
		if block.Multiple {
			switch fieldRv.Kind() {
			case reflect.Map:
				for it := fieldRv.MapRange(); it.Next(); {
					err := EncodeAsBlock(it.Value().Interface(), block.Name, body)
					if err != nil {
						return err
					}
				}
			case reflect.Slice:
				for i := 0; i < fieldRv.Len(); i++ {
					err := EncodeAsBlock(fieldRv.Index(i).Interface(), block.Name, body)
					if err != nil {
						return err
					}
				}
			default:
				return fmt.Errorf("unknown multiple blocks to encode %T as HCL expression", fieldVal.Interface())
			}
		} else {
			err := EncodeAsBlock(fieldVal.Interface(), block.Name, body)
			if err != nil {
				return err
			}
		}
	}

	for _, attr := range meta.Attrs {
		if attr.Ignore {
			continue
		}
		fieldVal := rv.FieldByIndex(attr.Reflect.Index)

		if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
			if !attr.Optional {
				return fmt.Errorf("missing attribute: %s", attr.Name)
			}
			continue
		}

		if fieldVal.IsZero() { // is null value or empty value
			if !attr.Optional {
				return fmt.Errorf("missing attribute: %s", attr.Name)
			}
			continue
		}

		var val cty.Value
		if err := mapper.Map(fieldVal.Interface(), &val); err != nil {
			return fmt.Errorf("cannot encode %T as HCL expression: %s", fieldVal.Interface(), err)
		}
		body.SetAttributeValue(attr.Name, val)
	}

	if n, ok := rv.Interface().(PostEncodeBody); ok {
		err := n.PostEncodeBody(body, rv.Interface())
		if err != nil {
			return err
		}
	}

	return nil
}
