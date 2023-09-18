package variables

import (
	"fmt"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

// Constants representing block and object names.
const (
	varBlockName  = "variables"
	varObjectName = "var"
)

// Variables is a custom block type that allows the definition of custom
// variables within the "variables" block. These variables can be referenced
// globally through the "var" object. Additionally, variables can reference
// each other within the "variables" block itself.
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
func Variables(ctx *hcl.EvalContext, body hcl.Body) (hcl.Body, hcl.Diagnostics) {
	if ctx.Variables == nil {
		ctx.Variables = map[string]cty.Value{}
	}

	// Decode the "variables" block.
	content, remain, diags := decodeVariablesBlock(body)
	if !handleDiagnostics(&diags, diags, !diags.HasErrors()) {
		return nil, diags
	}

	// Initialize and populate the variable resolver.
	resolver := initVariableResolver()
	if !handleDiagnostics(&diags, populateResolver(resolver, ctx, content), !diags.HasErrors()) {
		return nil, diags
	}

	// Resolve all variables and update the context with resolved variables.
	if !handleDiagnostics(&diags, resolveAllVariables(resolver, ctx), !diags.HasErrors()) {
		return nil, diags
	}

	return remain, diags
}

// decodeVariablesBlock decodes the "variables" block and returns the content
// and diagnostics.
func decodeVariablesBlock(body hcl.Body) (content *hcl.BodyContent, remain hcl.Body, diags hcl.Diagnostics) {
	content, remain, diags = body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: varBlockName}},
	})
	return
}

// initVariableResolver initializes a new variable resolver.
func initVariableResolver() *variableResolver {
	return &variableResolver{
		rootName:  varObjectName,
		variables: &variables{},
	}
}

// populateResolver populates the resolver with variables from the content.
func populateResolver(resolver *variableResolver, ctx *hcl.EvalContext, content *hcl.BodyContent) (diags hcl.Diagnostics) {
	for _, block := range content.Blocks.OfType(varBlockName) {
		attrs, attrsDiags := block.Body.JustAttributes()
		if !handleDiagnostics(&diags, attrsDiags, !diags.HasErrors()) {
			return
		}
		for _, attr := range sortAttributes(attrs) {
			if !handleDiagnostics(&diags, resolver.add(ctx, []cty.Value{cty.StringVal(attr.Name)}, attr.Expr), !diags.HasErrors()) {
				return
			}
		}
	}
	return diags
}

// resolveAllVariables resolves all variables and updates the context with the
// resolved variables.
func resolveAllVariables(resolver *variableResolver, ctx *hcl.EvalContext) (diags hcl.Diagnostics) {
	resDiags := resolver.resolveAll(ctx)
	if !handleDiagnostics(&diags, resDiags, !diags.HasErrors()) {
		return
	}
	ctx.Variables[varObjectName], diags = resolver.toValue(ctx)
	return diags
}

