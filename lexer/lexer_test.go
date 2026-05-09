package lexer

import "testing"

func TestLexerOperators(t *testing.T) {
	input := `! != == < <= > >= + - -> =`

	expected := []TokenType{
		BANG,
		NEQ,
		EQ,
		LT,
		LTE,
		GT,
		GTE,
		PLUS,
		MINUS,
		ARROW,
		ASSIGN,
		EOF,
	}

	l := New(input)

	for i, want := range expected {
		tok := l.NextToken()
		if tok.Type != want {
			t.Fatalf("token %d: got %q, want %q", i, tok.Type, want)
		}
	}
}

func TestLexerStage1Operators(t *testing.T) {
	input := `a * b & c | d ^ e`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{IDENT, "a"},
		{ASTERISK, "*"},
		{IDENT, "b"},
		{AMP, "&"},
		{IDENT, "c"},
		{PIPE, "|"},
		{IDENT, "d"},
		{CARET, "^"},
		{IDENT, "e"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] token type wrong. expected=%q got=%q", i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] literal wrong. expected=%q got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}
