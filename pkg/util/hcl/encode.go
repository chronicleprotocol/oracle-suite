package hcl

import (
	"fmt"
	"reflect"

	"github.com/defiweb/go-eth/types"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/origin"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
)

type PreEncodeBody interface {
	PreEncodeBody(block *hclwrite.Body, val interface{}) error
}

type PostEncodeBody interface {
	PostEncodeBody(block *hclwrite.Body, val interface{}) error
}

func Encode(val interface{}, body *hclwrite.Body) error {
	rv := reflect.ValueOf(val)
	ty := rv.Type()
	if ty.Kind() == reflect.Ptr {
		rv = rv.Elem()
		ty = rv.Type()
	}
	if ty.Kind() != reflect.Struct {
		return fmt.Errorf("value is %s, not struct", ty.Kind())
	}
	return populateBody(rv, body)
}

func EncodeAsBlock(val interface{}, blockType string, body *hclwrite.Body) error {
	rv := reflect.ValueOf(val)
	ty := rv.Type()
	if ty.Kind() == reflect.Ptr {
		if !rv.IsValid() {
			return nil
		}
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
		ty = rv.Type()
	}
	if ty.Kind() != reflect.Struct {
		return fmt.Errorf("value is %s, not struct", ty.Kind())
	}

	meta, diags := getStructMeta(ty)
	if diags.HasErrors() {
		return fmt.Errorf(diags.Error())
	}

	labels := make([]string, len(meta.Labels))
	for i, lf := range meta.Labels {
		fieldVal := rv.FieldByIndex(lf.Reflect.Index)
		if label, ok := fieldVal.Interface().(string); ok {
			labels[i] = label
		}
		if pair, ok := fieldVal.Interface().(value.Pair); ok {
			labels[i] = pair.String()
		}
	}

	newBlock := hclwrite.NewBlock(blockType, labels)
	err := populateBody(rv, newBlock.Body())
	if err != nil {
		return err
	}
	body.AppendBlock(newBlock)
	return nil
}

func populateBody(rv reflect.Value, body *hclwrite.Body) error { //nolint:gocyclo
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
		fieldRv := reflect.ValueOf(fieldVal.Interface())
		if block.Multiple {
			for i := 0; i < fieldRv.Len(); i++ {
				err := EncodeAsBlock(fieldRv.Index(i).Interface(), block.Name, body)
				if err != nil {
					return err
				}
				body.AppendNewline()
			}
		} else {
			err := EncodeAsBlock(fieldVal.Interface(), block.Name, body)
			if err != nil {
				return err
			}
			body.AppendNewline()
		}
	}

	for _, attr := range meta.Attrs {
		fieldVal := rv.FieldByIndex(attr.Reflect.Index)

		// todo, rewrite ctyMapper, dstType is ctyValue
		// todo, big.Int, big.Float
		var val cty.Value
		switch fieldVal.Type() {
		case reflect.TypeOf((*types.Address)(nil)).Elem():
			val = cty.StringVal(fieldVal.Interface().(types.Address).String())
		case reflect.TypeOf((*origin.ContractAddresses)(nil)).Elem():
			if addresses, ok := fieldVal.Interface().(origin.ContractAddresses); ok {
				mapAddresses := make(map[string]cty.Value)
				for key, value := range addresses {
					pairs := key.String()
					mapAddresses[pairs] = cty.StringVal(value.String())
				}
				val = cty.MapVal(mapAddresses)
			}
		// case reflect.TypeOf((*config.URL)(nil)).Elem():
		//	val = cty.StringVal(fieldVal.Interface().(*config.URL).String())
		default:
			valTy, err := gocty.ImpliedType(fieldVal.Interface())
			if err != nil {
				// return fmt.Errorf("cannot encode %T as HCL expression: %s", fieldVal.Interface(), err)
				continue
			}
			val, err = gocty.ToCtyValue(fieldVal.Interface(), valTy)
			if err != nil {
				// This should never happen, since we should always be able
				// to decode into the implied type.
				// return fmt.Errorf("failed to encode %T as %#v: %s", fieldVal.Interface(), valTy, err)
				continue
			}
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
