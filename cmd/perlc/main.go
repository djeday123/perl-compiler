// perlc is the Perl compiler that compiles Perl source code to machine code.
// It implements a lexer, parser, and code generator targeting x86-64 assembly.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/djeday123/perl-compiler/pkg/codegen"
	"github.com/djeday123/perl-compiler/pkg/lexer"
	"github.com/djeday123/perl-compiler/pkg/parser"
)

const version = "0.1.0"

func main() {
	// Command line flags
	outputFile := flag.String("o", "", "Output file name (default: input file with .asm extension)")
	showAST := flag.Bool("ast", false, "Print the AST and exit")
	showTokens := flag.Bool("tokens", false, "Print tokens and exit")
	showVersion := flag.Bool("version", false, "Show version information")
	showHelp := flag.Bool("help", false, "Show help information")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "perlc - Perl compiler to machine code\n\n")
		fmt.Fprintf(os.Stderr, "Usage: perlc [options] <source.pl>\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  perlc hello.pl           Compile hello.pl to hello.asm\n")
		fmt.Fprintf(os.Stderr, "  perlc -o out.asm test.pl Compile test.pl to out.asm\n")
		fmt.Fprintf(os.Stderr, "  perlc -ast hello.pl      Show the AST of hello.pl\n")
		fmt.Fprintf(os.Stderr, "  perlc -tokens hello.pl   Show the tokens of hello.pl\n")
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("perlc version %s\n", version)
		os.Exit(0)
	}

	if *showHelp || flag.NArg() < 1 {
		flag.Usage()
		os.Exit(0)
	}

	inputFile := flag.Arg(0)

	// Read input file
	source, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Create lexer
	l := lexer.New(string(source))

	// Token mode - just print tokens
	if *showTokens {
		printTokens(string(source))
		os.Exit(0)
	}

	// Create parser
	p := parser.New(l)

	// Parse the program
	program := p.ParseProgram()

	// Check for parse errors
	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "Parse errors:\n")
		for _, msg := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", msg)
		}
		os.Exit(1)
	}

	// AST mode - print AST and exit
	if *showAST {
		fmt.Println("Abstract Syntax Tree:")
		fmt.Println(program.String())
		os.Exit(0)
	}

	// Generate code
	gen := codegen.New()
	assembly := gen.Generate(program)

	// Check for code generation errors
	if len(gen.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "Code generation errors:\n")
		for _, msg := range gen.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", msg)
		}
		os.Exit(1)
	}

	// Determine output file name
	outFile := *outputFile
	if outFile == "" {
		ext := filepath.Ext(inputFile)
		outFile = strings.TrimSuffix(inputFile, ext) + ".asm"
	}

	// Write output
	err = os.WriteFile(outFile, []byte(assembly), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Compiled %s -> %s\n", inputFile, outFile)
	fmt.Println("\nTo assemble and link (Linux x86-64):")
	fmt.Printf("  nasm -f elf64 %s -o %s.o\n", outFile, strings.TrimSuffix(outFile, ".asm"))
	fmt.Printf("  ld %s.o -o %s -lc --dynamic-linker /lib64/ld-linux-x86-64.so.2\n",
		strings.TrimSuffix(outFile, ".asm"), strings.TrimSuffix(outFile, ".asm"))
}

func printTokens(source string) {
	l := lexer.New(source)
	for {
		tok := l.NextToken()
		fmt.Printf("%+v\n", tok)
		if tok.Type == "EOF" {
			break
		}
	}
}
