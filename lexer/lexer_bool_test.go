package lexer

import "testing"

func TestLexerBoolLiterals(t *testing.T) {
	input := `true false`

	tests := []struct {
		typ TokenType
		lit string
	}{
		{TRUE, "true"},
		{FALSE, "false"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.typ {
			t.Fatalf("test %d: token type wrong: got %q, want %q", i, tok.Type, tt.typ)
		}

		if tok.Literal != tt.lit {
			t.Fatalf("test %d: literal wrong: got %q, want %q", i, tok.Literal, tt.lit)
		}
	}
}
