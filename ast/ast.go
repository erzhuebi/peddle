package ast

type Program struct {
	Structs   []*StructDecl
	Functions []*FunctionDecl
}

type StructDecl struct {
	Name   string
	Fields []FieldDecl
}

type FieldDecl struct {
	Name string
	Type Type
}

type FunctionDecl struct {
	Name       string
	Params     []Param
	ReturnType Type
	Locals     []*VarDecl
	Body       []Stmt
}

type Param struct {
	Name string
	Type Type
}

type VarDecl struct {
	Name string
	Type Type
}

type Type struct {
	Name     string
	ArrayLen int
	IsArray  bool
}

func (t Type) String() string {
	if t.IsArray {
		return t.Name + "[]"
	}
	return t.Name
}

type Stmt interface {
	stmtNode()
}

type AssignStmt struct {
	Target LValue
	Value  Expr
}

func (*AssignStmt) stmtNode() {}

type CallStmt struct {
	Name string
	Args []Expr
}

func (*CallStmt) stmtNode() {}

type WhileStmt struct {
	Cond Expr
	Body []Stmt
}

func (*WhileStmt) stmtNode() {}

type IfStmt struct {
	Cond Expr
	Then []Stmt
	Else []Stmt
}

func (*IfStmt) stmtNode() {}

type ReturnStmt struct {
	Value Expr
}

func (*ReturnStmt) stmtNode() {}

type LValue interface {
	lvalueNode()
}

type VarLValue struct {
	Name string
}

func (*VarLValue) lvalueNode() {}

type IndexLValue struct {
	Name  string
	Index Expr
}

func (*IndexLValue) lvalueNode() {}

type FieldLValue struct {
	Base  string
	Field string
}

func (*FieldLValue) lvalueNode() {}

type Expr interface {
	exprNode()
}

type IdentExpr struct {
	Name string
}

func (*IdentExpr) exprNode() {}

type IndexExpr struct {
	Name  string
	Index Expr
}

func (*IndexExpr) exprNode() {}

type FieldExpr struct {
	Base  string
	Field string
}

func (*FieldExpr) exprNode() {}

type NumberExpr struct {
	Value string
}

func (*NumberExpr) exprNode() {}

type StringExpr struct {
	Value string
}

func (*StringExpr) exprNode() {}

type UnaryExpr struct {
	Op   string
	Expr Expr
}

func (*UnaryExpr) exprNode() {}

type BinaryExpr struct {
	Left  Expr
	Op    string
	Right Expr
}

func (*BinaryExpr) exprNode() {}

type CallExpr struct {
	Name string
	Args []Expr
}

func (*CallExpr) exprNode() {}
