package eval

import (
	"bytes"
	"testing"

	"perlc/pkg/lexer"
	"perlc/pkg/parser"
)

func evalInput(input string) (string, *Interpreter) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	interp := New()
	var buf bytes.Buffer
	interp.SetStdout(&buf)

	interp.Eval(program)
	return buf.String(), interp
}

// ============================================================
// Basic Tests
// ============================================================

func TestPrint(t *testing.T) {
	output, _ := evalInput(`print "Hello, World!";`)
	if output != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got %q", output)
	}
}

func TestSay(t *testing.T) {
	output, _ := evalInput(`say "Hello";`)
	if output != "Hello\n" {
		t.Errorf("expected 'Hello\\n', got %q", output)
	}
}

func TestArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`say 2 + 3;`, "5\n"},
		{`say 10 - 4;`, "6\n"},
		{`say 3 * 4;`, "12\n"},
		{`say 15 / 3;`, "5\n"},
		{`say 17 % 5;`, "2\n"},
		{`say 2 ** 10;`, "1024\n"},
	}

	for _, tt := range tests {
		output, _ := evalInput(tt.input)
		if output != tt.expected {
			t.Errorf("for %q: expected %q, got %q", tt.input, tt.expected, output)
		}
	}
}

func TestStringOps(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`say "Hello" . " " . "World";`, "Hello World\n"},
		{`say "ab" x 3;`, "ababab\n"},
		{`say length("hello");`, "5\n"},
		{`say uc("hello");`, "HELLO\n"},
		{`say lc("HELLO");`, "hello\n"},
	}

	for _, tt := range tests {
		output, _ := evalInput(tt.input)
		if output != tt.expected {
			t.Errorf("for %q: expected %q, got %q", tt.input, tt.expected, output)
		}
	}
}

// ============================================================
// Variable Tests
// ============================================================

func TestScalarVar(t *testing.T) {
	output, _ := evalInput(`
		my $x = 42;
		say $x;
	`)
	if output != "42\n" {
		t.Errorf("expected '42\\n', got %q", output)
	}
}

func TestVarAssignment(t *testing.T) {
	output, _ := evalInput(`
		my $x = 10;
		$x = $x + 5;
		say $x;
	`)
	if output != "15\n" {
		t.Errorf("expected '15\\n', got %q", output)
	}
}

func TestCompoundAssignment(t *testing.T) {
	output, _ := evalInput(`
		my $x = 10;
		$x += 5;
		say $x;
	`)
	if output != "15\n" {
		t.Errorf("expected '15\\n', got %q", output)
	}
}

func TestStringConcat(t *testing.T) {
	output, _ := evalInput(`
		my $s = "Hello";
		$s .= " World";
		say $s;
	`)
	if output != "Hello World\n" {
		t.Errorf("expected 'Hello World\\n', got %q", output)
	}
}

// ============================================================
// Control Flow Tests
// ============================================================

func TestIfTrue(t *testing.T) {
	output, _ := evalInput(`
		my $x = 1;
		if ($x) {
			say "yes";
		}
	`)
	if output != "yes\n" {
		t.Errorf("expected 'yes\\n', got %q", output)
	}
}

func TestIfFalse(t *testing.T) {
	output, _ := evalInput(`
		my $x = 0;
		if ($x) {
			say "yes";
		} else {
			say "no";
		}
	`)
	if output != "no\n" {
		t.Errorf("expected 'no\\n', got %q", output)
	}
}

func TestIfElsif(t *testing.T) {
	output, _ := evalInput(`
		my $x = 2;
		if ($x == 1) {
			say "one";
		} elsif ($x == 2) {
			say "two";
		} else {
			say "other";
		}
	`)
	if output != "two\n" {
		t.Errorf("expected 'two\\n', got %q", output)
	}
}

func TestUnless(t *testing.T) {
	output, _ := evalInput(`
		my $x = 0;
		unless ($x) {
			say "false";
		}
	`)
	if output != "false\n" {
		t.Errorf("expected 'false\\n', got %q", output)
	}
}

func TestWhile(t *testing.T) {
	output, _ := evalInput(`
		my $i = 0;
		while ($i < 3) {
			say $i;
			$i++;
		}
	`)
	if output != "0\n1\n2\n" {
		t.Errorf("expected '0\\n1\\n2\\n', got %q", output)
	}
}

func TestFor(t *testing.T) {
	output, _ := evalInput(`
		for (my $i = 0; $i < 3; $i++) {
			say $i;
		}
	`)
	if output != "0\n1\n2\n" {
		t.Errorf("expected '0\\n1\\n2\\n', got %q", output)
	}
}

