package variables

import (
	"fmt"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

const (
	varBlockName  = "variables"
	varObjectName = "var"
)

// Variables is a custom block type that allows to define custom variables
// in the "variables" block.
//
// Variables may be referenced by using the global "var" object.
//
// Example:
//
//	variables {
//	  var "foo" {
//	    value = "bar"
//	  }
//	}
//
//	block "example" {
//	  foo = var.foo.value
//	}
//
// Variables may be referenced in the "variables" block itself.
func Variables(ctx *hcl.EvalContext, body hcl.Body) (hcl.Body, hcl.Diagnostics) {
	if ctx.Variables == nil {
		ctx.Variables = map[string]cty.Value{}
	}

	// Decode the "variables" block.
	content, remain, diags := body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: varBlockName}},
	})
	if diags.HasErrors() {
		return nil, diags
	}

	// Add all variables to the resolver.
	resolver := &variableResolver{
		rootName:  varObjectName,
		variables: &variables{},
	}
	for _, block := range content.Blocks.OfType(varBlockName) {
		attrs, attrsDiags := block.Body.JustAttributes()
		if diags := diags.Extend(attrsDiags); diags.HasErrors() {
			return nil, diags
		}
		for _, a := range sortAttrs(attrs) {
			if diags := resolver.add(ctx, []cty.Value{cty.StringVal(a.Name)}, a.Expr); diags.HasErrors() {
				return nil, diags
			}
		}
	}

	// Resolve all variables.
	resDiags := resolver.resolveAll(ctx)
	if diags := diags.Extend(resDiags); diags.HasErrors() {
		return nil, diags
	}

	// Add all resolved variables to the context.
	ctx.Variables[varObjectName], diags = resolver.toValue(ctx)

	return remain, diags
}

// variableResolver recursively resolves variables.
type variableResolver struct {
	// rootName is the name of the root variable. This is used to detect
	// self-references in expressions.
	rootName string

	// variables is a tree structure that contains information required to
	// resolve variables.
	variables *variables
}

// add adds a new variable to the resolver.
//
// path is the path of the variable without the root name.
func (r *variableResolver) add(ctx *hcl.EvalContext, path []cty.Value, expr hcl.Expression) hcl.Diagnostics {
	return r.traverse(ctx, path, expr, func(ctx *hcl.EvalContext, path []cty.Value, expr hcl.Expression) (bool, hcl.Diagnostics) {
		v := r.variables.path(path)
		v.expression = expr
		switch exprTyp := expr.(type) {
		case exprMap:
			v.size = len(exprTyp.ExprMap())
		case exprList:
			v.size = len(exprTyp.ExprList())
		default:
			v.size = -1
		}
		// There is no need to traverse further if the expression does not
		// contain any self-references. Such expressions can be evaluated
		// immediately.
		return r.hasSelfReference(v), nil
	})
}

// resolveAll resolves all variables.
func (r *variableResolver) resolveAll(ctx *hcl.EvalContext) hcl.Diagnostics {
	for _, kv := range r.variables.variables {
		if diags := r.resolveSingle(ctx, []cty.Value{kv.key}); diags.HasErrors() {
			return diags
		}
	}
	return nil
}

