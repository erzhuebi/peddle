package parser

import (
	"fmt"
	"strconv"
	"strings"

	"peddle/ast"
	"peddle/lexer"
)

type Parser struct {
	l *lexer.Lexer

	cur  lexer.Token
	peek lexer.Token

	errors []string

	lines []string
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:     l,
		lines: strings.Split(l.Input(), "\n"),
	}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.cur = p.peek
	p.peek = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	prog := &ast.Program{}

	for p.cur.Type != lexer.EOF {
		switch p.cur.Type {
		case lexer.STRUCT:
			s := p.parseStruct()
			if s != nil {
				prog.Structs = append(prog.Structs, s)
			}

		case lexer.FN:
			fn := p.parseFunction()
			if fn != nil {
				prog.Functions = append(prog.Functions, fn)
			}

		default:
			p.errorf("expected struct or function declaration, got %s", p.cur.Type)
			p.nextToken()
		}
	}

	return prog
}

func (p *Parser) parseStruct() *ast.StructDecl {
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	s := &ast.StructDecl{Name: p.cur.Literal}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	p.nextToken()

	for p.cur.Type != lexer.RBRACE && p.cur.Type != lexer.EOF {
		if p.cur.Type != lexer.IDENT {
			p.errorf("expected field name, got %s", p.cur.Type)
			p.nextToken()
			continue
		}

		name := p.cur.Literal

		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		p.nextToken()
		typ := p.parseType()

		s.Fields = append(s.Fields, ast.FieldDecl{
			Name: name,
			Type: typ,
		})

		p.nextToken()
	}

	if !p.expectCur(lexer.RBRACE) {
		return nil
	}

	p.nextToken()
	return s
}

func (p *Parser) parseFunction() *ast.FunctionDecl {
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	fn := &ast.FunctionDecl{Name: p.cur.Literal}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	fn.Params = p.parseParams()

	if !p.expectCur(lexer.RPAREN) {
		return nil
	}

	if p.peek.Type == lexer.ARROW {
		p.nextToken()
		p.nextToken()
		fn.ReturnType = p.parseType()
	}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	p.nextToken()

	fn.Locals = p.parseVarDeclsAtBlockStart()
	fn.Body = p.parseStatementsUntil(lexer.RBRACE)

	if !p.expectCur(lexer.RBRACE) {
		return nil
	}

	p.nextToken()
	return fn
}

func (p *Parser) parseParams() []ast.Param {
	var params []ast.Param

	p.nextToken()

	if p.cur.Type == lexer.RPAREN {
		return params
	}

	for {
		if p.cur.Type != lexer.IDENT {
			p.errorf("expected parameter name, got %s", p.cur.Type)
			return params
		}

		name := p.cur.Literal

		if !p.expectPeek(lexer.COLON) {
			return params
		}

		p.nextToken()
		typ := p.parseType()

		params = append(params, ast.Param{
			Name: name,
			Type: typ,
		})

		if p.peek.Type != lexer.COMMA {
			break
		}

		p.nextToken()
		p.nextToken()
	}

	if !p.expectPeek(lexer.RPAREN) {
		return params
	}

	return params
}

func (p *Parser) parseVarDeclsAtBlockStart() []*ast.VarDecl {
	var locals []*ast.VarDecl

	for p.cur.Type == lexer.VAR {
		decls := p.parseVarDecls()
		locals = append(locals, decls...)
		p.nextToken()
	}

	return locals
}

func (p *Parser) parseVarDecls() []*ast.VarDecl {
	var names []string

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	names = append(names, p.cur.Literal)

	for p.peek.Type == lexer.COMMA {
		p.nextToken()
		p.nextToken()

		if p.cur.Type != lexer.IDENT {
			p.errorf("expected identifier after comma in var declaration")
			return nil
		}

		names = append(names, p.cur.Literal)
	}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	p.nextToken()
	typ := p.parseType()

	var decls []*ast.VarDecl

	for _, name := range names {
		decls = append(decls, &ast.VarDecl{
			Name: name,
			Type: typ,
		})
	}

	return decls
}

