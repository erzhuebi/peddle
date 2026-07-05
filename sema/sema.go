package sema

import (
	"fmt"
	"strings"

	"peddle/ast"
)

type Checker struct {
	constants map[string]*ast.ConstDecl
	globals   map[string]*ast.VarDecl
	functions map[string]*ast.FunctionDecl
	structs   map[string]*ast.StructDecl
}

func New() *Checker {
	return &Checker{
		constants: map[string]*ast.ConstDecl{},
		globals:   map[string]*ast.VarDecl{},
		functions: map[string]*ast.FunctionDecl{},
		structs:   map[string]*ast.StructDecl{},
	}
}

func (c *Checker) Check(p *ast.Program) error {
	for _, cn := range p.Consts {
		if _, exists := c.constants[cn.Name]; exists {
			return fmt.Errorf("duplicate const %q", cn.Name)
		}
		if _, exists := c.functions[cn.Name]; exists {
			return fmt.Errorf("const %q conflicts with function", cn.Name)
		}
		if _, exists := c.structs[cn.Name]; exists {
			return fmt.Errorf("const %q conflicts with struct", cn.Name)
		}
		c.constants[cn.Name] = cn
	}

	for _, s := range p.Structs {
		if _, exists := c.structs[s.Name]; exists {
			return fmt.Errorf("duplicate struct %q", s.Name)
		}
		if _, exists := c.constants[s.Name]; exists {
			return fmt.Errorf("struct %q conflicts with const", s.Name)
		}
		c.structs[s.Name] = s
	}

	for _, global := range p.Globals {
		if _, exists := c.globals[global.Name]; exists {
			return fmt.Errorf("duplicate global %q", global.Name)
		}
		if _, exists := c.constants[global.Name]; exists {
			return fmt.Errorf("global %q conflicts with const", global.Name)
		}
		if _, exists := c.structs[global.Name]; exists {
			return fmt.Errorf("global %q conflicts with struct", global.Name)
		}
		c.globals[global.Name] = global
	}

	for _, s := range p.Structs {
		for _, field := range s.Fields {
			if field.Type.IsMem {
				return fmt.Errorf("struct %s field %q: mem fields are not supported", s.Name, field.Name)
			}
			if err := c.checkType(field.Type); err != nil {
				return fmt.Errorf("struct %s field %q: %w", s.Name, field.Name, err)
			}
		}
	}

	for _, fn := range p.Functions {
		if _, exists := c.functions[fn.Name]; exists {
			return fmt.Errorf("duplicate function %q", fn.Name)
		}
		if _, exists := c.constants[fn.Name]; exists {
			return fmt.Errorf("function %q conflicts with const", fn.Name)
		}
		if _, exists := c.structs[fn.Name]; exists {
			return fmt.Errorf("function %q conflicts with struct", fn.Name)
		}
		if _, exists := c.globals[fn.Name]; exists {
			return fmt.Errorf("function %q conflicts with global", fn.Name)
		}
		c.functions[fn.Name] = fn
	}

	if _, ok := c.functions["main"]; !ok {
		return fmt.Errorf("missing main function")
	}

	for _, global := range p.Globals {
		if err := c.checkVarDecl(global, "global", nil, true); err != nil {
			return err
		}
	}

	for _, fn := range p.Functions {
		if err := c.checkFunction(fn); err != nil {
			return fmt.Errorf("function %s: %w", fn.Name, err)
		}
	}

	return nil
}

func (c *Checker) checkExpr(scope map[string]ast.Type, e ast.Expr) (ast.Type, error) {
	switch expr := e.(type) {
	case *ast.IdentExpr:
		if _, ok := c.constants[expr.Name]; ok {
			return ast.Type{Name: "int"}, nil
		}

		t, ok := scope[expr.Name]
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", expr.Name)
		}
		if isScalarPointerType(t) {
			return pointerPointeeType(t), nil
		}
		return t, nil

	case *ast.IndexExpr:
		t, ok := scope[expr.Name]
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", expr.Name)
		}
		if t.IsMem {
			if err := c.checkIndexForType(scope, t, expr.Index); err != nil {
				return ast.Type{}, err
			}
			return ast.Type{Name: "byte"}, nil
		}
		if !t.IsArray {
			return ast.Type{}, fmt.Errorf("%q is not an array", expr.Name)
		}

		if err := c.checkIndexForType(scope, t, expr.Index); err != nil {
			return ast.Type{}, err
		}

		return ast.Type{Name: t.Name}, nil

	case *ast.IndexFieldExpr:
		t, ok := scope[expr.Name]
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", expr.Name)
		}
		if !t.IsArray {
			return ast.Type{}, fmt.Errorf("%q is not an array", expr.Name)
		}

		idxType, err := c.checkExpr(scope, expr.Index)
		if err != nil {
			return ast.Type{}, err
		}
		if idxType.Name != "byte" && idxType.Name != "int" && idxType.Name != "uint" {
			return ast.Type{}, fmt.Errorf("array index must be byte, int, or uint")
		}

		elemType := ast.Type{Name: t.Name}
		return c.fieldType(elemType, expr.Field)

	case *ast.FieldExpr:
		baseType, ok := scope[expr.Base]
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", expr.Base)
		}
		return c.fieldType(baseType, expr.Field)

	case *ast.NumberExpr:
		return ast.Type{Name: "int"}, nil

	case *ast.CharExpr:
		return ast.Type{Name: "char"}, nil

	case *ast.BoolExpr:
		return ast.Type{Name: "bool"}, nil

	case *ast.StringExpr:
		return ast.Type{Name: "char", IsArray: true, ArrayLen: len(expr.Value)}, nil

	case *ast.ArrayLiteralExpr:
		return ast.Type{}, fmt.Errorf("array literal requires a target type")

	case *ast.StructLiteralExpr:
		return ast.Type{}, fmt.Errorf("struct literal requires a target type")

	case *ast.UnaryExpr:
		if expr.Op == "&" {
			return c.checkAddressOf(scope, expr.Expr)
		}

		t, err := c.checkExpr(scope, expr.Expr)
		if err != nil {
			return ast.Type{}, err
		}

		switch expr.Op {
		case "-":
			if !isNumeric(t) {
				return ast.Type{}, fmt.Errorf("unary - requires numeric operand")
			}
			if t.Name == "int" || t.Name == "uint" {
				return ast.Type{Name: "int"}, nil
			}
			return ast.Type{Name: "byte"}, nil

		case "!":
			if !isNumeric(t) {
				return ast.Type{}, fmt.Errorf("unary ! requires numeric or bool operand")
			}
			return ast.Type{Name: "bool"}, nil

		default:
			return ast.Type{}, fmt.Errorf("unsupported unary operator %q", expr.Op)
		}

	case *ast.BinaryExpr:
		left, err := c.checkExpr(scope, expr.Left)
		if err != nil {
			return ast.Type{}, err
		}

		right, err := c.checkExpr(scope, expr.Right)
		if err != nil {
			return ast.Type{}, err
		}

		switch expr.Op {
		case "+", "-", "*", "/", "%", "<<", ">>":
			if !isNumeric(left) || !isNumeric(right) {
				return ast.Type{}, fmt.Errorf("operator %s requires numeric operands", expr.Op)
			}
			return numericResultType(left, right), nil

		case "&", "|", "^":
			if !isNumeric(left) || !isNumeric(right) {
				return ast.Type{}, fmt.Errorf("operator %s requires numeric operands", expr.Op)
			}
			return numericResultType(left, right), nil

		case "==", "!=", "<", "<=", ">", ">=":
			if !compatibleComparable(left, right) {
				return ast.Type{}, fmt.Errorf("cannot compare %s and %s", left.String(), right.String())
			}
			return ast.Type{Name: "bool"}, nil

		default:
			return ast.Type{}, fmt.Errorf("unsupported operator %q", expr.Op)
		}

	case *ast.CallExpr:
		return c.checkCall(scope, expr.Name, expr.Args)

	default:
		return ast.Type{}, fmt.Errorf("unsupported expression")
	}
}

