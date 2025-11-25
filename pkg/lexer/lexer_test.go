package lexer

import (
	"testing"

	"github.com/djeday123/perl-compiler/pkg/token"
)

func TestNextToken(t *testing.T) {
	input := `my $x = 10;
my $y = 20;
my $sum = $x + $y;
print $sum;`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.MY, "my"},
		{token.SCALAR, "$x"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.MY, "my"},
		{token.SCALAR, "$y"},
		{token.ASSIGN, "="},
		{token.INT, "20"},
		{token.SEMICOLON, ";"},
		{token.MY, "my"},
		{token.SCALAR, "$sum"},
		{token.ASSIGN, "="},
		{token.SCALAR, "$x"},
		{token.PLUS, "+"},
		{token.SCALAR, "$y"},
		{token.SEMICOLON, ";"},
		{token.PRINT, "print"},
		{token.SCALAR, "$sum"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestOperators(t *testing.T) {
	input := `+ - * / % ** == != < > <= >= <=> && || ! & | ^ ~ << >>`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.PLUS, "+"},
		{token.MINUS, "-"},
		{token.ASTERISK, "*"},
		{token.SLASH, "/"},
		{token.MODULO, "%"},
		{token.POWER, "**"},
		{token.EQ, "=="},
		{token.NE, "!="},
		{token.LT, "<"},
		{token.GT, ">"},
		{token.LE, "<="},
		{token.GE, ">="},
		{token.NUMEQ, "<=>"},
		{token.AND, "&&"},
		{token.OR, "||"},
		{token.NOT, "!"},
		{token.BITAND, "&"},
		{token.BITOR, "|"},
		{token.BITXOR, "^"},
		{token.BITNOT, "~"},
		{token.LSHIFT, "<<"},
		{token.RSHIFT, ">>"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestStrings(t *testing.T) {
	input := `"hello" 'world' "with \"quotes\""`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.STRING, "hello"},
		{token.STRING, "world"},
		{token.STRING, `with \"quotes\"`},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNumbers(t *testing.T) {
	input := `123 456.789 1e10 3.14159`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.INT, "123"},
		{token.FLOAT, "456.789"},
		{token.FLOAT, "1e10"},
		{token.FLOAT, "3.14159"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestVariables(t *testing.T) {
	input := `$scalar @array %hash`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.SCALAR, "$scalar"},
		{token.ARRAY, "@array"},
		{token.HASH, "%hash"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestComments(t *testing.T) {
	input := `# this is a comment
my $x = 10; # inline comment`

	l := New(input)

	// First token should be a comment
	tok := l.NextToken()
	if tok.Type != token.COMMENT {
		t.Fatalf("expected COMMENT, got %s", tok.Type)
	}

	// Then we should get my
	tok = l.NextToken()
	if tok.Type != token.MY {
		t.Fatalf("expected MY, got %s", tok.Type)
	}
}

func TestDelimiters(t *testing.T) {
	input := `( ) { } [ ] ; , : -> =>`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.LBRACKET, "["},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},
		{token.COMMA, ","},
		{token.COLON, ":"},
		{token.ARROW, "->"},
		{token.FATCOMMA, "=>"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestIncrementDecrement(t *testing.T) {
	input := `++ -- += -= *= /=`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.INC, "++"},
		{token.DEC, "--"},
		{token.PLUSEQ, "+="},
		{token.MINUSEQ, "-="},
		{token.MULEQ, "*="},
		{token.DIVEQ, "/="},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestRange(t *testing.T) {
	input := `1..10 ...`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.INT, "1"},
		{token.RANGE, ".."},
		{token.INT, "10"},
		{token.ELLIPSIS, "..."},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLineAndColumn(t *testing.T) {
	input := `my $x = 10;
my $y = 20;`

	l := New(input)

	// First token should be at line 1
	tok := l.NextToken()
	if tok.Line != 1 {
		t.Fatalf("expected line 1, got %d", tok.Line)
	}

	// Skip to second line
	for tok.Type != token.SEMICOLON {
		tok = l.NextToken()
	}

	// Next token should be at line 2
	tok = l.NextToken()
	if tok.Line != 2 {
		t.Fatalf("expected line 2, got %d", tok.Line)
	}
}
