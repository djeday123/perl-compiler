# Perl Compiler Tests

## Overview

This directory contains integration tests for the Perl compiler. Tests verify both:
1. **Interpreter mode** (`./perlc script.pl`)
2. **Compiled mode** (`./perlc -r script.pl`)

## Test Structure

```
tests/
├── integration_test.go    # Go test framework (comprehensive)
├── fileio_test.go         # Detailed File I/O tests
├── run_tests.sh          # Run Go tests
├── quick_test.sh         # Run quick Perl tests
└── quick/                # Quick Perl test scripts
    ├── fileio_basic.pl
    ├── fileio_append.pl
    ├── regex_basic.pl
    ├── data_structures.pl
    ├── subroutines.pl
    ├── control_flow.pl
    └── references.pl
```

## Running Tests

### Quick Tests (Perl scripts)

Fast verification using Perl test scripts:

```bash
# Run all quick tests (both modes)
./tests/quick_test.sh

# Run only interpreter tests
./tests/quick_test.sh interp

# Run only compilation tests  
./tests/quick_test.sh compile
```

### Go Integration Tests

Comprehensive tests using Go test framework:

```bash
# Run all tests
cd tests && go test -v

# Run specific test group
cd tests && go test -v -run TestFileIO

# Run with coverage
cd tests && go test -cover
```

Available test groups:
- `TestBasicPrint` - print/say
- `TestVariables` - scalar variables
- `TestArithmetic` - math operations
- `TestStringOperations` - string functions
- `TestArrays` - array operations
- `TestHashes` - hash operations
- `TestControlFlow` - if/while/for/foreach
- `TestComparisons` - numeric/string comparison
- `TestSubroutines` - sub definitions, @_, return
- `TestReferences` - refs, derefs, anonymous
- `TestRegex` - =~, s///, capture groups
- `TestFileIO` - open, close, <>, print to file
- `TestBuiltinFunctions` - abs, sqrt, chr, etc.
- `TestEdgeCases` - truthiness, autovivification
- `TestIntegration` - complex programs (FizzBuzz, etc.)

### Individual Perl Tests

Run a single test file:

```bash
# Interpreter
./perlc tests/quick/fileio_basic.pl

# Compiled
./perlc -r tests/quick/fileio_basic.pl
```

## Test Output Format

Quick tests use this format:
- `PASS: description` - test passed
- `FAIL: description` - test failed

The test runner checks for "FAIL" in output to determine success.

## Adding New Tests

### Quick Perl Test

Create a `.pl` file in `tests/quick/`:

```perl
# tests/quick/my_feature.pl

say "Testing my feature...";

# Test 1
if (some_condition) {
    say "PASS: Test 1";
} else {
    say "FAIL: Test 1";
}

say "Done!";
```

### Go Integration Test

Add to `integration_test.go`:

```go
func TestMyFeature(t *testing.T) {
    tests := []TestCase{
        {
            Name:           "test description",
            Code:           `perl code here`,
            ExpectedOutput: "expected output",
        },
    }
    
    for _, tc := range tests {
        t.Run(tc.Name, func(t *testing.T) {
            runTest(t, tc)
        })
    }
}
```

## Test Case Structure

```go
type TestCase struct {
    Name           string            // Test name
    Code           string            // Perl code to run
    ExpectedOutput string            // Exact expected output
    ExpectedMatch  string            // Regex pattern (alternative)
    SetupFiles     map[string]string // Files to create before test
    CleanupFiles   []string          // Files to delete after test
    SkipCompile    bool              // Skip compilation test
    SkipInterpret  bool              // Skip interpreter test
}
```

## Current Test Coverage

| Feature | Interpreter | Compiler |
|---------|-------------|----------|
| print/say | ✅ | ✅ |
| Variables | ✅ | ✅ |
| Arithmetic | ✅ | ✅ |
| Strings | ✅ | ✅ |
| Arrays | ✅ | ✅ |
| Hashes | ✅ | ✅ |
| if/elsif/else | ✅ | ✅ |
| while/until | ✅ | ✅ |
| for/foreach | ✅ | ✅ |
| Subroutines | ✅ | ✅ |
| References | ✅ | ✅ |
| Regex match | ✅ | ✅ |
| Regex subst | ✅ | ✅ |
| File I/O | ✅ | ✅ |

## Known Limitations

1. `while (my $line = <$fh>)` - assignment in condition not yet supported
2. Some complex nested structures may have issues in compiled mode
3. Closures have limited support

## CI Integration

Example GitHub Actions workflow:

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go build -o perlc ./cmd/perlc
      - run: cd tests && go test -v
```
