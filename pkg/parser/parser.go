// Package parser implements a parser for Perl source code.
package parser

import (
	"fmt"
	"strconv"

	"github.com/djeday123/perl-compiler/pkg/ast"
	"github.com/djeday123/perl-compiler/pkg/lexer"
	"github.com/djeday123/perl-compiler/pkg/token"
)

// Operator precedence levels
const (
	_ int = iota
	LOWEST
	ORASSIGN    // or
	ANDASSIGN   // and
	NOTP        // not
	ASSIGN      // =
	TERNARY     // ?:
	RANGE       // ..
	LOGOR       // ||
	LOGAND      // &&
	BITOR       // |
	BITXOR      // ^
	BITAND      // &
	EQUALS      // == !=
	LESSGREATER // < > <= >=
	SHIFT       // << >>
	SUM         // + -
	PRODUCT     // * / %
	CONCAT      // .
	PREFIX      // -X !X ~X
	POWER       // **
	INCREMENT   // ++ --
	CALL        // myFunction(X)
	INDEX       // array[index]
)

var precedences = map[token.TokenType]int{
	token.ORKW:     ORASSIGN,
	token.ANDKW:    ANDASSIGN,
	token.NOTKW:    NOTP,
	token.ASSIGN:   ASSIGN,
	token.PLUSEQ:   ASSIGN,
	token.MINUSEQ:  ASSIGN,
	token.MULEQ:    ASSIGN,
	token.DIVEQ:    ASSIGN,
	token.MODEQ:    ASSIGN,
	token.POWEQ:    ASSIGN,
	token.CONCATEQ: ASSIGN,
	token.TERNARY:  TERNARY,
	token.RANGE:    RANGE,
	token.OR:       LOGOR,
	token.AND:      LOGAND,
	token.BITOR:    BITOR,
	token.BITXOR:   BITXOR,
	token.BITAND:   BITAND,
	token.EQ:       EQUALS,
	token.NE:       EQUALS,
	token.STREQ:    EQUALS,
	token.STRNE:    EQUALS,
	token.STRCMP:   EQUALS,
	token.NUMEQ:    EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.LE:       LESSGREATER,
	token.GE:       LESSGREATER,
	token.STRLT:    LESSGREATER,
	token.STRGT:    LESSGREATER,
	token.STRLE:    LESSGREATER,
	token.STRGE:    LESSGREATER,
	token.LSHIFT:   SHIFT,
	token.RSHIFT:   SHIFT,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.ASTERISK: PRODUCT,
	token.SLASH:    PRODUCT,
	token.MODULO:   PRODUCT,
	token.CONCAT:   CONCAT,
	token.REPEAT:   CONCAT,
	token.POWER:    POWER,
	token.INC:      INCREMENT,
	token.DEC:      INCREMENT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
	token.LBRACE:   INDEX,
	token.ARROW:    INDEX,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// Parser represents a parser for Perl source code.
type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

// New creates a new Parser for the given Lexer.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.SCALAR, p.parseScalarVariable)
	p.registerPrefix(token.ARRAY, p.parseArrayVariable)
	p.registerPrefix(token.HASH, p.parseHashVariable)
	p.registerPrefix(token.NOT, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.BITNOT, p.parsePrefixExpression)
	p.registerPrefix(token.NOTKW, p.parsePrefixExpression)
	p.registerPrefix(token.INC, p.parsePrefixExpression)
	p.registerPrefix(token.DEC, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.DEFINED, p.parseDefinedExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.MODULO, p.parseInfixExpression)
	p.registerInfix(token.POWER, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NE, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LE, p.parseInfixExpression)
	p.registerInfix(token.GE, p.parseInfixExpression)
	p.registerInfix(token.NUMEQ, p.parseInfixExpression)
	p.registerInfix(token.STREQ, p.parseInfixExpression)
	p.registerInfix(token.STRNE, p.parseInfixExpression)
	p.registerInfix(token.STRLT, p.parseInfixExpression)
	p.registerInfix(token.STRGT, p.parseInfixExpression)
	p.registerInfix(token.STRLE, p.parseInfixExpression)
	p.registerInfix(token.STRGE, p.parseInfixExpression)
	p.registerInfix(token.STRCMP, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.ANDKW, p.parseInfixExpression)
	p.registerInfix(token.ORKW, p.parseInfixExpression)
	p.registerInfix(token.BITAND, p.parseInfixExpression)
	p.registerInfix(token.BITOR, p.parseInfixExpression)
	p.registerInfix(token.BITXOR, p.parseInfixExpression)
	p.registerInfix(token.LSHIFT, p.parseInfixExpression)
	p.registerInfix(token.RSHIFT, p.parseInfixExpression)
	p.registerInfix(token.CONCAT, p.parseInfixExpression)
	p.registerInfix(token.REPEAT, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN, p.parseInfixExpression)
	p.registerInfix(token.PLUSEQ, p.parseInfixExpression)
	p.registerInfix(token.MINUSEQ, p.parseInfixExpression)
	p.registerInfix(token.MULEQ, p.parseInfixExpression)
	p.registerInfix(token.DIVEQ, p.parseInfixExpression)
	p.registerInfix(token.MODEQ, p.parseInfixExpression)
	p.registerInfix(token.POWEQ, p.parseInfixExpression)
	p.registerInfix(token.CONCATEQ, p.parseInfixExpression)
	p.registerInfix(token.RANGE, p.parseRangeExpression)
	p.registerInfix(token.TERNARY, p.parseTernaryExpression)
	p.registerInfix(token.INC, p.parsePostfixExpression)
	p.registerInfix(token.DEC, p.parsePostfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.ARROW, p.parseArrowExpression)

	// Read two tokens to initialize curToken and peekToken
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// Errors returns the list of parsing errors.
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("line %d: expected next token to be %s, got %s instead",
		p.peekToken.Line, t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("line %d: no prefix parse function for %s found",
		p.curToken.Line, t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// ParseProgram parses the entire program and returns the AST.
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.MY:
		return p.parseMyStatement()
	case token.SUB:
		return p.parseSubroutineStatement()
	case token.IF:
		return p.parseIfStatement()
	case token.UNLESS:
		return p.parseUnlessStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.UNTIL:
		return p.parseUntilStatement()
	case token.FOR, token.FOREACH:
		return p.parseForStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.PRINT, token.SAY:
		return p.parsePrintStatement()
	case token.COMMENT:
		// Skip comments
		return nil
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseMyStatement() *ast.MyStatement {
	stmt := &ast.MyStatement{Token: p.curToken}

	p.nextToken()

	// Parse just the variable name at high precedence to avoid capturing the assignment
	stmt.Name = p.parseExpression(ASSIGN + 1)

	if p.peekTokenIs(token.ASSIGN) {
		p.nextToken()
		p.nextToken()
		stmt.Value = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseSubroutineStatement() *ast.SubroutineStatement {
	stmt := &ast.SubroutineStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSIF) {
		p.nextToken()
		stmt.Alternative = p.parseElsifStatement()
	} else if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		stmt.Alternative = p.parseBlockStatement()
	}

	return stmt
}

func (p *Parser) parseElsifStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSIF) {
		p.nextToken()
		stmt.Alternative = p.parseElsifStatement()
	} else if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		stmt.Alternative = p.parseBlockStatement()
	}

	return stmt
}