func (c *Checker) checkFunction(fn *ast.FunctionDecl) error {
	scope := map[string]ast.Type{}
	for _, global := range c.globals {
		scope[global.Name] = global.Type
	}

	for _, param := range fn.Params {
		if _, exists := c.globals[param.Name]; exists {
			return fmt.Errorf("parameter %q conflicts with global", param.Name)
		}
		if _, exists := scope[param.Name]; exists {
			return fmt.Errorf("duplicate parameter %q", param.Name)
		}
		if err := c.checkParamType(param.Type); err != nil {
			return fmt.Errorf("parameter %q: %w", param.Name, err)
		}
		scope[param.Name] = param.Type
	}

	for _, local := range fn.Locals {
		if _, exists := c.globals[local.Name]; exists {
			return fmt.Errorf("local %q conflicts with global", local.Name)
		}
		if _, exists := scope[local.Name]; exists {
			return fmt.Errorf("duplicate local %q", local.Name)
		}
		if err := c.checkVarDecl(local, "local", scope, false); err != nil {
			return err
		}
		scope[local.Name] = local.Type
	}

	for i, returnType := range functionReturnTypes(fn) {
		if returnType.IsMem {
			return fmt.Errorf("return type %d: mem return values are not supported", i+1)
		}
		if err := c.checkType(returnType); err != nil {
			return fmt.Errorf("return type %d: %w", i+1, err)
		}
	}

	for _, stmt := range fn.Body {
		if err := c.checkStmt(scope, fn, stmt, 0); err != nil {
			return err
		}
	}

	return nil
}

func (c *Checker) checkVarDecl(v *ast.VarDecl, kind string, scope map[string]ast.Type, requireConst bool) error {
	if v.Type.IsMem && !v.HasAtAddress {
		return fmt.Errorf("%s %q: mem declarations require at address", kind, v.Name)
	}
	if v.HasAtAddress && !v.Type.IsMem {
		return fmt.Errorf("%s %q: at is only supported for mem declarations", kind, v.Name)
	}
	if err := c.checkType(v.Type); err != nil {
		return fmt.Errorf("%s %q: %w", kind, v.Name, err)
	}
	if v.HasAtAddress {
		if v.AtAddress < 0 || v.AtAddress > 65535 {
			return fmt.Errorf("%s %q: at address %d out of range 0..65535", kind, v.Name, v.AtAddress)
		}
		if v.AtAddress+v.Type.ArrayLen-1 > 65535 {
			return fmt.Errorf("%s %q: mem window exceeds address range", kind, v.Name)
		}
	}
	if v.Init != nil {
		if v.Type.IsMem {
			return fmt.Errorf("%s %q: mem declarations cannot have initializers", kind, v.Name)
		}
		if err := c.checkInitializer(scope, v.Type, v.Init, requireConst); err != nil {
			return fmt.Errorf("%s %q initializer: %w", kind, v.Name, err)
		}
	}
	return nil
}

func (c *Checker) checkInitializer(scope map[string]ast.Type, target ast.Type, init ast.Expr, requireConst bool) error {
	if target.IsArray {
		if target.Name == "char" {
			if str, ok := init.(*ast.StringExpr); ok {
				if len(str.Value) > target.ArrayLen {
					return fmt.Errorf("string length %d exceeds array capacity %d", len(str.Value), target.ArrayLen)
				}
				return nil
			}
		}

		lit, ok := init.(*ast.ArrayLiteralExpr)
		if !ok {
			return fmt.Errorf("array initializer must be an array literal")
		}
		if len(lit.Values) > target.ArrayLen {
			return fmt.Errorf("array literal has %d values but capacity is %d", len(lit.Values), target.ArrayLen)
		}

		elemType := ast.Type{Name: target.Name}
		for i, value := range lit.Values {
			if err := c.checkInitializer(scope, elemType, value, requireConst); err != nil {
				return fmt.Errorf("element %d: %w", i, err)
			}
		}
		return nil
	}

	if s, ok := c.structs[target.Name]; ok && !target.IsPointer && !target.IsMem {
		lit, ok := init.(*ast.StructLiteralExpr)
		if !ok {
			return fmt.Errorf("struct initializer for %s must be a struct literal", target.Name)
		}

		fieldTypes := map[string]ast.Type{}
		for _, field := range s.Fields {
			fieldTypes[field.Name] = field.Type
		}

		seen := map[string]bool{}
		for _, field := range lit.Fields {
			fieldType, ok := fieldTypes[field.Name]
			if !ok {
				return fmt.Errorf("type %s has no field %q", target.Name, field.Name)
			}
			if seen[field.Name] {
				return fmt.Errorf("field %q initialized more than once", field.Name)
			}
			seen[field.Name] = true

			if err := c.checkInitializer(scope, fieldType, field.Value, requireConst); err != nil {
				return fmt.Errorf("field %q: %w", field.Name, err)
			}
		}
		return nil
	}

	if _, ok := init.(*ast.ArrayLiteralExpr); ok {
		return fmt.Errorf("array literal cannot initialize %s", target.String())
	}
	if _, ok := init.(*ast.StructLiteralExpr); ok {
		return fmt.Errorf("struct literal cannot initialize %s", target.String())
	}
	if requireConst {
		if _, ok := c.constIntValue(init); !ok {
			return fmt.Errorf("global initializer must be a constant expression")
		}
	}

	valueType, err := c.checkExpr(scope, init)
	if err != nil {
		return err
	}
	if !sameType(target, valueType) && !c.canAssignExpr(target, valueType, init) {
		return fmt.Errorf("cannot initialize %s with %s", target.String(), valueType.String())
	}
	return c.checkAssignmentValue(target, init)
}

func (c *Checker) checkType(t ast.Type) error {
	return c.checkTypeFor(t, false)
}

func (c *Checker) checkParamType(t ast.Type) error {
	return c.checkTypeFor(t, true)
}

func (c *Checker) checkTypeFor(t ast.Type, allowPointer bool) error {
	if t.IsPointer {
		if !allowPointer {
			return fmt.Errorf("pointer types are only supported for parameters")
		}
		if t.IsArray || t.IsMem {
			return fmt.Errorf("pointer parameters cannot point to array or mem types")
		}
		if isScalarTypeName(t.Name) {
			return nil
		}
		if _, ok := c.structs[t.Name]; !ok {
			return fmt.Errorf("pointer parameters must point to scalar or struct types")
		}
		return nil
	}

	if t.IsMem {
		if t.Name != "mem" {
			return fmt.Errorf("invalid mem type")
		}
		if t.ArrayLen <= 0 {
			return fmt.Errorf("mem length must be positive")
		}
		return nil
	}

	switch t.Name {
	case "byte", "char", "bool", "int", "uint":
		if t.IsArray && t.ArrayLen <= 0 {
			return fmt.Errorf("array length must be positive")
		}
		return nil
	default:
		if _, ok := c.structs[t.Name]; ok {
			if t.IsArray && t.ArrayLen <= 0 {
				return fmt.Errorf("array length must be positive")
			}
			return nil
		}
		return fmt.Errorf("unknown type %q", t.Name)
	}
}

