// perlc is a Perl compiler that compiles Perl source code to machine code.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/djeday123/perl-compiler/pkg/codegen"
	"github.com/djeday123/perl-compiler/pkg/lexer"
	"github.com/djeday123/perl-compiler/pkg/parser"
)

var (
	version = "0.1.0"
)

func main() {
	// Parse command line flags
	outputFile := flag.String("o", "", "Output file name")
	emitAsm := flag.Bool("S", false, "Emit assembly code instead of executable")
	showVersion := flag.Bool("version", false, "Show version information")
	showTokens := flag.Bool("tokens", false, "Show lexer tokens (for debugging)")
	showAST := flag.Bool("ast", false, "Show AST (for debugging)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <source.pl>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "A Perl compiler that compiles Perl source code to machine code.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s hello.pl              # Compile to executable 'hello'\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -o myprogram hello.pl # Compile to executable 'myprogram'\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -S hello.pl           # Emit assembly to 'hello.s'\n", os.Args[0])
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("perlc version %s\n", version)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	sourceFile := flag.Arg(0)

	// Read source file
	source, err := os.ReadFile(sourceFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Tokenize
	l := lexer.New(string(source))

	if *showTokens {
		fmt.Println("=== Tokens ===")
		for {
			tok := l.NextToken()
			fmt.Printf("%s: %q (line %d, col %d)\n", tok.Type.String(), tok.Literal, tok.Line, tok.Column)
			if tok.Type == lexer.EOF {
				break
			}
		}
		// Re-create lexer for parsing
		l = lexer.New(string(source))
	}

	// Parse
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "Parser errors:\n")
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", err)
		}
		os.Exit(1)
	}

	if *showAST {
		fmt.Println("=== AST ===")
		fmt.Println(program.String())
	}

	// Generate code
	cg := codegen.New()
	asm := cg.Generate(program)

	// Determine output file name
	baseName := strings.TrimSuffix(filepath.Base(sourceFile), filepath.Ext(sourceFile))
	asmFile := baseName + ".s"

	if *outputFile == "" {
		*outputFile = baseName
	}

	// Write assembly file
	if err := os.WriteFile(asmFile, []byte(asm), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing assembly file: %v\n", err)
		os.Exit(1)
	}

	if *emitAsm {
		fmt.Printf("Assembly written to %s\n", asmFile)
		os.Exit(0)
	}

	// Compile assembly to object file
	objFile := baseName + ".o"
	if err := runCommand("as", "-o", objFile, asmFile); err != nil {
		fmt.Fprintf(os.Stderr, "Assembly error: %v\n", err)
		os.Exit(1)
	}

	// Link to executable
	if err := runCommand("gcc", "-no-pie", "-o", *outputFile, objFile); err != nil {
		fmt.Fprintf(os.Stderr, "Linking error: %v\n", err)
		os.Exit(1)
	}

	// Clean up intermediate files
	os.Remove(asmFile)
	os.Remove(objFile)

	fmt.Printf("Compiled to %s\n", *outputFile)
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
