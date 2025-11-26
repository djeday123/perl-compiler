package ast

import (
	"fmt"
	"strings"

	"perlc/pkg/lexer"
)

// ============================================================
// Literal Expressions
// Literal İfadeler
// ============================================================

// IntegerLiteral represents an integer literal.
// IntegerLiteral, bir tamsayı literalini temsil eder.
type IntegerLiteral struct {
	Token lexer.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Value }
func (il *IntegerLiteral) String() string       { return il.Token.Value }

// FloatLiteral represents a floating-point literal.
// FloatLiteral, bir kayan nokta literalini temsil eder.
type FloatLiteral struct {
	Token lexer.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Value }
func (fl *FloatLiteral) String() string       { return fl.Token.Value }

// StringLiteral represents a string literal.
// StringLiteral, bir string literalini temsil eder.
type StringLiteral struct {
	Token        lexer.Token
	Value        string
	Interpolated bool // true for "", false for ''
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Value }
func (sl *StringLiteral) String() string {
	if sl.Interpolated {
		return fmt.Sprintf(`"%s"`, sl.Value)
	}
	return fmt.Sprintf(`'%s'`, sl.Value)
}

// RegexLiteral represents a regex literal /pattern/flags.
// RegexLiteral, bir regex literalini temsil eder /pattern/flags.
type RegexLiteral struct {
	Token   lexer.Token
	Pattern string
	Flags   string
}

func (rl *RegexLiteral) expressionNode()      {}
func (rl *RegexLiteral) TokenLiteral() string { return rl.Token.Value }
func (rl *RegexLiteral) String() string       { return fmt.Sprintf("/%s/%s", rl.Pattern, rl.Flags) }

// UndefLiteral represents undef.
// UndefLiteral, undef'i temsil eder.
type UndefLiteral struct {
	Token lexer.Token
}

func (ul *UndefLiteral) expressionNode()      {}
func (ul *UndefLiteral) TokenLiteral() string { return "undef" }
func (ul *UndefLiteral) String() string       { return "undef" }

// ============================================================
// Variable Expressions
// Değişken İfadeleri
// ============================================================

// ScalarVar represents $var.
// ScalarVar, $var'ı temsil eder.
type ScalarVar struct {
	Token lexer.Token
	Name  string // Without sigil / Sigil olmadan
}

func (sv *ScalarVar) expressionNode()      {}
func (sv *ScalarVar) TokenLiteral() string { return sv.Token.Value }
func (sv *ScalarVar) String() string       { return "$" + sv.Name }

// ArrayVar represents @arr.
// ArrayVar, @arr'ı temsil eder.
type ArrayVar struct {
	Token lexer.Token
	Name  string
}

func (av *ArrayVar) expressionNode()      {}
func (av *ArrayVar) TokenLiteral() string { return av.Token.Value }
func (av *ArrayVar) String() string       { return "@" + av.Name }

// HashVar represents %hash.
// HashVar, %hash'i temsil eder.
type HashVar struct {
	Token lexer.Token
	Name  string
}

func (hv *HashVar) expressionNode()      {}
func (hv *HashVar) TokenLiteral() string { return hv.Token.Value }
func (hv *HashVar) String() string       { return "%" + hv.Name }

// CodeVar represents &sub.
// CodeVar, &sub'ı temsil eder.
type CodeVar struct {
	Token lexer.Token
	Name  string
}

func (cv *CodeVar) expressionNode()      {}
func (cv *CodeVar) TokenLiteral() string { return cv.Token.Value }
func (cv *CodeVar) String() string       { return "&" + cv.Name }

// GlobVar represents *glob.
// GlobVar, *glob'u temsil eder.
type GlobVar struct {
	Token lexer.Token
	Name  string
}

func (gv *GlobVar) expressionNode()      {}
func (gv *GlobVar) TokenLiteral() string { return gv.Token.Value }
func (gv *GlobVar) String() string       { return "*" + gv.Name }

