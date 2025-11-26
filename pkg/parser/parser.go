package parser

// Package parser implements Perl parsing using Pratt parsing.
// Paket parser, Pratt ayrıştırma kullanarak Perl ayrıştırmasını uygular.

import (
	"fmt"
	"strconv"
	"strings"

	"perlc/pkg/ast"
	"perlc/pkg/lexer"
)

// Precedence levels for operators.
// Operatörler için öncelik seviyeleri.
const (
	_ int = iota
	LOWEST
	COMMA          // ,
	ASSIGN         // = += -= etc.
	TERNARY        // ?:
	OR             // || or //
	AND            // &&
	BITOR          // |
	BITXOR         // ^
	BITAND         // &
	EQUALITY       // == != eq ne
	COMPARISON     // < > <= >= lt gt le ge <=>
	SHIFT          // << >>
	ADDITIVE       // + - .
	MULTIPLICATIVE // * / % x
	UNARY          // ! - ~ not \
	POWER          // **
	POSTFIX        // ++ --
	ARROW          // ->
	CALL           // ()
	INDEX          // [] {}
)

// precedences maps token types to precedence.
// precedences, token türlerini önceliğe eşler.
var precedences = map[lexer.TokenType]int{
	lexer.TokComma:       COMMA,
	lexer.TokFatArrow:    COMMA,
	lexer.TokAssign:      ASSIGN,
	lexer.TokPlusEq:      ASSIGN,
	lexer.TokMinusEq:     ASSIGN,
	lexer.TokStarEq:      ASSIGN,
	lexer.TokSlashEq:     ASSIGN,
	lexer.TokPercentEq:   ASSIGN,
	lexer.TokStarStarEq:  ASSIGN,
	lexer.TokDotEq:       ASSIGN,
	lexer.TokAndEq:       ASSIGN,
	lexer.TokOrEq:        ASSIGN,
	lexer.TokDefinedOrEq: ASSIGN,
	lexer.TokQuestion:    TERNARY,
	lexer.TokOr:          OR,
	lexer.TokDefinedOr:   OR,
	lexer.TokOrWord:      OR,
	lexer.TokAnd:         AND,
	lexer.TokAndWord:     AND,
	lexer.TokBitOr:       BITOR,
	lexer.TokBitXor:      BITXOR,
	lexer.TokBitAnd:      BITAND,
	lexer.TokEq:          EQUALITY,
	lexer.TokNe:          EQUALITY,
	lexer.TokStrEq:       EQUALITY,
	lexer.TokStrNe:       EQUALITY,
	lexer.TokLt:          COMPARISON,
	lexer.TokLe:          COMPARISON,
	lexer.TokGt:          COMPARISON,
	lexer.TokGe:          COMPARISON,
	lexer.TokStrLt:       COMPARISON,
	lexer.TokStrLe:       COMPARISON,
	lexer.TokStrGt:       COMPARISON,
	lexer.TokStrGe:       COMPARISON,
	lexer.TokSpaceship:   COMPARISON,
	lexer.TokCmp:         COMPARISON,
	lexer.TokLeftShift:   SHIFT,
	lexer.TokRightShift:  SHIFT,
	lexer.TokPlus:        ADDITIVE,
	lexer.TokMinus:       ADDITIVE,
	lexer.TokDot:         ADDITIVE,
	lexer.TokStar:        MULTIPLICATIVE,
	lexer.TokSlash:       MULTIPLICATIVE,
	lexer.TokPercent:     MULTIPLICATIVE,
	lexer.TokX:           MULTIPLICATIVE,
	lexer.TokStarStar:    POWER,
	lexer.TokIncr:        POSTFIX,
	lexer.TokDecr:        POSTFIX,
	lexer.TokArrow:       ARROW,
	lexer.TokLParen:      CALL,
	lexer.TokLBracket:    INDEX,
	lexer.TokLBrace:      INDEX,
	lexer.TokRange:       COMPARISON,
	lexer.TokRange3:      COMPARISON,
	lexer.TokMatch:       COMPARISON,
	lexer.TokNotMatch:    COMPARISON,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// Parser parses Perl source code into an AST.
// Parser, Perl kaynak kodunu AST'ye ayrıştırır.
type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  lexer.Token
	peekToken lexer.Token

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

// New creates a new parser.
// New, yeni bir ayrıştırıcı oluşturur.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)

	// Register prefix parsers
	// Önek ayrıştırıcıları kaydet
	p.registerPrefix(lexer.TokInteger, p.parseIntegerLiteral)
	p.registerPrefix(lexer.TokFloat, p.parseFloatLiteral)
	p.registerPrefix(lexer.TokString, p.parseStringLiteral)
	p.registerPrefix(lexer.TokRawString, p.parseRawStringLiteral)
	p.registerPrefix(lexer.TokScalar, p.parseScalarVar)
	p.registerPrefix(lexer.TokArray, p.parseArrayVar)
	p.registerPrefix(lexer.TokHash, p.parseHashVar)
	p.registerPrefix(lexer.TokCode, p.parseCodeVar)
	p.registerPrefix(lexer.TokArrayLen, p.parseArrayLengthVar)
	p.registerPrefix(lexer.TokSpecialVar, p.parseSpecialVar)
	p.registerPrefix(lexer.TokIdent, p.parseIdentifier)
	p.registerPrefix(lexer.TokUndef, p.parseUndef)
	p.registerPrefix(lexer.TokLParen, p.parseGroupedExpression)
	p.registerPrefix(lexer.TokLBracket, p.parseArrayLiteral)
	p.registerPrefix(lexer.TokLBrace, p.parseHashLiteral)
	p.registerPrefix(lexer.TokBackslash, p.parseRefExpr)
	p.registerPrefix(lexer.TokRegex, p.parseRegexLiteral)
	p.registerPrefix(lexer.TokSub, p.parseAnonSub)

	// Prefix operators
	// Önek operatörleri
	p.registerPrefix(lexer.TokMinus, p.parsePrefixExpression)
	p.registerPrefix(lexer.TokNot, p.parsePrefixExpression)
	p.registerPrefix(lexer.TokBitNot, p.parsePrefixExpression)
	p.registerPrefix(lexer.TokNotWord, p.parsePrefixExpression)
	p.registerPrefix(lexer.TokIncr, p.parsePrefixExpression)
	p.registerPrefix(lexer.TokDecr, p.parsePrefixExpression)

	// Register infix parsers
	// Ara ek ayrıştırıcıları kaydet
	p.registerInfix(lexer.TokPlus, p.parseInfixExpression)
	p.registerInfix(lexer.TokMinus, p.parseInfixExpression)
	p.registerInfix(lexer.TokStar, p.parseInfixExpression)
	p.registerInfix(lexer.TokSlash, p.parseInfixExpression)
	p.registerInfix(lexer.TokPercent, p.parseInfixExpression)
	p.registerInfix(lexer.TokStarStar, p.parseInfixExpression)
	p.registerInfix(lexer.TokDot, p.parseInfixExpression)
	p.registerInfix(lexer.TokX, p.parseInfixExpression)
	p.registerInfix(lexer.TokEq, p.parseInfixExpression)
	p.registerInfix(lexer.TokNe, p.parseInfixExpression)
	p.registerInfix(lexer.TokLt, p.parseInfixExpression)
	p.registerInfix(lexer.TokLe, p.parseInfixExpression)
	p.registerInfix(lexer.TokGt, p.parseInfixExpression)
	p.registerInfix(lexer.TokGe, p.parseInfixExpression)
	p.registerInfix(lexer.TokSpaceship, p.parseInfixExpression)
	p.registerInfix(lexer.TokStrEq, p.parseInfixExpression)
	p.registerInfix(lexer.TokStrNe, p.parseInfixExpression)
	p.registerInfix(lexer.TokStrLt, p.parseInfixExpression)
	p.registerInfix(lexer.TokStrLe, p.parseInfixExpression)
	p.registerInfix(lexer.TokStrGt, p.parseInfixExpression)
	p.registerInfix(lexer.TokStrGe, p.parseInfixExpression)
	p.registerInfix(lexer.TokCmp, p.parseInfixExpression)
	p.registerInfix(lexer.TokAnd, p.parseInfixExpression)
	p.registerInfix(lexer.TokOr, p.parseInfixExpression)
	p.registerInfix(lexer.TokDefinedOr, p.parseInfixExpression)
	p.registerInfix(lexer.TokAndWord, p.parseInfixExpression)
	p.registerInfix(lexer.TokOrWord, p.parseInfixExpression)
	p.registerInfix(lexer.TokBitAnd, p.parseInfixExpression)
	p.registerInfix(lexer.TokBitOr, p.parseInfixExpression)
	p.registerInfix(lexer.TokBitXor, p.parseInfixExpression)
	p.registerInfix(lexer.TokLeftShift, p.parseInfixExpression)
	p.registerInfix(lexer.TokRightShift, p.parseInfixExpression)
	p.registerInfix(lexer.TokRange, p.parseRangeExpression)
	p.registerInfix(lexer.TokRange3, p.parseRangeExpression)

	// Assignment
	// Atama
	p.registerInfix(lexer.TokAssign, p.parseAssignExpression)
	p.registerInfix(lexer.TokPlusEq, p.parseAssignExpression)
	p.registerInfix(lexer.TokMinusEq, p.parseAssignExpression)
	p.registerInfix(lexer.TokStarEq, p.parseAssignExpression)
	p.registerInfix(lexer.TokSlashEq, p.parseAssignExpression)
	p.registerInfix(lexer.TokPercentEq, p.parseAssignExpression)
	p.registerInfix(lexer.TokStarStarEq, p.parseAssignExpression)
	p.registerInfix(lexer.TokDotEq, p.parseAssignExpression)
	p.registerInfix(lexer.TokAndEq, p.parseAssignExpression)
	p.registerInfix(lexer.TokOrEq, p.parseAssignExpression)
	p.registerInfix(lexer.TokDefinedOrEq, p.parseAssignExpression)

	// Special infix
	// Özel ara ek
	p.registerInfix(lexer.TokQuestion, p.parseTernaryExpression)
	p.registerInfix(lexer.TokLParen, p.parseCallExpression)
	p.registerInfix(lexer.TokLBracket, p.parseIndexExpression)
	p.registerInfix(lexer.TokLBrace, p.parseHashAccessExpression)
	p.registerInfix(lexer.TokArrow, p.parseArrowExpression)
	p.registerInfix(lexer.TokIncr, p.parsePostfixExpression)
	p.registerInfix(lexer.TokDecr, p.parsePostfixExpression)
	p.registerInfix(lexer.TokMatch, p.parseMatchExpression)
	p.registerInfix(lexer.TokNotMatch, p.parseMatchExpression)

	p.registerInfix(lexer.TokFatArrow, p.parseFatArrowExpression)
	p.registerPrefix(lexer.TokBless, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokShift, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokPop, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokPush, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokPrint, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokSay, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokDie, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokWarn, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokDefined, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokRef, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokScalarKw, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokKeys, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokValues, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokEach, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokExists, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokDelete, p.parseBuiltinCall)

	// Array/Hash builtins
	p.registerPrefix(lexer.TokShift, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokUnshift, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokPop, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokPush, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokSplice, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokKeys, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokValues, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokEach, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokExists, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokDelete, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokSort, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokReverse, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokMap, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokGrep, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokJoin, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokSplit, p.parseBuiltinCall)

	// String builtins
	p.registerPrefix(lexer.TokLength, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokSubstr, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokIndex, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokRindex, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokLc, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokUc, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokLcfirst, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokUcfirst, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokChomp, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokChop, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokChr, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokOrd, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokHex, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokOct, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokPack, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokUnpack, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokSprintf, p.parseBuiltinCall)

	// Numeric builtins
	p.registerPrefix(lexer.TokAbs, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokInt, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokSqrt, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokRand, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokSrand, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokSin, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokCos, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokAtan2, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokExp, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokLog, p.parseBuiltinCall)

	// Misc builtins
	p.registerPrefix(lexer.TokLocaltime, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokGmtime, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokTime, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokSleep, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokExit, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokSystem, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokExec, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokFork, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokWait, p.parseBuiltinCall)
	p.registerPrefix(lexer.TokKill, p.parseBuiltinCall)

	// Read two tokens to initialize curToken and peekToken
	// curToken ve peekToken'ı başlatmak için iki token oku
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()

	// Skip newlines in most contexts
	// Çoğu bağlamda satır sonlarını atla
	for p.peekToken.Type == lexer.TokNewline {
		p.peekToken = p.l.NextToken()
	}
}

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
	msg := fmt.Sprintf("line %d: no prefix parse function for %v found",
		p.curToken.Line, t)
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

