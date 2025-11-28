package ast

import (
	"fmt"
	"strings"

	"perlc/pkg/lexer"
)

// ============================================================
// Variable Declarations
// Değişken Bildirimleri
// ============================================================

// VarDecl represents my/our/local/state declaration.
// VarDecl, my/our/local/state bildirimini temsil eder.
type VarDecl struct {
	Token  lexer.Token
	Kind   string       // "my", "our", "local", "state"
	Names  []Expression // Variables being declared
	Value  Expression   // Optional initializer
	IsList bool         // true if declared with parentheses: my ($x) vs my $x
}

func (vd *VarDecl) statementNode()       {}
func (vd *VarDecl) declarationNode()     {}
func (vd *VarDecl) TokenLiteral() string { return vd.Token.Value }
func (vd *VarDecl) String() string {
	names := make([]string, len(vd.Names))
	for i, n := range vd.Names {
		names[i] = n.String()
	}

	var out string
	if len(names) == 1 {
		out = fmt.Sprintf("%s %s", vd.Kind, names[0])
	} else {
		out = fmt.Sprintf("%s (%s)", vd.Kind, strings.Join(names, ", "))
	}

	if vd.Value != nil {
		out += " = " + vd.Value.String()
	}
	return out + ";"
}

// ============================================================
// Subroutine Declaration
// Altyordam Bildirimi
// ============================================================

// SubDecl represents sub name { }.
// SubDecl, sub name { }'ı temsil eder.
type SubDecl struct {
	Token      lexer.Token
	Name       string
	Prototype  string // Optional prototype
	Attributes []string
	Params     []*Param // With signatures
	Body       *BlockStmt
}

func (sd *SubDecl) statementNode()       {}
func (sd *SubDecl) declarationNode()     {}
func (sd *SubDecl) TokenLiteral() string { return sd.Token.Value }
func (sd *SubDecl) String() string {
	var out strings.Builder
	out.WriteString("sub ")
	out.WriteString(sd.Name)
	if sd.Prototype != "" {
		out.WriteString(fmt.Sprintf("(%s)", sd.Prototype))
	}
	for _, attr := range sd.Attributes {
		out.WriteString(" :" + attr)
	}
	out.WriteString(" ")
	out.WriteString(sd.Body.String())
	return out.String()
}

// ============================================================
// Package Declaration
// Paket Bildirimi
// ============================================================

// PackageDecl represents package Name;
// PackageDecl, package Name;'i temsil eder.
type PackageDecl struct {
	Token   lexer.Token
	Name    string
	Version string     // Optional version
	Block   *BlockStmt // Optional block form: package Foo { }
}

func (pd *PackageDecl) statementNode()       {}
func (pd *PackageDecl) declarationNode()     {}
func (pd *PackageDecl) TokenLiteral() string { return pd.Token.Value }
func (pd *PackageDecl) String() string {
	out := "package " + pd.Name
	if pd.Version != "" {
		out += " " + pd.Version
	}
	if pd.Block != nil {
		out += " " + pd.Block.String()
	} else {
		out += ";"
	}
	return out
}

// ============================================================
// Use/No/Require Declarations
// Use/No/Require Bildirimleri
// ============================================================

// UseDecl represents use Module [VERSION] [LIST];
// UseDecl, use Module [VERSION] [LIST];'ı temsil eder.
type UseDecl struct {
	Token   lexer.Token
	Module  string
	Version string
	Args    []Expression // Import list
}

func (ud *UseDecl) statementNode()       {}
func (ud *UseDecl) declarationNode()     {}
func (ud *UseDecl) TokenLiteral() string { return ud.Token.Value }
func (ud *UseDecl) String() string {
	out := "use " + ud.Module
	if ud.Version != "" {
		out += " " + ud.Version
	}
	if len(ud.Args) > 0 {
		args := make([]string, len(ud.Args))
		for i, a := range ud.Args {
			args[i] = a.String()
		}
		out += " " + strings.Join(args, ", ")
	}
	return out + ";"
}

// NoDecl represents no Module;
// NoDecl, no Module;'u temsil eder.
type NoDecl struct {
	Token  lexer.Token
	Module string
	Args   []Expression
}

func (nd *NoDecl) statementNode()       {}
func (nd *NoDecl) declarationNode()     {}
func (nd *NoDecl) TokenLiteral() string { return nd.Token.Value }
func (nd *NoDecl) String() string {
	out := "no " + nd.Module
	if len(nd.Args) > 0 {
		args := make([]string, len(nd.Args))
		for i, a := range nd.Args {
			args[i] = a.String()
		}
		out += " " + strings.Join(args, ", ")
	}
	return out + ";"
}

// RequireDecl represents require Module or require "file".
// RequireDecl, require Module veya require "file"'ı temsil eder.
type RequireDecl struct {
	Token  lexer.Token
	Module string     // Module name
	Expr   Expression // Or expression (require $var)
}

func (rd *RequireDecl) statementNode()       {}
func (rd *RequireDecl) declarationNode()     {}
func (rd *RequireDecl) TokenLiteral() string { return rd.Token.Value }
func (rd *RequireDecl) String() string {
	if rd.Module != "" {
		return "require " + rd.Module + ";"
	}
	return "require " + rd.Expr.String() + ";"
}

// ============================================================
// Special Blocks
// Özel Bloklar
// ============================================================

// SpecialBlock represents BEGIN/END/CHECK/INIT/UNITCHECK blocks.
// SpecialBlock, BEGIN/END/CHECK/INIT/UNITCHECK bloklarını temsil eder.
type SpecialBlock struct {
	Token lexer.Token
	Kind  string // "BEGIN", "END", "CHECK", "INIT", "UNITCHECK"
	Body  *BlockStmt
}

func (sb *SpecialBlock) statementNode()       {}
func (sb *SpecialBlock) declarationNode()     {}
func (sb *SpecialBlock) TokenLiteral() string { return sb.Token.Value }
func (sb *SpecialBlock) String() string {
	return sb.Kind + " " + sb.Body.String()
}