func (c *Checker) checkStmt(scope map[string]ast.Type, fn *ast.FunctionDecl, s ast.Stmt, loopDepth int) error {
	switch stmt := s.(type) {
	case *ast.AssignStmt:
		if len(stmt.Targets) > 0 {
			return c.checkMultiAssign(scope, stmt)
		}

		targetType, err := c.checkLValue(scope, stmt.Target)
		if err != nil {
			return err
		}
		if targetType.IsPointer {
			return fmt.Errorf("cannot assign to pointer parameter")
		}

		valueType, err := c.checkExpr(scope, stmt.Value)
		if err != nil {
			return err
		}

		if !sameType(targetType, valueType) && !c.canAssignExpr(targetType, valueType, stmt.Value) {
			return fmt.Errorf("cannot assign %s to %s", valueType.String(), targetType.String())
		}
		if err := c.checkAssignmentValue(targetType, stmt.Value); err != nil {
			return err
		}

	case *ast.CallStmt:
		if _, err := c.checkCall(scope, stmt.Name, stmt.Args); err != nil {
			return err
		}

	case *ast.WhileStmt:
		condType, err := c.checkExpr(scope, stmt.Cond)
		if err != nil {
			return err
		}
		if condType.IsPointer {
			return fmt.Errorf("pointer value cannot be used as condition")
		}

		for _, inner := range stmt.Body {
			if err := c.checkStmt(scope, fn, inner, loopDepth+1); err != nil {
				return err
			}
		}

	case *ast.ForStmt:
		if stmt.IsCounted {
			counterType, ok := scope[stmt.Counter]
			if !ok {
				return fmt.Errorf("unknown loop variable %q", stmt.Counter)
			}
			if counterType.IsArray || (counterType.Name != "byte" && counterType.Name != "int") {
				return fmt.Errorf("for loop variable must be byte or int")
			}

			startType, err := c.checkExpr(scope, stmt.Start)
			if err != nil {
				return err
			}
			if !isNumeric(startType) {
				return fmt.Errorf("for loop start must be numeric")
			}
			if !sameType(counterType, startType) && !canAssign(counterType, startType) {
				return fmt.Errorf("cannot assign for loop start %s to %s", startType.String(), counterType.String())
			}

			endType, err := c.checkExpr(scope, stmt.End)
			if err != nil {
				return err
			}
			if !isNumeric(endType) {
				return fmt.Errorf("for loop end must be numeric")
			}
			if !sameType(counterType, endType) && !canAssign(counterType, endType) {
				return fmt.Errorf("cannot assign for loop end %s to %s", endType.String(), counterType.String())
			}
		} else if stmt.Cond != nil {
			condType, err := c.checkExpr(scope, stmt.Cond)
			if err != nil {
				return err
			}
			if condType.IsPointer {
				return fmt.Errorf("pointer value cannot be used as condition")
			}
		}

		for _, inner := range stmt.Body {
			if err := c.checkStmt(scope, fn, inner, loopDepth+1); err != nil {
				return err
			}
		}

	case *ast.IfStmt:
		condType, err := c.checkExpr(scope, stmt.Cond)
		if err != nil {
			return err
		}
		if condType.IsPointer {
			return fmt.Errorf("pointer value cannot be used as condition")
		}

		for _, inner := range stmt.Then {
			if err := c.checkStmt(scope, fn, inner, loopDepth); err != nil {
				return err
			}
		}

		for _, inner := range stmt.Else {
			if err := c.checkStmt(scope, fn, inner, loopDepth); err != nil {
				return err
			}
		}

	case *ast.ReturnStmt:
		returnTypes := functionReturnTypes(fn)
		values := returnValues(stmt)

		if len(returnTypes) == 0 {
			if len(values) > 0 {
				return fmt.Errorf("function has no return type but returns a value")
			}
			return nil
		}

		if len(values) == 0 {
			return fmt.Errorf("function must return %s", formatTypeList(returnTypes))
		}

		if len(values) != len(returnTypes) {
			return fmt.Errorf("function returns %d values, return statement has %d", len(returnTypes), len(values))
		}

		for i, value := range values {
			valueType, err := c.checkExpr(scope, value)
			if err != nil {
				return err
			}

			returnType := returnTypes[i]
			if !sameType(returnType, valueType) && !c.canAssignExpr(returnType, valueType, value) {
				return fmt.Errorf("cannot return %s as value %d from function returning %s", valueType.String(), i+1, formatTypeList(returnTypes))
			}
			if err := c.checkAssignmentValue(returnType, value); err != nil {
				return err
			}
		}

	case *ast.BreakStmt:
		if loopDepth == 0 {
			return fmt.Errorf("break outside loop")
		}

	case *ast.ContinueStmt:
		if loopDepth == 0 {
			return fmt.Errorf("continue outside loop")
		}

	default:
		return fmt.Errorf("unsupported statement")
	}

	return nil
}

func (c *Checker) checkLValue(scope map[string]ast.Type, lv ast.LValue) (ast.Type, error) {
	switch v := lv.(type) {
	case *ast.VarLValue:
		t, ok := scope[v.Name]
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", v.Name)
		}
		if isScalarPointerType(t) {
			return pointerPointeeType(t), nil
		}
		return t, nil

	case *ast.IndexLValue:
		t, ok := scope[v.Name]
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", v.Name)
		}
		if t.IsMem {
			if err := c.checkIndexForType(scope, t, v.Index); err != nil {
				return ast.Type{}, err
			}
			return ast.Type{Name: "byte"}, nil
		}
		if !t.IsArray {
			return ast.Type{}, fmt.Errorf("%q is not an array", v.Name)
		}

		if err := c.checkIndexForType(scope, t, v.Index); err != nil {
			return ast.Type{}, err
		}

		return ast.Type{Name: t.Name}, nil

	case *ast.IndexFieldLValue:
		t, ok := scope[v.Name]
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", v.Name)
		}
		if !t.IsArray {
			return ast.Type{}, fmt.Errorf("%q is not an array", v.Name)
		}

		idxType, err := c.checkExpr(scope, v.Index)
		if err != nil {
			return ast.Type{}, err
		}
		if idxType.Name != "byte" && idxType.Name != "int" && idxType.Name != "uint" {
			return ast.Type{}, fmt.Errorf("array index must be byte, int, or uint")
		}

		elemType := ast.Type{Name: t.Name}
		return c.fieldType(elemType, v.Field)

	case *ast.FieldLValue:
		baseType, ok := scope[v.Base]
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", v.Base)
		}
		return c.fieldType(baseType, v.Field)

	default:
		return ast.Type{}, fmt.Errorf("unsupported assignment target")
	}
}

func (c *Checker) checkIndexForType(scope map[string]ast.Type, container ast.Type, index ast.Expr) error {
	idxType, err := c.checkExpr(scope, index)
	if err != nil {
		return err
	}
	if idxType.Name != "byte" && idxType.Name != "int" && idxType.Name != "uint" {
		return fmt.Errorf("index must be byte, int, or uint")
	}

	if !container.IsMem {
		return nil
	}

	v, ok := c.constIntValue(index)
	if !ok {
		return nil
	}
	if v < 0 || v >= container.ArrayLen {
		return fmt.Errorf("mem index %d out of range 0..%d", v, container.ArrayLen-1)
	}

	return nil
}

