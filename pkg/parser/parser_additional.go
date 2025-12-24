package parser

// Package parser implements Perl parsing using Pratt parsing.
// Paket parser, Pratt ayrıştırma kullanarak Perl ayrıştırmasını uygular.

import (
	"fmt"
	"strings"

	"perlc/pkg/ast"
	"perlc/pkg/lexer"
)

// ============================================================
// Control Flow Parsers
// Kontrol Akışı Ayrıştırıcıları
// ============================================================

func (p *Parser) parseIfStmt() ast.Statement {
	stmt := &ast.IfStmt{Token: p.curToken}
	stmt.Unless = p.curToken.Type == lexer.TokUnless

	if !p.expectPeek(lexer.TokLParen) {
		return nil
	}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(lexer.TokRParen) {
		return nil
	}

	if !p.expectPeek(lexer.TokLBrace) {
		return nil
	}
	stmt.Then = p.parseBlockStmt()

	// Elsif clauses
	for p.peekTokenIs(lexer.TokElsif) {
		p.nextToken()
		clause := &ast.ElsifClause{}

		if !p.expectPeek(lexer.TokLParen) {
			return nil
		}
		p.nextToken()
		clause.Condition = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.TokRParen) {
			return nil
		}

		if !p.expectPeek(lexer.TokLBrace) {
			return nil
		}
		clause.Body = p.parseBlockStmt()
		stmt.Elsif = append(stmt.Elsif, clause)
	}

	// Else clause
	if p.peekTokenIs(lexer.TokElse) {
		p.nextToken()
		if !p.expectPeek(lexer.TokLBrace) {
			return nil
		}
		stmt.Else = p.parseBlockStmt()
	}

	return stmt
}

func (p *Parser) parseWhileStmt() ast.Statement {
	stmt := &ast.WhileStmt{Token: p.curToken}
	stmt.Until = p.curToken.Type == lexer.TokUntil

	if !p.expectPeek(lexer.TokLParen) {
		return nil
	}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(lexer.TokRParen) {
		return nil
	}

	if !p.expectPeek(lexer.TokLBrace) {
		return nil
	}
	stmt.Body = p.parseBlockStmt()

	return stmt
}

func (p *Parser) parseForStmt() ast.Statement {
	token := p.curToken

	if !p.expectPeek(lexer.TokLParen) {
		return nil
	}

	p.nextToken() // skip (, now at first token inside

	// Check if it's foreach-style: for my $x (@arr) or for $x (@arr)
	// Need to look ahead to distinguish from C-style: for (my $i = 0; ...)
	if p.curTokenIs(lexer.TokMy) || p.curTokenIs(lexer.TokOur) || p.curTokenIs(lexer.TokLocal) {
		// Save position to check what follows the variable
		// If "my $x (" -> foreach style
		// If "my $x =" -> C-style

		// Peek: my $var ... what's next?
		// For C-style: my $i = 0; -> after $i comes =
		// For foreach: my $x (@arr) -> after $x comes ( but we're already past outer (

		// Actually in "for (my $i = 0; ...)" we're inside parens
		// In "for my $x (@arr)" the my is OUTSIDE parens
		// But our current position is AFTER (, so this must be C-style!

		// So if we're here (after opening paren) and see "my", it's C-style init
		// Fall through to C-style parsing
	} else if p.curTokenIs(lexer.TokScalar) {
		// for ($x ...) - need to check if it's foreach or C-style
		// For now, assume C-style if inside parens
	}

	// C-style for: for (init; cond; post) { body }
	stmt := &ast.ForStmt{Token: token}

	// Init
	if !p.curTokenIs(lexer.TokSemi) {
		stmt.Init = p.parseStatement()
	}
	// After parseStatement, we might be on ; or need to advance
	if p.peekTokenIs(lexer.TokSemi) {
		p.nextToken()
	}
	if p.curTokenIs(lexer.TokSemi) {
		p.nextToken() // skip ;
	}

	// Condition
	if !p.curTokenIs(lexer.TokSemi) {
		stmt.Condition = p.parseExpression(LOWEST)
	}
	if !p.expectPeek(lexer.TokSemi) {
		return nil
	}
	p.nextToken() // skip ;

	// Post
	if !p.curTokenIs(lexer.TokRParen) {
		stmt.Post = p.parseExpression(LOWEST)
	}
	if !p.expectPeek(lexer.TokRParen) {
		return nil
	}

	// Body
	if !p.expectPeek(lexer.TokLBrace) {
		return nil
	}
	stmt.Body = p.parseBlockStmt()

	return stmt
}