// ============================================================
// Main Parse Function
// Ana Ayrıştırma Fonksiyonu
// ============================================================

// ParseProgram parses the entire program.
// ParseProgram, tüm programı ayrıştırır.
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(lexer.TokEOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

// ============================================================
// Statement Parsing
// Deyim Ayrıştırma
// ============================================================

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case lexer.TokMy, lexer.TokOur, lexer.TokLocal, lexer.TokState:
		return p.parseVarDecl()
	case lexer.TokSub:
		return p.parseSubDecl()
	case lexer.TokPackage:
		return p.parsePackageDecl()
	case lexer.TokUse:
		return p.parseUseDecl()
	case lexer.TokNo:
		return p.parseNoDecl()
	case lexer.TokRequire:
		return p.parseRequireDecl()
	case lexer.TokIf:
		return p.parseIfStmt()
	case lexer.TokUnless:
		return p.parseIfStmt() // Same parser, different flag
	case lexer.TokWhile:
		return p.parseWhileStmt()
	case lexer.TokUntil:
		return p.parseWhileStmt()
	case lexer.TokFor:
		return p.parseForStmt()
	case lexer.TokForeach:
		return p.parseForeachStmt()
	case lexer.TokLast:
		return p.parseLastStmt()
	case lexer.TokNext:
		return p.parseNextStmt()
	case lexer.TokRedo:
		return p.parseRedoStmt()
	case lexer.TokReturn:
		return p.parseReturnStmt()
	case lexer.TokLBrace:
		return p.parseBlockStmt()
	case lexer.TokBEGIN, lexer.TokEND, lexer.TokCHECK, lexer.TokINIT, lexer.TokUNITCHECK:
		return p.parseSpecialBlock()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseExpressionStatement() *ast.ExprStmt {
	stmt := &ast.ExprStmt{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	// Optional semicolon
	// Opsiyonel noktalı virgül
	if p.peekTokenIs(lexer.TokSemi) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseBlockStmt() *ast.BlockStmt {
	block := &ast.BlockStmt{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken() // skip {

	for !p.curTokenIs(lexer.TokRBrace) && !p.curTokenIs(lexer.TokEOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// ============================================================
// Expression Parsing (Pratt Parser)
// İfade Ayrıştırma (Pratt Ayrıştırıcı)
// ============================================================

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(lexer.TokSemi) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

// ============================================================
// Literal Parsers
// Literal Ayrıştırıcıları
// ============================================================

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Value, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("line %d: could not parse %q as integer",
			p.curToken.Line, p.curToken.Value)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Value, 64)
	if err != nil {
		msg := fmt.Sprintf("line %d: could not parse %q as float",
			p.curToken.Line, p.curToken.Value)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{
		Token:        p.curToken,
		Value:        p.curToken.Value,
		Interpolated: true,
	}
}

func (p *Parser) parseRawStringLiteral() ast.Expression {
	return &ast.StringLiteral{
		Token:        p.curToken,
		Value:        p.curToken.Value,
		Interpolated: false,
	}
}

func (p *Parser) parseRegexLiteral() ast.Expression {
	lit := &ast.RegexLiteral{Token: p.curToken}

	// Value may contain pattern/flags
	// Değer pattern/flags içerebilir
	parts := strings.SplitN(p.curToken.Value, "/", 2)
	lit.Pattern = parts[0]
	if len(parts) > 1 {
		lit.Flags = parts[1]
	}

	return lit
}

func (p *Parser) parseUndef() ast.Expression {
	return &ast.UndefLiteral{Token: p.curToken}
}

// ============================================================
// Variable Parsers
// Değişken Ayrıştırıcıları
// ============================================================

func (p *Parser) parseScalarVar() ast.Expression {
	name := p.curToken.Value
	name = strings.TrimPrefix(name, "$")
	return &ast.ScalarVar{Token: p.curToken, Name: name}
}

func (p *Parser) parseArrayVar() ast.Expression {
	name := p.curToken.Value
	name = strings.TrimPrefix(name, "@")
	return &ast.ArrayVar{Token: p.curToken, Name: name}
}

func (p *Parser) parseHashVar() ast.Expression {
	name := p.curToken.Value
	name = strings.TrimPrefix(name, "%")
	return &ast.HashVar{Token: p.curToken, Name: name}
}

func (p *Parser) parseCodeVar() ast.Expression {
	name := p.curToken.Value
	name = strings.TrimPrefix(name, "&")
	return &ast.CodeVar{Token: p.curToken, Name: name}
}

func (p *Parser) parseArrayLengthVar() ast.Expression {
	name := p.curToken.Value
	name = strings.TrimPrefix(name, "$#")
	return &ast.ArrayLengthVar{Token: p.curToken, Name: name}
}

func (p *Parser) parseSpecialVar() ast.Expression {
	return &ast.SpecialVar{Token: p.curToken, Name: p.curToken.Value}
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Value}
}

// ============================================================
// Operator Parsers
// Operatör Ayrıştırıcıları
// ============================================================

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpr{
		Token:    p.curToken,
		Operator: p.curToken.Value,
	}
	p.nextToken()
	expression.Right = p.parseExpression(UNARY)
	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpr{
		Token:    p.curToken,
		Operator: p.curToken.Value,
		Left:     left,
	}
	precedence := p.curPrecedence()

	// Right associative for **
	// ** için sağdan ilişkili
	if p.curToken.Type == lexer.TokStarStar {
		precedence--
	}

	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}

func (p *Parser) parsePostfixExpression(left ast.Expression) ast.Expression {
	return &ast.PostfixExpr{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Value,
	}
}

func (p *Parser) parseAssignExpression(left ast.Expression) ast.Expression {
	expression := &ast.AssignExpr{
		Token:    p.curToken,
		Operator: p.curToken.Value,
		Left:     left,
	}
	p.nextToken()
	// Right associative
	expression.Right = p.parseExpression(ASSIGN - 1)
	return expression
}

func (p *Parser) parseTernaryExpression(condition ast.Expression) ast.Expression {
	expression := &ast.TernaryExpr{
		Token:     p.curToken,
		Condition: condition,
	}
	p.nextToken()
	expression.Then = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TokColon) {
		return nil
	}
	p.nextToken()
	expression.Else = p.parseExpression(TERNARY - 1)
	return expression
}

func (p *Parser) parseRangeExpression(left ast.Expression) ast.Expression {
	expression := &ast.RangeExpr{
		Token:    p.curToken,
		Start:    left,
		ThreeDot: p.curToken.Type == lexer.TokRange3,
	}
	p.nextToken()
	expression.End = p.parseExpression(COMPARISON)
	return expression
}

// ============================================================
// Access Parsers
// Erişim Ayrıştırıcıları
// ============================================================

func (p *Parser) parseGroupedExpression() ast.Expression {
	startToken := p.curToken
	p.nextToken()

	// Empty parens
	if p.curTokenIs(lexer.TokRParen) {
		return &ast.ArrayExpr{Token: startToken, Elements: []ast.Expression{}}
	}

	exp := p.parseExpression(LOWEST)

	// Check for list: (1, 2, 3)
	// Liste kontrol et: (1, 2, 3)
	if p.peekTokenIs(lexer.TokComma) {
		elements := []ast.Expression{exp}
		for p.peekTokenIs(lexer.TokComma) {
			p.nextToken() // move to ,
			if p.peekTokenIs(lexer.TokRParen) {
				break // trailing comma
			}
			p.nextToken() // skip ,
			elements = append(elements, p.parseExpression(LOWEST))
		}
		if !p.expectPeek(lexer.TokRParen) {
			return nil
		}
		return &ast.ArrayExpr{Token: startToken, Elements: elements}
	}

	if !p.expectPeek(lexer.TokRParen) {
		return nil
	}
	return exp
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.ArrayAccess{Token: p.curToken, Array: left}
	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(lexer.TokRBracket) {
		return nil
	}
	return exp
}

func (p *Parser) parseHashAccessExpression(left ast.Expression) ast.Expression {
	exp := &ast.HashAccess{Token: p.curToken, Hash: left}
	p.nextToken()
	exp.Key = p.parseExpression(LOWEST)
	if !p.expectPeek(lexer.TokRBrace) {
		return nil
	}
	return exp
}

func (p *Parser) parseArrowExpression(left ast.Expression) ast.Expression {
	token := p.curToken
	p.nextToken()

	// Check what follows ->
	// -> sonrasını kontrol et
	switch p.curToken.Type {
	case lexer.TokLBracket:
		// ->[]
		p.nextToken()
		index := p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.TokRBracket) {
			return nil
		}
		return &ast.ArrowAccess{
			Token: token,
			Left:  left,
			Right: &ast.ArrayAccess{Token: p.curToken, Index: index},
		}
	case lexer.TokLBrace:
		// ->{}
		p.nextToken()
		key := p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.TokRBrace) {
			return nil
		}
		return &ast.ArrowAccess{
			Token: token,
			Left:  left,
			Right: &ast.HashAccess{Token: p.curToken, Key: key},
		}
	case lexer.TokIdent:
		// ->method or ->method()
		method := p.curToken.Value
		if p.peekTokenIs(lexer.TokLParen) {
			p.nextToken()
			args := p.parseExpressionList(lexer.TokRParen)
			return &ast.MethodCall{
				Token:  token,
				Object: left,
				Method: method,
				Args:   args,
			}
		}
		return &ast.MethodCall{
			Token:  token,
			Object: left,
			Method: method,
			Args:   nil,
		}
	default:
		return &ast.ArrowAccess{Token: token, Left: left}
	}
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpr{Token: p.curToken, Function: function}
	exp.Args = p.parseExpressionList(lexer.TokRParen)
	return exp
}

