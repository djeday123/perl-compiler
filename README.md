# perl-compiler

Perl compiler to machine code written in Go.

## Overview

This is a Perl compiler that translates Perl source code into x86-64 assembly code, which can then be assembled and linked into native machine code executables.

## Features

- **Lexer**: Tokenizes Perl source code supporting:
  - Variables: scalars (`$var`), arrays (`@arr`), hashes (`%hash`)
  - Literals: integers, floats, strings
  - Operators: arithmetic, comparison, logical, bitwise
  - Keywords: `my`, `sub`, `if`, `elsif`, `else`, `while`, `for`, `foreach`, `return`, `print`, etc.

- **Parser**: Builds an Abstract Syntax Tree (AST) supporting:
  - Variable declarations and assignments
  - Arithmetic and logical expressions
  - Control flow: if/elsif/else, while, for/foreach
  - Subroutine definitions
  - Function calls

- **Code Generator**: Produces x86-64 NASM assembly targeting:
  - System V AMD64 ABI calling convention
  - Linux syscalls for exit
  - Printf for output

## Installation

### Prerequisites

- Go 1.21 or later
- NASM (Netwide Assembler) - for assembling the generated code
- GCC or LD - for linking

### Building

```bash
# Clone the repository
git clone https://github.com/djeday123/perl-compiler.git
cd perl-compiler

# Build the compiler
go build -o perlc ./cmd/perlc

# Run tests
go test ./...
```

## Usage

### Basic Compilation

```bash
# Compile a Perl file to assembly
./perlc hello.pl

# This generates hello.asm
```

### Command Line Options

```
Usage: perlc [options] <source.pl>

Options:
  -o string       Output file name (default: input file with .asm extension)
  -ast            Print the AST and exit
  -tokens         Print tokens and exit
  -version        Show version information
  -help           Show help information
```

### Examples

```bash
# Compile with custom output name
./perlc -o output.asm source.pl

# View the AST
./perlc -ast hello.pl

# View tokens
./perlc -tokens hello.pl
```

### Assembling and Linking (Linux x86-64)

```bash
# Compile Perl to assembly
./perlc hello.pl

# Assemble with NASM
nasm -f elf64 hello.asm -o hello.o

# Link with libc
ld hello.o -o hello -lc --dynamic-linker /lib64/ld-linux-x86-64.so.2

# Run
./hello
```

## Project Structure

```
perl-compiler/
├── cmd/
│   └── perlc/          # Main compiler entry point
│       └── main.go
├── pkg/
│   ├── token/          # Token definitions
│   │   ├── token.go
│   │   └── token_test.go
│   ├── lexer/          # Lexical analyzer
│   │   ├── lexer.go
│   │   └── lexer_test.go
│   ├── ast/            # Abstract Syntax Tree
│   │   └── ast.go
│   ├── parser/         # Parser
│   │   ├── parser.go
│   │   └── parser_test.go
│   └── codegen/        # Code generator
│       ├── codegen.go
│       └── codegen_test.go
├── internal/
│   └── errors/         # Error handling
│       └── errors.go
├── examples/           # Example Perl programs
│   ├── hello.pl
│   ├── arithmetic.pl
│   ├── control_flow.pl
│   └── subroutines.pl
├── go.mod
├── LICENSE
└── README.md
```

## Supported Perl Features

### Variables
- Scalar variables: `my $x = 10;`
- Array variables: `my @arr;` (basic support)
- Hash variables: `my %hash;` (basic support)

### Operators
- Arithmetic: `+`, `-`, `*`, `/`, `%`, `**`
- Comparison: `==`, `!=`, `<`, `>`, `<=`, `>=`, `<=>`
- String comparison: `eq`, `ne`, `lt`, `gt`, `le`, `ge`, `cmp`
- Logical: `&&`, `||`, `!`, `and`, `or`, `not`
- Bitwise: `&`, `|`, `^`, `~`, `<<`, `>>`
- Assignment: `=`, `+=`, `-=`, `*=`, `/=`, `%=`, `.=`
- Increment/Decrement: `++`, `--`
- String: `.` (concatenation), `x` (repetition)
- Range: `..`
- Ternary: `?:`

### Control Flow
- `if` / `elsif` / `else`
- `unless`
- `while`
- `until`
- `for` / `foreach`

### Subroutines
- `sub name { ... }`
- `return value;`

### I/O
- `print`

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./pkg/lexer
go test ./pkg/parser
go test ./pkg/codegen
```

## Limitations

This is an educational project and has several limitations:

- Limited subset of Perl supported
- No object-oriented features
- No regular expressions
- No modules/packages support
- No file I/O
- x86-64 Linux only

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
