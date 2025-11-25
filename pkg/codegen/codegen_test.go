package codegen

import (
	"strings"
	"testing"

	"github.com/djeday123/perl-compiler/pkg/lexer"
	"github.com/djeday123/perl-compiler/pkg/parser"
)

func TestGenerateSimpleProgram(t *testing.T) {
	input := `my $x = 42;
print $x;`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			t.Errorf("parser error: %s", err)
		}
		t.FailNow()
	}

	cg := New()
	asm := cg.Generate(program)

	// Check that the output contains expected assembly sections
	if !strings.Contains(asm, ".section .data") {
		t.Error("output should contain .data section")
	}

	if !strings.Contains(asm, ".section .text") {
		t.Error("output should contain .text section")
	}

	if !strings.Contains(asm, ".globl main") {
		t.Error("output should contain main entry point")
	}

	if !strings.Contains(asm, "main:") {
		t.Error("output should contain main label")
	}
}

func TestGenerateWithSubroutine(t *testing.T) {
	input := `sub add {
    return 42;
}`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			t.Errorf("parser error: %s", err)
		}
		t.FailNow()
	}

	cg := New()
	asm := cg.Generate(program)

	// Check that the subroutine is generated
	if !strings.Contains(asm, "add:") {
		t.Error("output should contain 'add' subroutine label")
	}
}

func TestGenerateArithmetic(t *testing.T) {
	input := `my $x = 5 + 3;`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			t.Errorf("parser error: %s", err)
		}
		t.FailNow()
	}

	cg := New()
	asm := cg.Generate(program)

	// Check that addition instruction is present
	if !strings.Contains(asm, "addq") {
		t.Error("output should contain 'addq' instruction for addition")
	}
}

func TestGenerateIfStatement(t *testing.T) {
	input := `if ($x == 5) {
    print "five";
}`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			t.Errorf("parser error: %s", err)
		}
		t.FailNow()
	}

	cg := New()
	asm := cg.Generate(program)

	// Check that comparison and jump instructions are present
	if !strings.Contains(asm, "cmpq") {
		t.Error("output should contain 'cmpq' instruction for comparison")
	}

	if !strings.Contains(asm, "je ") {
		t.Error("output should contain 'je' instruction for conditional jump")
	}
}

func TestGenerateWhileLoop(t *testing.T) {
	input := `while ($i < 10) {
    my $x = $i;
}`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			t.Errorf("parser error: %s", err)
		}
		t.FailNow()
	}

	cg := New()
	asm := cg.Generate(program)

	// Check that loop structure is present
	if !strings.Contains(asm, "jmp ") {
		t.Error("output should contain 'jmp' instruction for loop")
	}
}

func TestGenerateString(t *testing.T) {
	input := `print "Hello, World!";`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			t.Errorf("parser error: %s", err)
		}
		t.FailNow()
	}

	cg := New()
	asm := cg.Generate(program)

	// Check that string constant is in data section
	if !strings.Contains(asm, "Hello, World!") {
		t.Error("output should contain string constant")
	}

	// Check that printf is called
	if !strings.Contains(asm, "call printf") || !strings.Contains(asm, "call puts") {
		t.Error("output should contain call to printf or puts")
	}
}
