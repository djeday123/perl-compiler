package ast

import (
	"fmt"
	"strings"

	"perlc/pkg/lexer"
)

// ============================================================
// Block Statement
// Blok İfadesi
// ============================================================

// BlockStmt represents { statements }.
// BlockStmt, { statements }'ı temsil eder.
type BlockStmt struct {
	Token      lexer.Token
	Statements []Statement
}

func (bs *BlockStmt) statementNode()       {}
func (bs *BlockStmt) TokenLiteral() string { return bs.Token.Value }
func (bs *BlockStmt) String() string {
	var out strings.Builder
	out.WriteString("{ ")
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	out.WriteString(" }")
	return out.String()
}

// ============================================================
// Expression Statement
// İfade Deyimi
// ============================================================

// ExprStmt wraps an expression as a statement.
// ExprStmt, bir ifadeyi deyim olarak sarar.
type ExprStmt struct {
	Token      lexer.Token
	Expression Expression
}

func (es *ExprStmt) statementNode()       {}
func (es *ExprStmt) TokenLiteral() string { return es.Token.Value }
func (es *ExprStmt) String() string {
	if es.Expression != nil {
		return es.Expression.String() + ";"
	}
	return ""
}

// ============================================================
// Control Flow Statements
// Kontrol Akışı Deyimleri
// ============================================================

// IfStmt represents if/unless/elsif/else.
// IfStmt, if/unless/elsif/else'i temsil eder.
type IfStmt struct {
	Token     lexer.Token
	Unless    bool // true for 'unless'
	Condition Expression
	Then      *BlockStmt
	Elsif     []*ElsifClause
	Else      *BlockStmt
}

type ElsifClause struct {
	Condition Expression
	Body      *BlockStmt
}

func (is *IfStmt) statementNode()       {}
func (is *IfStmt) TokenLiteral() string { return is.Token.Value }
func (is *IfStmt) String() string {
	var out strings.Builder
	kw := "if"
	if is.Unless {
		kw = "unless"
	}
	out.WriteString(fmt.Sprintf("%s (%s) %s", kw, is.Condition.String(), is.Then.String()))
	for _, elsif := range is.Elsif {
		out.WriteString(fmt.Sprintf(" elsif (%s) %s", elsif.Condition.String(), elsif.Body.String()))
	}
	if is.Else != nil {
		out.WriteString(" else ")
		out.WriteString(is.Else.String())
	}
	return out.String()
}

// WhileStmt represents while/until loops.
// WhileStmt, while/until döngülerini temsil eder.
type WhileStmt struct {
	Token     lexer.Token
	Until     bool // true for 'until'
	Condition Expression
	Body      *BlockStmt
	Continue  *BlockStmt // continue block
}

func (ws *WhileStmt) statementNode()       {}
func (ws *WhileStmt) TokenLiteral() string { return ws.Token.Value }
func (ws *WhileStmt) String() string {
	kw := "while"
	if ws.Until {
		kw = "until"
	}
	out := fmt.Sprintf("%s (%s) %s", kw, ws.Condition.String(), ws.Body.String())
	if ws.Continue != nil {
		out += " continue " + ws.Continue.String()
	}
	return out
}

// ForStmt represents C-style for loop.
// ForStmt, C-tarzı for döngüsünü temsil eder.
type ForStmt struct {
	Token     lexer.Token
	Init      Statement
	Condition Expression
	Post      Expression
	Body      *BlockStmt
}

func (fs *ForStmt) statementNode()       {}
func (fs *ForStmt) TokenLiteral() string { return fs.Token.Value }
func (fs *ForStmt) String() string {
	init := ""
	if fs.Init != nil {
		init = fs.Init.String()
	}
	cond := ""
	if fs.Condition != nil {
		cond = fs.Condition.String()
	}
	post := ""
	if fs.Post != nil {
		post = fs.Post.String()
	}
	return fmt.Sprintf("for (%s; %s; %s) %s", init, cond, post, fs.Body.String())
}

// ForeachStmt represents foreach loop.
// ForeachStmt, foreach döngüsünü temsil eder.
type ForeachStmt struct {
	Token    lexer.Token
	Variable Expression // $item
	List     Expression // @array or (list)
	Body     *BlockStmt
	Continue *BlockStmt
}

func (fs *ForeachStmt) statementNode()       {}
func (fs *ForeachStmt) TokenLiteral() string { return fs.Token.Value }
func (fs *ForeachStmt) String() string {
	out := fmt.Sprintf("foreach %s (%s) %s",
		fs.Variable.String(), fs.List.String(), fs.Body.String())
	if fs.Continue != nil {
		out += " continue " + fs.Continue.String()
	}
	return out
}

// ============================================================
// Loop Control Statements
// Döngü Kontrol Deyimleri
// ============================================================

// LastStmt represents 'last' (break).
// LastStmt, 'last' (break)'i temsil eder.
type LastStmt struct {
	Token lexer.Token
	Label string // Optional label
}

func (ls *LastStmt) statementNode()       {}
func (ls *LastStmt) TokenLiteral() string { return ls.Token.Value }
func (ls *LastStmt) String() string {
	if ls.Label != "" {
		return "last " + ls.Label
	}
	return "last"
}

