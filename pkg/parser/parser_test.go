package parser

// Package parser tests

import (
	"testing"

	"perlc/pkg/ast"
	"perlc/pkg/lexer"
)

// ============================================================
// Helper Functions
// Yardımcı Fonksiyonlar
// ============================================================

func parseProgram(t *testing.T, input string) *ast.Program {
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	return program
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}
	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

// ============================================================
// Literal Tests
// Literal Testleri
// ============================================================

func TestIntegerLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"42;", 42},
		{"0;", 0},
		{"-1;", -1}, // This is actually prefix minus + 1
		{"0x1F;", 31},
		{"0b1010;", 10},
		{"0777;", 511},
	}

	for _, tt := range tests {
		program := parseProgram(t, tt.input)
		if len(program.Statements) != 1 {
			t.Fatalf("expected 1 statement, got %d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExprStmt)
		if !ok {
			t.Fatalf("not ExprStmt, got %T", program.Statements[0])
		}

		testIntegerLiteral(t, stmt.Expression, tt.expected)
	}
}

func testIntegerLiteral(t *testing.T, exp ast.Expression, expected int64) {
	// Handle negative numbers (prefix expression)
	if prefix, ok := exp.(*ast.PrefixExpr); ok && prefix.Operator == "-" {
		lit, ok := prefix.Right.(*ast.IntegerLiteral)
		if !ok {
			t.Fatalf("not IntegerLiteral, got %T", prefix.Right)
		}
		if lit.Value != -expected {
			t.Errorf("value not %d, got %d", -expected, lit.Value)
		}
		return
	}

	lit, ok := exp.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("not IntegerLiteral, got %T", exp)
	}
	if lit.Value != expected {
		t.Errorf("value not %d, got %d", expected, lit.Value)
	}
}

func TestStringLiteral(t *testing.T) {
	input := `"hello world";`
	program := parseProgram(t, input)

	stmt := program.Statements[0].(*ast.ExprStmt)
	lit, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("not StringLiteral, got %T", stmt.Expression)
	}
	if lit.Value != "hello world" {
		t.Errorf("value not %q, got %q", "hello world", lit.Value)
	}
}

// ============================================================
// Variable Tests
// Değişken Testleri
// ============================================================

func TestScalarVar(t *testing.T) {
	input := `$foo;`
	program := parseProgram(t, input)

	stmt := program.Statements[0].(*ast.ExprStmt)
	v, ok := stmt.Expression.(*ast.ScalarVar)
	if !ok {
		t.Fatalf("not ScalarVar, got %T", stmt.Expression)
	}
	if v.Name != "foo" {
		t.Errorf("name not foo, got %s", v.Name)
	}
}

func TestArrayVar(t *testing.T) {
	input := `@arr;`
	program := parseProgram(t, input)

	stmt := program.Statements[0].(*ast.ExprStmt)
	v, ok := stmt.Expression.(*ast.ArrayVar)
	if !ok {
		t.Fatalf("not ArrayVar, got %T", stmt.Expression)
	}
	if v.Name != "arr" {
		t.Errorf("name not arr, got %s", v.Name)
	}
}

func TestHashVar(t *testing.T) {
	input := `%hash;`
	program := parseProgram(t, input)

	stmt := program.Statements[0].(*ast.ExprStmt)
	v, ok := stmt.Expression.(*ast.HashVar)
	if !ok {
		t.Fatalf("not HashVar, got %T", stmt.Expression)
	}
	if v.Name != "hash" {
		t.Errorf("name not hash, got %s", v.Name)
	}
}

// ============================================================
// Operator Tests
// Operatör Testleri
// ============================================================

func TestInfixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		left     int64
		operator string
		right    int64
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 % 5;", 5, "%", 5},
		{"5 ** 2;", 5, "**", 2},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 <= 5;", 5, "<=", 5},
		{"5 >= 5;", 5, ">=", 5},
		{"5 <=> 5;", 5, "<=>", 5},
	}

	for _, tt := range tests {
		program := parseProgram(t, tt.input)
		stmt := program.Statements[0].(*ast.ExprStmt)
		exp, ok := stmt.Expression.(*ast.InfixExpr)
		if !ok {
			t.Fatalf("not InfixExpr for %q, got %T", tt.input, stmt.Expression)
		}

		testIntegerLiteral(t, exp.Left, tt.left)
		if exp.Operator != tt.operator {
			t.Errorf("operator not %s, got %s", tt.operator, exp.Operator)
		}
		testIntegerLiteral(t, exp.Right, tt.right)
	}
}

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1 + 2 * 3;", "(1 + (2 * 3))"},
		{"1 * 2 + 3;", "((1 * 2) + 3)"},
		{"2 ** 3 ** 2;", "(2 ** (3 ** 2))"}, // Right associative
		{"1 + 2 + 3;", "((1 + 2) + 3)"},     // Left associative
		{"1 < 2 == 3 > 4;", "((1 < 2) == (3 > 4))"},
		{"1 && 2 || 3;", "((1 && 2) || 3)"},
	}

	for _, tt := range tests {
		program := parseProgram(t, tt.input)
		stmt := program.Statements[0].(*ast.ExprStmt)
		actual := stmt.Expression.String()
		if actual != tt.expected {
			t.Errorf("for %q: expected %s, got %s", tt.input, tt.expected, actual)
		}
	}
}

func TestTernaryExpression(t *testing.T) {
	input := `$x ? 1 : 2;`
	program := parseProgram(t, input)

	stmt := program.Statements[0].(*ast.ExprStmt)
	exp, ok := stmt.Expression.(*ast.TernaryExpr)
	if !ok {
		t.Fatalf("not TernaryExpr, got %T", stmt.Expression)
	}

	if _, ok := exp.Condition.(*ast.ScalarVar); !ok {
		t.Errorf("condition not ScalarVar")
	}
	testIntegerLiteral(t, exp.Then, 1)
	testIntegerLiteral(t, exp.Else, 2)
}

// ============================================================
// Declaration Tests
// Bildirim Testleri
// ============================================================

func TestMyDecl(t *testing.T) {
	input := `my $x = 42;`
	program := parseProgram(t, input)

	decl, ok := program.Statements[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("not VarDecl, got %T", program.Statements[0])
	}
	if decl.Kind != "my" {
		t.Errorf("kind not my, got %s", decl.Kind)
	}
	if len(decl.Names) != 1 {
		t.Fatalf("expected 1 name, got %d", len(decl.Names))
	}
}

func TestMyListDecl(t *testing.T) {
	input := `my ($x, $y) = (1, 2);`
	program := parseProgram(t, input)

	decl, ok := program.Statements[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("not VarDecl, got %T", program.Statements[0])
	}
	if len(decl.Names) != 2 {
		t.Errorf("expected 2 names, got %d", len(decl.Names))
	}
}

func TestSubDecl(t *testing.T) {
	input := `sub foo { return 42; }`
	program := parseProgram(t, input)

	decl, ok := program.Statements[0].(*ast.SubDecl)
	if !ok {
		t.Fatalf("not SubDecl, got %T", program.Statements[0])
	}
	if decl.Name != "foo" {
		t.Errorf("name not foo, got %s", decl.Name)
	}
	if decl.Body == nil {
		t.Error("body is nil")
	}
}

func TestPackageDecl(t *testing.T) {
	input := `package Foo::Bar;`
	program := parseProgram(t, input)

	decl, ok := program.Statements[0].(*ast.PackageDecl)
	if !ok {
		t.Fatalf("not PackageDecl, got %T", program.Statements[0])
	}
	if decl.Name != "Foo::Bar" {
		t.Errorf("name not Foo::Bar, got %s", decl.Name)
	}
}