func (p *Parser) parseType() ast.Type {
	if p.cur.Type != lexer.IDENT {
		p.errorf("expected type name, got %s", p.cur.Type)
		return ast.Type{}
	}

	t := ast.Type{Name: p.cur.Literal}

	if p.peek.Type == lexer.LBRACK {
		p.nextToken()

		if !p.expectPeek(lexer.NUMBER) {
			return t
		}

		n, err := strconv.Atoi(p.cur.Literal)
		if err != nil {
			p.errorf("invalid array length %q", p.cur.Literal)
			return t
		}

		t.IsArray = true
		t.ArrayLen = n

		if !p.expectPeek(lexer.RBRACK) {
			return t
		}
	}

	return t
}

func (p *Parser) parseStatementsUntil(end lexer.TokenType) []ast.Stmt {
	var stmts []ast.Stmt

	for p.cur.Type != end && p.cur.Type != lexer.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}

	return stmts
}

func (p *Parser) parseStatement() ast.Stmt {
	switch p.cur.Type {
	case lexer.IDENT:
		return p.parseIdentStatement()

	case lexer.WHILE:
		return p.parseWhile()

	case lexer.IF:
		return p.parseIf()

	case lexer.RETURN:
		return p.parseReturn()

	case lexer.VAR:
		p.errorf("var declarations are only allowed at the beginning of a block")
		p.nextToken()
		return nil

	default:
		p.errorf("unexpected token in statement: %s", p.cur.Type)
		p.nextToken()
		return nil
	}
}

func (p *Parser) parseIdentStatement() ast.Stmt {
	name := p.cur.Literal

	switch p.peek.Type {
	case lexer.ASSIGN, lexer.LBRACK, lexer.DOT:
		target := p.parseLValue()

		if !p.expectCur(lexer.ASSIGN) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)
		p.nextToken()

		return &ast.AssignStmt{
			Target: target,
			Value:  value,
		}

	case lexer.LPAREN:
		args := p.parseCallArgs()
		p.nextToken()

		return &ast.CallStmt{Name: name, Args: args}

	default:
		p.errorf("expected assignment or call after identifier %q", name)
		p.nextToken()
		return nil
	}
}

func (p *Parser) parseLValue() ast.LValue {
	name := p.cur.Literal

	if p.peek.Type == lexer.LBRACK {
		p.nextToken()
		p.nextToken()

		index := p.parseExpression(LOWEST)

		if p.cur.Type != lexer.RBRACK {
			if !p.expectPeek(lexer.RBRACK) {
				return &ast.VarLValue{Name: name}
			}
		}

		p.nextToken()

		if p.cur.Type == lexer.DOT {
			if !p.expectPeek(lexer.IDENT) {
				return &ast.IndexLValue{Name: name, Index: index}
			}

			field := p.cur.Literal
			p.nextToken()

			return &ast.IndexFieldLValue{
				Name:  name,
				Index: index,
				Field: field,
			}
		}

		return &ast.IndexLValue{Name: name, Index: index}
	}

	if p.peek.Type == lexer.DOT {
		p.nextToken()

		if !p.expectPeek(lexer.IDENT) {
			return &ast.VarLValue{Name: name}
		}

		field := p.cur.Literal
		p.nextToken()

		return &ast.FieldLValue{Base: name, Field: field}
	}

	p.nextToken()
	return &ast.VarLValue{Name: name}
}

func (p *Parser) parseWhile() ast.Stmt {
	p.nextToken()
	cond := p.parseExpression(LOWEST)

	if p.cur.Type != lexer.LBRACE {
		if !p.expectPeek(lexer.LBRACE) {
			return nil
		}
	}

	p.nextToken()
	body := p.parseStatementsUntil(lexer.RBRACE)

	if !p.expectCur(lexer.RBRACE) {
		return nil
	}

	p.nextToken()

	return &ast.WhileStmt{
		Cond: cond,
		Body: body,
	}
}