func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(lexer.TokComma) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseMatchExpression(left ast.Expression) ast.Expression {
	exp := &ast.MatchExpr{
		Token:  p.curToken,
		Target: left,
		Negate: p.curToken.Type == lexer.TokNotMatch,
	}
	p.nextToken()
	if p.curToken.Type == lexer.TokRegex {
		exp.Pattern = p.parseRegexLiteral().(*ast.RegexLiteral)
	}
	return exp
}

// ============================================================
// Composite Literal Parsers
// Bileşik Literal Ayrıştırıcıları
// ============================================================

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayExpr{Token: p.curToken}
	array.Elements = p.parseExpressionList(lexer.TokRBracket)
	return array
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashExpr{Token: p.curToken}
	hash.Pairs = []*ast.HashPair{}

	if p.peekTokenIs(lexer.TokRBrace) {
		p.nextToken()
		return hash
	}

	p.nextToken() // move to first key
	pair := p.parseHashPair()
	if pair != nil {
		hash.Pairs = append(hash.Pairs, pair)
	}

	for p.peekTokenIs(lexer.TokComma) {
		p.nextToken() // move to comma
		if p.peekTokenIs(lexer.TokRBrace) {
			break // trailing comma
		}
		p.nextToken() // skip comma, move to next key
		pair := p.parseHashPair()
		if pair != nil {
			hash.Pairs = append(hash.Pairs, pair)
		}
	}

	if !p.expectPeek(lexer.TokRBrace) {
		return nil
	}

	return hash
}