func (c *Checker) checkMultiAssign(scope map[string]ast.Type, stmt *ast.AssignStmt) error {
	call, ok := stmt.Value.(*ast.CallExpr)
	if !ok {
		return fmt.Errorf("multi-assignment requires a function call")
	}

	returnTypes, err := c.checkCallReturnTypes(scope, call.Name, call.Args)
	if err != nil {
		return err
	}

	if len(returnTypes) != len(stmt.Targets) {
		return fmt.Errorf("multi-assignment has %d targets but %s returns %d values", len(stmt.Targets), call.Name, len(returnTypes))
	}

	for i, target := range stmt.Targets {
		if target == "_" {
			continue
		}

		targetType, err := c.checkLValue(scope, &ast.VarLValue{Name: target})
		if err != nil {
			return err
		}
		if targetType.IsPointer {
			return fmt.Errorf("cannot assign to pointer parameter")
		}

		returnType := returnTypes[i]
		if !sameType(targetType, returnType) && !c.canAssignExpr(targetType, returnType, stmt.Value) {
			return fmt.Errorf("cannot assign return value %d (%s) to %s", i+1, returnType.String(), targetType.String())
		}
	}

	return nil
}

func (c *Checker) checkCall(scope map[string]ast.Type, name string, args []ast.Expr) (ast.Type, error) {
	switch name {
	case "ticks":
		if len(args) != 0 {
			return ast.Type{}, fmt.Errorf("ticks expects no arguments")
		}
		return ast.Type{Name: "int"}, nil

	case "elapsed":
		if len(args) != 1 {
			return ast.Type{}, fmt.Errorf("elapsed expects one argument")
		}

		t, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(t) {
			return ast.Type{}, fmt.Errorf("elapsed argument must be numeric")
		}

		return ast.Type{Name: "int"}, nil

	case "tickdue":
		if len(args) != 2 {
			return ast.Type{}, fmt.Errorf("tickdue expects two arguments")
		}

		lastType, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(lastType) {
			return ast.Type{}, fmt.Errorf("tickdue last argument must be numeric")
		}

		intervalType, err := c.checkExpr(scope, args[1])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(intervalType) {
			return ast.Type{}, fmt.Errorf("tickdue interval argument must be numeric")
		}

		return ast.Type{Name: "bool"}, nil
	case "key":
		if len(args) != 0 {
			return ast.Type{}, fmt.Errorf("key expects no arguments")
		}
		return ast.Type{Name: "char"}, nil

	case "waitkey":
		if len(args) != 0 {
			return ast.Type{}, fmt.Errorf("%s expects no arguments", name)
		}
		return ast.Type{Name: "char"}, nil

	case "readline":
		if len(args) != 3 {
			return ast.Type{}, fmt.Errorf("readline expects three arguments")
		}

		if _, ok := args[0].(*ast.StringExpr); ok {
			return ast.Type{}, fmt.Errorf("readline buffer must be char array variable")
		}

		bufferType, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !(bufferType.IsArray && bufferType.Name == "char") {
			return ast.Type{}, fmt.Errorf("readline buffer must be char array")
		}

		echoType, err := c.checkExpr(scope, args[1])
		if err != nil {
			return ast.Type{}, err
		}
		if echoType.Name != "bool" && !isNumeric(echoType) {
			return ast.Type{}, fmt.Errorf("readline echo argument must be bool or numeric")
		}

		maxType, err := c.checkExpr(scope, args[2])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(maxType) {
			return ast.Type{}, fmt.Errorf("readline max argument must be numeric")
		}

		return ast.Type{Name: "int"}, nil

	case "print":
		if len(args) != 1 {
			return ast.Type{}, fmt.Errorf("print expects one argument")
		}
		t, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if t.IsPointer {
			return ast.Type{}, fmt.Errorf("print argument cannot be pointer")
		}
		return ast.Type{}, nil

	case "poke":
		if len(args) != 2 {
			return ast.Type{}, fmt.Errorf("poke expects two arguments")
		}
		for _, arg := range args {
			t, err := c.checkExpr(scope, arg)
			if err != nil {
				return ast.Type{}, err
			}
			if !isNumeric(t) {
				return ast.Type{}, fmt.Errorf("poke arguments must be numeric")
			}
		}
		return ast.Type{}, nil

	case "peek":
		if len(args) != 1 {
			return ast.Type{}, fmt.Errorf("peek expects one argument")
		}
		t, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(t) {
			return ast.Type{}, fmt.Errorf("peek argument must be numeric")
		}
		return ast.Type{Name: "byte"}, nil

	case "joy":
		if len(args) != 1 {
			return ast.Type{}, fmt.Errorf("joy expects one argument")
		}
		t, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(t) {
			return ast.Type{}, fmt.Errorf("joy argument must be numeric")
		}
		return ast.Type{Name: "byte"}, nil

	case "cls":
		if len(args) != 0 {
			return ast.Type{}, fmt.Errorf("cls expects no arguments")
		}
		return ast.Type{}, nil

	case "asciifont":
		if len(args) != 0 {
			return ast.Type{}, fmt.Errorf("asciifont expects no arguments")
		}
		return ast.Type{}, nil

	case "toascii", "topetscii":
		if len(args) != 1 {
			return ast.Type{}, fmt.Errorf("%s expects one argument", name)
		}

		t, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !(t.IsArray && t.Name == "char") {
			return ast.Type{}, fmt.Errorf("%s expects char array", name)
		}

		return ast.Type{}, nil

	case "border", "background", "textcolor":
		if len(args) != 1 {
			return ast.Type{}, fmt.Errorf("%s expects one argument", name)
		}

		t, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}

		if !isNumeric(t) {
			return ast.Type{}, fmt.Errorf("%s argument must be numeric", name)
		}

		return ast.Type{}, nil

	case "gotoxy":
		if len(args) != 2 {
			return ast.Type{}, fmt.Errorf("gotoxy expects two arguments")
		}

		for _, arg := range args {
			t, err := c.checkExpr(scope, arg)
			if err != nil {
				return ast.Type{}, err
			}

			if !isNumeric(t) {
				return ast.Type{}, fmt.Errorf("gotoxy arguments must be numeric")
			}
		}

		return ast.Type{}, nil

	case "putraw", "putchar", "putcolor":
		if len(args) != 3 {
			return ast.Type{}, fmt.Errorf("%s expects three arguments", name)
		}

		for _, arg := range args {
			t, err := c.checkExpr(scope, arg)
			if err != nil {
				return ast.Type{}, err
			}

			if !isNumeric(t) {
				return ast.Type{}, fmt.Errorf("%s arguments must be numeric", name)
			}
		}

		return ast.Type{}, nil

	case "putcharcolor":
		if len(args) != 4 {
			return ast.Type{}, fmt.Errorf("putcharcolor expects four arguments")
		}

		for _, arg := range args {
			t, err := c.checkExpr(scope, arg)
			if err != nil {
				return ast.Type{}, err
			}

			if !isNumeric(t) {
				return ast.Type{}, fmt.Errorf("putcharcolor arguments must be numeric")
			}
		}

		return ast.Type{}, nil

	case "putstr":
		if len(args) != 3 {
			return ast.Type{}, fmt.Errorf("putstr expects three arguments")
		}

		xType, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(xType) {
			return ast.Type{}, fmt.Errorf("putstr x argument must be numeric")
		}

		yType, err := c.checkExpr(scope, args[1])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(yType) {
			return ast.Type{}, fmt.Errorf("putstr y argument must be numeric")
		}

		if _, ok := args[2].(*ast.StringExpr); ok {
			return ast.Type{}, nil
		}

		textType, err := c.checkExpr(scope, args[2])
		if err != nil {
			return ast.Type{}, err
		}

		if !(textType.IsArray && textType.Name == "char") {
			return ast.Type{}, fmt.Errorf("putstr expects string literal or char array")
		}

		return ast.Type{}, nil

	case "putstrcolor":
		if len(args) != 4 {
			return ast.Type{}, fmt.Errorf("putstrcolor expects four arguments")
		}

		xType, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(xType) {
			return ast.Type{}, fmt.Errorf("putstrcolor x argument must be numeric")
		}

		yType, err := c.checkExpr(scope, args[1])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(yType) {
			return ast.Type{}, fmt.Errorf("putstrcolor y argument must be numeric")
		}

		if _, ok := args[2].(*ast.StringExpr); !ok {
			textType, err := c.checkExpr(scope, args[2])
			if err != nil {
				return ast.Type{}, err
			}

			if !(textType.IsArray && textType.Name == "char") {
				return ast.Type{}, fmt.Errorf("putstrcolor expects string literal or char array")
			}
		}

		colorType, err := c.checkExpr(scope, args[3])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(colorType) {
			return ast.Type{}, fmt.Errorf("putstrcolor color argument must be numeric")
		}

		return ast.Type{}, nil
	case "netconnect":
		if len(args) != 2 {
			return ast.Type{}, fmt.Errorf("netconnect expects two arguments")
		}

		addrType, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !(addrType.IsArray && addrType.Name == "char") {
			return ast.Type{}, fmt.Errorf("netconnect address must be char array")
		}

		portType, err := c.checkExpr(scope, args[1])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(portType) {
			return ast.Type{}, fmt.Errorf("netconnect port must be numeric")
		}

		return ast.Type{Name: "bool"}, nil

	case "netbuffer":
		if len(args) != 1 {
			return ast.Type{}, fmt.Errorf("netbuffer expects one argument")
		}

		bufferType, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !(bufferType.IsArray && bufferType.Name == "byte") {
			return ast.Type{}, fmt.Errorf("netbuffer buffer must be byte array")
		}

		return ast.Type{}, nil

	case "netavailable":
		if len(args) != 0 {
			return ast.Type{}, fmt.Errorf("netavailable expects no arguments")
		}
		return ast.Type{Name: "int"}, nil

	case "netread":
		if len(args) != 3 {
			return ast.Type{}, fmt.Errorf("netread expects three arguments")
		}

		bufferType, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !(bufferType.IsArray && (bufferType.Name == "byte" || bufferType.Name == "char")) {
			return ast.Type{}, fmt.Errorf("netread buffer must be byte array or char array")
		}

		maxType, err := c.checkExpr(scope, args[1])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(maxType) {
			return ast.Type{}, fmt.Errorf("netread max argument must be numeric")
		}

		timeoutType, err := c.checkExpr(scope, args[2])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(timeoutType) {
			return ast.Type{}, fmt.Errorf("netread timeout argument must be numeric")
		}

		return ast.Type{Name: "int"}, nil

	case "netreadlf":
		if len(args) != 3 {
			return ast.Type{}, fmt.Errorf("netreadlf expects three arguments")
		}

		bufferType, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !(bufferType.IsArray && (bufferType.Name == "byte" || bufferType.Name == "char")) {
			return ast.Type{}, fmt.Errorf("netreadlf buffer must be byte array or char array")
		}

		maxType, err := c.checkExpr(scope, args[1])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(maxType) {
			return ast.Type{}, fmt.Errorf("netreadlf max argument must be numeric")
		}

		timeoutType, err := c.checkExpr(scope, args[2])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(timeoutType) {
			return ast.Type{}, fmt.Errorf("netreadlf timeout argument must be numeric")
		}

		return ast.Type{Name: "bool"}, nil

	case "netwrite":
		if len(args) != 2 {
			return ast.Type{}, fmt.Errorf("netwrite expects two arguments")
		}

		bufferType, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !(bufferType.IsArray && (bufferType.Name == "byte" || bufferType.Name == "char")) {
			return ast.Type{}, fmt.Errorf("netwrite buffer must be byte array or char array")
		}

		lenType, err := c.checkExpr(scope, args[1])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(lenType) {
			return ast.Type{}, fmt.Errorf("netwrite len argument must be numeric")
		}

		return ast.Type{Name: "int"}, nil

	case "netclose":
		if len(args) != 0 {
			return ast.Type{}, fmt.Errorf("netclose expects no arguments")
		}
		return ast.Type{}, nil

	case "netconnected":
		if len(args) != 0 {
			return ast.Type{}, fmt.Errorf("netconnected expects no arguments")
		}
		return ast.Type{Name: "bool"}, nil
	case "fileopen":
		if len(args) != 3 {
			return ast.Type{}, fmt.Errorf("fileopen expects three arguments")
		}

		nameType, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !(nameType.IsArray && nameType.Name == "char") {
			return ast.Type{}, fmt.Errorf("fileopen name must be char array")
		}

		modeType, err := c.checkExpr(scope, args[1])
		if err != nil {
			return ast.Type{}, err
		}
		if !(modeType.IsArray && modeType.Name == "char") {
			return ast.Type{}, fmt.Errorf("fileopen mode must be char array")
		}

		deviceType, err := c.checkExpr(scope, args[2])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(deviceType) {
			return ast.Type{}, fmt.Errorf("fileopen device must be numeric")
		}

		return ast.Type{Name: "byte"}, nil

	case "fileclose":
		if len(args) != 1 {
			return ast.Type{}, fmt.Errorf("fileclose expects one argument")
		}

		handleType, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(handleType) {
			return ast.Type{}, fmt.Errorf("fileclose handle must be numeric")
		}

		return ast.Type{}, nil

	case "fileread":
		if len(args) != 3 {
			return ast.Type{}, fmt.Errorf("fileread expects three arguments")
		}

		handleType, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(handleType) {
			return ast.Type{}, fmt.Errorf("fileread handle must be numeric")
		}

		bufferType, err := c.checkExpr(scope, args[1])
		if err != nil {
			return ast.Type{}, err
		}
		if !(bufferType.IsArray && (bufferType.Name == "byte" || bufferType.Name == "char")) {
			return ast.Type{}, fmt.Errorf("fileread buffer must be byte array or char array")
		}

		maxType, err := c.checkExpr(scope, args[2])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(maxType) {
			return ast.Type{}, fmt.Errorf("fileread max argument must be numeric")
		}

		return ast.Type{Name: "int"}, nil

	case "filewrite":
		if len(args) != 3 {
			return ast.Type{}, fmt.Errorf("filewrite expects three arguments")
		}

		handleType, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(handleType) {
			return ast.Type{}, fmt.Errorf("filewrite handle must be numeric")
		}

		bufferType, err := c.checkExpr(scope, args[1])
		if err != nil {
			return ast.Type{}, err
		}
		if !(bufferType.IsArray && (bufferType.Name == "byte" || bufferType.Name == "char")) {
			return ast.Type{}, fmt.Errorf("filewrite buffer must be byte array or char array")
		}

		lenType, err := c.checkExpr(scope, args[2])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(lenType) {
			return ast.Type{}, fmt.Errorf("filewrite len argument must be numeric")
		}

		return ast.Type{Name: "int"}, nil

	case "fileload":
		if len(args) != 3 {
			return ast.Type{}, fmt.Errorf("fileload expects three arguments")
		}

		nameType, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !(nameType.IsArray && nameType.Name == "char") {
			return ast.Type{}, fmt.Errorf("fileload name must be char array")
		}

		bufferType, err := c.checkExpr(scope, args[1])
		if err != nil {
			return ast.Type{}, err
		}
		if !(bufferType.IsArray && (bufferType.Name == "byte" || bufferType.Name == "char")) {
			return ast.Type{}, fmt.Errorf("fileload buffer must be byte array or char array")
		}

		deviceType, err := c.checkExpr(scope, args[2])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(deviceType) {
			return ast.Type{}, fmt.Errorf("fileload device must be numeric")
		}

		return ast.Type{Name: "int"}, nil

	case "filesave":
		if len(args) != 4 {
			return ast.Type{}, fmt.Errorf("filesave expects four arguments")
		}

		nameType, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !(nameType.IsArray && nameType.Name == "char") {
			return ast.Type{}, fmt.Errorf("filesave name must be char array")
		}

		bufferType, err := c.checkExpr(scope, args[1])
		if err != nil {
			return ast.Type{}, err
		}
		if !(bufferType.IsArray && (bufferType.Name == "byte" || bufferType.Name == "char")) {
			return ast.Type{}, fmt.Errorf("filesave buffer must be byte array or char array")
		}

		lenType, err := c.checkExpr(scope, args[2])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(lenType) {
			return ast.Type{}, fmt.Errorf("filesave len argument must be numeric")
		}

		deviceType, err := c.checkExpr(scope, args[3])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(deviceType) {
			return ast.Type{}, fmt.Errorf("filesave device must be numeric")
		}

		return ast.Type{Name: "int"}, nil
	case "sound_init":
		if len(args) != 1 {
			return ast.Type{}, fmt.Errorf("sound_init expects one argument")
		}
		if err := c.checkSoundByteArrayArg(scope, args[0], "sound_init pool"); err != nil {
			return ast.Type{}, err
		}
		return ast.Type{}, nil

	case "sound_reset":
		if len(args) != 0 {
			return ast.Type{}, fmt.Errorf("sound_reset expects no arguments")
		}
		return ast.Type{}, nil

	case "sound_load":
		if err := c.checkSoundLoadArgs(scope, args); err != nil {
			return ast.Type{}, err
		}
		return ast.Type{}, fmt.Errorf("sound_load returns multiple values; use multi-assignment")

	case "sound_play":
		if len(args) != 4 {
			return ast.Type{}, fmt.Errorf("sound_play expects four arguments")
		}
		for i, arg := range args {
			t, err := c.checkExpr(scope, arg)
			if err != nil {
				return ast.Type{}, err
			}
			if !isNumeric(t) {
				return ast.Type{}, fmt.Errorf("sound_play argument %d must be numeric", i+1)
			}
		}
		return ast.Type{Name: "int"}, nil

	case "sound_stop":
		if len(args) != 1 {
			return ast.Type{}, fmt.Errorf("%s expects one argument", name)
		}
		t, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(t) {
			return ast.Type{}, fmt.Errorf("%s handle must be numeric", name)
		}
		return ast.Type{}, nil

	case "sound_stop_voices":
		if len(args) != 1 {
			return ast.Type{}, fmt.Errorf("%s expects one argument", name)
		}
		t, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !isNumeric(t) {
			return ast.Type{}, fmt.Errorf("%s voices must be numeric", name)
		}
		return ast.Type{}, nil

	case "sound_num", "sound_memfree":
		if len(args) != 0 {
			return ast.Type{}, fmt.Errorf("%s expects no arguments", name)
		}
		return ast.Type{Name: "int"}, nil

	case "itoa":
		if len(args) != 1 {
			return ast.Type{}, fmt.Errorf("itoa expects one argument")
		}

		t, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}

		if !isNumeric(t) {
			return ast.Type{}, fmt.Errorf("itoa argument must be numeric")
		}

		return ast.Type{
			Name:     "char",
			IsArray:  true,
			ArrayLen: 6,
		}, nil
	case "itox":
		if len(args) != 1 {
			return ast.Type{}, fmt.Errorf("itox expects one argument")
		}

		t, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}

		if !isNumeric(t) {
			return ast.Type{}, fmt.Errorf("itox argument must be numeric")
		}

		width := 2
		if t.Name == "int" {
			width = 4
		}

		return ast.Type{
			Name:     "char",
			IsArray:  true,
			ArrayLen: width,
		}, nil
	case "len", "size":
		if len(args) != 1 {
			return ast.Type{}, fmt.Errorf("%s expects one argument", name)
		}

		t, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}

		if !t.IsArray && !t.IsMem {
			return ast.Type{}, fmt.Errorf("%s expects array or mem", name)
		}

		return ast.Type{Name: "int"}, nil

	case "append":
		if len(args) != 2 {
			return ast.Type{}, fmt.Errorf("append expects two arguments")
		}

		dst, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}

		if !dst.IsArray {
			return ast.Type{}, fmt.Errorf("append destination must be array")
		}

		valueType, err := c.checkExpr(scope, args[1])
		if err != nil {
			return ast.Type{}, err
		}

		if dst.Name == "char" && valueType.Name == "char" && valueType.IsArray {
			return ast.Type{}, nil
		}

		elemType := ast.Type{Name: dst.Name}
		if !sameType(elemType, valueType) && !c.canAssignExpr(elemType, valueType, args[1]) {
			return ast.Type{}, fmt.Errorf("append value cannot be %s for %s[]", valueType.String(), dst.Name)
		}
		if err := c.checkAssignmentValue(elemType, args[1]); err != nil {
			return ast.Type{}, err
		}

		return ast.Type{}, nil

	case "copy":
		if len(args) != 2 {
			return ast.Type{}, fmt.Errorf("copy expects two arguments")
		}

		dst, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}

		src, err := c.checkExpr(scope, args[1])
		if err != nil {
			return ast.Type{}, err
		}

		if !dst.IsArray {
			return ast.Type{}, fmt.Errorf("copy destination must be array")
		}

		if !src.IsArray {
			return ast.Type{}, fmt.Errorf("copy source must be array")
		}

		if dst.Name != src.Name {
			return ast.Type{}, fmt.Errorf("copy requires arrays with same element type")
		}

		if src.ArrayLen > dst.ArrayLen {
			return ast.Type{}, fmt.Errorf("source array does not fit destination")
		}

		return ast.Type{}, nil

	case "fill":
		if len(args) != 2 {
			return ast.Type{}, fmt.Errorf("fill expects two arguments")
		}

		dst, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}

		if !dst.IsArray {
			return ast.Type{}, fmt.Errorf("fill destination must be array")
		}

		valueType, err := c.checkExpr(scope, args[1])
		if err != nil {
			return ast.Type{}, err
		}

		elemType := ast.Type{Name: dst.Name}
		if !sameType(elemType, valueType) && !c.canAssignExpr(elemType, valueType, args[1]) {
			return ast.Type{}, fmt.Errorf("fill value cannot be %s for %s[]", valueType.String(), dst.Name)
		}
		if err := c.checkAssignmentValue(elemType, args[1]); err != nil {
			return ast.Type{}, err
		}

		return ast.Type{}, nil

	case "clear":
		if len(args) != 1 {
			return ast.Type{}, fmt.Errorf("clear expects one argument")
		}

		dst, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}

		if !dst.IsArray {
			return ast.Type{}, fmt.Errorf("clear expects array")
		}

		return ast.Type{}, nil
	}

	fn, ok := c.functions[name]
	if !ok {
		return ast.Type{}, fmt.Errorf("unknown function %q", name)
	}

	returnTypes, err := c.checkUserCall(scope, fn, name, args)
	if err != nil {
		return ast.Type{}, err
	}

	if len(returnTypes) > 1 {
		return ast.Type{}, fmt.Errorf("function %s returns multiple values; use multi-assignment", name)
	}
	if len(returnTypes) == 0 {
		return ast.Type{}, nil
	}
	return returnTypes[0], nil
}

