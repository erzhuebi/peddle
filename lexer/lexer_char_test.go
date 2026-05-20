package lexer

import "testing"

func TestLexerCharLiterals(t *testing.T) {
	input := `
'A'
' '
'\n'
'\''
'\\'
'\0'
`

	tests := []struct {
		wantType    TokenType
		wantLiteral string
	}{
		{CHAR, "65"},
		{CHAR, "32"},
		{CHAR, "10"},
		{CHAR, "39"},
		{CHAR, "92"},
		{CHAR, "0"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.wantType {
			t.Fatalf("token %d: got type %q, want %q", i, tok.Type, tt.wantType)
		}

		if tok.Literal != tt.wantLiteral {
			t.Fatalf("token %d: got literal %q, want %q", i, tok.Literal, tt.wantLiteral)
		}
	}
}

func TestLexerCharLiteralInsideCall(t *testing.T) {
	input := `putchar(0, 0, 'P')`

	tests := []struct {
		wantType    TokenType
		wantLiteral string
	}{
		{IDENT, "putchar"},
		{LPAREN, "("},
		{NUMBER, "0"},
		{COMMA, ","},
		{NUMBER, "0"},
		{COMMA, ","},
		{CHAR, "80"},
		{RPAREN, ")"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.wantType {
			t.Fatalf("token %d: got type %q, want %q", i, tok.Type, tt.wantType)
		}

		if tok.Literal != tt.wantLiteral {
			t.Fatalf("token %d: got literal %q, want %q", i, tok.Literal, tt.wantLiteral)
		}
	}
}
