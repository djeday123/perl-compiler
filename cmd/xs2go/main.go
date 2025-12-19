// cmd/xs2go/main.go
package main

import (
	"fmt"
	"os"
	"perlc/pkg/xs2go"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: xs2go <file.xs>")
		os.Exit(1)
	}

	translator := xs2go.New()

	code, err := translator.Translate(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(code)
}