func (p *Parser) parseHashPair() *ast.HashPair {
	key := p.parseExpression(COMMA + 1) // Higher than comma to stop at =>

	// Expect =>
	if !p.expectPeek(lexer.TokFatArrow) {
		return nil
	}
	p.nextToken() // move to value
	value := p.parseExpression(COMMA)

	return &ast.HashPair{Key: key, Value: value}
}

func (p *Parser) ParseHashPair_old() *ast.HashPair {
	key := p.parseExpression(LOWEST)

	// Expect => or ,
	// => veya , bekle
	if p.peekTokenIs(lexer.TokFatArrow) {
		p.nextToken() // consume =>
		p.nextToken() // move to value
		value := p.parseExpression(LOWEST)
		return &ast.HashPair{Key: key, Value: value}
	}

	// Comma-separated pair (old style)
	if p.peekTokenIs(lexer.TokComma) {
		p.nextToken() // consume ,
		p.nextToken() // move to value
		value := p.parseExpression(LOWEST)
		return &ast.HashPair{Key: key, Value: value}
	}

	return &ast.HashPair{Key: key, Value: nil}
}

func (p *Parser) parseRefExpr() ast.Expression {
	exp := &ast.RefExpr{Token: p.curToken}
	p.nextToken()
	exp.Value = p.parseExpression(UNARY)
	return exp
}

