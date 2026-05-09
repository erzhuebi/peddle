package parser

import (
	"testing"

	"peddle/ast"
	"peddle/lexer"
)

func parseExprFromMain(t *testing.T, src string) ast.Expr {
	t.Helper()

	l := lexer.New(src)
	p := New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	if len(prog.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Functions))
	}

	fn := prog.Functions[0]
	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(fn.Body))
	}

	assign, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}

	return assign.Value
}

func TestParseUnaryMinus(t *testing.T) {
	expr := parseExprFromMain(t, `
fn main() {
    var x: int
    x = -1
}
`)

	u, ok := expr.(*ast.UnaryExpr)
	if !ok {
		t.Fatalf("expected UnaryExpr, got %T", expr)
	}

	if u.Op != "-" {
		t.Fatalf("got op %q, want -", u.Op)
	}
}

func TestParseUnaryBang(t *testing.T) {
	expr := parseExprFromMain(t, `
fn main() {
    var x: bool
    x = !0
}
`)

	u, ok := expr.(*ast.UnaryExpr)
	if !ok {
		t.Fatalf("expected UnaryExpr, got %T", expr)
	}

	if u.Op != "!" {
		t.Fatalf("got op %q, want !", u.Op)
	}
}

func parseProgramForTest(t *testing.T, src string) *ast.Program {
	t.Helper()

	l := lexer.New(src)
	p := New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	return prog
}

func TestParseIfElse(t *testing.T) {
	prog := parseProgramForTest(t, `
fn main() {
    var x: byte

    if x == 0 {
        x = 1
    } else {
        x = 2
    }
}
`)

	fn := prog.Functions[0]
	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(fn.Body))
	}

	stmt, ok := fn.Body[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", fn.Body[0])
	}

	if len(stmt.Then) != 1 {
		t.Fatalf("expected 1 then stmt, got %d", len(stmt.Then))
	}

	if len(stmt.Else) != 1 {
		t.Fatalf("expected 1 else stmt, got %d", len(stmt.Else))
	}
}

func TestParseWhile(t *testing.T) {
	prog := parseProgramForTest(t, `
fn main() {
    var i: byte

    while i < 10 {
        i = i + 1
    }
}
`)

	fn := prog.Functions[0]
	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(fn.Body))
	}

	_, ok := fn.Body[0].(*ast.WhileStmt)
	if !ok {
		t.Fatalf("expected WhileStmt, got %T", fn.Body[0])
	}
}

func TestParseUserFunctionCall(t *testing.T) {
	expr := parseExprFromMain(t, `
fn main() {
    var x: int
    x = add(1, 2)
}
`)

	call, ok := expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", expr)
	}

	if call.Name != "add" {
		t.Fatalf("got call name %q, want add", call.Name)
	}

	if len(call.Args) != 2 {
		t.Fatalf("got %d args, want 2", len(call.Args))
	}
}

func TestParseArrayIndexAssignment(t *testing.T) {
	prog := parseProgramForTest(t, `
fn main() {
    var a: int[4]

    a[0] = 1
}
`)

	fn := prog.Functions[0]
	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(fn.Body))
	}

	assign, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}

	_, ok = assign.Target.(*ast.IndexLValue)
	if !ok {
		t.Fatalf("expected IndexLValue, got %T", assign.Target)
	}
}

func TestParseFunctionReturnArrow(t *testing.T) {
	prog := parseProgramForTest(t, `
fn add(a: int, b: int) -> int {
    return a + b
}
`)

	if len(prog.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Functions))
	}

	fn := prog.Functions[0]

	if fn.Name != "add" {
		t.Fatalf("got function name %q, want add", fn.Name)
	}

	if fn.ReturnType.Name != "int" {
		t.Fatalf("got return type %q, want int", fn.ReturnType.Name)
	}
}

func TestParseComments(t *testing.T) {
	input := `
fn main() {
    # this is a comment
    var x: int
    x = 1 # trailing comment
}
`

	l := lexer.New(input)
	p := New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	if len(prog.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Functions))
	}

	fn := prog.Functions[0]

	if len(fn.Locals) != 1 {
		t.Fatalf("expected 1 local, got %d", len(fn.Locals))
	}

	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(fn.Body))
	}
}

func TestParseMultipleVarDecls(t *testing.T) {
	prog := parseProgramForTest(t, `
fn main() {
    var x, y, z: int
    var a, b: byte[16]

    x = 1
}
`)

	if len(prog.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Functions))
	}

	fn := prog.Functions[0]

	if len(fn.Locals) != 5 {
		t.Fatalf("expected 5 locals, got %d", len(fn.Locals))
	}

	expected := []struct {
		name     string
		typeName string
		isArray  bool
		arrayLen int
	}{
		{name: "x", typeName: "int"},
		{name: "y", typeName: "int"},
		{name: "z", typeName: "int"},
		{name: "a", typeName: "byte", isArray: true, arrayLen: 16},
		{name: "b", typeName: "byte", isArray: true, arrayLen: 16},
	}

	for i, want := range expected {
		got := fn.Locals[i]

		if got.Name != want.name {
			t.Fatalf("local %d: got name %q, want %q", i, got.Name, want.name)
		}

		if got.Type.Name != want.typeName {
			t.Fatalf("local %d: got type %q, want %q", i, got.Type.Name, want.typeName)
		}

		if got.Type.IsArray != want.isArray {
			t.Fatalf("local %d: got IsArray %v, want %v", i, got.Type.IsArray, want.isArray)
		}

		if got.Type.ArrayLen != want.arrayLen {
			t.Fatalf("local %d: got ArrayLen %d, want %d", i, got.Type.ArrayLen, want.arrayLen)
		}
	}
}