// resolveSingle resolves a single variable.
func (r *variableResolver) resolveSingle(ctx *hcl.EvalContext, path []cty.Value) (diags hcl.Diagnostics) {
	_, variable := r.variables.closest(path)

	// Before resolving the variable, try to optimize variable.
	// If successful, then we can skip traversing the expression.
	optDiags := variable.optimize(ctx)
	if diags = diags.Extend(optDiags); diags.HasErrors() {
		return diags
	}

	// Skip if already resolved.
	if variable.resolved != nil {
		return nil
	}

	// If is not already resolved but is being visited, then we have a circular reference.
	if variable.visited {
		return hcl.Diagnostics{{
			Severity:    hcl.DiagError,
			Summary:     "Circular reference detected",
			Detail:      "Variable refers to itself through a circular reference.",
			Subject:     variable.expression.Range().Ptr(),
			Expression:  variable.expression,
			EvalContext: ctx,
		}}
	}
	variable.visited = true

	// If the variable does not have any self reference, then we can simply
	// evaluate the expression.
	if !r.hasSelfReference(variable) {
		value, valDiags := variable.expression.Value(ctx)
		if diags := diags.Extend(valDiags); diags.HasErrors() {
			return diags
		}
		variable.resolved = &value
		return nil
	}

	// Resolve the variable.
	return r.traverse(ctx, path, variable.expression, func(ctx *hcl.EvalContext, path []cty.Value, expr hcl.Expression) (bool, hcl.Diagnostics) {
		switch expr.(type) {
		case exprMap, exprList:
			return true, nil
		default:
			// Resolve referenced variables first.
			varRefs := variable.expression.Variables()
			varRefPaths := make([][]cty.Value, 0, len(varRefs))
			for _, varRef := range varRefs {
				if varRef.RootName() != varObjectName {
					continue
				}
				var path []cty.Value
				for _, t := range varRef[1:] {
					switch t := t.(type) {
					case hcl.TraverseRoot:
						// Skip.
					case hcl.TraverseAttr:
						path = append(path, cty.StringVal(t.Name))
					case hcl.TraverseIndex:
						path = append(path, t.Key)
					default:
						return false, hcl.Diagnostics{{
							Severity:    hcl.DiagError,
							Summary:     "Invalid variable reference",
							Detail:      "Invalid variable reference.",
							Subject:     expr.Range().Ptr(),
							Expression:  expr,
							EvalContext: ctx,
						}}
					}
				}
				resDiags := r.resolveSingle(ctx, path)
				if diags := diags.Extend(resDiags); diags.HasErrors() {
					return false, diags
				}
				varRefPaths = append(varRefPaths, path)
			}

			// To evaluate the expression that contains self references, we need to
			// add those self references to the evaluation context. Unfortunately,
			// the cty.Value is immutable, so we need to rebuild the variable
			// every time we need to resolve a self reference. To avoid rebuilding
			// the entire variable, we create a new variable that only contains
			// variables that are referenced by the expression.
			varRefFiltered := &variables{}
			for _, path := range varRefPaths {
				n, v := r.variables.closest(path)
				fv := varRefFiltered.path(path[0 : len(path)-n])
				fv.expression = v.expression
				fv.resolved = v.resolved
				fv.variables = v.variables
			}

			// Convert the filtered variable to a cty.Value and add it to the\
			// evaluation context.
			ref, refDiags := varRefFiltered.toValue(ctx)
			if diags = diags.Extend(refDiags); diags.HasErrors() {
				return false, diags
			}
			ctx.Variables[r.rootName] = ref

			// Evaluate the expression.
			value, valDiags := variable.expression.Value(ctx)
			if diags = diags.Extend(valDiags); diags.HasErrors() {
				return false, diags
			}
			variable.resolved = &value
			ctx.Variables[r.rootName] = cty.NilVal
		}
		return true, nil
	})
}

// toValue returns resolved variables as a single cty.Value.
func (r *variableResolver) toValue(ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	return r.variables.toValue(ctx)
}

// traverse traverses all map and list expressions in the given expression.
//
// The given function is called for each expression. If the function returns
// false, then the traversal is stopped.
func (r *variableResolver) traverse(
	ctx *hcl.EvalContext,
	path []cty.Value,
	expr hcl.Expression,
	fn func(ctx *hcl.EvalContext, path []cty.Value, expr hcl.Expression) (bool, hcl.Diagnostics),
) hcl.Diagnostics {

	var diags hcl.Diagnostics
	switch exprTyp := expr.(type) {
	case exprMap:
		cont, fnDiags := fn(ctx, path, expr)
		if diags = diags.Extend(fnDiags); diags.HasErrors() {
			return diags
		}
		if !cont {
			return diags
		}
		for _, kv := range exprTyp.ExprMap() {
			if len(kv.Key.Variables()) > 0 {
				diags = diags.Append(&hcl.Diagnostic{
					Severity:    hcl.DiagError,
					Summary:     "Variables inside map keys are not supported",
					Detail:      "Variable contains a map key with a variable inside.",
					Subject:     kv.Value.Range().Ptr(),
					Expression:  kv.Value,
					EvalContext: ctx,
				})
			}
			if diags.HasErrors() {
				continue
			}
			key, keyDiags := kv.Key.Value(ctx)
			if diags = diags.Extend(keyDiags); diags.HasErrors() {
				return diags
			}
			subDiags := r.traverse(ctx, append(path, key), kv.Value, fn)
			if diags = diags.Extend(subDiags); diags.HasErrors() {
				return diags
			}
		}
	case exprList:
		cont, fnDiags := fn(ctx, path, expr)
		if diags = diags.Extend(fnDiags); diags.HasErrors() {
			return diags
		}
		if !cont {
			return diags
		}
		for i, e := range exprTyp.ExprList() {
			subDiags := r.traverse(ctx, append(path, cty.NumberIntVal(int64(i))), e, fn)
			if diags = diags.Extend(subDiags); diags.HasErrors() {
				return diags
			}
		}
	default:
		cont, fnDiags := fn(ctx, path, exprTyp)
		if diags = diags.Extend(fnDiags); diags.HasErrors() {
			return diags
		}
		if !cont {
			return diags
		}
	}
	return diags
}