func (p *Parser) parseAnonSub() ast.Expression {
	exp := &ast.AnonSubExpr{Token: p.curToken}

	if !p.expectPeek(lexer.TokLBrace) {
		return nil
	}

	exp.Body = p.parseBlockStmt()
	return exp
}

// ============================================================
// Declaration Parsers
// Bildirim Ayrıştırıcıları
// ============================================================

func (p *Parser) parseVarDecl() ast.Statement {
	decl := &ast.VarDecl{Token: p.curToken, Kind: p.curToken.Value}
	decl.Names = []ast.Expression{}

	p.nextToken() // skip my/our/local/state

	if p.curTokenIs(lexer.TokLParen) {
		// List declaration: my ($x, $y)
		decl.Names = p.parseExpressionList(lexer.TokRParen)
	} else {
		// Single variable: my $x
		decl.Names = append(decl.Names, p.parseExpression(ASSIGN))
	}

	// Optional initializer
	if p.peekTokenIs(lexer.TokAssign) {
		p.nextToken()
		p.nextToken()
		decl.Value = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(lexer.TokSemi) {
		p.nextToken()
	}

	return decl
}

func (p *Parser) parseSubDecl() ast.Statement {
	decl := &ast.SubDecl{Token: p.curToken}

	if !p.expectPeek(lexer.TokIdent) {
		return nil
	}
	decl.Name = p.curToken.Value

	// Optional prototype
	// Opsiyonel prototip
	if p.peekTokenIs(lexer.TokLParen) {
		p.nextToken()
		p.nextToken()
		// Read prototype until )
		var proto strings.Builder
		for !p.curTokenIs(lexer.TokRParen) && !p.curTokenIs(lexer.TokEOF) {
			proto.WriteString(p.curToken.Value)
			p.nextToken()
		}
		decl.Prototype = proto.String()
	}

	// Optional attributes
	// Opsiyonel özellikler
	for p.peekTokenIs(lexer.TokColon) {
		p.nextToken()
		p.nextToken()
		decl.Attributes = append(decl.Attributes, p.curToken.Value)
	}

	if !p.expectPeek(lexer.TokLBrace) {
		return nil
	}

	decl.Body = p.parseBlockStmt()
	return decl
}

func (p *Parser) parsePackageDecl() ast.Statement {
	decl := &ast.PackageDecl{Token: p.curToken}

	p.nextToken()
	decl.Name = p.curToken.Value

	// Handle Package::Name
	for p.peekTokenIs(lexer.TokDoubleColon) {
		p.nextToken()
		decl.Name += p.curToken.Value
		p.nextToken()
		decl.Name += p.curToken.Value
	}

	// Optional version
	// Opsiyonel versiyon
	if p.peekTokenIs(lexer.TokFloat) || p.peekTokenIs(lexer.TokVersion) {
		p.nextToken()
		decl.Version = p.curToken.Value
	}

	// Block form or semicolon
	// Blok formu veya noktalı virgül
	if p.peekTokenIs(lexer.TokLBrace) {
		p.nextToken()
		decl.Block = p.parseBlockStmt()
	} else if p.peekTokenIs(lexer.TokSemi) {
		p.nextToken()
	}

	return decl
}

func (p *Parser) parseUseDecl() ast.Statement {
	decl := &ast.UseDecl{Token: p.curToken}

	p.nextToken()
	decl.Module = p.curToken.Value

	// Handle Module::Name
	for p.peekTokenIs(lexer.TokDoubleColon) {
		p.nextToken()
		decl.Module += p.curToken.Value
		p.nextToken()
		decl.Module += p.curToken.Value
	}

	// Optional version
	if p.peekTokenIs(lexer.TokFloat) || p.peekTokenIs(lexer.TokVersion) {
		p.nextToken()
		decl.Version = p.curToken.Value
	}

	// Optional import list
	if p.peekTokenIs(lexer.TokQw) || p.peekTokenIs(lexer.TokLParen) {
		p.nextToken()
		// TODO: Parse qw() or import list
	}

	if p.peekTokenIs(lexer.TokSemi) {
		p.nextToken()
	}

	return decl
}

func (p *Parser) parseNoDecl() ast.Statement {
	decl := &ast.NoDecl{Token: p.curToken}

	p.nextToken()
	decl.Module = p.curToken.Value

	if p.peekTokenIs(lexer.TokSemi) {
		p.nextToken()
	}

	return decl
}

func (p *Parser) parseRequireDecl() ast.Statement {
	decl := &ast.RequireDecl{Token: p.curToken}

	p.nextToken()

	if p.curTokenIs(lexer.TokString) || p.curTokenIs(lexer.TokRawString) {
		decl.Expr = p.parseExpression(LOWEST)
	} else {
		decl.Module = p.curToken.Value
	}

	if p.peekTokenIs(lexer.TokSemi) {
		p.nextToken()
	}

	return decl
}

func (p *Parser) parseSpecialBlock() ast.Statement {
	block := &ast.SpecialBlock{Token: p.curToken, Kind: p.curToken.Value}

	if !p.expectPeek(lexer.TokLBrace) {
		return nil
	}

	block.Body = p.parseBlockStmt()
	return block
}

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
	exp := &ast.CallExpr{
		Token:    p.curToken,
		Function: &ast.Identifier{Token: p.curToken, Value: p.curToken.Value},
	}

	// Check for parentheses or direct args
	// Parantez veya doğrudan argümanlar kontrol et
	if p.peekTokenIs(lexer.TokLParen) {
		p.nextToken()
		exp.Args = p.parseExpressionList(lexer.TokRParen)
	} else if !p.peekTokenIs(lexer.TokSemi) && !p.peekTokenIs(lexer.TokRBrace) &&
		!p.peekTokenIs(lexer.TokEOF) && !p.peekTokenIs(lexer.TokComma) &&
		!p.peekTokenIs(lexer.TokRParen) {
		// Parse arguments without parens until statement end
		// Parantez olmadan argümanları deyim sonuna kadar ayrıştır
		p.nextToken()
		exp.Args = []ast.Expression{p.parseExpression(COMMA)}

		// Handle multiple args separated by comma
		for p.peekTokenIs(lexer.TokComma) {
			p.nextToken()
			p.nextToken()
			exp.Args = append(exp.Args, p.parseExpression(COMMA))
		}
	}

	return exp
}
