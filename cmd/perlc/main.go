package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"perlc/pkg/codegen"
	"perlc/pkg/eval"
	"perlc/pkg/lexer"
	"perlc/pkg/parser"
)

func main() {
	compile := flag.Bool("c", false, "Compile to Go code")
	output := flag.String("o", "", "Output file name")
	run := flag.Bool("r", false, "Compile and run")
	flag.Parse()

	if flag.NArg() < 1 {
		repl()
		return
	}

	filename := flag.Arg(0)
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	input := string(data)

	if *compile || *run {
		compileToGo(input, filename, *output, *run)
	} else {
		interpret(input)
	}
}

func interpret(input string) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			fmt.Fprintf(os.Stderr, "Parse error: %s\n", e)
		}
		os.Exit(1)
	}

	interp := eval.New()
	interp.Eval(program)
}

func compileToGo(input, filename, outputName string, runAfter bool) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			fmt.Fprintf(os.Stderr, "Parse error: %s\n", e)
		}
		os.Exit(1)
	}

	gen := codegen.New()
	goCode := gen.Generate(program)

	fmt.Println("=== Generated Go Code ===")
	fmt.Println(goCode)
	fmt.Println("=== End Generated Code ===")

	// Determine output filename
	if outputName == "" {
		base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
		outputName = base
	}

	// Create temp directory for compilation
	tmpDir, err := os.MkdirTemp("", "perlc-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	goFile := filepath.Join(tmpDir, "main.go")

	// Write Go file
	err = os.WriteFile(goFile, []byte(goCode), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing Go file: %v\n", err)
		os.Exit(1)
	}

	// Compile with go build
	exeName := outputName
	if os.PathSeparator == '\\' {
		exeName += ".exe"
	}

	// Get absolute path for output
	absExe, _ := filepath.Abs(exeName)

	cmd := exec.Command("go", "build", "-o", absExe, goFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error compiling: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Compiled: %s\n", exeName)

	// Run if requested
	if runAfter {
		fmt.Println("---")
		cmd = exec.Command(absExe)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
}
func repl() {
	fmt.Println("perlc REPL (type 'exit' to quit)")
	interp := eval.New()

	for {
		fmt.Print("perl> ")
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		if input == "exit" || input == "quit" {
			break
		}

		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			for _, e := range p.Errors() {
				fmt.Printf("Error: %s\n", e)
			}
			continue
		}

		interp.Eval(program)
	}
}

func Run(input string) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			fmt.Fprintf(os.Stderr, "Parse error: %s\n", e)
		}
		os.Exit(1)
	}

	interp := eval.New()
	interp.Eval(program)
}
