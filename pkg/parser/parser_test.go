package parser

import (
	"testing"

	"github.com/djeday123/perl-compiler/pkg/ast"
	"github.com/djeday123/perl-compiler/pkg/lexer"
)

func TestMyStatement(t *testing.T) {
	input := `my $x = 5;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.MyStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.MyStatement. got=%T",
			program.Statements[0])
	}

	if stmt.Name.Value != "x" {
		t.Fatalf("stmt.Name.Value not 'x'. got=%s", stmt.Name.Value)
	}

	if stmt.Name.Sigil != "$" {
		t.Fatalf("stmt.Name.Sigil not '$'. got=%s", stmt.Name.Sigil)
	}
}

func TestSubroutineStatement(t *testing.T) {
	input := `sub hello {
    print "Hello, World!";
}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.SubroutineStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.SubroutineStatement. got=%T",
			program.Statements[0])
	}

	if stmt.Name != "hello" {
		t.Fatalf("stmt.Name not 'hello'. got=%s", stmt.Name)
	}

	if len(stmt.Body.Statements) != 1 {
		t.Fatalf("stmt.Body.Statements does not contain 1 statement. got=%d",
			len(stmt.Body.Statements))
	}
}

func TestIfStatement(t *testing.T) {
	input := `if ($x == 5) {
    print "five";
}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.IfStatement. got=%T",
			program.Statements[0])
	}

	if stmt.Consequence == nil {
		t.Fatal("stmt.Consequence is nil")
	}
}

func TestIfElseStatement(t *testing.T) {
	input := `if ($x == 5) {
    print "five";
} else {
    print "not five";
}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.IfStatement. got=%T",
			program.Statements[0])
	}

	if stmt.Alternative == nil {
		t.Fatal("stmt.Alternative is nil")
	}
}

func TestWhileStatement(t *testing.T) {
	input := `while ($i < 10) {
    print $i;
}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.WhileStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.WhileStatement. got=%T",
			program.Statements[0])
	}

	if stmt.Body == nil {
		t.Fatal("stmt.Body is nil")
	}
}

func TestPrintStatement(t *testing.T) {
	input := `print "Hello";`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.PrintStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.PrintStatement. got=%T",
			program.Statements[0])
	}

	if len(stmt.Arguments) != 1 {
		t.Fatalf("stmt.Arguments does not contain 1 argument. got=%d",
			len(stmt.Arguments))
	}
}

func TestReturnStatement(t *testing.T) {
	input := `return 42;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ReturnStatement. got=%T",
			program.Statements[0])
	}

	if stmt.ReturnValue == nil {
		t.Fatal("stmt.ReturnValue is nil")
	}
}

func TestIntegerExpression(t *testing.T) {
	input := `42;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Expression)
	}

	if literal.Value != 42 {
		t.Fatalf("literal.Value not 42. got=%d", literal.Value)
	}
}

func TestStringExpression(t *testing.T) {
	input := `"hello world";`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("exp not *ast.StringLiteral. got=%T", stmt.Expression)
	}

	if literal.Value != "hello world" {
		t.Fatalf("literal.Value not 'hello world'. got=%s", literal.Value)
	}
}

func TestInfixExpressions(t *testing.T) {
	tests := []struct {
		input      string
		leftValue  int64
		operator   string
		rightValue int64
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statement. got=%d",
				len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("exp not *ast.InfixExpression. got=%T", stmt.Expression)
		}

		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s", tt.operator, exp.Operator)
		}
	}
}

func TestPrefixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
	}{
		{"!5;", "!"},
		{"-15;", "-"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statement. got=%d",
				len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("exp not *ast.PrefixExpression. got=%T", stmt.Expression)
		}

		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s", tt.operator, exp.Operator)
		}
	}
}

func TestVariableExpression(t *testing.T) {
	input := `$foo;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got=%T", stmt.Expression)
	}

	if ident.Value != "foo" {
		t.Fatalf("ident.Value not 'foo'. got=%s", ident.Value)
	}

	if ident.Sigil != "$" {
		t.Fatalf("ident.Sigil not '$'. got=%s", ident.Sigil)
	}
}

func TestAssignmentStatement(t *testing.T) {
	input := `$x = 42;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.AssignmentStatement. got=%T",
			program.Statements[0])
	}

	if stmt.Name.Value != "x" {
		t.Fatalf("stmt.Name.Value not 'x'. got=%s", stmt.Name.Value)
	}

	if stmt.Name.Sigil != "$" {
		t.Fatalf("stmt.Name.Sigil not '$'. got=%s", stmt.Name.Sigil)
	}

	if stmt.Value == nil {
		t.Fatal("stmt.Value is nil")
	}
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
