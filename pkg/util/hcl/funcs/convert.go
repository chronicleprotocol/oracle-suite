package funcs

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
)

// MakeToFunc returns a function that converts its argument to the given type
// using the cty/convert package.
//
// It should work just like the "to*" functions in Terraform.
func MakeToFunc(wantTyp cty.Type) function.Function {
	return function.New(&function.Spec{
		Description: fmt.Sprintf("Converts the given value to %s type.", wantTyp.FriendlyName()),
		Params: []function.Parameter{
			{
				Name:             "value",
				Description:      "The value to convert.",
				Type:             cty.DynamicPseudoType,
				AllowNull:        true,
				AllowUnknown:     false,
				AllowMarked:      true,
				AllowDynamicType: true,
			},
		},
		Type: func(args []cty.Value) (cty.Type, error) {
			valTyp := args[0].Type()
			if valTyp.Equals(wantTyp) {
				return wantTyp, nil
			}
			if convert.GetConversionUnsafe(args[0].Type(), wantTyp) == nil {
				return cty.NilType, function.NewArgErrorf(
					0,
					fmt.Sprintf(
						"cannot convert %s to %s: ",
						valTyp.FriendlyNameForConstraint(),
						wantTyp.FriendlyNameForConstraint(),
					),
					convert.MismatchMessage(valTyp, wantTyp),
				)
			}
			return wantTyp, nil
		},
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			val, err := convert.Convert(args[0], retType)
			if err != nil {
				return cty.NilVal, function.NewArgErrorf(
					0,
					"cannot convert %s to %s: %s",
					args[0].Type().FriendlyName(),
					retType.FriendlyName(),
					err,
				)
			}
			return val, nil
		},
	})
}
