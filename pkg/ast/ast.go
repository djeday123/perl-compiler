// Package ast defines the Abstract Syntax Tree for Perl source code.
package ast

import (
	"bytes"
	"strings"
)

// Node is the interface that all AST nodes must implement.
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement is a node that represents a statement.
type Statement interface {
	Node
	statementNode()
}

// Expression is a node that represents an expression.
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of every AST.
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
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// Identifier represents a variable or function name.
type Identifier struct {
	Token string // the token literal
	Value string // the identifier's value
	Sigil string // $, @, or % for variables
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token }
func (i *Identifier) String() string       { return i.Sigil + i.Value }

// IntegerLiteral represents an integer value.
type IntegerLiteral struct {
	Token string
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token }
func (il *IntegerLiteral) String() string       { return il.Token }

// FloatLiteral represents a floating point value.
type FloatLiteral struct {
	Token string
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token }
func (fl *FloatLiteral) String() string       { return fl.Token }

// StringLiteral represents a string value.
type StringLiteral struct {
	Token string
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token }
func (sl *StringLiteral) String() string       { return "\"" + sl.Value + "\"" }

// PrefixExpression represents a prefix operator expression like !expr or -expr.
type PrefixExpression struct {
	Token    string     // The prefix token, e.g. !
	Operator string     // The operator
	Right    Expression // The expression to the right of the operator
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

// InfixExpression represents an infix operator expression like a + b.
type InfixExpression struct {
	Token    string     // The operator token, e.g. +
	Left     Expression // The left-hand side expression
	Operator string     // The operator
	Right    Expression // The right-hand side expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

// AssignmentStatement represents a variable assignment.
type AssignmentStatement struct {
	Token string      // the '=' token
	Name  *Identifier // the variable name
	Value Expression  // the value to assign
}

func (as *AssignmentStatement) statementNode()       {}
func (as *AssignmentStatement) TokenLiteral() string { return as.Token }
func (as *AssignmentStatement) String() string {
	var out bytes.Buffer
	out.WriteString(as.Name.String())
	out.WriteString(" = ")
	if as.Value != nil {
		out.WriteString(as.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// MyStatement represents a 'my' variable declaration.
type MyStatement struct {
	Token string      // the 'my' token
	Name  *Identifier // the variable name
	Value Expression  // optional initial value
}

func (ms *MyStatement) statementNode()       {}
func (ms *MyStatement) TokenLiteral() string { return ms.Token }
func (ms *MyStatement) String() string {
	var out bytes.Buffer
	out.WriteString("my ")
	out.WriteString(ms.Name.String())
	if ms.Value != nil {
		out.WriteString(" = ")
		out.WriteString(ms.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// ReturnStatement represents a return statement.
type ReturnStatement struct {
	Token       string     // the 'return' token
	ReturnValue Expression // the value to return
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString("return ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

// ExpressionStatement represents a statement consisting of a single expression.
type ExpressionStatement struct {
	Token      string     // the first token of the expression
	Expression Expression // the expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String() + ";"
	}
	return ""
}

// BlockStatement represents a block of statements enclosed in braces.
type BlockStatement struct {
	Token      string      // the '{' token
	Statements []Statement // the statements in the block
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	out.WriteString("{ ")
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	out.WriteString(" }")
	return out.String()
}

// IfStatement represents an if/elsif/else statement.
type IfStatement struct {
	Token       string          // the 'if' token
	Condition   Expression      // the condition
	Consequence *BlockStatement // the block to execute if condition is true
	Alternative Statement       // optional else/elsif
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token }
func (is *IfStatement) String() string {
	var out bytes.Buffer
	out.WriteString("if (")
	out.WriteString(is.Condition.String())
	out.WriteString(") ")
	out.WriteString(is.Consequence.String())
	if is.Alternative != nil {
		out.WriteString(" else ")
		out.WriteString(is.Alternative.String())
	}
	return out.String()
}

// WhileStatement represents a while loop.
type WhileStatement struct {
	Token     string          // the 'while' token
	Condition Expression      // the condition
	Body      *BlockStatement // the loop body
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token }
func (ws *WhileStatement) String() string {
	var out bytes.Buffer
	out.WriteString("while (")
	out.WriteString(ws.Condition.String())
	out.WriteString(") ")
	out.WriteString(ws.Body.String())
	return out.String()
}

// ForStatement represents a for/foreach loop.
type ForStatement struct {
	Token    string          // the 'for'/'foreach' token
	Variable *Identifier     // the loop variable
	List     Expression      // the list to iterate over
	Body     *BlockStatement // the loop body
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token }
func (fs *ForStatement) String() string {
	var out bytes.Buffer
	out.WriteString("for ")
	out.WriteString(fs.Variable.String())
	out.WriteString(" (")
	out.WriteString(fs.List.String())
	out.WriteString(") ")
	out.WriteString(fs.Body.String())
	return out.String()
}

// SubroutineStatement represents a subroutine definition.
type SubroutineStatement struct {
	Token      string          // the 'sub' token
	Name       string          // the subroutine name
	Parameters []*Identifier   // the parameters
	Body       *BlockStatement // the subroutine body
}

func (ss *SubroutineStatement) statementNode()       {}
func (ss *SubroutineStatement) TokenLiteral() string { return ss.Token }
func (ss *SubroutineStatement) String() string {
	var out bytes.Buffer
	out.WriteString("sub ")
	out.WriteString(ss.Name)
	out.WriteString(" ")
	out.WriteString(ss.Body.String())
	return out.String()
}

// CallExpression represents a subroutine or function call.
type CallExpression struct {
	Token     string       // the function name token
	Function  string       // the function name
	Arguments []Expression // the arguments
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Function)
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// PrintStatement represents a print statement.
type PrintStatement struct {
	Token     string       // the 'print' token
	Arguments []Expression // the arguments to print
}

func (ps *PrintStatement) statementNode()       {}
func (ps *PrintStatement) TokenLiteral() string { return ps.Token }
func (ps *PrintStatement) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range ps.Arguments {
		args = append(args, a.String())
	}
	out.WriteString("print ")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(";")
	return out.String()
}

// ArrayLiteral represents an array literal.
type ArrayLiteral struct {
	Token    string       // the '[' or '(' token
	Elements []Expression // the array elements
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer
	elements := []string{}
	for _, e := range al.Elements {
		elements = append(elements, e.String())
	}
	out.WriteString("(")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString(")")
	return out.String()
}

// HashLiteral represents a hash literal.
type HashLiteral struct {
	Token string                    // the '{' or '(' token
	Pairs map[Expression]Expression // the key-value pairs
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token }
func (hl *HashLiteral) String() string {
	var out bytes.Buffer
	pairs := []string{}
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+" => "+value.String())
	}
	out.WriteString("(")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString(")")
	return out.String()
}

// IndexExpression represents an array or hash access.
type IndexExpression struct {
	Token string     // the '[' or '{' token
	Left  Expression // the array or hash
	Index Expression // the index or key
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")
	return out.String()
}

// Boolean represents a boolean value (though Perl doesn't have explicit booleans).
type Boolean struct {
	Token string
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token }
func (b *Boolean) String() string       { return b.Token }

// RangeExpression represents a range expression like 1..10.
type RangeExpression struct {
	Token string     // the '..' token
	Start Expression // start of range
	End   Expression // end of range
}

func (re *RangeExpression) expressionNode()      {}
func (re *RangeExpression) TokenLiteral() string { return re.Token }
func (re *RangeExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(re.Start.String())
	out.WriteString("..")
	out.WriteString(re.End.String())
	out.WriteString(")")
	return out.String()
}