// ArrayLengthVar represents $#arr.
// ArrayLengthVar, $#arr'ı temsil eder.
type ArrayLengthVar struct {
	Token lexer.Token
	Name  string
}

func (al *ArrayLengthVar) expressionNode()      {}
func (al *ArrayLengthVar) TokenLiteral() string { return al.Token.Value }
func (al *ArrayLengthVar) String() string       { return "$#" + al.Name }

// SpecialVar represents special variables like $_, $@, etc.
// SpecialVar, $_, $@ gibi özel değişkenleri temsil eder.
type SpecialVar struct {
	Token lexer.Token
	Name  string // Full name including sigil / Sigil dahil tam isim
}

func (sv *SpecialVar) expressionNode()      {}
func (sv *SpecialVar) TokenLiteral() string { return sv.Token.Value }
func (sv *SpecialVar) String() string       { return sv.Name }

// ============================================================
// Operator Expressions
// Operatör İfadeleri
// ============================================================

// PrefixExpr represents prefix operators like !$x, -$x, ++$x.
// PrefixExpr, !$x, -$x, ++$x gibi önek operatörlerini temsil eder.
type PrefixExpr struct {
	Token    lexer.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpr) expressionNode()      {}
func (pe *PrefixExpr) TokenLiteral() string { return pe.Token.Value }
func (pe *PrefixExpr) String() string {
	return fmt.Sprintf("(%s%s)", pe.Operator, pe.Right.String())
}

// PostfixExpr represents postfix operators like $x++, $x--.
// PostfixExpr, $x++, $x-- gibi sonek operatörlerini temsil eder.
type PostfixExpr struct {
	Token    lexer.Token
	Left     Expression
	Operator string
}

func (pe *PostfixExpr) expressionNode()      {}
func (pe *PostfixExpr) TokenLiteral() string { return pe.Token.Value }
func (pe *PostfixExpr) String() string {
	return fmt.Sprintf("(%s%s)", pe.Left.String(), pe.Operator)
}

// InfixExpr represents binary operators like $x + $y.
// InfixExpr, $x + $y gibi ikili operatörleri temsil eder.
type InfixExpr struct {
	Token    lexer.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpr) expressionNode()      {}
func (ie *InfixExpr) TokenLiteral() string { return ie.Token.Value }
func (ie *InfixExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", ie.Left.String(), ie.Operator, ie.Right.String())
}

// TernaryExpr represents $cond ? $then : $else.
// TernaryExpr, $cond ? $then : $else'i temsil eder.
type TernaryExpr struct {
	Token     lexer.Token
	Condition Expression
	Then      Expression
	Else      Expression
}

func (te *TernaryExpr) expressionNode()      {}
func (te *TernaryExpr) TokenLiteral() string { return te.Token.Value }
func (te *TernaryExpr) String() string {
	return fmt.Sprintf("(%s ? %s : %s)",
		te.Condition.String(), te.Then.String(), te.Else.String())
}

// AssignExpr represents assignment $x = $y.
// AssignExpr, $x = $y atamasını temsil eder.
type AssignExpr struct {
	Token    lexer.Token
	Left     Expression
	Operator string // =, +=, -=, etc.
	Right    Expression
}

func (ae *AssignExpr) expressionNode()      {}
func (ae *AssignExpr) TokenLiteral() string { return ae.Token.Value }
func (ae *AssignExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", ae.Left.String(), ae.Operator, ae.Right.String())
}

// ============================================================
// Access Expressions
// Erişim İfadeleri
// ============================================================

// ArrayAccess represents $arr[index].
// ArrayAccess, $arr[index]'i temsil eder.
type ArrayAccess struct {
	Token lexer.Token
	Array Expression
	Index Expression
}

func (aa *ArrayAccess) expressionNode()      {}
func (aa *ArrayAccess) TokenLiteral() string { return aa.Token.Value }
func (aa *ArrayAccess) String() string {
	return fmt.Sprintf("%s[%s]", aa.Array.String(), aa.Index.String())
}