// hasSelfReference returns whether the given variable has a reference to
// other variables.
func (r *variableResolver) hasSelfReference(v *variables) bool {
	if v.expression == nil {
		return false
	}
	for _, ref := range v.expression.Variables() {
		if ref.RootName() == r.rootName {
			return true
		}
	}
	return false
}

type exprMap interface {
	ExprMap() []hcl.KeyValuePair
}

type exprList interface {
	ExprList() []hcl.Expression
}

// variables describes a single variable in the variables block.
//
// A variable may contain other variables if it is a map or a list because
// each element of a map or a list is treated as a separate variable.
type variables struct {
	// expression that defines this variable.
	expression hcl.Expression

	// resolved is the resolved value of this variable.
	resolved *cty.Value

	// visited is whether this variable is already visited during the
	// resolution process.
	visited bool

	// size is the number of elements in this variable, -1 if the variable
	// is not a map or a list.
	size int

	// variables contained in this variable (if the expression is a map or a list).
	variables []kv
}

type kv struct {
	key   cty.Value
	value *variables
}

// path returns the variables at the given path. If a variable at the given
// path does not exist, it is created.
func (t *variables) path(path []cty.Value) *variables {
	if len(path) == 0 {
		return t
	}
	for _, i := range t.variables {
		if i.key.Equals(path[0]).True() {
			return i.value.path(path[1:])
		}
	}
	kv := kv{
		key:   path[0],
		value: &variables{},
	}
	t.variables = append(t.variables, kv)
	return kv.value.path(path[1:])
}

// closest returns the first already defined variable that shares the
// longest common prefix with the given path. It returns the distance
// between the given path and the found variable and the found variable.
func (t *variables) closest(path []cty.Value) (int, *variables) {
	if len(path) == 0 {
		return 0, t
	}
	for _, i := range t.variables {
		if i.key.Equals(path[0]).True() {
			return i.value.closest(path[1:])
		}
	}
	return len(path), t
}

// optimize sets the resolved value if all sub variables are already resolved.
func (t *variables) optimize(cty *hcl.EvalContext) hcl.Diagnostics {
	if t.resolved != nil {
		return nil
	}
	if t.expression == nil && t.resolved == nil && len(t.variables) == 0 {
		return nil
	}
	for _, v := range t.variables {
		if diags := v.value.optimize(cty); diags.HasErrors() {
			return diags
		}
	}
	if t.size != len(t.variables) {
		return nil
	}
	resolved := true
	for _, v := range t.variables {
		if v.value.resolved == nil {
			resolved = false
		}
	}
	if resolved {
		value, diags := t.toValue(cty)
		if diags.HasErrors() {
			return diags
		}
		t.resolved = &value
		t.variables = nil
	}
	return nil
}

// toValue converts the variables to a cty value.
//
// It tries to use a list or map whenever possible because it allows to
// access the values using the index operator instead of the attribute
// operator. This is important because the index operator can be used
// in more places than the attribute operator.
func (t *variables) toValue(ctx *hcl.EvalContext) (value cty.Value, diags hcl.Diagnostics) {
	if t.resolved != nil {
		return *t.resolved, nil
	}
	if t.expression == nil && t.resolved == nil && len(t.variables) == 0 {
		return cty.NilVal, nil
	}
	switch t.expression.(type) {
	case exprList:
		vals := make([]cty.Value, len(t.variables))
		for i, v := range t.variables {
			value, valDiags := v.value.toValue(ctx)
			if diags = diags.Extend(valDiags); diags.HasErrors() {
				continue
			}
			vals[i] = value
		}
		if diags.HasErrors() {
			return cty.NilVal, diags
		}
		return cty.TupleVal(vals), diags
	case exprMap, nil:
		vals := make(map[string]cty.Value)
		for _, v := range t.variables {
			key, err := convert.Convert(v.key, cty.String)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity:    hcl.DiagError,
					Summary:     "Incorrect key type",
					Detail:      fmt.Sprintf("Can't use this value as a key: %s.", err.Error()),
					Subject:     v.value.expression.Range().Ptr(),
					Expression:  v.value.expression,
					EvalContext: ctx,
				})
				continue
			}
			if diags.HasErrors() {
				continue
			}
			value, valDiags := v.value.toValue(ctx)
			if diags = diags.Extend(valDiags); diags.HasErrors() {
				continue
			}
			vals[key.AsString()] = value
		}
		if diags.HasErrors() {
			return cty.NilVal, diags
		}
		return cty.ObjectVal(vals), diags
	}
	return cty.NilVal, diags
}

func sortAttrs(attrs hcl.Attributes) []*hcl.Attribute {
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sorted := make([]*hcl.Attribute, 0, len(attrs))
	for _, k := range keys {
		sorted = append(sorted, attrs[k])
	}
	return sorted
}
