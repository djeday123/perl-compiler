package lexer

import (
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `my $x = 5;
my $y = 10;
my $result = $x + $y;
print $result;
`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{MY, "my"},
		{SCALAR, "$"},
		{IDENT, "x"},
		{ASSIGN, "="},
		{INT, "5"},
		{SEMICOLON, ";"},
		{MY, "my"},
		{SCALAR, "$"},
		{IDENT, "y"},
		{ASSIGN, "="},
		{INT, "10"},
		{SEMICOLON, ";"},
		{MY, "my"},
		{SCALAR, "$"},
		{IDENT, "result"},
		{ASSIGN, "="},
		{SCALAR, "$"},
		{IDENT, "x"},
		{PLUS, "+"},
		{SCALAR, "$"},
		{IDENT, "y"},
		{SEMICOLON, ";"},
		{PRINT, "print"},
		{SCALAR, "$"},
		{IDENT, "result"},
		{SEMICOLON, ";"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType.String(), tok.Type.String())
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestOperators(t *testing.T) {
	input := `== != < > <= >= && || ++ -- -> => =~ !~`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{EQ, "=="},
		{NOT_EQ, "!="},
		{LT, "<"},
		{GT, ">"},
		{LT_EQ, "<="},
		{GT_EQ, ">="},
		{AND, "&&"},
		{OR, "||"},
		{PLUSPLUS, "++"},
		{MINUSMINUS, "--"},
		{ARROW, "->"},
		{FATARROW, "=>"},
		{MATCH, "=~"},
		{NOTMATCH, "!~"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType.String(), tok.Type.String())
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestStrings(t *testing.T) {
	input := `"hello world" 'single quotes'`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{STRING, "hello world"},
		{STRING, "single quotes"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType.String(), tok.Type.String())
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNumbers(t *testing.T) {
	input := `42 3.14 0 100`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{INT, "42"},
		{FLOAT, "3.14"},
		{INT, "0"},
		{INT, "100"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType.String(), tok.Type.String())
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestComments(t *testing.T) {
	input := `# this is a comment
my $x = 5; # inline comment`

	l := New(input)

	// First token should be the comment
	tok := l.NextToken()
	if tok.Type != COMMENT {
		t.Fatalf("expected COMMENT, got %s", tok.Type.String())
	}

	// Then my
	tok = l.NextToken()
	if tok.Type != MY {
		t.Fatalf("expected MY, got %s", tok.Type.String())
	}
}

func TestKeywords(t *testing.T) {
	input := `sub my our if elsif else while for foreach return print`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{SUB, "sub"},
		{MY, "my"},
		{OUR, "our"},
		{IF, "if"},
		{ELSIF, "elsif"},
		{ELSE, "else"},
		{WHILE, "while"},
		{FOR, "for"},
		{FOREACH, "foreach"},
		{RETURN, "return"},
		{PRINT, "print"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType.String(), tok.Type.String())
		}
	}
}

func TestSubroutine(t *testing.T) {
	input := `sub add {
    my $a = shift;
    my $b = shift;
    return $a + $b;
}`

	l := New(input)

	// Just verify we can tokenize the whole thing without errors
	for {
		tok := l.NextToken()
		if tok.Type == EOF {
			break
		}
		if tok.Type == ILLEGAL {
			t.Fatalf("unexpected ILLEGAL token: %q", tok.Literal)
		}
	}
}
