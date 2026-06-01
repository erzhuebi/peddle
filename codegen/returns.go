package codegen

import "peddle/ast"

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
