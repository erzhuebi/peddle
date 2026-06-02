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
	CHAR   = "CHAR"

	CONST    = "CONST"
	FN       = "FN"
	VAR      = "VAR"
	ARROW    = "->"
	WHILE    = "WHILE"
	FOR      = "FOR"
	TO       = "TO"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	STRUCT   = "STRUCT"
	BREAK    = "BREAK"
	CONTINUE = "CONTINUE"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	AT       = "AT"

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
	SLASH    = "/"
	PERCENT  = "%"
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
	"const":    CONST,
	"fn":       FN,
	"var":      VAR,
	"while":    WHILE,
	"for":      FOR,
	"to":       TO,
	"if":       IF,
	"else":     ELSE,
	"return":   RETURN,
	"struct":   STRUCT,
	"break":    BREAK,
	"continue": CONTINUE,
	"true":     TRUE,
	"false":    FALSE,
	"at":       AT,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
