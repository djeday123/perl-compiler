package codegen

import (
	"strings"
	"testing"

	"github.com/djeday123/perl-compiler/pkg/lexer"
	"github.com/djeday123/perl-compiler/pkg/parser"
)

func TestGenerateIntegerLiteral(t *testing.T) {
	input := `my $x = 42;`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	gen := New()
	output := gen.Generate(program)

	if !strings.Contains(output, "mov rax, 42") {
		t.Errorf("expected 'mov rax, 42' in output, got:\n%s", output)
	}
}

func TestGeneratePrintStatement(t *testing.T) {
	input := `print "hello";`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	gen := New()
	output := gen.Generate(program)

	// Should contain format strings
	if !strings.Contains(output, "section .data") {
		t.Errorf("expected 'section .data' in output")
	}

	// Should contain text section
	if !strings.Contains(output, "section .text") {
		t.Errorf("expected 'section .text' in output")
	}
}

func TestGenerateArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my $x = 5 + 3;", "add rax, rcx"},
		{"my $x = 10 - 5;", "sub rax, rcx"},
		{"my $x = 4 * 3;", "imul rax, rcx"},
		{"my $x = 10 / 2;", "idiv rcx"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		gen := New()
		output := gen.Generate(program)

		if !strings.Contains(output, tt.expected) {
			t.Errorf("for input %q, expected %q in output, got:\n%s",
				tt.input, tt.expected, output)
		}
	}
}

func TestGenerateIfStatement(t *testing.T) {
	input := `my $x = 10; if ($x > 5) { print $x; }`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	gen := New()
	output := gen.Generate(program)

	// Should contain conditional jump
	if !strings.Contains(output, "jz") {
		t.Errorf("expected conditional jump in output")
	}

	// Should contain labels
	if !strings.Contains(output, ".else") {
		t.Errorf("expected else label in output")
	}
}

func TestGenerateWhileLoop(t *testing.T) {
	input := `my $i = 0; while ($i < 10) { $i++; }`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	gen := New()
	output := gen.Generate(program)

	// Should contain loop labels
	if !strings.Contains(output, ".while_start") {
		t.Errorf("expected while_start label in output")
	}

	if !strings.Contains(output, ".while_end") {
		t.Errorf("expected while_end label in output")
	}

	// Should contain jump back to start
	if !strings.Contains(output, "jmp .while_start") {
		t.Errorf("expected jump to while_start in output")
	}
}

func TestGenerateSubroutine(t *testing.T) {
	input := `sub add { return 42; }`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	gen := New()
	output := gen.Generate(program)

	// Should contain subroutine label
	if !strings.Contains(output, "add:") {
		t.Errorf("expected 'add:' label in output")
	}

	// Should contain function prologue/epilogue
	if !strings.Contains(output, "push rbp") {
		t.Errorf("expected function prologue in output")
	}

	if !strings.Contains(output, "ret") {
		t.Errorf("expected 'ret' instruction in output")
	}
}

func TestGenerateComparison(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my $x = 5 == 5;", "sete al"},
		{"my $x = 5 != 5;", "setne al"},
		{"my $x = 5 < 10;", "setl al"},
		{"my $x = 10 > 5;", "setg al"},
		{"my $x = 5 <= 10;", "setle al"},
		{"my $x = 10 >= 5;", "setge al"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		gen := New()
		output := gen.Generate(program)

		if !strings.Contains(output, tt.expected) {
			t.Errorf("for input %q, expected %q in output, got:\n%s",
				tt.input, tt.expected, output)
		}
	}
}

func TestGenerateBitwiseOperations(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my $x = 5 & 3;", "and rax, rcx"},
		{"my $x = 5 | 3;", "or rax, rcx"},
		{"my $x = 5 ^ 3;", "xor rax, rcx"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		gen := New()
		output := gen.Generate(program)

		if !strings.Contains(output, tt.expected) {
			t.Errorf("for input %q, expected %q in output, got:\n%s",
				tt.input, tt.expected, output)
		}
	}
}

func TestGeneratePrefixOperations(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my $x = -5;", "neg rax"},
		{"my $x = !1;", "setz al"},
		{"my $x = ~5;", "not rax"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		gen := New()
		output := gen.Generate(program)

		if !strings.Contains(output, tt.expected) {
			t.Errorf("for input %q, expected %q in output, got:\n%s",
				tt.input, tt.expected, output)
		}
	}
}

func TestGenerateStringLiteral(t *testing.T) {
	input := `my $msg = "Hello, World!";`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	gen := New()
	output := gen.Generate(program)

	// Should contain string in data section
	if !strings.Contains(output, "Hello, World!") {
		t.Errorf("expected string literal in output")
	}

	// Should contain lea instruction for string address
	if !strings.Contains(output, "lea rax") {
		t.Errorf("expected 'lea rax' for string address in output")
	}
}

func TestEscapeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"line1\nline2", `line1", 10, "line2`},
		{`quote"here`, `quote", 34, "here`},
	}

	for _, tt := range tests {
		result := escapeString(tt.input)
		if result != tt.expected {
			t.Errorf("escapeString(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGeneratorErrors(t *testing.T) {
	gen := New()

	if len(gen.Errors()) != 0 {
		t.Errorf("new generator should have no errors")
	}
}

func TestCompleteProgram(t *testing.T) {
	input := `
my $a = 10;
my $b = 20;
my $sum = $a + $b;
print $sum;
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	gen := New()
	output := gen.Generate(program)

	// Should have all required sections
	sections := []string{
		"section .data",
		"section .text",
		"global _start",
		"_start:",
	}

	for _, section := range sections {
		if !strings.Contains(output, section) {
			t.Errorf("expected %q in output", section)
		}
	}

	// Should have proper exit
	if !strings.Contains(output, "mov rax, 60") {
		t.Errorf("expected exit syscall in output")
	}
}