func (p *Parser) parseForeachStyleFor(token lexer.Token) ast.Statement {
	stmt := &ast.ForeachStmt{Token: token}

	if p.curTokenIs(lexer.TokMy) {
		p.nextToken()
	}
	stmt.Variable = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TokLParen) {
		return nil
	}
	p.nextToken()
	stmt.List = p.parseExpression(LOWEST)
	if !p.expectPeek(lexer.TokRParen) {
		return nil
	}
	if !p.expectPeek(lexer.TokRParen) {
		return nil
	}

	if !p.expectPeek(lexer.TokLBrace) {
		return nil
	}
	stmt.Body = p.parseBlockStmt()

	return stmt
}

func (p *Parser) parseForeachStmt() ast.Statement {
	stmt := &ast.ForeachStmt{Token: p.curToken}

	p.nextToken() // skip foreach

	// Optional my/our/local
	if p.curTokenIs(lexer.TokMy) || p.curTokenIs(lexer.TokOur) || p.curTokenIs(lexer.TokLocal) {
		p.nextToken()
	}

	// Variable - parse with high precedence to stop before (
	stmt.Variable = p.parseExpression(CALL)

	// List in parentheses
	if !p.expectPeek(lexer.TokLParen) {
		return nil
	}
	p.nextToken() // skip (
	stmt.List = p.parseExpression(LOWEST)
	if !p.expectPeek(lexer.TokRParen) {
		return nil
	}

	// Body
	if !p.expectPeek(lexer.TokLBrace) {
		return nil
	}
	stmt.Body = p.parseBlockStmt()

	return stmt
}

