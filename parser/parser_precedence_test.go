package parser

import (
	"fmt"
	"strings"
	"testing"

	"peddle/ast"
	"peddle/lexer"
)

func TestParserExpressionPrecedence(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want string
	}{
		{
			name: "grouped bitwise before comparison",
			expr: "(j & 4) == 0",
			want: "((j & 4) == 0)",
		},
		{
			name: "bitwise and before comparison without parentheses",
			expr: "j & 4 == 0",
			want: "((j & 4) == 0)",
		},
		{
			name: "explicit comparison inside bitwise and",
			expr: "j & (4 == 0)",
			want: "(j & (4 == 0))",
		},
		{
			name: "product before sum",
			expr: "a + b * c",
			want: "(a + (b * c))",
		},
		{
			name: "left associative sum",
			expr: "a + b - c",
			want: "((a + b) - c)",
		},
		{
			name: "left associative product",
			expr: "a * b / c % d",
			want: "(((a * b) / c) % d)",
		},
		{
			name: "sum before shift",
			expr: "a + b << c",
			want: "((a + b) << c)",
		},
		{
			name: "shift before bitwise and",
			expr: "a & b << c",
			want: "(a & (b << c))",
		},
		{
			name: "bitwise and before xor before or",
			expr: "a | b ^ c & d",
			want: "(a | (b ^ (c & d)))",
		},
		{
			name: "relational before equality",
			expr: "a < b == c",
			want: "((a < b) == c)",
		},
		{
			name: "grouped comparison before equality",
			expr: "(a < b) == true",
			want: "((a < b) == true)",
		},
		{
			name: "call before product",
			expr: "foo(1 + 2) * 3",
			want: "(foo((1 + 2)) * 3)",
		},
		{
			name: "index before product",
			expr: "arr[1 + 2] * 3",
			want: "(arr[(1 + 2)] * 3)",
		},
		{
			name: "field before equality",
			expr: "player.hp == 10",
			want: "(player.hp == 10)",
		},
		{
			name: "indexed field before equality",
			expr: "players[0].hp == 10",
			want: "(players[0].hp == 10)",
		},
		{
			name: "unary minus before product",
			expr: "-a * b",
			want: "((-a) * b)",
		},
		{
			name: "unary bang before equality",
			expr: "!done == true",
			want: "((!done) == true)",
		},
		{
			name: "address-of before call argument",
			expr: "addr(&player)",
			want: "addr((&player))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAssignedExpressionString(t, tt.expr)
			if got != tt.want {
				t.Fatalf("wrong parse for %q:\n  got  %s\n  want %s", tt.expr, got, tt.want)
			}
		})
	}
}

func parseAssignedExpressionString(t *testing.T, expr string) string {
	t.Helper()

	src := fmt.Sprintf(`
fn main() {
    var result int
    result = %s
}
`, expr)

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

	return exprString(assign.Value)
}

func exprString(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.IdentExpr:
		return v.Name

	case *ast.IndexExpr:
		return fmt.Sprintf("%s[%s]", v.Name, exprString(v.Index))

	case *ast.FieldExpr:
		return fmt.Sprintf("%s.%s", v.Base, v.Field)

	case *ast.IndexFieldExpr:
		return fmt.Sprintf("%s[%s].%s", v.Name, exprString(v.Index), v.Field)

	case *ast.NumberExpr:
		return v.Value

	case *ast.CharExpr:
		return "'" + v.Value + "'"

	case *ast.BoolExpr:
		if v.Value {
			return "true"
		}
		return "false"

	case *ast.StringExpr:
		return `"` + v.Value + `"`

	case *ast.UnaryExpr:
		return fmt.Sprintf("(%s%s)", v.Op, exprString(v.Expr))

	case *ast.BinaryExpr:
		return fmt.Sprintf("(%s %s %s)", exprString(v.Left), v.Op, exprString(v.Right))

	case *ast.CallExpr:
		args := make([]string, 0, len(v.Args))
		for _, arg := range v.Args {
			args = append(args, exprString(arg))
		}
		return fmt.Sprintf("%s(%s)", v.Name, strings.Join(args, ", "))

	default:
		return fmt.Sprintf("<unknown %T>", e)
	}
}