// NextStmt represents 'next' (continue).
// NextStmt, 'next' (continue)'i temsil eder.
type NextStmt struct {
	Token lexer.Token
	Label string
}

func (ns *NextStmt) statementNode()       {}
func (ns *NextStmt) TokenLiteral() string { return ns.Token.Value }
func (ns *NextStmt) String() string {
	if ns.Label != "" {
		return "next " + ns.Label
	}
	return "next"
}

// RedoStmt represents 'redo'.
// RedoStmt, 'redo'yu temsil eder.
type RedoStmt struct {
	Token lexer.Token
	Label string
}

func (rs *RedoStmt) statementNode()       {}
func (rs *RedoStmt) TokenLiteral() string { return rs.Token.Value }
func (rs *RedoStmt) String() string {
	if rs.Label != "" {
		return "redo " + rs.Label
	}
	return "redo"
}

// ReturnStmt represents 'return'.
// ReturnStmt, 'return'ü temsil eder.
type ReturnStmt struct {
	Token lexer.Token
	Value Expression
}

func (rs *ReturnStmt) statementNode()       {}
func (rs *ReturnStmt) TokenLiteral() string { return rs.Token.Value }
func (rs *ReturnStmt) String() string {
	if rs.Value != nil {
		return "return " + rs.Value.String()
	}
	return "return"
}

// ============================================================
// Modifier Statements (postfix if/unless/while/until/for)
// Değiştirici Deyimler (sonek if/unless/while/until/for)
// ============================================================

// ModifierStmt represents statement if/unless/while/until condition.
// ModifierStmt, statement if/unless/while/until condition'ı temsil eder.
type ModifierStmt struct {
	Token     lexer.Token
	Statement Statement
	Modifier  string // "if", "unless", "while", "until", "for", "foreach"
	Condition Expression
}

func (ms *ModifierStmt) statementNode()       {}
func (ms *ModifierStmt) TokenLiteral() string { return ms.Token.Value }
func (ms *ModifierStmt) String() string {
	return fmt.Sprintf("%s %s %s",
		ms.Statement.String(), ms.Modifier, ms.Condition.String())
}

// ============================================================
// Do/Eval Statements
// Do/Eval Deyimleri
// ============================================================

// DoStmt represents do { } while/until or do EXPR.
// DoStmt, do { } while/until veya do EXPR'i temsil eder.
type DoStmt struct {
	Token     lexer.Token
	Body      *BlockStmt
	Condition Expression // for do-while
	Until     bool       // true for do-until
}

func (ds *DoStmt) statementNode()       {}
func (ds *DoStmt) TokenLiteral() string { return ds.Token.Value }
func (ds *DoStmt) String() string {
	if ds.Condition != nil {
		kw := "while"
		if ds.Until {
			kw = "until"
		}
		return fmt.Sprintf("do %s %s (%s)", ds.Body.String(), kw, ds.Condition.String())
	}
	return "do " + ds.Body.String()
}

// EvalStmt represents eval { } or eval EXPR.
// EvalStmt, eval { } veya eval EXPR'i temsil eder.
type EvalStmt struct {
	Token lexer.Token
	Body  *BlockStmt // eval { }
	Expr  Expression // eval EXPR
}

func (es *EvalStmt) statementNode()       {}
func (es *EvalStmt) TokenLiteral() string { return es.Token.Value }
func (es *EvalStmt) String() string {
	if es.Body != nil {
		return "eval " + es.Body.String()
	}
	return "eval " + es.Expr.String()
}

// ============================================================
// Label Statement
// Etiket Deyimi
// ============================================================

// LabelStmt represents LABEL: statement.
// LabelStmt, LABEL: statement'ı temsil eder.
type LabelStmt struct {
	Token     lexer.Token
	Label     string
	Statement Statement
}

func (ls *LabelStmt) statementNode()       {}
func (ls *LabelStmt) TokenLiteral() string { return ls.Token.Value }
func (ls *LabelStmt) String() string {
	return ls.Label + ": " + ls.Statement.String()
}

// ============================================================
// Given/When (switch)
// Given/When (switch)
// ============================================================

// GivenStmt represents given/when/default.
// GivenStmt, given/when/default'u temsil eder.
type GivenStmt struct {
	Token   lexer.Token
	Topic   Expression
	Clauses []*WhenClause
	Default *BlockStmt
}

type WhenClause struct {
	Condition Expression
	Body      *BlockStmt
}

func (gs *GivenStmt) statementNode()       {}
func (gs *GivenStmt) TokenLiteral() string { return gs.Token.Value }
func (gs *GivenStmt) String() string {
	var out strings.Builder
	out.WriteString(fmt.Sprintf("given (%s) { ", gs.Topic.String()))
	for _, w := range gs.Clauses {
		out.WriteString(fmt.Sprintf("when (%s) %s ", w.Condition.String(), w.Body.String()))
	}
	if gs.Default != nil {
		out.WriteString("default ")
		out.WriteString(gs.Default.String())
	}
	out.WriteString("}")
	return out.String()
}