func (p *Parser) parseLastStmt() ast.Statement {
	stmt := &ast.LastStmt{Token: p.curToken}
	if p.peekTokenIs(lexer.TokIdent) {
		p.nextToken()
		stmt.Label = p.curToken.Value
	}
	if p.peekTokenIs(lexer.TokSemi) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseNextStmt() ast.Statement {
	stmt := &ast.NextStmt{Token: p.curToken}
	if p.peekTokenIs(lexer.TokIdent) {
		p.nextToken()
		stmt.Label = p.curToken.Value
	}
	if p.peekTokenIs(lexer.TokSemi) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseRedoStmt() ast.Statement {
	stmt := &ast.RedoStmt{Token: p.curToken}
	if p.peekTokenIs(lexer.TokIdent) {
		p.nextToken()
		stmt.Label = p.curToken.Value
	}
	if p.peekTokenIs(lexer.TokSemi) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseReturnStmt() ast.Statement {
	stmt := &ast.ReturnStmt{Token: p.curToken}

	if !p.peekTokenIs(lexer.TokSemi) && !p.peekTokenIs(lexer.TokRBrace) {
		p.nextToken()
		stmt.Value = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(lexer.TokSemi) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseFatArrowExpression(left ast.Expression) ast.Expression {
	// fat arrow used in hash context, return as-is for hash parsing
	// hash bağlamında kullanılan fat arrow, hash ayrıştırma için olduğu gibi döndür
	return left
}

func (p *Parser) parseBuiltinCall() ast.Expression {
	tok := p.curToken
	name := tok.Value

	// Special handling for print/say with filehandle: print $fh "text"
	if name == "print" || name == "say" {
		return p.parsePrintCall(tok, name)
	}

	expr := &ast.CallExpr{
		Token:    tok,
		Function: &ast.Identifier{Token: tok, Value: name},
	}

	// Check for parentheses
	if p.peekTokenIs(lexer.TokLParen) {
		p.nextToken()
		expr.Args = p.parseExpressionList(lexer.TokRParen)
	} else {
		// No parentheses - parse arguments
		p.nextToken()
		expr.Args = p.parseListExpression()
	}

	return expr
}

func (p *Parser) parsePrintCall(tok lexer.Token, name string) ast.Expression {
	expr := &ast.CallExpr{
		Token:    tok,
		Function: &ast.Identifier{Token: tok, Value: name},
	}

	// Check for parentheses
	if p.peekTokenIs(lexer.TokLParen) {
		p.nextToken()
		expr.Args = p.parseExpressionList(lexer.TokRParen)
		return expr
	}

	p.nextToken()

	// Check if first token is a scalar variable (potential filehandle)
	// Filehandle form: print $fh "text" or print $fh $var
	// But NOT: print $a + $b (that's an expression)
	// Filehandle is followed by a string or another scalar (not an operator)
	if p.curTokenIs(lexer.TokScalar) &&
		(p.peekTokenIs(lexer.TokString) || p.peekTokenIs(lexer.TokRawString) ||
			(p.peekTokenIs(lexer.TokScalar) && !p.isOperatorToken(p.peekToken.Type))) {
		// This is filehandle form: print $fh "text" or print $fh $var
		fhExpr := p.parseExpression(LOWEST)
		expr.Args = append(expr.Args, fhExpr)
		p.nextToken()
		expr.Args = append(expr.Args, p.parseListExpression()...)
		return expr
	}

	// Normal print - parse full expression list
	expr.Args = p.parseListExpression()
	return expr
}

func (p *Parser) ParsePrintCallComplex(tok lexer.Token, name string) ast.Expression {
	expr := &ast.CallExpr{
		Token:    tok,
		Function: &ast.Identifier{Token: tok, Value: name},
	}

	// Check for parentheses
	if p.peekTokenIs(lexer.TokLParen) {
		p.nextToken()
		expr.Args = p.parseExpressionList(lexer.TokRParen)
		return expr
	}

	p.nextToken()

	// Check if first token is a scalar variable (potential filehandle)
	// But NOT if it's followed by -> (that's arrow access, not filehandle)
	if p.curTokenIs(lexer.TokScalar) && !p.peekTokenIs(lexer.TokArrow) {

		fhExpr := p.parseExpression(CALL)

		// Check if next token is expression (not comma, not semicolon) - filehandle form
		if !p.peekTokenIs(lexer.TokComma) && !p.peekTokenIs(lexer.TokSemi) && !p.peekTokenIs(lexer.TokEOF) && !p.peekTokenIs(lexer.TokArrow) {
			expr.Args = append(expr.Args, fhExpr)
			p.nextToken()
			expr.Args = append(expr.Args, p.parseListExpression()...)
		} else {
			// Normal: print $var; or print $var, $var2;
			expr.Args = append(expr.Args, fhExpr)
			if p.peekTokenIs(lexer.TokComma) {
				p.nextToken()
				p.nextToken()
				expr.Args = append(expr.Args, p.parseListExpression()...)
			}
		}
		return expr
	}

	// Normal print without filehandle
	expr.Args = p.parseListExpression()
	return expr
}

func (p *Parser) ParsePrintCallComplex2(tok lexer.Token, name string) ast.Expression {
	expr := &ast.CallExpr{
		Token:    tok,
		Function: &ast.Identifier{Token: tok, Value: name},
	}

	// Check for parentheses
	if p.peekTokenIs(lexer.TokLParen) {
		p.nextToken()
		expr.Args = p.parseExpressionList(lexer.TokRParen)
		return expr
	}

	p.nextToken()

	// Check if first token is a scalar variable (potential filehandle)
	// Filehandle form: print $fh "text" (identifier after scalar, no operator)
	if p.curTokenIs(lexer.TokScalar) && p.peekTokenIs(lexer.TokString) {
		// This is filehandle form: print $fh "text"
		fhExpr := p.parseExpression(CALL)

		expr.Args = append(expr.Args, fhExpr)
		p.nextToken()

		expr.Args = append(expr.Args, p.parseListExpression()...)

		return expr
	}

	// Normal print - parse full expression list
	expr.Args = p.parseListExpression()
	return expr
}

func (p *Parser) parseSubstExpression(left ast.Expression) ast.Expression {
	tok := p.curToken
	// Parse s/pattern/replacement/flags from token value
	parts := strings.SplitN(tok.Value, "/", 3)
	pattern := ""
	replacement := ""
	flags := ""
	if len(parts) >= 1 {
		pattern = parts[0]
	}
	if len(parts) >= 2 {
		replacement = parts[1]
	}
	if len(parts) >= 3 {
		flags = parts[2]
	}

	return &ast.SubstExpr{
		Token:       tok,
		Target:      left,
		Pattern:     pattern,
		Replacement: replacement,
		Flags:       flags,
	}
}

func (p *Parser) parseOpenExpr() ast.Expression {
	tok := p.curToken

	if !p.expectPeek(lexer.TokLParen) {
		// open without parens: open FH, MODE, FILE
		p.nextToken()
	} else {
		p.nextToken() // skip (
	}

	// Filehandle
	var fh ast.Expression
	if p.curTokenIs(lexer.TokMy) {
		p.nextToken() // skip my
	}
	fh = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TokComma) {
		return nil
	}
	p.nextToken()

	// Mode
	mode := p.parseExpression(LOWEST)

	// Optional third argument (filename)
	var filename ast.Expression
	if p.peekTokenIs(lexer.TokComma) {
		p.nextToken() // skip comma
		p.nextToken()
		filename = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(lexer.TokRParen) {
		p.nextToken()
	}

	return &ast.CallExpr{
		Token:    tok,
		Function: &ast.Identifier{Token: tok, Value: "open"},
		Args:     []ast.Expression{fh, mode, filename},
	}
}

func (p *Parser) parseCloseExpr() ast.Expression {
	tok := p.curToken

	if !p.expectPeek(lexer.TokLParen) {
		p.nextToken()
	} else {
		p.nextToken() // skip (
	}

	fh := p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.TokRParen) {
		p.nextToken()
	}

	return &ast.CallExpr{
		Token:    tok,
		Function: &ast.Identifier{Token: tok, Value: "close"},
		Args:     []ast.Expression{fh},
	}
}

func (p *Parser) parseReadLineExpr() ast.Expression {
	tok := p.curToken

	expr := &ast.ReadLineExpr{Token: tok}

	if tok.Type == lexer.TokDiamond {
		// <> - STDIN/ARGV
		expr.Filehandle = nil
	} else {
		// <FH> or <$fh>
		if len(tok.Value) > 0 && tok.Value[0] == '$' {
			// Variable filehandle
			expr.Filehandle = &ast.ScalarVar{Token: tok, Name: tok.Value[1:]}
		} else {
			// Bareword filehandle
			expr.Filehandle = &ast.Identifier{Token: tok, Value: tok.Value}
		}
	}

	return expr
}

// parseListExpression parses comma-separated expressions until semicolon or EOF
func (p *Parser) parseListExpression() []ast.Expression {
	var list []ast.Expression

	if p.curTokenIs(lexer.TokSemi) || p.curTokenIs(lexer.TokEOF) {
		return list
	}

	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(lexer.TokComma) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	return list
}

// -----------------------------------------------------------------//
// ------------------------ Parsing Helpers ----------------------- //
// ------------------------ Ayrıştırma Yardımcıları ----------------------- //
// -----------------------------------------------------------------//
// Errors returns parsing errors.
// Errors, ayrıştırma hatalarını döndürür.
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("line %d: expected next token to be %v, got %v instead",
		p.peekToken.Line, t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("line %d: no prefix parse function for %v found (value=%q, peek=%v/%q)",
		p.curToken.Line, t, p.curToken.Value, p.peekToken.Type, p.peekToken.Value)
	p.errors = append(p.errors, msg)
}

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
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

// isOperatorToken checks if token is an operator
func (p *Parser) isOperatorToken(t lexer.TokenType) bool {
	switch t {
	case lexer.TokPlus, lexer.TokMinus, lexer.TokStar, lexer.TokSlash,
		lexer.TokPercent, lexer.TokStarStar, lexer.TokDot, lexer.TokX,
		lexer.TokEq, lexer.TokNe, lexer.TokLt, lexer.TokLe, lexer.TokGt, lexer.TokGe,
		lexer.TokStrEq, lexer.TokStrNe, lexer.TokStrLt, lexer.TokStrLe, lexer.TokStrGt, lexer.TokStrGe,
		lexer.TokAnd, lexer.TokOr, lexer.TokAndWord, lexer.TokOrWord,
		lexer.TokAssign, lexer.TokPlusEq, lexer.TokMinusEq,
		lexer.TokArrow, lexer.TokComma, lexer.TokSemi:
		return true
	default:
		return false
	}
}

// isBareword returns true if current token can be used as a hash key bareword.
// In Perl, keywords can be used as hash keys without quoting.
func (p *Parser) isBareword() bool {
	switch p.curToken.Type {
	case lexer.TokIdent:
		return true
	// Keywords that can be used as barewords in hash keys
	case lexer.TokX, lexer.TokIf, lexer.TokElse, lexer.TokFor, lexer.TokForeach,
		lexer.TokWhile, lexer.TokMy, lexer.TokOur, lexer.TokLocal, lexer.TokSub,
		lexer.TokUse, lexer.TokPackage, lexer.TokReturn, lexer.TokLast, lexer.TokNext,
		lexer.TokStrEq, lexer.TokStrNe, lexer.TokStrLt, lexer.TokStrLe, lexer.TokStrGt, lexer.TokStrGe,
		lexer.TokAndWord, lexer.TokOrWord, lexer.TokNotWord,
		lexer.TokPrint, lexer.TokSay, lexer.TokDefined, lexer.TokUndef, lexer.TokRef,
		lexer.TokLength, lexer.TokPush, lexer.TokPop, lexer.TokShift, lexer.TokUnshift,
		lexer.TokKeys, lexer.TokValues, lexer.TokJoin, lexer.TokSplit,
		lexer.TokAbs, lexer.TokInt, lexer.TokSqrt, lexer.TokChr, lexer.TokOrd,
		lexer.TokLc, lexer.TokUc, lexer.TokChomp, lexer.TokChop,
		lexer.TokOpen, lexer.TokClose, lexer.TokDie, lexer.TokWarn, lexer.TokExit:
		return true
	default:
		return false
	}
}

// parseBlockAsAnonSub парсит { block } как AnonSubExpr
func (p *Parser) parseBlockAsAnonSub() ast.Expression {
	tok := p.curToken // должен быть {

	if !p.curTokenIs(lexer.TokLBrace) {
		return nil
	}

	body := p.parseBlockStmt()

	return &ast.AnonSubExpr{
		Token:  tok,
		Params: nil,
		Body:   body,
	}
}

// ============================================================
// Также добавить новую функцию parseGrepMap для обработки
// grep { block } @arr и map { block } @arr синтаксиса:
// ============================================================

func (p *Parser) parseGrepMap() ast.Expression {
	tok := p.curToken
	funcName := tok.Value // "grep" или "map"

	call := &ast.CallExpr{
		Token:    tok,
		Function: &ast.Identifier{Token: tok, Value: funcName},
		Args:     []ast.Expression{},
	}

	// Проверяем следующий токен (peek, не next!)
	if p.peekTokenIs(lexer.TokLBrace) {
		// grep { ... } @arr или map { ... } @arr
		p.nextToken() // переходим на {

		// Парсим блок как анонимную функцию
		block := p.parseBlockAsAnonSub()
		call.Args = append(call.Args, block)

		p.nextToken() // переходим с } на @nums

		// После блока ожидаем массив или список
		// НЕ нужна запятая между блоком и массивом в Perl!
		if p.curTokenIs(lexer.TokArray) {
			arr := p.parseExpression(LOWEST)
			call.Args = append(call.Args, arr)
		} else if p.curTokenIs(lexer.TokLParen) {
			// grep { ... } (1, 2, 3)
			//p.nextToken() // skip (
			args := p.parseExpressionList(lexer.TokRParen)
			// Создаём массив из аргументов
			arrExpr := &ast.ArrayExpr{Token: tok, Elements: args}
			call.Args = append(call.Args, arrExpr)
		} else if p.curTokenIs(lexer.TokScalar) {
			// grep { ... } @$ref - разыменование
			arr := p.parseExpression(LOWEST)
			call.Args = append(call.Args, arr)
		}
	} else if p.peekTokenIs(lexer.TokLParen) {
		// grep(EXPR, @arr) - функциональный синтаксис с скобками
		p.nextToken() // переходим на (
		call.Args = p.parseExpressionList(lexer.TokRParen)
	} else {
		// grep EXPR, @arr - без скобок
		p.nextToken()
		call.Args = p.parseListExpression()
	}

	return call
}