func TestUseDecl(t *testing.T) {
	input := `use strict;`
	program := parseProgram(t, input)

	decl, ok := program.Statements[0].(*ast.UseDecl)
	if !ok {
		t.Fatalf("not UseDecl, got %T", program.Statements[0])
	}
	if decl.Module != "strict" {
		t.Errorf("module not strict, got %s", decl.Module)
	}
}

// ============================================================
// Control Flow Tests
// Kontrol Akışı Testleri
// ============================================================

func TestIfStmt(t *testing.T) {
	input := `if ($x) { 1; }`
	program := parseProgram(t, input)

	stmt, ok := program.Statements[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("not IfStmt, got %T", program.Statements[0])
	}
	if stmt.Unless {
		t.Error("should not be unless")
	}
	if stmt.Condition == nil {
		t.Error("condition is nil")
	}
	if stmt.Then == nil {
		t.Error("then block is nil")
	}
}

func TestIfElseStmt(t *testing.T) {
	input := `if ($x) { 1; } else { 2; }`
	program := parseProgram(t, input)

	stmt := program.Statements[0].(*ast.IfStmt)
	if stmt.Else == nil {
		t.Error("else block is nil")
	}
}

func TestIfElsifElseStmt(t *testing.T) {
	input := `if ($x) { 1; } elsif ($y) { 2; } else { 3; }`
	program := parseProgram(t, input)

	stmt := program.Statements[0].(*ast.IfStmt)
	if len(stmt.Elsif) != 1 {
		t.Errorf("expected 1 elsif, got %d", len(stmt.Elsif))
	}
	if stmt.Else == nil {
		t.Error("else block is nil")
	}
}

func TestWhileStmt(t *testing.T) {
	input := `while ($x) { $x--; }`
	program := parseProgram(t, input)

	stmt, ok := program.Statements[0].(*ast.WhileStmt)
	if !ok {
		t.Fatalf("not WhileStmt, got %T", program.Statements[0])
	}
	if stmt.Until {
		t.Error("should not be until")
	}
}

func TestForeachStmt(t *testing.T) {
	input := `foreach my $x (@arr) { print $x; }`
	program := parseProgram(t, input)

	stmt, ok := program.Statements[0].(*ast.ForeachStmt)
	if !ok {
		t.Fatalf("not ForeachStmt, got %T", program.Statements[0])
	}
	if stmt.Variable == nil {
		t.Error("variable is nil")
	}
	if stmt.List == nil {
		t.Error("list is nil")
	}
}

func TestReturnStmt(t *testing.T) {
	input := `return 42;`
	program := parseProgram(t, input)

	stmt, ok := program.Statements[0].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("not ReturnStmt, got %T", program.Statements[0])
	}
	testIntegerLiteral(t, stmt.Value, 42)
}

// ============================================================
// Complex Expression Tests
// Karmaşık İfade Testleri
// ============================================================

func TestArrayLiteral(t *testing.T) {
	input := `[1, 2, 3];`
	program := parseProgram(t, input)

	stmt := program.Statements[0].(*ast.ExprStmt)
	arr, ok := stmt.Expression.(*ast.ArrayExpr)
	if !ok {
		t.Fatalf("not ArrayExpr, got %T", stmt.Expression)
	}
	if len(arr.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr.Elements))
	}
}

func TestHashLiteral(t *testing.T) {
	input := `my $h = {a => 1, b => 2};`
	program := parseProgram(t, input)

	decl, ok := program.Statements[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("not VarDecl, got %T", program.Statements[0])
	}

	hash, ok := decl.Value.(*ast.HashExpr)
	if !ok {
		t.Fatalf("value not HashExpr, got %T", decl.Value)
	}
	if len(hash.Pairs) != 2 {
		t.Errorf("expected 2 pairs, got %d", len(hash.Pairs))
	}
}

