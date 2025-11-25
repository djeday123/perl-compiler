// Package ast defines the abstract syntax tree for Perl programs.
package ast

import (
	"bytes"
	"strings"

	"github.com/djeday123/perl-compiler/pkg/token"
)

// Node represents a node in the AST.
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement represents a statement node.
type Statement interface {
	Node
	statementNode()
}

// Expression represents an expression node.
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

// Identifier represents an identifier.
type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// IntegerLiteral represents an integer literal.
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FloatLiteral represents a floating-point literal.
type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

// StringLiteral represents a string literal.
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return "\"" + sl.Value + "\"" }

// ScalarVariable represents a scalar variable ($var).
type ScalarVariable struct {
	Token token.Token
	Name  string
}

func (sv *ScalarVariable) expressionNode()      {}
func (sv *ScalarVariable) TokenLiteral() string { return sv.Token.Literal }
func (sv *ScalarVariable) String() string       { return sv.Name }

// ArrayVariable represents an array variable (@arr).
type ArrayVariable struct {
	Token token.Token
	Name  string
}

func (av *ArrayVariable) expressionNode()      {}
func (av *ArrayVariable) TokenLiteral() string { return av.Token.Literal }
func (av *ArrayVariable) String() string       { return av.Name }

// HashVariable represents a hash variable (%hash).
type HashVariable struct {
	Token token.Token
	Name  string
}

func (hv *HashVariable) expressionNode()      {}
func (hv *HashVariable) TokenLiteral() string { return hv.Token.Literal }
func (hv *HashVariable) String() string       { return hv.Name }

// PrefixExpression represents a prefix expression (!expr, -expr, etc.).
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

// InfixExpression represents an infix expression (a + b, a * b, etc.).
type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

// PostfixExpression represents a postfix expression (expr++, expr--).
type PostfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
}

func (poe *PostfixExpression) expressionNode()      {}
func (poe *PostfixExpression) TokenLiteral() string { return poe.Token.Literal }
func (poe *PostfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(poe.Left.String())
	out.WriteString(poe.Operator)
	out.WriteString(")")
	return out.String()
}

// AssignmentStatement represents a variable assignment.
type AssignmentStatement struct {
	Token token.Token // the '=' token
	Name  Expression  // the variable being assigned to
	Value Expression  // the value being assigned
}

func (as *AssignmentStatement) statementNode()       {}
func (as *AssignmentStatement) TokenLiteral() string { return as.Token.Literal }
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

// MyStatement represents a my declaration (my $var = value;).
type MyStatement struct {
	Token token.Token // the 'my' token
	Name  Expression  // the variable being declared
	Value Expression  // the initial value (optional)
}

func (ms *MyStatement) statementNode()       {}
func (ms *MyStatement) TokenLiteral() string { return ms.Token.Literal }
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

// ExpressionStatement represents an expression as a statement.
type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// BlockStatement represents a block of statements { ... }.
type BlockStatement struct {
	Token      token.Token // the '{' token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
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
	Token       token.Token // the 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative Statement // can be BlockStatement or another IfStatement (elsif)
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
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
	Token     token.Token // the 'while' token
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
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
	Token    token.Token // the 'for' or 'foreach' token
	Variable Expression  // loop variable
	List     Expression  // the list to iterate over
	Body     *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	var out bytes.Buffer
	out.WriteString("for ")
	if fs.Variable != nil {
		out.WriteString(fs.Variable.String())
		out.WriteString(" ")
	}
	out.WriteString("(")
	out.WriteString(fs.List.String())
	out.WriteString(") ")
	out.WriteString(fs.Body.String())
	return out.String()
}

// SubroutineStatement represents a subroutine definition.
type SubroutineStatement struct {
	Token      token.Token // the 'sub' token
	Name       *Identifier
	Parameters []*ScalarVariable
	Body       *BlockStatement
}

func (ss *SubroutineStatement) statementNode()       {}
func (ss *SubroutineStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *SubroutineStatement) String() string {
	var out bytes.Buffer
	out.WriteString("sub ")
	out.WriteString(ss.Name.String())
	out.WriteString(" ")
	out.WriteString(ss.Body.String())
	return out.String()
}

// ReturnStatement represents a return statement.
type ReturnStatement struct {
	Token       token.Token // the 'return' token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString("return ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

// CallExpression represents a subroutine/function call.
type CallExpression struct {
	Token     token.Token  // the '(' token
	Function  Expression   // Identifier or other expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// PrintStatement represents a print statement.
type PrintStatement struct {
	Token     token.Token // the 'print' token
	Arguments []Expression
}

func (ps *PrintStatement) statementNode()       {}
func (ps *PrintStatement) TokenLiteral() string { return ps.Token.Literal }
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

// ArrayLiteral represents an array literal (list).
type ArrayLiteral struct {
	Token    token.Token // the '(' token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer
	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}
	out.WriteString("(")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString(")")
	return out.String()
}

// HashLiteral represents a hash literal.
type HashLiteral struct {
	Token token.Token
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }
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

// IndexExpression represents an array/hash index access.
type IndexExpression struct {
	Token token.Token // the '[' or '{' token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	if ie.Token.Type == token.LBRACKET {
		out.WriteString("[")
		out.WriteString(ie.Index.String())
		out.WriteString("]")
	} else {
		out.WriteString("{")
		out.WriteString(ie.Index.String())
		out.WriteString("}")
	}
	out.WriteString(")")
	return out.String()
}

// RangeExpression represents a range expression (1..10).
type RangeExpression struct {
	Token token.Token
	Start Expression
	End   Expression
}

func (re *RangeExpression) expressionNode()      {}
func (re *RangeExpression) TokenLiteral() string { return re.Token.Literal }
func (re *RangeExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(re.Start.String())
	out.WriteString("..")
	out.WriteString(re.End.String())
	out.WriteString(")")
	return out.String()
}

// TernaryExpression represents a ternary expression (cond ? true : false).
type TernaryExpression struct {
	Token       token.Token // the '?' token
	Condition   Expression
	Consequence Expression
	Alternative Expression
}

func (te *TernaryExpression) expressionNode()      {}
func (te *TernaryExpression) TokenLiteral() string { return te.Token.Literal }
func (te *TernaryExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(te.Condition.String())
	out.WriteString(" ? ")
	out.WriteString(te.Consequence.String())
	out.WriteString(" : ")
	out.WriteString(te.Alternative.String())
	out.WriteString(")")
	return out.String()
}