func (c *Checker) checkCallReturnTypes(scope map[string]ast.Type, name string, args []ast.Expr) ([]ast.Type, error) {
	if name == "sound_load" {
		if err := c.checkSoundLoadArgs(scope, args); err != nil {
			return nil, err
		}
		return []ast.Type{{Name: "uint"}, {Name: "int"}}, nil
	}

	if fn, ok := c.functions[name]; ok {
		return c.checkUserCall(scope, fn, name, args)
	}

	t, err := c.checkCall(scope, name, args)
	if err != nil {
		return nil, err
	}
	if t.Name == "" {
		return nil, nil
	}
	return []ast.Type{t}, nil
}

func (c *Checker) checkUserCall(scope map[string]ast.Type, fn *ast.FunctionDecl, name string, args []ast.Expr) ([]ast.Type, error) {
	if len(args) != len(fn.Params) {
		return nil, fmt.Errorf("function %s expects %d args, got %d", name, len(fn.Params), len(args))
	}

	for i, arg := range args {
		paramType := fn.Params[i].Type
		if paramType.IsPointer {
			if err := c.checkPointerArgument(scope, paramType, arg); err != nil {
				return nil, fmt.Errorf("argument %d to %s: %w", i+1, name, err)
			}
			continue
		}

		argType, err := c.checkExpr(scope, arg)
		if err != nil {
			return nil, err
		}

		if paramType.IsArray {
			switch arg.(type) {
			case *ast.IdentExpr, *ast.FieldExpr, *ast.IndexFieldExpr, *ast.CallExpr:
			default:
				return nil, fmt.Errorf("argument %d to %s: array parameter requires array storage", i+1, name)
			}
			if !sameType(paramType, argType) {
				return nil, fmt.Errorf("argument %d to %s: cannot pass %s as %s", i+1, name, argType.String(), paramType.String())
			}
			continue
		}

		if paramType.IsMem {
			if _, ok := arg.(*ast.IdentExpr); !ok {
				return nil, fmt.Errorf("argument %d to %s: mem parameter requires mem storage", i+1, name)
			}
			if !sameType(paramType, argType) {
				return nil, fmt.Errorf("argument %d to %s: cannot pass %s as %s", i+1, name, argType.String(), paramType.String())
			}
			continue
		}

		if !sameType(paramType, argType) && !c.canAssignExpr(paramType, argType, arg) {
			return nil, fmt.Errorf("argument %d to %s: cannot pass %s as %s", i+1, name, argType.String(), paramType.String())
		}
		if err := c.checkAssignmentValue(paramType, arg); err != nil {
			return nil, fmt.Errorf("argument %d to %s: %w", i+1, name, err)
		}
	}

	return functionReturnTypes(fn), nil
}

