package sema

import (
	"fmt"

	"peddle/ast"
)

type Checker struct {
	functions map[string]*ast.FunctionDecl
	structs   map[string]*ast.StructDecl
}

func New() *Checker {
	return &Checker{
		functions: map[string]*ast.FunctionDecl{},
		structs:   map[string]*ast.StructDecl{},
	}
}

func (c *Checker) Check(p *ast.Program) error {
	for _, s := range p.Structs {
		if _, exists := c.structs[s.Name]; exists {
			return fmt.Errorf("duplicate struct %q", s.Name)
		}
		c.structs[s.Name] = s
	}

	for _, fn := range p.Functions {
		if _, exists := c.functions[fn.Name]; exists {
			return fmt.Errorf("duplicate function %q", fn.Name)
		}
		c.functions[fn.Name] = fn
	}

	if _, ok := c.functions["main"]; !ok {
		return fmt.Errorf("missing main function")
	}

	for _, fn := range p.Functions {
		if err := c.checkFunction(fn); err != nil {
			return fmt.Errorf("function %s: %w", fn.Name, err)
		}
	}

	return nil
}

func (c *Checker) checkFunction(fn *ast.FunctionDecl) error {
	scope := map[string]ast.Type{}

	for _, param := range fn.Params {
		if _, exists := scope[param.Name]; exists {
			return fmt.Errorf("duplicate parameter %q", param.Name)
		}
		if err := c.checkType(param.Type); err != nil {
			return fmt.Errorf("parameter %q: %w", param.Name, err)
		}
		scope[param.Name] = param.Type
	}

	for _, local := range fn.Locals {
		if _, exists := scope[local.Name]; exists {
			return fmt.Errorf("duplicate local %q", local.Name)
		}
		if err := c.checkType(local.Type); err != nil {
			return fmt.Errorf("local %q: %w", local.Name, err)
		}
		scope[local.Name] = local.Type
	}

	if fn.ReturnType.Name != "" {
		if err := c.checkType(fn.ReturnType); err != nil {
			return fmt.Errorf("return type: %w", err)
		}
	}

	for _, stmt := range fn.Body {
		if err := c.checkStmt(scope, fn, stmt); err != nil {
			return err
		}
	}

	return nil
}