func (p *Parser) parseUnlessStatement() *ast.IfStatement {
	// Unless is just "if not"
	stmt := &ast.IfStatement{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	condition := p.parseExpression(LOWEST)

	// Wrap the condition with a NOT
	stmt.Condition = &ast.PrefixExpression{
		Token:    token.Token{Type: token.NOT, Literal: "!"},
		Operator: "!",
		Right:    condition,
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		stmt.Alternative = p.parseBlockStatement()
	}

	return stmt
}

func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseUntilStatement() *ast.WhileStatement {
	// Until is just "while not"
	stmt := &ast.WhileStatement{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	condition := p.parseExpression(LOWEST)

	// Wrap the condition with a NOT
	stmt.Condition = &ast.PrefixExpression{
		Token:    token.Token{Type: token.NOT, Literal: "!"},
		Operator: "!",
		Right:    condition,
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curToken}

	// Check for loop variable
	if p.peekTokenIs(token.MY) {
		p.nextToken()
		p.nextToken()
		// Parse just the variable - don't allow function calls or other high precedence operators
		stmt.Variable = p.parseExpression(CALL)
	} else if p.peekTokenIs(token.SCALAR) {
		p.nextToken()
		// Parse just the variable
		stmt.Variable = p.parseExpression(CALL)
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.List = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	if !p.curTokenIs(token.SEMICOLON) {
		stmt.ReturnValue = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parsePrintStatement() *ast.PrintStatement {
	stmt := &ast.PrintStatement{Token: p.curToken}
	stmt.Arguments = []ast.Expression{}

	p.nextToken()

	for !p.curTokenIs(token.SEMICOLON) && !p.curTokenIs(token.EOF) {
		arg := p.parseExpression(LOWEST)
		if arg != nil {
			stmt.Arguments = append(stmt.Arguments, arg)
		}
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
		} else {
			break
		}
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("line %d: could not parse %q as integer",
			p.curToken.Line, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("line %d: could not parse %q as float",
			p.curToken.Line, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseScalarVariable() ast.Expression {
	return &ast.ScalarVariable{Token: p.curToken, Name: p.curToken.Literal}
}

func (p *Parser) parseArrayVariable() ast.Expression {
	return &ast.ArrayVariable{Token: p.curToken, Name: p.curToken.Literal}
}

func (p *Parser) parseHashVariable() ast.Expression {
	return &ast.HashVariable{Token: p.curToken, Name: p.curToken.Literal}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parsePostfixExpression(left ast.Expression) ast.Expression {
	return &ast.PostfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
}

func (p *Parser) parseRangeExpression(left ast.Expression) ast.Expression {
	expression := &ast.RangeExpression{
		Token: p.curToken,
		Start: left,
	}

	p.nextToken()
	expression.End = p.parseExpression(RANGE)

	return expression
}

func (p *Parser) parseTernaryExpression(condition ast.Expression) ast.Expression {
	expression := &ast.TernaryExpression{
		Token:     p.curToken,
		Condition: condition,
	}

	p.nextToken()
	expression.Consequence = p.parseExpression(LOWEST)

	if !p.expectPeek(token.COLON) {
		return nil
	}

	p.nextToken()
	expression.Alternative = p.parseExpression(LOWEST)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if p.curToken.Type == token.LBRACKET {
		if !p.expectPeek(token.RBRACKET) {
			return nil
		}
	} else {
		if !p.expectPeek(token.RBRACE) {
			return nil
		}
	}

	return exp
}

func (p *Parser) parseArrowExpression(left ast.Expression) ast.Expression {
	// Handle -> for method calls and dereferencing
	p.nextToken()

	if p.curTokenIs(token.LBRACKET) {
		// Array dereference: $ref->[index]
		exp := &ast.IndexExpression{Token: p.curToken, Left: left}
		p.nextToken()
		exp.Index = p.parseExpression(LOWEST)
		if !p.expectPeek(token.RBRACKET) {
			return nil
		}
		return exp
	} else if p.curTokenIs(token.LBRACE) {
		// Hash dereference: $ref->{key}
		exp := &ast.IndexExpression{Token: p.curToken, Left: left}
		p.nextToken()
		exp.Index = p.parseExpression(LOWEST)
		if !p.expectPeek(token.RBRACE) {
			return nil
		}
		return exp
	} else if p.curTokenIs(token.IDENT) {
		// Method call: $obj->method(args)
		methodName := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if p.peekTokenIs(token.LPAREN) {
			p.nextToken()
			call := &ast.CallExpression{
				Token:    p.curToken,
				Function: methodName,
			}
			call.Arguments = append([]ast.Expression{left}, p.parseExpressionList(token.RPAREN)...)
			return call
		}

		return methodName
	}

	return nil
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseDefinedExpression() ast.Expression {
	expr := &ast.CallExpression{
		Token:    p.curToken,
		Function: &ast.Identifier{Token: p.curToken, Value: "defined"},
	}

	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		expr.Arguments = p.parseExpressionList(token.RPAREN)
	} else {
		p.nextToken()
		arg := p.parseExpression(PREFIX)
		if arg != nil {
			expr.Arguments = []ast.Expression{arg}
		}
	}

	return expr
}