func (c *Checker) checkSoundLoadArgs(scope map[string]ast.Type, args []ast.Expr) error {
	if len(args) != 2 {
		return fmt.Errorf("sound_load expects two arguments")
	}
	if err := c.checkSoundByteArrayArg(scope, args[0], "sound_load data"); err != nil {
		return err
	}

	kindType, err := c.checkExpr(scope, args[1])
	if err != nil {
		return err
	}
	if !isNumeric(kindType) {
		return fmt.Errorf("sound_load kind must be numeric")
	}

	return nil
}

func (c *Checker) checkSoundByteArrayArg(scope map[string]ast.Type, arg ast.Expr, name string) error {
	t, err := c.checkExpr(scope, arg)
	if err != nil {
		return err
	}
	if !(t.IsArray && t.Name == "byte") {
		return fmt.Errorf("%s must be byte array", name)
	}
	return nil
}

func (c *Checker) checkPointerArgument(scope map[string]ast.Type, paramType ast.Type, arg ast.Expr) error {
	u, ok := arg.(*ast.UnaryExpr)
	if !ok || u.Op != "&" {
		return fmt.Errorf("pointer parameter requires explicit &")
	}

	targetType, err := c.checkPointerArgumentTarget(scope, u.Expr)
	if err != nil {
		return err
	}

	want := pointerPointeeType(paramType)
	if !sameType(want, targetType) {
		return fmt.Errorf("cannot pass &%s as %s", targetType.String(), paramType.String())
	}

	return nil
}