// sortAttributes sorts the attributes by name.
func sortAttributes(attrs hcl.Attributes) []*hcl.Attribute {
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

// variableResolver is a type that recursively resolves variables
type variableResolver struct {
	rootName  string     // Name of the root variable, used to detect self-references in expressions.
	variables *variables // Tree structure containing information required to resolve variables.
}

// add adds a new variable into the resolver. The path parameter represents
// the path of the variable excluding the root name.
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

// resolveAll iterates through and resolves all variables added to the resolver.
func (r *variableResolver) resolveAll(ctx *hcl.EvalContext) hcl.Diagnostics {
	for _, kv := range r.variables.variables {
		if diags := r.resolveSingle(ctx, []cty.Value{kv.key}); diags.HasErrors() {
			return diags
		}
	}
	return nil
}

// toValue returns resolved variables as a single cty.Value.
func (r *variableResolver) toValue(ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	return r.variables.toValue(ctx)
}

// resolveSingle ensures that the value of the variable at the specified path
// has been resolved.
func (r *variableResolver) resolveSingle(ctx *hcl.EvalContext, path []cty.Value) (diags hcl.Diagnostics) {
	// If current variable is referenced by another variable, it may happen
	// that it will reference an element of a list or a map that is already
	// resolved at higher level. For this reason, we use the closest method
	// to find that map/list.
	_, variable := r.variables.closest(path)

	// Optimize the variable if possible
	if !handleDiagnostics(&diags, r.optimizeVariable(ctx, variable), !diags.HasErrors()) {
		return diags
	}

	// Check if the variable is already resolved or has a circular reference
	if !handleDiagnostics(&diags, r.checkVariableStatus(ctx, variable), !diags.HasErrors()) {
		return diags
	}

	// Resolve the variable if it does not have any self-reference
	if !r.hasSelfReference(variable) {
		if !handleDiagnostics(&diags, r.resolveVariableWithoutSelfReference(ctx, variable), !diags.HasErrors()) {
			return diags
		}
	}

	// Resolve the variable with self-reference
	return r.resolveVariableWithSelfReference(ctx, path, variable)
}

// optimizeVariable tries to optimize the variable before resolving it.
func (r *variableResolver) optimizeVariable(ctx *hcl.EvalContext, variable *variables) (diags hcl.Diagnostics) {
	optDiags := variable.optimize(ctx)
	return diags.Extend(optDiags)
}

// checkVariableStatus checks if the variable is already resolved or has a
// circular reference.
func (r *variableResolver) checkVariableStatus(ctx *hcl.EvalContext, variable *variables) (diags hcl.Diagnostics) {
	if variable.resolved != nil {
		return nil
	}
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
	return nil
}

// resolveVariableWithoutSelfReference resolves the variable if it does not
// have any self-reference.
func (r *variableResolver) resolveVariableWithoutSelfReference(ctx *hcl.EvalContext, variable *variables) (diags hcl.Diagnostics) {
	value, valDiags := variable.expression.Value(ctx)
	if diags := diags.Extend(valDiags); diags.HasErrors() {
		return diags
	}
	variable.resolved = &value
	return nil
}

// resolveVariableWithSelfReference resolves the variable if it has
// self-reference.
func (r *variableResolver) resolveVariableWithSelfReference(ctx *hcl.EvalContext, path []cty.Value, variable *variables) (diags hcl.Diagnostics) {
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

			// Create a new variable that only contains variables referenced by
			// the expression.
			//
			// To evaluate the expression that contains self references, we
			// need to add those self references to the evaluation context.
			// Unfortunately, the cty.Value is immutable, so we need to rebuild
			// the variable every time we need to resolve a self reference. To
			// avoid rebuilding the entire variable, we create a new variable
			// that only contains variables that are referenced by the expression.
			varRefFiltered := &variables{}
			for _, path := range varRefPaths {
				n, v := r.variables.closest(path)
				fv := varRefFiltered.path(path[0 : len(path)-n])
				fv.expression = v.expression
				fv.resolved = v.resolved
				fv.variables = v.variables
			}

			// Convert the filtered variable to a cty.Value and add it to the
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

// traverse iterates through all map and list expressions within the given
// expression. The specified function is invoked for each expression, halting
// the traversal if it returns false.
func (r *variableResolver) traverse(
	ctx *hcl.EvalContext,
	path []cty.Value,
	expr hcl.Expression,
	fn func(ctx *hcl.EvalContext, path []cty.Value, expr hcl.Expression) (bool, hcl.Diagnostics),
) hcl.Diagnostics {

	var diags hcl.Diagnostics
	switch exprType := expr.(type) {
	case exprMap:
		// Handle map expressions.
		cont, fnDiags := fn(ctx, path, expr)
		if !handleDiagnostics(&diags, fnDiags, cont) {
			return diags
		}
		for _, kv := range exprType.ExprMap() {
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
		// Handle list expressions.
		cont, fnDiags := fn(ctx, path, expr)
		if !handleDiagnostics(&diags, fnDiags, cont) {
			return diags
		}
		for i, e := range exprType.ExprList() {
			subDiags := r.traverse(ctx, append(path, cty.NumberIntVal(int64(i))), e, fn)
			if diags = diags.Extend(subDiags); diags.HasErrors() {
				return diags
			}
		}
	default:
		// Handle other types of expressions.
		cont, fnDiags := fn(ctx, path, exprType)
		if !handleDiagnostics(&diags, fnDiags, cont) {
			return diags
		}
	}

	return diags
}

// hasSelfReference determines whether the specified variable contains
// references to other variables.
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

// variables represents a single variable within the variables block. It may
// encapsulate other variables if it represents a map or a list.
type variables struct {
	// expression defines this variable.
	expression hcl.Expression

	// resolved holds the resolved value of this variable.
	resolved *cty.Value

	// visited indicates whether this variable has been visited during the
	// resolution process.
	visited bool

	// size indicates the number of elements in this variable, or -1 if the
	// variable is neither a map nor a list.
	size int

	// variables holds the variables contained within this variable (if the
	// expression is a map or a list).
	variables []kv
}

type kv struct {
	key   cty.Value
	value *variables
}

// path retrieves the variable at the specified path, creating it if it does
// not exist.
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

// closest finds the first defined variable that shares the longest common
// prefix with the given path. It returns the distance between the given path
// and the found variable, along with the found variable itself.
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

// optimize attempts to optimize the variable by setting its resolved value
// if it is a map or a list and all of its children are already resolved.
func (t *variables) optimize(ctx *hcl.EvalContext) hcl.Diagnostics {
	// If the variable is already resolved, return nil.
	if t.resolved != nil {
		return nil
	}

	// If the variable is empty, return a nil value.
	if t.expression == nil && t.resolved == nil && len(t.variables) == 0 {
		return nil
	}

	// Iterate over the child variables and attempt to optimize them first.
	for _, v := range t.variables {
		if diags := v.value.optimize(ctx); diags.HasErrors() {
			return diags
		}
	}

	// If the size of the variable does not match the number of child variables,
	// it means that the variable is not yet fully resolved, therefore cannot be
	// optimized.
	if t.size != len(t.variables) {
		return nil
	}

	// Check if all child variables are resolved.
	allResolved := true
	for _, v := range t.variables {
		if v.value.resolved == nil {
			allResolved = false
			break
		}
	}

	// If all child variables are resolved, convert them to a cty value and set
	// it as the resolved value of the current variable.
	if allResolved {
		value, diags := t.toValue(ctx)
		if diags.HasErrors() {
			return diags
		}
		t.resolved = &value
		t.variables = nil
	}

	return nil
}

// toValue converts the variables to a cty value.
func (t *variables) toValue(ctx *hcl.EvalContext) (value cty.Value, diags hcl.Diagnostics) {
	// If the variable is already resolved, return the resolved value.
	if t.resolved != nil {
		return *t.resolved, nil
	}

	// If the variable is empty, return a nil value.
	if t.expression == nil && t.resolved == nil && len(t.variables) == 0 {
		return cty.NilVal, nil
	}

	switch t.expression.(type) {
	case exprList:
		// If the variable is a list, convert each item in the list to a
		// cty value and return a tuple of these values.
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
		// If the variable is a map or nil, convert each key-value pair to a
		// cty value and return an object of these values.
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

// handleDiagnostics is a helper function to handle diagnostics and continue
// condition.
func handleDiagnostics(target *hcl.Diagnostics, diags hcl.Diagnostics, cont bool) bool {
	*target = diags.Extend(diags)
	if !cont || diags.HasErrors() {
		return false
	}
	return true
}