func TestForeach(t *testing.T) {
	output, _ := evalInput(`
		my @arr = (1, 2, 3);
		foreach my $x (@arr) {
			say $x;
		}
	`)
	if output != "1\n2\n3\n" {
		t.Errorf("expected '1\\n2\\n3\\n', got %q", output)
	}
}

func TestLast(t *testing.T) {
	output, _ := evalInput(`
		my $i = 0;
		while (1) {
			say $i;
			$i++;
			if ($i >= 3) {
				last;
			}
		}
	`)
	if output != "0\n1\n2\n" {
		t.Errorf("expected '0\\n1\\n2\\n', got %q", output)
	}
}

func TestNext(t *testing.T) {
	output, _ := evalInput(`
		for (my $i = 0; $i < 5; $i++) {
			if ($i == 2) {
				next;
			}
			say $i;
		}
	`)
	if output != "0\n1\n3\n4\n" {
		t.Errorf("expected '0\\n1\\n3\\n4\\n', got %q", output)
	}
}

// ============================================================
// Subroutine Tests
// ============================================================

func TestSubroutine(t *testing.T) {
	output, _ := evalInput(`
		sub greet {
			say "Hello";
		}
		greet();
	`)
	if output != "Hello\n" {
		t.Errorf("expected 'Hello\\n', got %q", output)
	}
}

func TestSubroutineArgs(t *testing.T) {
	output, _ := evalInput(`
		sub add {
			my $a = shift;
			my $b = shift;
			return $a + $b;
		}
		say add(3, 4);
	`)
	if output != "7\n" {
		t.Errorf("expected '7\\n', got %q", output)
	}
}

func TestSubroutineReturn(t *testing.T) {
	output, _ := evalInput(`
		sub double {
			my $x = shift;
			return $x * 2;
		}
		say double(21);
	`)
	if output != "42\n" {
		t.Errorf("expected '42\\n', got %q", output)
	}
}

// ============================================================
// Array Tests
// ============================================================

func TestArrayLiteral(t *testing.T) {
	output, _ := evalInput(`
		my @arr = (1, 2, 3);
		say $arr[0];
		say $arr[1];
		say $arr[2];
	`)
	if output != "1\n2\n3\n" {
		t.Errorf("expected '1\\n2\\n3\\n', got %q", output)
	}
}

func TestArrayPush(t *testing.T) {
	output, _ := evalInput(`
		my @arr = (1, 2);
		push @arr, 3;
		say $arr[2];
	`)
	if output != "3\n" {
		t.Errorf("expected '3\\n', got %q", output)
	}
}

func TestArrayPop(t *testing.T) {
	output, _ := evalInput(`
		my @arr = (1, 2, 3);
		my $x = pop @arr;
		say $x;
	`)
	if output != "3\n" {
		t.Errorf("expected '3\\n', got %q", output)
	}
}

// ============================================================
// Hash Tests
// ============================================================

func TestHashLiteral(t *testing.T) {
	output, _ := evalInput(`
		my $h = {a => 1, b => 2};
		say $h->{a};
		say $h->{b};
	`)
	if output != "1\n2\n" {
		t.Errorf("expected '1\\n2\\n', got %q", output)
	}
}

// ============================================================
// Ternary Operator Test
// ============================================================

func TestTernary(t *testing.T) {
	output, _ := evalInput(`
		my $x = 1;
		say $x ? "yes" : "no";
	`)
	if output != "yes\n" {
		t.Errorf("expected 'yes\\n', got %q", output)
	}
}

// ============================================================
// Logical Operators Test
// ============================================================

func TestLogicalOr(t *testing.T) {
	output, _ := evalInput(`
		my $x = 0 || 42;
		say $x;
	`)
	if output != "42\n" {
		t.Errorf("expected '42\\n', got %q", output)
	}
}

func TestDefinedOr(t *testing.T) {
	output, _ := evalInput(`
		my $x;
		my $y = $x // "default";
		say $y;
	`)
	if output != "default\n" {
		t.Errorf("expected 'default\\n', got %q", output)
	}
}

func TestForDebug(t *testing.T) {
	input := `for (my $i = 0; $i < 3; $i++) { say $i; }`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			t.Logf("parse error: %s", e)
		}
	}

	t.Logf("statements: %d", len(program.Statements))
	for i, stmt := range program.Statements {
		t.Logf("stmt[%d]: %T = %s", i, stmt, stmt.String())
	}
}
