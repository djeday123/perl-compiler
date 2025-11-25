# perl-compiler

A Perl compiler to machine code, written in Go. This project compiles a subset of Perl source code directly to x86-64 assembly, which is then assembled and linked to produce native executables.

## Features

- **Lexer**: Tokenizes Perl source code
- **Parser**: Builds an Abstract Syntax Tree (AST) from tokens
- **Code Generator**: Generates x86-64 assembly code
- **Native Compilation**: Produces standalone executables

### Supported Perl Constructs

- Variable declarations (`my $var = value;`)
- Scalar variables (`$var`)
- Integer and string literals
- Arithmetic operators (`+`, `-`, `*`, `/`, `%`)
- Comparison operators (`==`, `!=`, `<`, `>`, `<=`, `>=`)
- Logical operators (`&&`, `||`, `!`)
- String concatenation (`.`)
- Print statements (`print`)
- Subroutine definitions (`sub name { ... }`)
- Control flow: `if`/`elsif`/`else`
- Loops: `while`, `for`/`foreach`
- Return statements
- Comments (`# comment`)

## Installation

### Prerequisites

- Go 1.18 or later
- GCC (for linking)
- GNU Assembler (`as`)

### Building

```bash
go build -o perlc ./cmd/perlc
```

Or install directly:

```bash
go install github.com/djeday123/perl-compiler/cmd/perlc@latest
```

## Usage

```bash
# Compile a Perl file to an executable
perlc hello.pl              # Creates executable 'hello'

# Specify output file name
perlc -o myprogram hello.pl

# Emit assembly code only
perlc -S hello.pl           # Creates 'hello.s'

# Show tokens (for debugging)
perlc -tokens hello.pl

# Show AST (for debugging)
perlc -ast hello.pl

# Show version
perlc --version
```

## Example

Create a file `hello.pl`:

```perl
#!/usr/bin/perl

my $message = "Hello, World!";
print $message;

my $x = 10;
my $y = 20;
my $sum = $x + $y;
print $sum;

if ($x < $y) {
    print "x is less than y";
}

sub add {
    return 42;
}
```

Compile and run:

```bash
./perlc hello.pl
./hello
```

Output:
```
Hello, World!
30
x is less than y
```

## Project Structure

```
perl-compiler/
├── cmd/
│   └── perlc/           # Main compiler CLI
│       └── main.go
├── pkg/
│   ├── lexer/           # Lexical analyzer (tokenizer)
│   │   ├── lexer.go
│   │   ├── lexer_test.go
│   │   └── token.go
│   ├── parser/          # Syntax analyzer (parser)
│   │   ├── parser.go
│   │   └── parser_test.go
│   ├── ast/             # Abstract Syntax Tree definitions
│   │   └── ast.go
│   └── codegen/         # Code generator (x86-64 assembly)
│       ├── codegen.go
│       └── codegen_test.go
├── go.mod
├── LICENSE
└── README.md
```

## Testing

Run the test suite:

```bash
go test ./...
```

Run with verbose output:

```bash
go test -v ./...
```

## Architecture

1. **Lexer** (`pkg/lexer`): Reads Perl source code and produces tokens
2. **Parser** (`pkg/parser`): Consumes tokens and builds an AST
3. **Code Generator** (`pkg/codegen`): Traverses the AST and emits x86-64 assembly
4. **Assembler/Linker**: The generated assembly is assembled using `as` and linked with `gcc`

## Limitations

This is a proof-of-concept compiler that supports a subset of Perl. Notable limitations include:

- No regular expressions
- No hash support (partial)
- No array operations beyond literals
- No file I/O
- No modules/packages
- No object-oriented features
- Limited string operations

## License

MIT License - see [LICENSE](LICENSE) for details.
