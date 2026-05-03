package lexer

import "testing"

func TestLexerOperators(t *testing.T) {
	input := `! != == < <= > >= + - =`

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