func TestParserStage1OperatorPrecedence(t *testing.T) {
	input := `
fn main() {
    var a, b, c, d, e, x: int
    x = a | b ^ c & d + e * 2
}
`

	l := lexer.New(input)
	p := New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	if len(prog.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Functions))
	}

	fn := prog.Functions[0]

	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(fn.Body))
	}

	stmt, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}

	expr, ok := stmt.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", stmt.Value)
	}

	if expr.Op != "|" {
		t.Fatalf("expected top-level operator |, got %q", expr.Op)
	}

	right, ok := expr.Right.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected right side BinaryExpr, got %T", expr.Right)
	}

	if right.Op != "^" {
		t.Fatalf("expected second-level operator ^, got %q", right.Op)
	}

	rightRight, ok := right.Right.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected right-right BinaryExpr, got %T", right.Right)
	}

	if rightRight.Op != "&" {
		t.Fatalf("expected third-level operator &, got %q", rightRight.Op)
	}

	add, ok := rightRight.Right.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected add BinaryExpr, got %T", rightRight.Right)
	}

	if add.Op != "+" {
		t.Fatalf("expected fourth-level operator +, got %q", add.Op)
	}

	mul, ok := add.Right.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected multiply BinaryExpr, got %T", add.Right)
	}

	if mul.Op != "*" {
		t.Fatalf("expected deepest operator *, got %q", mul.Op)
	}
}

func TestParserStage2ShiftOperatorPrecedence(t *testing.T) {
	input := `
fn main() {
    var a, b, c, d, x: int
    x = a | b & c << d + 1
}
`

	l := lexer.New(input)
	p := New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	fn := prog.Functions[0]

	stmt, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}

	expr, ok := stmt.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", stmt.Value)
	}

	if expr.Op != "|" {
		t.Fatalf("expected top-level operator |, got %q", expr.Op)
	}

	andExpr, ok := expr.Right.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected right side BinaryExpr, got %T", expr.Right)
	}

	if andExpr.Op != "&" {
		t.Fatalf("expected second-level operator &, got %q", andExpr.Op)
	}

	shiftExpr, ok := andExpr.Right.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected shift BinaryExpr, got %T", andExpr.Right)
	}

	if shiftExpr.Op != "<<" {
		t.Fatalf("expected shift operator <<, got %q", shiftExpr.Op)
	}

	addExpr, ok := shiftExpr.Right.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected add BinaryExpr as shift count, got %T", shiftExpr.Right)
	}

	if addExpr.Op != "+" {
		t.Fatalf("expected + to bind tighter than <<, got %q", addExpr.Op)
	}
}

func TestParserStage3DivisionModuloPrecedence(t *testing.T) {
	input := `
fn main() {
    var a, b, c, d, x: int
    x = a + b / c % d
}
`

	l := lexer.New(input)
	p := New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	fn := prog.Functions[0]

	stmt, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}

	expr, ok := stmt.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", stmt.Value)
	}

	if expr.Op != "+" {
		t.Fatalf("expected top-level operator +, got %q", expr.Op)
	}

	modExpr, ok := expr.Right.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected right side BinaryExpr, got %T", expr.Right)
	}

	if modExpr.Op != "%" {
		t.Fatalf("expected right side operator %%, got %q", modExpr.Op)
	}

	divExpr, ok := modExpr.Left.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected division BinaryExpr, got %T", modExpr.Left)
	}

	if divExpr.Op != "/" {
		t.Fatalf("expected division operator /, got %q", divExpr.Op)
	}
}

func TestParseConstDeclarations(t *testing.T) {
	input := `
const BORDER = $d020
const BG = 0xd021
const MASK = %1111_0000
const BIG = 1_000

fn main() {
    var x: int
    x = BIG
}
`

	l := lexer.New(input)
	p := New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	if len(prog.Consts) != 4 {
		t.Fatalf("expected 4 consts, got %d", len(prog.Consts))
	}

	tests := []struct {
		name  string
		value string
	}{
		{"BORDER", "53280"},
		{"BG", "53281"},
		{"MASK", "240"},
		{"BIG", "1000"},
	}

	for i, tt := range tests {
		if prog.Consts[i].Name != tt.name {
			t.Fatalf("const[%d] name wrong. expected=%q got=%q", i, tt.name, prog.Consts[i].Name)
		}

		if prog.Consts[i].Value != tt.value {
			t.Fatalf("const[%d] value wrong. expected=%q got=%q", i, tt.value, prog.Consts[i].Value)
		}
	}

	if len(prog.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Functions))
	}
}

func TestParseConstBeforeStructAndFunction(t *testing.T) {
	input := `
const DEFAULT_HP = 100

struct Player {
    hp: byte
}

fn main() {
    var p: Player
    p.hp = DEFAULT_HP
}
`

	l := lexer.New(input)
	p := New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	if len(prog.Consts) != 1 {
		t.Fatalf("expected 1 const, got %d", len(prog.Consts))
	}

	if len(prog.Structs) != 1 {
		t.Fatalf("expected 1 struct, got %d", len(prog.Structs))
	}

	if len(prog.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Functions))
	}
}