func (c *Checker) checkPointerArgumentTarget(scope map[string]ast.Type, e ast.Expr) (ast.Type, error) {
	switch expr := e.(type) {
	case *ast.IdentExpr:
		t, ok := scope[expr.Name]
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", expr.Name)
		}
		if t.IsPointer {
			return ast.Type{}, fmt.Errorf("cannot take address of pointer parameter")
		}
		if t.IsArray {
			return ast.Type{}, fmt.Errorf("cannot pass array storage as pointer parameter")
		}
		if t.IsMem {
			return ast.Type{}, fmt.Errorf("cannot pass mem storage as pointer parameter")
		}
		return t, nil

	case *ast.IndexExpr:
		t, ok := scope[expr.Name]
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", expr.Name)
		}
		if t.IsMem {
			if err := c.checkIndexForType(scope, t, expr.Index); err != nil {
				return ast.Type{}, err
			}
			return ast.Type{Name: "byte"}, nil
		}
		if !t.IsArray {
			return ast.Type{}, fmt.Errorf("%q is not an array", expr.Name)
		}

		if err := c.checkIndexForType(scope, t, expr.Index); err != nil {
			return ast.Type{}, err
		}

		return ast.Type{Name: t.Name}, nil

	default:
		return ast.Type{}, fmt.Errorf("can only take address of variables or array elements")
	}
}

func (c *Checker) fieldType(base ast.Type, field string) (ast.Type, error) {
	if base.IsPointer {
		base = ast.Type{Name: base.Name}
	}

	if base.IsArray {
		return ast.Type{}, fmt.Errorf("cannot access field %q on array type %s", field, base.String())
	}

	s, ok := c.structs[base.Name]
	if !ok {
		return ast.Type{}, fmt.Errorf("type %s has no fields", base.Name)
	}

	for _, f := range s.Fields {
		if f.Name == field {
			return f.Type, nil
		}
	}

	return ast.Type{}, fmt.Errorf("type %s has no field %q", base.Name, field)
}