func TestArrayAccess(t *testing.T) {
	input := `$arr[0];`
	program := parseProgram(t, input)

	stmt := program.Statements[0].(*ast.ExprStmt)
	acc, ok := stmt.Expression.(*ast.ArrayAccess)
	if !ok {
		t.Fatalf("not ArrayAccess, got %T", stmt.Expression)
	}
	if acc.Array == nil {
		t.Error("array is nil")
	}
	testIntegerLiteral(t, acc.Index, 0)
}

func TestHashAccess(t *testing.T) {
	input := `$hash{key};`
	program := parseProgram(t, input)

	stmt := program.Statements[0].(*ast.ExprStmt)
	acc, ok := stmt.Expression.(*ast.HashAccess)
	if !ok {
		t.Fatalf("not HashAccess, got %T", stmt.Expression)
	}
	if acc.Hash == nil {
		t.Error("hash is nil")
	}
}

func TestMethodCall(t *testing.T) {
	input := `$obj->method(1, 2);`
	program := parseProgram(t, input)

	stmt := program.Statements[0].(*ast.ExprStmt)
	call, ok := stmt.Expression.(*ast.MethodCall)
	if !ok {
		t.Fatalf("not MethodCall, got %T", stmt.Expression)
	}
	if call.Method != "method" {
		t.Errorf("method not method, got %s", call.Method)
	}
	if len(call.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(call.Args))
	}
}

func TestCallExpr(t *testing.T) {
	input := `foo(1, 2, 3);`
	program := parseProgram(t, input)

	stmt := program.Statements[0].(*ast.ExprStmt)
	call, ok := stmt.Expression.(*ast.CallExpr)
	if !ok {
		t.Fatalf("not CallExpr, got %T", stmt.Expression)
	}
	if len(call.Args) != 3 {
		t.Errorf("expected 3 args, got %d", len(call.Args))
	}
}

func TestRefExpr(t *testing.T) {
	input := `\@arr;`
	program := parseProgram(t, input)

	stmt := program.Statements[0].(*ast.ExprStmt)
	ref, ok := stmt.Expression.(*ast.RefExpr)
	if !ok {
		t.Fatalf("not RefExpr, got %T", stmt.Expression)
	}
	if _, ok := ref.Value.(*ast.ArrayVar); !ok {
		t.Errorf("value not ArrayVar, got %T", ref.Value)
	}
}

func TestAnonSub(t *testing.T) {
	input := `my $f = sub { return 1; };`
	program := parseProgram(t, input)

	decl := program.Statements[0].(*ast.VarDecl)
	anon, ok := decl.Value.(*ast.AnonSubExpr)
	if !ok {
		t.Fatalf("not AnonSubExpr, got %T", decl.Value)
	}
	if anon.Body == nil {
		t.Error("body is nil")
	}
}

// ============================================================
// Real Perl Code Test
// Gerçek Perl Kodu Testi
// ============================================================

func TestRealPerlCode(t *testing.T) {
	input := `
use strict;
use warnings;

package Calculator;

sub new {
    my ($class, $value) = @_;
    return bless { value => $value }, $class;
}

sub add {
    my ($self, $n) = @_;
    $self->{value} += $n;
    return $self;
}

sub value {
    my $self = shift;
    return $self->{value};
}

1;
`
	program := parseProgram(t, input)

	if len(program.Statements) < 5 {
		t.Errorf("expected at least 5 statements, got %d", len(program.Statements))
	}
}

// temporaly at parser_test.go
func TestDebugTokens(t *testing.T) {
	inputs := []string{
		"a => 1",
		"my ($x, $y)",
		"foreach my $x (@arr) { }",
		"bless {}",
	}

	for _, input := range inputs {
		t.Logf("=== %s ===", input)
		l := lexer.New(input)
		for {
			tok := l.NextToken()
			t.Logf("  %d: %q", tok.Type, tok.Value)
			if tok.Type == lexer.TokEOF {
				break
			}
		}
	}
}