// HashAccess represents $hash{key}.
// HashAccess, $hash{key}'i temsil eder.
type HashAccess struct {
	Token lexer.Token
	Hash  Expression
	Key   Expression
}

func (ha *HashAccess) expressionNode()      {}
func (ha *HashAccess) TokenLiteral() string { return ha.Token.Value }
func (ha *HashAccess) String() string {
	return fmt.Sprintf("%s{%s}", ha.Hash.String(), ha.Key.String())
}

// ArrowAccess represents $ref->[index] or $ref->{key} or $obj->method.
// ArrowAccess, $ref->[index], $ref->{key} veya $obj->method'u temsil eder.
type ArrowAccess struct {
	Token lexer.Token
	Left  Expression
	Right Expression // ArrayAccess, HashAccess, or CallExpr
}

func (aa *ArrowAccess) expressionNode()      {}
func (aa *ArrowAccess) TokenLiteral() string { return aa.Token.Value }
func (aa *ArrowAccess) String() string {
	return fmt.Sprintf("%s->%s", aa.Left.String(), aa.Right.String())
}

// ============================================================
// Call Expressions
// Çağrı İfadeleri
// ============================================================

// CallExpr represents function/method call.
// CallExpr, fonksiyon/method çağrısını temsil eder.
type CallExpr struct {
	Token    lexer.Token
	Function Expression // Identifier or expression
	Args     []Expression
}

func (ce *CallExpr) expressionNode()      {}
func (ce *CallExpr) TokenLiteral() string { return ce.Token.Value }
func (ce *CallExpr) String() string {
	args := make([]string, len(ce.Args))
	for i, a := range ce.Args {
		args[i] = a.String()
	}
	return fmt.Sprintf("%s(%s)", ce.Function.String(), strings.Join(args, ", "))
}

// MethodCall represents $obj->method(args).
// MethodCall, $obj->method(args)'ı temsil eder.
type MethodCall struct {
	Token  lexer.Token
	Object Expression
	Method string
	Args   []Expression
}

func (mc *MethodCall) expressionNode()      {}
func (mc *MethodCall) TokenLiteral() string { return mc.Token.Value }
func (mc *MethodCall) String() string {
	args := make([]string, len(mc.Args))
	for i, a := range mc.Args {
		args[i] = a.String()
	}
	return fmt.Sprintf("%s->%s(%s)", mc.Object.String(), mc.Method, strings.Join(args, ", "))
}

// ============================================================
// Composite Expressions
// Bileşik İfadeler
// ============================================================

// ArrayExpr represents [...] array literal.
// ArrayExpr, [...] dizi literalini temsil eder.
type ArrayExpr struct {
	Token    lexer.Token
	Elements []Expression
}

func (ae *ArrayExpr) expressionNode()      {}
func (ae *ArrayExpr) TokenLiteral() string { return ae.Token.Value }
func (ae *ArrayExpr) String() string {
	elements := make([]string, len(ae.Elements))
	for i, e := range ae.Elements {
		elements[i] = e.String()
	}
	return "[" + strings.Join(elements, ", ") + "]"
}

// HashExpr represents {...} hash literal.
// HashExpr, {...} hash literalini temsil eder.
type HashExpr struct {
	Token lexer.Token
	Pairs []*HashPair
}

type HashPair struct {
	Key   Expression
	Value Expression
}

