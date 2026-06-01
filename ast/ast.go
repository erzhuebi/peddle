package ast

type Program struct {
	Consts    []*ConstDecl
	Structs   []*StructDecl
	Functions []*FunctionDecl
}

type ConstDecl struct {
	Name  string
	Value string
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
	Name        string
	Params      []Param
	ReturnType  Type
	ReturnTypes []Type
	Locals      []*VarDecl
	Body        []Stmt
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
	Name      string
	ArrayLen  int
	IsArray   bool
	IsPointer bool
}

func (t Type) String() string {
	s := t.Name
	if t.IsArray {
		s += "[]"
	}
	if t.IsPointer {
		s = "*" + s
	}
	return s
}

type Stmt interface {
	stmtNode()
}

type AssignStmt struct {
	Target  LValue
	Targets []string
	Value   Expr
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

type ForStmt struct {
	Cond Expr
	Body []Stmt

	IsCounted bool
	Counter   string
	Start     Expr
	End       Expr
}

func (*ForStmt) stmtNode() {}

type IfStmt struct {
	Cond Expr
	Then []Stmt
	Else []Stmt
}

func (*IfStmt) stmtNode() {}

type ReturnStmt struct {
	Value  Expr
	Values []Expr
}

func (*ReturnStmt) stmtNode() {}

type BreakStmt struct{}

func (*BreakStmt) stmtNode() {}

type ContinueStmt struct{}

func (*ContinueStmt) stmtNode() {}

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

type IndexFieldLValue struct {
	Name  string
	Index Expr
	Field string
}

func (*IndexFieldLValue) lvalueNode() {}

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

type IndexFieldExpr struct {
	Name  string
	Index Expr
	Field string
}

func (*IndexFieldExpr) exprNode() {}

type NumberExpr struct {
	Value string
}

func (*NumberExpr) exprNode() {}

type CharExpr struct {
	Value string
}

func (*CharExpr) exprNode() {}

type BoolExpr struct {
	Value bool
}

func (*BoolExpr) exprNode() {}

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