func (p *Parser) parseIf() ast.Stmt {
	p.nextToken()
	cond := p.parseExpression(LOWEST)

	if p.cur.Type != lexer.LBRACE {
		if !p.expectPeek(lexer.LBRACE) {
			return nil
		}
	}

	p.nextToken()
	thenBody := p.parseStatementsUntil(lexer.RBRACE)

	if !p.expectCur(lexer.RBRACE) {
		return nil
	}

	var elseBody []ast.Stmt

	if p.peek.Type == lexer.ELSE {
		p.nextToken()

		if !p.expectPeek(lexer.LBRACE) {
			return nil
		}

		p.nextToken()
		elseBody = p.parseStatementsUntil(lexer.RBRACE)

		if !p.expectCur(lexer.RBRACE) {
			return nil
		}
	}

	p.nextToken()

	return &ast.IfStmt{
		Cond: cond,
		Then: thenBody,
		Else: elseBody,
	}
}

func (p *Parser) parseReturn() ast.Stmt {
	p.nextToken()

	if p.cur.Type == lexer.RBRACE {
		return &ast.ReturnStmt{}
	}

	value := p.parseExpression(LOWEST)
	p.nextToken()

	return &ast.ReturnStmt{Value: value}
}

const (
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	BITOR
	BITXOR
	BITAND
	SHIFT
	SUM
	PRODUCT
	CALL
	INDEX
	FIELD
)

var precedences = map[lexer.TokenType]int{
	lexer.EQ:       EQUALS,
	lexer.NEQ:      EQUALS,
	lexer.LT:       LESSGREATER,
	lexer.LTE:      LESSGREATER,
	lexer.GT:       LESSGREATER,
	lexer.GTE:      LESSGREATER,
	lexer.PIPE:     BITOR,
	lexer.CARET:    BITXOR,
	lexer.AMP:      BITAND,
	lexer.SHL:      SHIFT,
	lexer.SHR:      SHIFT,
	lexer.PLUS:     SUM,
	lexer.MINUS:    SUM,
	lexer.ASTERISK: PRODUCT,
	lexer.SLASH:    PRODUCT,
	lexer.PERCENT:  PRODUCT,
	lexer.LPAREN:   CALL,
	lexer.LBRACK:   INDEX,
	lexer.DOT:      FIELD,
}

func (p *Parser) parseExpression(precedence int) ast.Expr {
	var left ast.Expr

	switch p.cur.Type {
	case lexer.MINUS:
		p.nextToken()
		left = &ast.UnaryExpr{
			Op:   "-",
			Expr: p.parseExpression(SUM),
		}

	case lexer.BANG:
		p.nextToken()
		left = &ast.UnaryExpr{
			Op:   "!",
			Expr: p.parseExpression(SUM),
		}

	case lexer.IDENT:
		left = &ast.IdentExpr{Name: p.cur.Literal}

	case lexer.NUMBER:
		left = &ast.NumberExpr{Value: p.cur.Literal}

	case lexer.STRING:
		left = &ast.StringExpr{Value: p.cur.Literal}

	case lexer.LPAREN:
		p.nextToken()
		left = p.parseExpression(LOWEST)

		if p.cur.Type != lexer.RPAREN {
			if !p.expectPeek(lexer.RPAREN) {
				return left
			}
		}

	default:
		p.errorf("expected expression, got %s", p.cur.Type)
		return nil
	}

	for p.peek.Type != lexer.EOF &&
		p.peek.Type != lexer.RBRACE &&
		p.peek.Type != lexer.COMMA &&
		p.peek.Type != lexer.RPAREN &&
		precedence < p.peekPrecedence() {

		switch p.peek.Type {
		case lexer.LPAREN:
			ident, ok := left.(*ast.IdentExpr)
			if !ok {
				p.errorf("only identifier calls are supported")
				return left
			}
			left = p.parseCallExpr(ident.Name)

		case lexer.LBRACK:
			ident, ok := left.(*ast.IdentExpr)
			if !ok {
				p.errorf("only identifier array indexing is supported")
				return left
			}
			left = p.parseIndexExpr(ident.Name)

		case lexer.DOT:
			switch v := left.(type) {
			case *ast.IdentExpr:
				left = p.parseFieldExpr(v.Name)

			case *ast.IndexExpr:
				left = p.parseIndexFieldExpr(v.Name, v.Index)

			default:
				p.errorf("only identifier or indexed field access is supported")
				return left
			}

		default:
			op := p.peek
			p.nextToken()
			p.nextToken()

			right := p.parseExpression(precedences[op.Type])

			left = &ast.BinaryExpr{
				Left:  left,
				Op:    op.Literal,
				Right: right,
			}
		}
	}

	return left
}