func (he *HashExpr) expressionNode()      {}
func (he *HashExpr) TokenLiteral() string { return he.Token.Value }
func (he *HashExpr) String() string {
	pairs := make([]string, len(he.Pairs))
	for i, p := range he.Pairs {
		pairs[i] = fmt.Sprintf("%s => %s", p.Key.String(), p.Value.String())
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}

// RangeExpr represents $a .. $b or $a ... $b.
// RangeExpr, $a .. $b veya $a ... $b'yi temsil eder.
type RangeExpr struct {
	Token    lexer.Token
	Start    Expression
	End      Expression
	ThreeDot bool // ... vs ..
}

func (re *RangeExpr) expressionNode()      {}
func (re *RangeExpr) TokenLiteral() string { return re.Token.Value }
func (re *RangeExpr) String() string {
	op := ".."
	if re.ThreeDot {
		op = "..."
	}
	return fmt.Sprintf("(%s %s %s)", re.Start.String(), op, re.End.String())
}

// ============================================================
// Reference Expressions
// Referans İfadeleri
// ============================================================

// RefExpr represents \$var (reference creation).
// RefExpr, \$var (referans oluşturma) temsil eder.
type RefExpr struct {
	Token lexer.Token
	Value Expression
}

func (re *RefExpr) expressionNode()      {}
func (re *RefExpr) TokenLiteral() string { return re.Token.Value }
func (re *RefExpr) String() string       { return "\\" + re.Value.String() }

// DerefExpr represents $$ref, @$ref, %$ref.
// DerefExpr, $$ref, @$ref, %$ref'i temsil eder.
type DerefExpr struct {
	Token lexer.Token
	Sigil string // $, @, %, &, *
	Value Expression
}

func (de *DerefExpr) expressionNode()      {}
func (de *DerefExpr) TokenLiteral() string { return de.Token.Value }
func (de *DerefExpr) String() string       { return de.Sigil + de.Value.String() }

// ============================================================
// Anonymous Sub Expression
// Anonim Sub İfadesi
// ============================================================

// AnonSubExpr represents sub { ... }.
// AnonSubExpr, sub { ... }'u temsil eder.
type AnonSubExpr struct {
	Token  lexer.Token
	Params []*Param
	Body   *BlockStmt
}

func (as *AnonSubExpr) expressionNode()      {}
func (as *AnonSubExpr) TokenLiteral() string { return as.Token.Value }
func (as *AnonSubExpr) String() string {
	return fmt.Sprintf("sub { %s }", as.Body.String())
}

// Param represents a subroutine parameter.
// Param, bir altyordam parametresini temsil eder.
type Param struct {
	Name    string
	Sigil   string // $, @, %
	Default Expression
}

// ============================================================
// Regex Expressions
// Regex İfadeleri
// ============================================================

// MatchExpr represents $str =~ /pattern/.
// MatchExpr, $str =~ /pattern/'ı temsil eder.
type MatchExpr struct {
	Token   lexer.Token
	Target  Expression
	Pattern *RegexLiteral
	Negate  bool // !~
}

func (me *MatchExpr) expressionNode()      {}
func (me *MatchExpr) TokenLiteral() string { return me.Token.Value }
func (me *MatchExpr) String() string {
	op := "=~"
	if me.Negate {
		op = "!~"
	}
	return fmt.Sprintf("(%s %s %s)", me.Target.String(), op, me.Pattern.String())
}

// SubstExpr represents $str =~ s/pattern/replacement/.
// SubstExpr, $str =~ s/pattern/replacement/'i temsil eder.
type SubstExpr struct {
	Token       lexer.Token
	Target      Expression
	Pattern     string
	Replacement string
	Flags       string
}

func (se *SubstExpr) expressionNode()      {}
func (se *SubstExpr) TokenLiteral() string { return se.Token.Value }
func (se *SubstExpr) String() string {
	return fmt.Sprintf("(%s =~ s/%s/%s/%s)",
		se.Target.String(), se.Pattern, se.Replacement, se.Flags)
}

// ============================================================
// Identifier
// Tanımlayıcı
// ============================================================

// Identifier represents a bare identifier (sub name, label, etc.).
// Identifier, bir çıplak tanımlayıcıyı temsil eder (sub adı, etiket, vb.).
type Identifier struct {
	Token lexer.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Value }
func (i *Identifier) String() string       { return i.Value }
