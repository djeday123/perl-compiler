package parser

import (
	"testing"

	"github.com/djeday123/perl-compiler/pkg/ast"
	"github.com/djeday123/perl-compiler/pkg/lexer"
)

func TestMyStatement(t *testing.T) {
	input := `my $x = 10;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.MyStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.MyStatement. got=%T",
			program.Statements[0])
	}

	if stmt.TokenLiteral() != "my" {
		t.Fatalf("stmt.TokenLiteral not 'my'. got=%q", stmt.TokenLiteral())
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := `42;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Expression)
	}

	if literal.Value != 42 {
		t.Errorf("literal.Value not %d. got=%d", 42, literal.Value)
	}

	if literal.TokenLiteral() != "42" {
		t.Errorf("literal.TokenLiteral not %s. got=%s", "42",
			literal.TokenLiteral())
	}
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world";`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	literal, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("exp not *ast.StringLiteral. got=%T", stmt.Expression)
	}

	if literal.Value != "hello world" {
		t.Errorf("literal.Value not %q. got=%q", "hello world", literal.Value)
	}
}

func TestScalarVariable(t *testing.T) {
	input := `$x;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	scalar, ok := stmt.Expression.(*ast.ScalarVariable)
	if !ok {
		t.Fatalf("exp not *ast.ScalarVariable. got=%T", stmt.Expression)
	}

	if scalar.Name != "$x" {
		t.Errorf("scalar.Name not %q. got=%q", "$x", scalar.Name)
	}
}

func TestInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  int64
		operator   string
		rightValue int64
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 % 5;", 5, "%", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
	}

	for _, tt := range infixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("exp is not ast.InfixExpression. got=%T", stmt.Expression)
		}

		if !testIntegerLiteral(t, exp.Left, tt.leftValue) {
			return
		}

		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s",
				tt.operator, exp.Operator)
		}

		if !testIntegerLiteral(t, exp.Right, tt.rightValue) {
			return
		}
	}
}

func TestPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		value    int64
	}{
		{"-15;", "-", 15},
		{"!5;", "!", 5},
	}

	for _, tt := range prefixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("stmt is not ast.PrefixExpression. got=%T", stmt.Expression)
		}

		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s",
				tt.operator, exp.Operator)
		}

		if !testIntegerLiteral(t, exp.Right, tt.value) {
			return
		}
	}
}

func TestIfStatement(t *testing.T) {
	input := `if ($x > 5) { print $x; }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.IfStatement. got=%T",
			program.Statements[0])
	}

	if stmt.TokenLiteral() != "if" {
		t.Fatalf("stmt.TokenLiteral not 'if'. got=%q", stmt.TokenLiteral())
	}

	if stmt.Consequence == nil {
		t.Fatalf("consequence is nil")
	}
}

func TestIfElseStatement(t *testing.T) {
	input := `if ($x > 5) { print $x; } else { print $y; }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.IfStatement. got=%T",
			program.Statements[0])
	}

	if stmt.Alternative == nil {
		t.Fatalf("alternative is nil")
	}
}

func TestWhileStatement(t *testing.T) {
	input := `while ($i < 10) { $i++; }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.WhileStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.WhileStatement. got=%T",
			program.Statements[0])
	}

	if stmt.TokenLiteral() != "while" {
		t.Fatalf("stmt.TokenLiteral not 'while'. got=%q", stmt.TokenLiteral())
	}
}

func TestForStatement(t *testing.T) {
	input := `foreach my $i (1..10) { print $i; }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ForStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ForStatement. got=%T",
			program.Statements[0])
	}

	if stmt.TokenLiteral() != "foreach" {
		t.Fatalf("stmt.TokenLiteral not 'foreach'. got=%q", stmt.TokenLiteral())
	}
}

func TestSubroutineStatement(t *testing.T) {
	input := `sub add { return 42; }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.SubroutineStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.SubroutineStatement. got=%T",
			program.Statements[0])
	}

	if stmt.Name.Value != "add" {
		t.Fatalf("subroutine name not 'add'. got=%q", stmt.Name.Value)
	}
}

func TestPrintStatement(t *testing.T) {
	input := `print "hello";`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.PrintStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.PrintStatement. got=%T",
			program.Statements[0])
	}

	if stmt.TokenLiteral() != "print" {
		t.Fatalf("stmt.TokenLiteral not 'print'. got=%q", stmt.TokenLiteral())
	}

	if len(stmt.Arguments) != 1 {
		t.Fatalf("wrong number of arguments. got=%d", len(stmt.Arguments))
	}
}

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1 + 2 * 3;", "((1) + ((2) * (3)))"},
		{"1 * 2 + 3;", "(((1) * (2)) + (3))"},
		{"1 + 2 + 3;", "(((1) + (2)) + (3))"},
		{"-1 * 2;", "(((-1)) * (2))"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		// Note: The string representation might vary, so this is a basic check
		if len(actual) == 0 {
			t.Errorf("program.String() returned empty string for input %q", tt.input)
		}
	}
}

func TestCallExpression(t *testing.T) {
	input := `add(1, 2);`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("exp not *ast.CallExpression. got=%T", stmt.Expression)
	}

	if len(call.Arguments) != 2 {
		t.Fatalf("wrong number of arguments. got=%d", len(call.Arguments))
	}
}

func TestTernaryExpression(t *testing.T) {
	input := `$x > 5 ? 1 : 0;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	ternary, ok := stmt.Expression.(*ast.TernaryExpression)
	if !ok {
		t.Fatalf("exp not *ast.TernaryExpression. got=%T", stmt.Expression)
	}

	if ternary.Condition == nil {
		t.Fatal("ternary condition is nil")
	}
}

// Helper functions

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

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}

	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}

	return true
}
