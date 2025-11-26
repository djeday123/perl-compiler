// Package ast defines the Abstract Syntax Tree for Perl.
// Paket ast, Perl için Soyut Sözdizimi Ağacını tanımlar.
package ast

import "perlc/pkg/lexer"

// Node is the interface for all AST nodes.
// Node, tüm AST düğümleri için arayüzdür.
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement is an AST node that represents a statement.
// Statement, bir ifadeyi temsil eden AST düğümüdür.
type Statement interface {
	Node
	statementNode()
}

// Expression is an AST node that represents an expression.
// Expression, bir ifadeyi temsil eden AST düğümüdür.
type Expression interface {
	Node
	expressionNode()
}

// Declaration is an AST node that represents a declaration.
// Declaration, bir bildirimi temsil eden AST düğümüdür.
type Declaration interface {
	Node
	declarationNode()
}

// Program is the root node of the AST.
// Program, AST'nin kök düğümüdür.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out string
	for _, s := range p.Statements {
		out += s.String()
	}
	return out
}

// Position stores source location information.
// Position, kaynak konum bilgisini saklar.
type Position struct {
	File   string
	Line   int
	Column int
}

// FromToken creates Position from a Token.
// FromToken, Token'dan Position oluşturur.
func FromToken(tok lexer.Token) Position {
	return Position{
		File:   tok.File,
		Line:   tok.Line,
		Column: tok.Column,
	}
}
