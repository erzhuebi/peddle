package lexer

import (
	"strconv"
	"strings"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte

	line   int
	column int
}

func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

func (l *Lexer) Input() string {
	return l.input
}

func (l *Lexer) readChar() {
	if l.ch == '\n' {
		l.line++
		l.column = 0
	}

	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}

	l.position = l.readPosition
	l.readPosition++
	l.column++
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()
	l.skipComment()
	l.skipWhitespace()

	line := l.line
	column := l.column

	var tok Token

	switch l.ch {
	case '(':
		tok = l.newToken(LPAREN, string(l.ch), line, column)
	case ')':
		tok = l.newToken(RPAREN, string(l.ch), line, column)
	case '{':
		tok = l.newToken(LBRACE, string(l.ch), line, column)
	case '}':
		tok = l.newToken(RBRACE, string(l.ch), line, column)
	case '[':
		tok = l.newToken(LBRACK, string(l.ch), line, column)
	case ']':
		tok = l.newToken(RBRACK, string(l.ch), line, column)
	case ',':
		tok = l.newToken(COMMA, string(l.ch), line, column)
	case ':':
		tok = l.newToken(COLON, string(l.ch), line, column)
	case '.':
		tok = l.newToken(DOT, string(l.ch), line, column)
	case '+':
		tok = l.newToken(PLUS, string(l.ch), line, column)
	case '*':
		tok = l.newToken(ASTERISK, string(l.ch), line, column)
	case '/':
		tok = l.newToken(SLASH, string(l.ch), line, column)
	case '&':
		tok = l.newToken(AMP, string(l.ch), line, column)
	case '|':
		tok = l.newToken(PIPE, string(l.ch), line, column)
	case '^':
		tok = l.newToken(CARET, string(l.ch), line, column)
	case '$':
		lit := l.readDollarHexNumber()
		return Token{Type: NUMBER, Literal: lit, Line: line, Column: column}

	case '%':
		if l.peekChar() == '0' || l.peekChar() == '1' {
			lit := l.readBinaryNumber()
			return Token{Type: NUMBER, Literal: lit, Line: line, Column: column}
		}
		tok = l.newToken(PERCENT, string(l.ch), line, column)

	case '-':
		if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(ARROW, string(ch)+string(l.ch), line, column)
		} else {
			tok = l.newToken(MINUS, string(l.ch), line, column)
		}

	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(EQ, string(ch)+string(l.ch), line, column)
		} else {
			tok = l.newToken(ASSIGN, string(l.ch), line, column)
		}

	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(NEQ, string(ch)+string(l.ch), line, column)
		} else {
			tok = l.newToken(BANG, string(l.ch), line, column)
		}

	case '<':
		if l.peekChar() == '<' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(SHL, string(ch)+string(l.ch), line, column)
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(LTE, string(ch)+string(l.ch), line, column)
		} else {
			tok = l.newToken(LT, string(l.ch), line, column)
		}

	case '>':
		if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(SHR, string(ch)+string(l.ch), line, column)
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(GTE, string(ch)+string(l.ch), line, column)
		} else {
			tok = l.newToken(GT, string(l.ch), line, column)
		}

	case '"':
		lit := l.readString()
		return Token{
			Type:    STRING,
			Literal: lit,
			Line:    line,
			Column:  column,
		}

	case 0:
		tok = l.newToken(EOF, "", line, column)

	default:
		if isLetter(l.ch) {
			lit := l.readIdentifier()
			return Token{
				Type:    LookupIdent(lit),
				Literal: lit,
				Line:    line,
				Column:  column,
			}
		}

		if isDigit(l.ch) {
			lit := l.readNumber()
			return Token{
				Type:    NUMBER,
				Literal: lit,
				Line:    line,
				Column:  column,
			}
		}

		tok = l.newToken(ILLEGAL, string(l.ch), line, column)
	}

	l.readChar()
	return tok
}

func (l *Lexer) newToken(t TokenType, lit string, line int, column int) Token {
	return Token{
		Type:    t,
		Literal: lit,
		Line:    line,
		Column:  column,
	}
}

func (l *Lexer) skipWhitespace() {
	for {
		for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
			l.readChar()
		}

		if l.ch == '#' {
			for l.ch != '\n' && l.ch != 0 {
				l.readChar()
			}
			continue
		}

		break
	}
}

func (l *Lexer) skipComment() {
	if l.ch == '#' {
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) readIdentifier() string {
	pos := l.position

	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}

	return l.input[pos:l.position]
}

func (l *Lexer) readNumber() string {
	if l.ch == '0' && (l.peekChar() == 'x' || l.peekChar() == 'X') {
		l.readChar()
		l.readChar()

		var hex strings.Builder
		for isHexDigit(l.ch) || l.ch == '_' {
			if l.ch != '_' {
				hex.WriteByte(l.ch)
			}
			l.readChar()
		}

		n, err := strconv.ParseInt(hex.String(), 16, 64)
		if err != nil {
			return "0"
		}
		return strconv.FormatInt(n, 10)
	}

	var dec strings.Builder
	for isDigit(l.ch) || l.ch == '_' {
		if l.ch != '_' {
			dec.WriteByte(l.ch)
		}
		l.readChar()
	}

	n, err := strconv.ParseInt(dec.String(), 10, 64)
	if err != nil {
		return "0"
	}
	return strconv.FormatInt(n, 10)
}

func (l *Lexer) readDollarHexNumber() string {
	l.readChar()

	var hex strings.Builder
	for isHexDigit(l.ch) || l.ch == '_' {
		if l.ch != '_' {
			hex.WriteByte(l.ch)
		}
		l.readChar()
	}

	n, err := strconv.ParseInt(hex.String(), 16, 64)
	if err != nil {
		return "0"
	}
	return strconv.FormatInt(n, 10)
}

func (l *Lexer) readBinaryNumber() string {
	l.readChar()

	var bin strings.Builder
	for l.ch == '0' || l.ch == '1' || l.ch == '_' {
		if l.ch != '_' {
			bin.WriteByte(l.ch)
		}
		l.readChar()
	}

	n, err := strconv.ParseInt(bin.String(), 2, 64)
	if err != nil {
		return "0"
	}
	return strconv.FormatInt(n, 10)
}

func (l *Lexer) readString() string {
	var out strings.Builder

	l.readChar()

	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()

			switch l.ch {
			case 'n':
				out.WriteByte('\n')
			case '"':
				out.WriteByte('"')
			case '\\':
				out.WriteByte('\\')
			case '0':
				out.WriteByte(0)
			default:
				out.WriteByte('\\')
				if l.ch != 0 {
					out.WriteByte(l.ch)
				}
			}

			if l.ch != 0 {
				l.readChar()
			}
			continue
		}

		out.WriteByte(l.ch)
		l.readChar()
	}

	if l.ch == '"' {
		l.readChar()
	}

	return out.String()
}

func isLetter(ch byte) bool {
	return ch == '_' ||
		('a' <= ch && ch <= 'z') ||
		('A' <= ch && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return isDigit(ch) ||
		('a' <= ch && ch <= 'f') ||
		('A' <= ch && ch <= 'F')
}