func (c *Checker) checkAddressOf(scope map[string]ast.Type, e ast.Expr) (ast.Type, error) {
	switch expr := e.(type) {
	case *ast.IdentExpr:
		t, ok := scope[expr.Name]
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", expr.Name)
		}
		if t.IsPointer {
			return ast.Type{}, fmt.Errorf("cannot take address of pointer parameter")
		}
		if t.IsMem {
			return ast.Type{Name: "uint"}, nil
		}
		if t.IsArray {
			return ast.Type{Name: "uint"}, nil
		}
		if _, ok := c.structs[t.Name]; !ok {
			return ast.Type{Name: "uint"}, nil
		}
		return ast.Type{Name: t.Name, IsPointer: true}, nil

	case *ast.IndexExpr:
		t, ok := scope[expr.Name]
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", expr.Name)
		}
		if t.IsMem {
			if err := c.checkIndexForType(scope, t, expr.Index); err != nil {
				return ast.Type{}, err
			}
			return ast.Type{Name: "uint"}, nil
		}
		if !t.IsArray {
			return ast.Type{}, fmt.Errorf("%q is not an array", expr.Name)
		}

		if err := c.checkIndexForType(scope, t, expr.Index); err != nil {
			return ast.Type{}, err
		}

		if _, ok := c.structs[t.Name]; !ok {
			return ast.Type{Name: "uint"}, nil
		}
		return ast.Type{Name: t.Name, IsPointer: true}, nil

	default:
		return ast.Type{}, fmt.Errorf("can only take address of variables or array elements")
	}
}

func sameType(a, b ast.Type) bool {
	return a.Name == b.Name && a.IsArray == b.IsArray && a.IsMem == b.IsMem && a.ArrayLen == b.ArrayLen && a.IsPointer == b.IsPointer
}

func functionReturnTypes(fn *ast.FunctionDecl) []ast.Type {
	if len(fn.ReturnTypes) > 0 {
		return fn.ReturnTypes
	}
	if fn.ReturnType.Name != "" {
		return []ast.Type{fn.ReturnType}
	}
	return nil
}

func returnValues(stmt *ast.ReturnStmt) []ast.Expr {
	if len(stmt.Values) > 0 {
		return stmt.Values
	}
	if stmt.Value != nil {
		return []ast.Expr{stmt.Value}
	}
	return nil
}

func formatTypeList(types []ast.Type) string {
	if len(types) == 0 {
		return "no values"
	}

	parts := make([]string, len(types))
	for i, typ := range types {
		parts[i] = typ.String()
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func pointerPointeeType(t ast.Type) ast.Type {
	return ast.Type{Name: t.Name}
}

func isScalarPointerType(t ast.Type) bool {
	return t.IsPointer && !t.IsArray && !t.IsMem && isScalarTypeName(t.Name)
}

func isScalarTypeName(name string) bool {
	switch name {
	case "byte", "char", "bool", "int", "uint":
		return true
	default:
		return false
	}
}

func numericResultType(a, b ast.Type) ast.Type {
	if a.Name == "uint" || b.Name == "uint" {
		return ast.Type{Name: "uint"}
	}
	if a.Name == "int" || b.Name == "int" {
		return ast.Type{Name: "int"}
	}
	return ast.Type{Name: "byte"}
}

func (c *Checker) canAssignExpr(dst, src ast.Type, expr ast.Expr) bool {
	if canAssign(dst, src) {
		return true
	}

	if dst.Name == "uint" && !dst.IsArray && !dst.IsMem && !dst.IsPointer && src.Name == "int" && !src.IsArray && !src.IsMem && !src.IsPointer {
		_, ok := c.constIntValue(expr)
		return ok
	}

	return false
}

func canAssign(dst, src ast.Type) bool {
	if dst.IsPointer {
		return false
	}

	if src.IsPointer {
		return dst.Name == "uint" && !dst.IsArray
	}

	if dst.IsMem || src.IsMem {
		return false
	}

	if dst.IsArray || src.IsArray {
		if dst.Name == "char" && dst.IsArray && src.Name == "char" && src.IsArray {
			return src.ArrayLen <= dst.ArrayLen
		}
		return false
	}

	if dst.Name == "int" && (src.Name == "byte" || src.Name == "char" || src.Name == "bool") {
		return true
	}

	if dst.Name == "uint" && (src.Name == "byte" || src.Name == "char" || src.Name == "bool") {
		return true
	}

	if dst.Name == "byte" && (src.Name == "int" || src.Name == "uint") {
		// Allow explicit low-byte truncation. Warnings/casts may be added later.
		return true
	}

	if dst.Name == "char" && (src.Name == "int" || src.Name == "uint") {
		return true
	}

	if dst.Name == "bool" && (src.Name == "int" || src.Name == "uint") {
		// bool is represented as one byte: 0 = false, non-zero = true.
		return true
	}

	return false
}

func isNumeric(t ast.Type) bool {
	return !t.IsPointer && !t.IsArray && !t.IsMem && isScalarTypeName(t.Name)
}

func compatibleComparable(a, b ast.Type) bool {
	if a.IsArray || b.IsArray || a.IsMem || b.IsMem {
		return false
	}
	return isNumeric(a) && isNumeric(b)
}

func (c *Checker) checkAssignmentValue(dst ast.Type, expr ast.Expr) error {
	if dst.Name != "uint" || dst.IsArray || dst.IsMem || dst.IsPointer {
		return nil
	}

	v, ok := c.constIntValue(expr)
	if !ok {
		return nil
	}

	if v < 0 || v > 65535 {
		return fmt.Errorf("uint value %d out of range 0..65535", v)
	}

	return nil
}

func (c *Checker) constIntValue(expr ast.Expr) (int, bool) {
	switch e := expr.(type) {
	case *ast.NumberExpr:
		var v int
		if _, err := fmt.Sscanf(e.Value, "%d", &v); err != nil {
			return 0, false
		}
		return v, true

	case *ast.CharExpr:
		var v int
		if _, err := fmt.Sscanf(e.Value, "%d", &v); err != nil {
			return 0, false
		}
		return v, true

	case *ast.BoolExpr:
		if e.Value {
			return 1, true
		}
		return 0, true

	case *ast.IdentExpr:
		cn, ok := c.constants[e.Name]
		if !ok {
			return 0, false
		}
		var v int
		if _, err := fmt.Sscanf(cn.Value, "%d", &v); err != nil {
			return 0, false
		}
		return v, true

	case *ast.UnaryExpr:
		v, ok := c.constIntValue(e.Expr)
		if !ok {
			return 0, false
		}
		switch e.Op {
		case "-":
			return -v, true
		case "!":
			if v == 0 {
				return 1, true
			}
			return 0, true
		default:
			return 0, false
		}

	case *ast.BinaryExpr:
		left, ok := c.constIntValue(e.Left)
		if !ok {
			return 0, false
		}
		right, ok := c.constIntValue(e.Right)
		if !ok {
			return 0, false
		}

		switch e.Op {
		case "+":
			return left + right, true
		case "-":
			return left - right, true
		case "*":
			return left * right, true
		case "/":
			if right == 0 {
				return 0, true
			}
			return left / right, true
		case "%":
			if right == 0 {
				return left, true
			}
			return left % right, true
		case "&":
			return left & right, true
		case "|":
			return left | right, true
		case "^":
			return left ^ right, true
		case "<<":
			if right < 0 {
				right = 0
			}
			return left << right, true
		case ">>":
			if right < 0 {
				right = 0
			}
			return left >> right, true
		default:
			return 0, false
		}
	}

	return 0, false
}
