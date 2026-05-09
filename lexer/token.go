package lexer

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT  = "IDENT"
	NUMBER = "NUMBER"
	STRING = "STRING"

	FN     = "FN"
	VAR    = "VAR"
	ARROW  = "->"
	WHILE  = "WHILE"
	IF     = "IF"
	ELSE   = "ELSE"
	RETURN = "RETURN"
	STRUCT = "STRUCT"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"
	LBRACK = "["
	RBRACK = "]"

	COMMA  = ","
	COLON  = ":"
	DOT    = "."
	ASSIGN = "="

	PLUS     = "+"
	MINUS    = "-"
	ASTERISK = "*"
	AMP      = "&"
	PIPE     = "|"
	CARET    = "^"
	SHL      = "<<"
	SHR      = ">>"
	BANG     = "!"

	EQ  = "=="
	NEQ = "!="
	LT  = "<"
	LTE = "<="
	GT  = ">"
	GTE = ">="
)

var keywords = map[string]TokenType{
	"fn":     FN,
	"var":    VAR,
	"while":  WHILE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"struct": STRUCT,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