func (c *Checker) checkType(t ast.Type) error {
	switch t.Name {
	case "byte", "char", "bool", "int":
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

func (c *Checker) checkStmt(scope map[string]ast.Type, fn *ast.FunctionDecl, s ast.Stmt) error {
	switch stmt := s.(type) {
	case *ast.AssignStmt:
		targetType, err := c.checkLValue(scope, stmt.Target)
		if err != nil {
			return err
		}

		valueType, err := c.checkExpr(scope, stmt.Value)
		if err != nil {
			return err
		}

		if !sameType(targetType, valueType) && !canAssign(targetType, valueType) {
			return fmt.Errorf("cannot assign %s to %s", valueType.String(), targetType.String())
		}

	case *ast.CallStmt:
		if _, err := c.checkCall(scope, stmt.Name, stmt.Args); err != nil {
			return err
		}

	case *ast.WhileStmt:
		_, err := c.checkExpr(scope, stmt.Cond)
		if err != nil {
			return err
		}

		for _, inner := range stmt.Body {
			if err := c.checkStmt(scope, fn, inner); err != nil {
				return err
			}
		}

	case *ast.IfStmt:
		_, err := c.checkExpr(scope, stmt.Cond)
		if err != nil {
			return err
		}

		for _, inner := range stmt.Then {
			if err := c.checkStmt(scope, fn, inner); err != nil {
				return err
			}
		}

		for _, inner := range stmt.Else {
			if err := c.checkStmt(scope, fn, inner); err != nil {
				return err
			}
		}

	case *ast.ReturnStmt:
		if fn.ReturnType.Name == "" {
			if stmt.Value != nil {
				return fmt.Errorf("function has no return type but returns a value")
			}
			return nil
		}

		if stmt.Value == nil {
			return fmt.Errorf("function must return %s", fn.ReturnType.String())
		}

		valueType, err := c.checkExpr(scope, stmt.Value)
		if err != nil {
			return err
		}

		if !sameType(fn.ReturnType, valueType) && !canAssign(fn.ReturnType, valueType) {
			return fmt.Errorf("cannot return %s from function returning %s", valueType.String(), fn.ReturnType.String())
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
		return t, nil

	case *ast.IndexLValue:
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
		if idxType.Name != "byte" && idxType.Name != "int" {
			return ast.Type{}, fmt.Errorf("array index must be byte or int")
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
		if idxType.Name != "byte" && idxType.Name != "int" {
			return ast.Type{}, fmt.Errorf("array index must be byte or int")
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

func (c *Checker) checkExpr(scope map[string]ast.Type, e ast.Expr) (ast.Type, error) {
	switch expr := e.(type) {
	case *ast.IdentExpr:
		t, ok := scope[expr.Name]
		if !ok {
			return ast.Type{}, fmt.Errorf("unknown variable %q", expr.Name)
		}
		return t, nil

	case *ast.IndexExpr:
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
		if idxType.Name != "byte" && idxType.Name != "int" {
			return ast.Type{}, fmt.Errorf("array index must be byte or int")
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
		if idxType.Name != "byte" && idxType.Name != "int" {
			return ast.Type{}, fmt.Errorf("array index must be byte or int")
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

	case *ast.StringExpr:
		return ast.Type{Name: "char", IsArray: true, ArrayLen: len(expr.Value)}, nil

	case *ast.UnaryExpr:
		t, err := c.checkExpr(scope, expr.Expr)
		if err != nil {
			return ast.Type{}, err
		}

		switch expr.Op {
		case "-":
			if !isNumeric(t) {
				return ast.Type{}, fmt.Errorf("unary - requires numeric operand")
			}
			if t.Name == "int" {
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
		case "+", "-":
			if !isNumeric(left) || !isNumeric(right) {
				return ast.Type{}, fmt.Errorf("operator %s requires numeric operands", expr.Op)
			}
			if left.Name == "int" || right.Name == "int" {
				return ast.Type{Name: "int"}, nil
			}
			return ast.Type{Name: "byte"}, nil

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

func (c *Checker) checkCall(scope map[string]ast.Type, name string, args []ast.Expr) (ast.Type, error) {
	switch name {
	case "print":
		if len(args) != 1 {
			return ast.Type{}, fmt.Errorf("print expects one argument")
		}
		_, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
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

	case "strlen":
		if len(args) != 1 {
			return ast.Type{}, fmt.Errorf("strlen expects one argument")
		}
		t, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}
		if !(t.IsArray && t.Name == "char") {
			return ast.Type{}, fmt.Errorf("strlen expects char array")
		}
		return ast.Type{Name: "int"}, nil

	case "strcpy", "stradd":
		if len(args) != 2 {
			return ast.Type{}, fmt.Errorf("%s expects two arguments", name)
		}

		dst, err := c.checkExpr(scope, args[0])
		if err != nil {
			return ast.Type{}, err
		}

		src, err := c.checkExpr(scope, args[1])
		if err != nil {
			return ast.Type{}, err
		}

		if !(dst.IsArray && dst.Name == "char") {
			return ast.Type{}, fmt.Errorf("%s destination must be char array", name)
		}

		if !(src.IsArray && src.Name == "char") {
			return ast.Type{}, fmt.Errorf("%s source must be char array", name)
		}

		if name == "strcpy" && src.ArrayLen > dst.ArrayLen {
			return ast.Type{}, fmt.Errorf("source string does not fit destination")
		}

		return ast.Type{}, nil
	}

	fn, ok := c.functions[name]
	if !ok {
		return ast.Type{}, fmt.Errorf("unknown function %q", name)
	}

	if len(args) != len(fn.Params) {
		return ast.Type{}, fmt.Errorf("function %s expects %d args, got %d", name, len(fn.Params), len(args))
	}

	for i, arg := range args {
		argType, err := c.checkExpr(scope, arg)
		if err != nil {
			return ast.Type{}, err
		}

		paramType := fn.Params[i].Type
		if !sameType(paramType, argType) && !canAssign(paramType, argType) {
			return ast.Type{}, fmt.Errorf("argument %d to %s: cannot pass %s as %s", i+1, name, argType.String(), paramType.String())
		}
	}

	return fn.ReturnType, nil
}

func (c *Checker) fieldType(base ast.Type, field string) (ast.Type, error) {
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

func sameType(a, b ast.Type) bool {
	return a.Name == b.Name && a.IsArray == b.IsArray && a.ArrayLen == b.ArrayLen
}

func canAssign(dst, src ast.Type) bool {
	if dst.IsArray || src.IsArray {
		if dst.Name == "char" && dst.IsArray && src.Name == "char" && src.IsArray {
			return src.ArrayLen <= dst.ArrayLen
		}
		return false
	}

	if dst.Name == "int" && (src.Name == "byte" || src.Name == "char" || src.Name == "bool") {
		return true
	}

	if dst.Name == "byte" && src.Name == "int" {
		// Allow explicit low-byte truncation. Warnings/casts may be added later.
		return true
	}

	if dst.Name == "char" && src.Name == "int" {
		return true
	}

	if dst.Name == "bool" && src.Name == "int" {
		// bool is represented as one byte: 0 = false, non-zero = true.
		return true
	}

	return false
}

func isNumeric(t ast.Type) bool {
	return !t.IsArray && (t.Name == "byte" || t.Name == "char" || t.Name == "int" || t.Name == "bool")
}

func compatibleComparable(a, b ast.Type) bool {
	if a.IsArray || b.IsArray {
		return false
	}
	return isNumeric(a) && isNumeric(b)
}