func (p *Parser) parseCallExpr(name string) ast.Expr {
	args := p.parseCallArgs()
	return &ast.CallExpr{Name: name, Args: args}
}

func (p *Parser) parseIndexExpr(name string) ast.Expr {
	p.nextToken()
	p.nextToken()

	index := p.parseExpression(LOWEST)

	if p.cur.Type != lexer.RBRACK {
		if !p.expectPeek(lexer.RBRACK) {
			return &ast.IdentExpr{Name: name}
		}
	}

	return &ast.IndexExpr{Name: name, Index: index}
}

func (p *Parser) parseFieldExpr(base string) ast.Expr {
	p.nextToken()

	if !p.expectPeek(lexer.IDENT) {
		return &ast.IdentExpr{Name: base}
	}

	return &ast.FieldExpr{
		Base:  base,
		Field: p.cur.Literal,
	}
}

func (p *Parser) parseIndexFieldExpr(name string, index ast.Expr) ast.Expr {
	p.nextToken()

	if !p.expectPeek(lexer.IDENT) {
		return &ast.IndexExpr{Name: name, Index: index}
	}

	return &ast.IndexFieldExpr{
		Name:  name,
		Index: index,
		Field: p.cur.Literal,
	}
}

func (p *Parser) parseCallArgs() []ast.Expr {
	var args []ast.Expr

	if !p.expectPeek(lexer.LPAREN) {
		return args
	}

	p.nextToken()

	if p.cur.Type == lexer.RPAREN {
		return args
	}

	for {
		arg := p.parseExpression(LOWEST)
		if arg != nil {
			args = append(args, arg)
		}

		if p.cur.Type == lexer.RPAREN {
			break
		}

		if p.peek.Type == lexer.COMMA {
			p.nextToken()
			p.nextToken()
			continue
		}

		if p.peek.Type == lexer.RPAREN {
			p.nextToken()
			break
		}

		p.errorf("expected comma or closing paren in argument list, got %s", p.peek.Type)
		break
	}

	return args
}

func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peek.Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peek.Type != t {
		p.errorf("expected next token %s, got %s", t, p.peek.Type)
		return false
	}

	p.nextToken()
	return true
}

func (p *Parser) expectCur(t lexer.TokenType) bool {
	if p.cur.Type != t {
		p.errorf("expected current token %s, got %s", t, p.cur.Type)
		return false
	}

	return true
}

func (p *Parser) errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)

	line := p.cur.Line
	col := p.cur.Column

	var srcLine string
	if line-1 >= 0 && line-1 < len(p.lines) {
		srcLine = p.lines[line-1]
	}

	caret := ""
	if col > 0 {
		caret = strings.Repeat(" ", col-1) + "^^^"
	}

	full := fmt.Sprintf("%d:%d: %s\n\n    %s\n    %s",
		line, col, msg, srcLine, caret,
	)

	p.errors = append(p.errors, full)
}
