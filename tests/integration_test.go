package tests

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestCase represents a single integration test
type TestCase struct {
	Name           string
	Code           string
	ExpectedOutput string
	ExpectedMatch  string // Regex pattern for flexible matching
	SetupFiles     map[string]string // Files to create before test
	CleanupFiles   []string          // Files to remove after test
	SkipCompile    bool              // Skip compilation test
	SkipInterpret  bool              // Skip interpretation test
}

// runInterpreter runs perlc in interpreter mode
func runInterpreter(t *testing.T, code string) (string, error) {
	// Create temp file with test code
	tmpFile, err := os.CreateTemp("", "test_*.pl")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(code); err != nil {
		return "", err
	}
	tmpFile.Close()

	// Determine exe name
	exeName := "./perlc"
	if os.PathSeparator == '\\' {
		exeName = "./perlc.exe"
	}

	// Run perlc
	cmd := exec.Command(exeName, tmpFile.Name())
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output += stderr.String()
	}

	return output, err
}

// runCompiled runs perlc with -r flag (compile and run)
func runCompiled(t *testing.T, code string) (string, error) {
	// Create temp file with test code
	tmpFile, err := os.CreateTemp("", "test_*.pl")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(code); err != nil {
		return "", err
	}
	tmpFile.Close()

	// Determine exe name
	exeName := "./perlc"
	if os.PathSeparator == '\\' {
		exeName = "./perlc.exe"
	}

	// Run perlc -r
	cmd := exec.Command(exeName, "-r", tmpFile.Name())
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	output := stdout.String()

	// Remove "Compiled: xxx.exe\n---\n" prefix
	if idx := strings.Index(output, "---\n"); idx != -1 {
		output = output[idx+4:]
	}

	// Cleanup exe
	base := strings.TrimSuffix(filepath.Base(tmpFile.Name()), ".pl")
	os.Remove(base + ".exe")
	os.Remove(base)

	return output, err
}

// checkOutput compares actual output with expected
func checkOutput(t *testing.T, name, mode, actual, expected, pattern string) {
	actual = strings.TrimSpace(actual)
	expected = strings.TrimSpace(expected)

	if pattern != "" {
		re := regexp.MustCompile(pattern)
		if !re.MatchString(actual) {
			t.Errorf("[%s] %s:\nExpected pattern: %s\nActual:\n%s", mode, name, pattern, actual)
		}
		return
	}

	if actual != expected {
		t.Errorf("[%s] %s:\nExpected:\n%s\n\nActual:\n%s", mode, name, expected, actual)
	}
}

// runTest executes a single test case
func runTest(t *testing.T, tc TestCase) {
	// Setup files
	for filename, content := range tc.SetupFiles {
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create setup file %s: %v", filename, err)
		}
	}

	// Cleanup
	defer func() {
		for _, f := range tc.CleanupFiles {
			os.Remove(f)
		}
		for f := range tc.SetupFiles {
			os.Remove(f)
		}
	}()

	// Test interpreter
	if !tc.SkipInterpret {
		output, err := runInterpreter(t, tc.Code)
		if err != nil && tc.ExpectedOutput != "" {
			t.Errorf("[INTERP] %s: interpreter error: %v\nOutput: %s", tc.Name, err, output)
		} else {
			checkOutput(t, tc.Name, "INTERP", output, tc.ExpectedOutput, tc.ExpectedMatch)
		}
	}

	// Test compilation
	if !tc.SkipCompile {
		output, err := runCompiled(t, tc.Code)
		if err != nil && tc.ExpectedOutput != "" {
			t.Errorf("[COMPILE] %s: compilation error: %v\nOutput: %s", tc.Name, err, output)
		} else {
			checkOutput(t, tc.Name, "COMPILE", output, tc.ExpectedOutput, tc.ExpectedMatch)
		}
	}
}

// ============================================================
// Basic Tests
// ============================================================

func TestBasicPrint(t *testing.T) {
	tests := []TestCase{
		{
			Name:           "print string",
			Code:           `print "Hello, World!\n";`,
			ExpectedOutput: "Hello, World!",
		},
		{
			Name:           "say string",
			Code:           `say "Hello";`,
			ExpectedOutput: "Hello",
		},
		{
			Name:           "multiple print",
			Code:           `print "One\n"; print "Two\n"; print "Three\n";`,
			ExpectedOutput: "One\nTwo\nThree",
		},
		{
			Name:           "print with variable",
			Code:           `my $x = "World"; print "Hello, $x!\n";`,
			ExpectedOutput: "Hello, World!",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

func TestVariables(t *testing.T) {
	tests := []TestCase{
		{
			Name:           "scalar assignment",
			Code:           `my $x = 42; say $x;`,
			ExpectedOutput: "42",
		},
		{
			Name:           "scalar reassignment",
			Code:           `my $x = 10; $x = 20; say $x;`,
			ExpectedOutput: "20",
		},
		{
			Name:           "multiple scalars",
			Code:           `my $a = 1; my $b = 2; my $c = 3; say "$a $b $c";`,
			ExpectedOutput: "1 2 3",
		},
		{
			Name:           "string variable",
			Code:           `my $name = "Alice"; say "Hello, $name";`,
			ExpectedOutput: "Hello, Alice",
		},
		{
			Name:           "variable interpolation",
			Code:           `my $x = 5; my $y = 10; say "x=$x, y=$y, sum=" . ($x + $y);`,
			ExpectedOutput: "x=5, y=10, sum=15",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

// ============================================================
// Arithmetic Tests
// ============================================================

func TestArithmetic(t *testing.T) {
	tests := []TestCase{
		{
			Name:           "addition",
			Code:           `say 2 + 3;`,
			ExpectedOutput: "5",
		},
		{
			Name:           "subtraction",
			Code:           `say 10 - 4;`,
			ExpectedOutput: "6",
		},
		{
			Name:           "multiplication",
			Code:           `say 6 * 7;`,
			ExpectedOutput: "42",
		},
		{
			Name:           "division",
			Code:           `say 15 / 3;`,
			ExpectedOutput: "5",
		},
		{
			Name:           "modulo",
			Code:           `say 17 % 5;`,
			ExpectedOutput: "2",
		},
		{
			Name:           "power",
			Code:           `say 2 ** 10;`,
			ExpectedOutput: "1024",
		},
		{
			Name:           "complex expression",
			Code:           `say (2 + 3) * 4 - 5;`,
			ExpectedOutput: "15",
		},
		{
			Name:           "negative numbers",
			Code:           `say -5 + 10;`,
			ExpectedOutput: "5",
		},
		{
			Name:           "float arithmetic",
			Code:           `say 3.14 * 2;`,
			ExpectedMatch:  `6\.28`,
		},
		{
			Name:           "increment",
			Code:           `my $x = 5; $x++; say $x;`,
			ExpectedOutput: "6",
		},
		{
			Name:           "decrement",
			Code:           `my $x = 5; $x--; say $x;`,
			ExpectedOutput: "4",
		},
		{
			Name:           "compound assignment +=",
			Code:           `my $x = 10; $x += 5; say $x;`,
			ExpectedOutput: "15",
		},
		{
			Name:           "compound assignment *=",
			Code:           `my $x = 3; $x *= 4; say $x;`,
			ExpectedOutput: "12",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

// ============================================================
// String Operations Tests
// ============================================================

func TestStringOperations(t *testing.T) {
	tests := []TestCase{
		{
			Name:           "concatenation",
			Code:           `say "Hello" . " " . "World";`,
			ExpectedOutput: "Hello World",
		},
		{
			Name:           "string repetition",
			Code:           `say "ab" x 3;`,
			ExpectedOutput: "ababab",
		},
		{
			Name:           "length",
			Code:           `say length("hello");`,
			ExpectedOutput: "5",
		},
		{
			Name:           "uc",
			Code:           `say uc("hello");`,
			ExpectedOutput: "HELLO",
		},
		{
			Name:           "lc",
			Code:           `say lc("HELLO");`,
			ExpectedOutput: "hello",
		},
		{
			Name:           "ucfirst",
			Code:           `say ucfirst("hello world");`,
			ExpectedOutput: "Hello world",
		},
		{
			Name:           "lcfirst",
			Code:           `say lcfirst("HELLO");`,
			ExpectedOutput: "hELLO",
		},
		{
			Name:           "reverse string",
			Code:           `say scalar(reverse("hello"));`,
			ExpectedOutput: "olleh",
		},
		{
			Name:           "substr get",
			Code:           `say substr("Hello World", 0, 5);`,
			ExpectedOutput: "Hello",
		},
		{
			Name:           "substr from end",
			Code:           `say substr("Hello World", -5);`,
			ExpectedOutput: "World",
		},
		{
			Name:           "index",
			Code:           `say index("Hello World", "World");`,
			ExpectedOutput: "6",
		},
		{
			Name:           "rindex",
			Code:           `say rindex("Hello Hello", "Hello");`,
			ExpectedOutput: "6",
		},
		{
			Name:           "chomp",
			Code:           `my $s = "hello\n"; chomp($s); say "[$s]";`,
			ExpectedOutput: "[hello]",
		},
		{
			Name:           "chop",
			Code:           `my $s = "hello"; chop($s); say $s;`,
			ExpectedOutput: "hell",
		},
		{
			Name:           "sprintf",
			Code:           `say sprintf("Name: %s, Age: %d", "Alice", 30);`,
			ExpectedOutput: "Name: Alice, Age: 30",
		},
		{
			Name:           "sprintf padding",
			Code:           `say sprintf("%05d", 42);`,
			ExpectedOutput: "00042",
		},
		{
			Name:           "sprintf float",
			Code:           `say sprintf("%.2f", 3.14159);`,
			ExpectedOutput: "3.14",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

// ============================================================
// Array Tests
// ============================================================

func TestArrays(t *testing.T) {
	tests := []TestCase{
		{
			Name:           "array creation and access",
			Code:           `my @arr = (1, 2, 3); say $arr[0];`,
			ExpectedOutput: "1",
		},
		{
			Name:           "array length",
			Code:           `my @arr = (1, 2, 3, 4, 5); say scalar(@arr);`,
			ExpectedOutput: "5",
		},
		{
			Name:           "array push",
			Code:           `my @arr = (1, 2); push(@arr, 3); say "@arr";`,
			ExpectedOutput: "1 2 3",
		},
		{
			Name:           "array pop",
			Code:           `my @arr = (1, 2, 3); my $x = pop(@arr); say "$x: @arr";`,
			ExpectedOutput: "3: 1 2",
		},
		{
			Name:           "array shift",
			Code:           `my @arr = (1, 2, 3); my $x = shift(@arr); say "$x: @arr";`,
			ExpectedOutput: "1: 2 3",
		},
		{
			Name:           "array unshift",
			Code:           `my @arr = (2, 3); unshift(@arr, 1); say "@arr";`,
			ExpectedOutput: "1 2 3",
		},
		{
			Name:           "array join",
			Code:           `my @arr = ("a", "b", "c"); say join("-", @arr);`,
			ExpectedOutput: "a-b-c",
		},
		{
			Name:           "array split",
			Code:           `my @arr = split(",", "a,b,c"); say "@arr";`,
			ExpectedOutput: "a b c",
		},
		{
			Name:           "array sort numeric",
			Code:           `my @arr = (3, 1, 4, 1, 5); my @sorted = sort { $a <=> $b } @arr; say "@sorted";`,
			ExpectedOutput: "1 1 3 4 5",
		},
		{
			Name:           "array sort string",
			Code:           `my @arr = ("banana", "apple", "cherry"); my @sorted = sort @arr; say "@sorted";`,
			ExpectedOutput: "apple banana cherry",
		},
		{
			Name:           "array reverse",
			Code:           `my @arr = (1, 2, 3); my @rev = reverse(@arr); say "@rev";`,
			ExpectedOutput: "3 2 1",
		},
		{
			Name:           "array slice",
			Code:           `my @arr = (10, 20, 30, 40, 50); my @slice = @arr[1, 3]; say "@slice";`,
			ExpectedOutput: "20 40",
		},
		{
			Name:           "array range",
			Code:           `my @arr = (1..5); say "@arr";`,
			ExpectedOutput: "1 2 3 4 5",
		},
		{
			Name:           "negative index",
			Code:           `my @arr = (1, 2, 3); say $arr[-1];`,
			ExpectedOutput: "3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

// ============================================================
// Hash Tests
// ============================================================

func TestHashes(t *testing.T) {
	tests := []TestCase{
		{
			Name:           "hash creation and access",
			Code:           `my %h = (a => 1, b => 2); say $h{a};`,
			ExpectedOutput: "1",
		},
		{
			Name:           "hash assignment",
			Code:           `my %h; $h{x} = 10; say $h{x};`,
			ExpectedOutput: "10",
		},
		{
			Name:           "hash keys",
			Code:           `my %h = (a => 1, b => 2); my @k = sort keys %h; say "@k";`,
			ExpectedOutput: "a b",
		},
		{
			Name:           "hash values",
			Code:           `my %h = (a => 1, b => 2); my @v = sort values %h; say "@v";`,
			ExpectedOutput: "1 2",
		},
		{
			Name:           "hash exists",
			Code:           `my %h = (a => 1); say exists $h{a} ? "yes" : "no";`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "hash delete",
			Code:           `my %h = (a => 1, b => 2); delete $h{a}; say exists $h{a} ? "yes" : "no";`,
			ExpectedOutput: "no",
		},
		{
			Name:           "hash each",
			Code: `my %h = (x => 10);
while (my ($k, $v) = each %h) {
    say "$k=$v";
}`,
			ExpectedOutput: "x=10",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

// ============================================================
// Control Flow Tests
// ============================================================

func TestControlFlow(t *testing.T) {
	tests := []TestCase{
		{
			Name:           "if true",
			Code:           `if (1) { say "yes"; }`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "if false",
			Code:           `if (0) { say "yes"; } else { say "no"; }`,
			ExpectedOutput: "no",
		},
		{
			Name:           "if-elsif-else",
			Code: `my $x = 2;
if ($x == 1) { say "one"; }
elsif ($x == 2) { say "two"; }
else { say "other"; }`,
			ExpectedOutput: "two",
		},
		{
			Name:           "unless",
			Code:           `unless (0) { say "yes"; }`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "statement modifier if",
			Code:           `say "yes" if 1;`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "statement modifier unless",
			Code:           `say "yes" unless 0;`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "while loop",
			Code: `my $i = 0;
while ($i < 3) {
    say $i;
    $i++;
}`,
			ExpectedOutput: "0\n1\n2",
		},
		{
			Name:           "until loop",
			Code: `my $i = 0;
until ($i >= 3) {
    say $i;
    $i++;
}`,
			ExpectedOutput: "0\n1\n2",
		},
		{
			Name:           "for loop C-style",
			Code: `for (my $i = 0; $i < 3; $i++) {
    say $i;
}`,
			ExpectedOutput: "0\n1\n2",
		},
		{
			Name:           "foreach array",
			Code: `my @arr = (1, 2, 3);
foreach my $x (@arr) {
    say $x;
}`,
			ExpectedOutput: "1\n2\n3",
		},
		{
			Name:           "foreach range",
			Code: `foreach my $i (1..3) {
    say $i;
}`,
			ExpectedOutput: "1\n2\n3",
		},
		{
			Name:           "for as foreach",
			Code: `for my $x (1, 2, 3) {
    say $x;
}`,
			ExpectedOutput: "1\n2\n3",
		},
		{
			Name:           "last in loop",
			Code: `foreach my $i (1..10) {
    last if $i > 3;
    say $i;
}`,
			ExpectedOutput: "1\n2\n3",
		},
		{
			Name:           "next in loop",
			Code: `foreach my $i (1..5) {
    next if $i % 2 == 0;
    say $i;
}`,
			ExpectedOutput: "1\n3\n5",
		},
		{
			Name:           "ternary operator",
			Code:           `my $x = 10; say $x > 5 ? "big" : "small";`,
			ExpectedOutput: "big",
		},
		{
			Name:           "logical and",
			Code:           `say 1 && 2;`,
			ExpectedOutput: "2",
		},
		{
			Name:           "logical or",
			Code:           `say 0 || "default";`,
			ExpectedOutput: "default",
		},
		{
			Name:           "defined-or",
			Code:           `my $x; say $x // "undefined";`,
			ExpectedOutput: "undefined",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

// ============================================================
// Comparison Tests
// ============================================================

func TestComparisons(t *testing.T) {
	tests := []TestCase{
		{
			Name:           "numeric equal",
			Code:           `say 5 == 5 ? "yes" : "no";`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "numeric not equal",
			Code:           `say 5 != 3 ? "yes" : "no";`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "numeric less than",
			Code:           `say 3 < 5 ? "yes" : "no";`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "numeric greater than",
			Code:           `say 5 > 3 ? "yes" : "no";`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "numeric less equal",
			Code:           `say 5 <= 5 ? "yes" : "no";`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "numeric greater equal",
			Code:           `say 5 >= 5 ? "yes" : "no";`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "string equal",
			Code:           `say "abc" eq "abc" ? "yes" : "no";`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "string not equal",
			Code:           `say "abc" ne "xyz" ? "yes" : "no";`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "string less than",
			Code:           `say "abc" lt "xyz" ? "yes" : "no";`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "string greater than",
			Code:           `say "xyz" gt "abc" ? "yes" : "no";`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "spaceship operator",
			Code:           `say 5 <=> 3;`,
			ExpectedOutput: "1",
		},
		{
			Name:           "cmp operator",
			Code:           `say "abc" cmp "xyz";`,
			ExpectedOutput: "-1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

// ============================================================
// Subroutine Tests
// ============================================================

func TestSubroutines(t *testing.T) {
	tests := []TestCase{
		{
			Name: "simple sub",
			Code: `sub greet { say "Hello"; }
greet();`,
			ExpectedOutput: "Hello",
		},
		{
			Name: "sub with args",
			Code: `sub add {
    my ($a, $b) = @_;
    return $a + $b;
}
say add(2, 3);`,
			ExpectedOutput: "5",
		},
		{
			Name: "sub with return",
			Code: `sub double {
    my ($x) = @_;
    return $x * 2;
}
my $result = double(21);
say $result;`,
			ExpectedOutput: "42",
		},
		{
			Name: "sub multiple returns",
			Code: `sub minmax {
    my @nums = @_;
    my $min = $nums[0];
    my $max = $nums[0];
    foreach my $n (@nums) {
        $min = $n if $n < $min;
        $max = $n if $n > $max;
    }
    return ($min, $max);
}
my ($min, $max) = minmax(5, 2, 8, 1, 9);
say "min=$min, max=$max";`,
			ExpectedOutput: "min=1, max=9",
		},
		{
			Name: "recursive sub",
			Code: `sub factorial {
    my ($n) = @_;
    return 1 if $n <= 1;
    return $n * factorial($n - 1);
}
say factorial(5);`,
			ExpectedOutput: "120",
		},
		{
			Name: "sub wantarray",
			Code: `sub context_test {
    if (wantarray()) {
        return (1, 2, 3);
    } else {
        return "scalar";
    }
}
my @arr = context_test();
my $s = context_test();
say "@arr";
say $s;`,
			ExpectedOutput: "1 2 3\nscalar",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

// ============================================================
// Reference Tests
// ============================================================

func TestReferences(t *testing.T) {
	tests := []TestCase{
		{
			Name: "scalar reference",
			Code: `my $x = 42;
my $ref = \$x;
say $$ref;`,
			ExpectedOutput: "42",
		},
		{
			Name: "array reference",
			Code: `my @arr = (1, 2, 3);
my $ref = \@arr;
say $ref->[1];`,
			ExpectedOutput: "2",
		},
		{
			Name: "hash reference",
			Code: `my %h = (a => 1, b => 2);
my $ref = \%h;
say $ref->{a};`,
			ExpectedOutput: "1",
		},
		{
			Name: "anonymous array",
			Code: `my $arr = [1, 2, 3];
say $arr->[2];`,
			ExpectedOutput: "3",
		},
		{
			Name: "anonymous hash",
			Code: `my $h = {x => 10, y => 20};
say $h->{y};`,
			ExpectedOutput: "20",
		},
		{
			Name: "ref function",
			Code: `my $arr = [1, 2, 3];
my $h = {a => 1};
my $x = 5;
say ref($arr);
say ref($h);
say ref(\$x);`,
			ExpectedOutput: "ARRAY\nHASH\nSCALAR",
		},
		{
			Name: "modify through reference",
			Code: `my $x = 10;
my $ref = \$x;
$$ref = 20;
say $x;`,
			ExpectedOutput: "20",
		},
		{
			Name: "nested structures",
			Code: `my $data = {
    name => "Alice",
    scores => [90, 85, 95]
};
say $data->{name};
say $data->{scores}[1];`,
			ExpectedOutput: "Alice\n85",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

// ============================================================
// Regex Tests
// ============================================================

func TestRegex(t *testing.T) {
	tests := []TestCase{
		{
			Name:           "simple match",
			Code:           `my $s = "Hello World"; say $s =~ /World/ ? "yes" : "no";`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "no match",
			Code:           `my $s = "Hello"; say $s =~ /World/ ? "yes" : "no";`,
			ExpectedOutput: "no",
		},
		{
			Name:           "negated match",
			Code:           `my $s = "Hello"; say $s !~ /World/ ? "yes" : "no";`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "case insensitive",
			Code:           `my $s = "HELLO"; say $s =~ /hello/i ? "yes" : "no";`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "capture groups",
			Code:           `my $s = "Name: Alice"; $s =~ /Name: (\w+)/; say $1;`,
			ExpectedOutput: "Alice",
		},
		{
			Name:           "substitution",
			Code:           `my $s = "Hello World"; $s =~ s/World/Perl/; say $s;`,
			ExpectedOutput: "Hello Perl",
		},
		{
			Name:           "global substitution",
			Code:           `my $s = "aaa"; $s =~ s/a/b/g; say $s;`,
			ExpectedOutput: "bbb",
		},
		{
			Name:           "substitution with capture",
			Code:           `my $s = "Hello World"; $s =~ s/(\w+) (\w+)/$2 $1/; say $s;`,
			ExpectedOutput: "World Hello",
		},
		{
			Name:           "match in list context",
			Code:           `my $s = "a1b2c3"; my @nums = $s =~ /(\d)/g; say "@nums";`,
			ExpectedOutput: "1 2 3",
		},
		{
			Name:           "anchors",
			Code:           `say "hello" =~ /^hello$/ ? "yes" : "no";`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "word boundary",
			Code:           `say "hello world" =~ /\bworld\b/ ? "yes" : "no";`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "character class",
			Code:           `say "abc123" =~ /[0-9]+/ ? "yes" : "no";`,
			ExpectedOutput: "yes",
		},
		{
			Name:           "tr transliteration",
			Code:           `my $s = "hello"; $s =~ tr/a-z/A-Z/; say $s;`,
			ExpectedOutput: "HELLO",
		},
		{
			Name:           "tr count",
			Code:           `my $s = "hello world"; my $cnt = ($s =~ tr/o/o/); say $cnt;`,
			ExpectedOutput: "2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

// ============================================================
// File I/O Tests
// ============================================================

func TestFileIO(t *testing.T) {
	tests := []TestCase{
		{
			Name: "write and read file",
			Code: `my $fh;
open($fh, ">", "test_io.txt");
print $fh "Line 1\n";
print $fh "Line 2\n";
close($fh);

open($fh, "<", "test_io.txt");
my $line1 = <$fh>;
my $line2 = <$fh>;
close($fh);

chomp($line1);
chomp($line2);
say "1: $line1";
say "2: $line2";`,
			ExpectedOutput: "1: Line 1\n2: Line 2",
			CleanupFiles:   []string{"test_io.txt"},
		},
		{
			Name: "read existing file",
			Code: `open(my $fh, "<", "input_test.txt");
my $content = <$fh>;
close($fh);
chomp($content);
say "Got: $content";`,
			ExpectedOutput: "Got: Hello from test file",
			SetupFiles: map[string]string{
				"input_test.txt": "Hello from test file\n",
			},
		},
		{
			Name: "append mode",
			Code: `open(my $fh, ">", "append_test.txt");
print $fh "Line 1\n";
close($fh);

open($fh, ">>", "append_test.txt");
print $fh "Line 2\n";
close($fh);

open($fh, "<", "append_test.txt");
my @lines;
while (my $line = <$fh>) {
    chomp($line);
    push(@lines, $line);
}
close($fh);
say join(", ", @lines);`,
			ExpectedOutput: "Line 1, Line 2",
			CleanupFiles:   []string{"append_test.txt"},
		},
		{
			Name: "say to filehandle",
			Code: `open(my $fh, ">", "say_test.txt");
say $fh "Hello";
say $fh "World";
close($fh);

open($fh, "<", "say_test.txt");
my $l1 = <$fh>;
my $l2 = <$fh>;
close($fh);
chomp($l1);
chomp($l2);
say "$l1 $l2";`,
			ExpectedOutput: "Hello World",
			CleanupFiles:   []string{"say_test.txt"},
		},
		{
			Name: "3-arg open read",
			Code: `open(my $fh, "<", "three_arg_test.txt");
my $line = <$fh>;
close($fh);
chomp($line);
say $line;`,
			ExpectedOutput: "Test content",
			SetupFiles: map[string]string{
				"three_arg_test.txt": "Test content\n",
			},
		},
		{
			Name: "open return value",
			Code: `my $result = open(my $fh, "<", "nonexistent_file_xyz.txt");
say $result ? "opened" : "failed";`,
			ExpectedOutput: "failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

// ============================================================
// Built-in Function Tests
// ============================================================

func TestBuiltinFunctions(t *testing.T) {
	tests := []TestCase{
		{
			Name:           "abs positive",
			Code:           `say abs(42);`,
			ExpectedOutput: "42",
		},
		{
			Name:           "abs negative",
			Code:           `say abs(-42);`,
			ExpectedOutput: "42",
		},
		{
			Name:           "int",
			Code:           `say int(3.7);`,
			ExpectedOutput: "3",
		},
		{
			Name:           "sqrt",
			Code:           `say sqrt(16);`,
			ExpectedOutput: "4",
		},
		{
			Name:           "sin",
			Code:           `say int(sin(0) * 100);`,
			ExpectedOutput: "0",
		},
		{
			Name:           "cos",
			Code:           `say int(cos(0) * 100);`,
			ExpectedOutput: "100",
		},
		{
			Name:           "log",
			Code:           `say int(log(2.718281828) * 100);`,
			ExpectedOutput: "100",
		},
		{
			Name:           "exp",
			Code:           `say int(exp(1) * 100);`,
			ExpectedOutput: "271",
		},
		{
			Name:           "rand",
			Code:           `my $r = rand(); say $r >= 0 && $r < 1 ? "ok" : "fail";`,
			ExpectedOutput: "ok",
		},
		{
			Name:           "defined",
			Code:           `my $x = 1; my $y; say defined($x) ? "yes" : "no"; say defined($y) ? "yes" : "no";`,
			ExpectedOutput: "yes\nno",
		},
		{
			Name:           "wantarray in scalar",
			Code:           `sub test { wantarray() ? "list" : "scalar" } my $x = test(); say $x;`,
			ExpectedOutput: "scalar",
		},
		{
			Name:           "scalar on array",
			Code:           `my @arr = (1, 2, 3, 4, 5); say scalar(@arr);`,
			ExpectedOutput: "5",
		},
		{
			Name:           "chr and ord",
			Code:           `say chr(65); say ord("A");`,
			ExpectedOutput: "A\n65",
		},
		{
			Name:           "hex",
			Code:           `say hex("ff");`,
			ExpectedOutput: "255",
		},
		{
			Name:           "oct",
			Code:           `say oct("77");`,
			ExpectedOutput: "63",
		},
		{
			Name:           "pack and unpack",
			Code:           `my $packed = pack("A3", "ABC"); say $packed;`,
			ExpectedOutput: "ABC",
		},
		{
			Name:           "lc and uc",
			Code:           `say lc("HELLO"); say uc("world");`,
			ExpectedOutput: "hello\nWORLD",
		},
		{
			Name:           "grep",
			Code:           `my @nums = (1, 2, 3, 4, 5); my @even = grep { $_ % 2 == 0 } @nums; say "@even";`,
			ExpectedOutput: "2 4",
		},
		{
			Name:           "map",
			Code:           `my @nums = (1, 2, 3); my @doubled = map { $_ * 2 } @nums; say "@doubled";`,
			ExpectedOutput: "2 4 6",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

// ============================================================
// Edge Cases and Special Features
// ============================================================

func TestEdgeCases(t *testing.T) {
	tests := []TestCase{
		{
			Name:           "empty string is false",
			Code:           `say "" ? "true" : "false";`,
			ExpectedOutput: "false",
		},
		{
			Name:           "zero string is false",
			Code:           `say "0" ? "true" : "false";`,
			ExpectedOutput: "false",
		},
		{
			Name:           "zero is false",
			Code:           `say 0 ? "true" : "false";`,
			ExpectedOutput: "false",
		},
		{
			Name:           "undef is false",
			Code:           `my $x; say $x ? "true" : "false";`,
			ExpectedOutput: "false",
		},
		{
			Name:           "empty array is false",
			Code:           `my @arr; say @arr ? "true" : "false";`,
			ExpectedOutput: "false",
		},
		{
			Name:           "non-empty array is true",
			Code:           `my @arr = (1); say @arr ? "true" : "false";`,
			ExpectedOutput: "true",
		},
		{
			Name:           "heredoc",
			Code: `my $text = <<END;
Hello
World
END
chomp($text);
say $text;`,
			ExpectedOutput: "Hello\nWorld",
		},
		{
			Name:           "qw operator",
			Code:           `my @arr = qw(one two three); say "@arr";`,
			ExpectedOutput: "one two three",
		},
		{
			Name:           "q and qq",
			Code:           `my $x = 5; say q(no $x); say qq(yes $x);`,
			ExpectedOutput: "no $x\nyes 5",
		},
		{
			Name:           "string numeric context",
			Code:           `my $s = "42abc"; say $s + 8;`,
			ExpectedOutput: "50",
		},
		{
			Name:           "autovivification array",
			Code:           `my @arr; $arr[5] = "x"; say scalar(@arr);`,
			ExpectedOutput: "6",
		},
		{
			Name:           "autovivification hash",
			Code:           `my %h; $h{a}{b} = 1; say $h{a}{b};`,
			ExpectedOutput: "1",
		},
		{
			Name:           "chained comparison",
			Code:           `my $x = 5; say 1 < $x && $x < 10 ? "in range" : "out";`,
			ExpectedOutput: "in range",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

// ============================================================
// Comprehensive Integration Tests
// ============================================================

func TestIntegration(t *testing.T) {
	tests := []TestCase{
		{
			Name: "FizzBuzz",
			Code: `foreach my $i (1..15) {
    if ($i % 15 == 0) {
        say "FizzBuzz";
    } elsif ($i % 3 == 0) {
        say "Fizz";
    } elsif ($i % 5 == 0) {
        say "Buzz";
    } else {
        say $i;
    }
}`,
			ExpectedOutput: "1\n2\nFizz\n4\nBuzz\nFizz\n7\n8\nFizz\nBuzz\n11\nFizz\n13\n14\nFizzBuzz",
		},
		{
			Name: "Fibonacci",
			Code: `sub fib {
    my ($n) = @_;
    return $n if $n < 2;
    return fib($n-1) + fib($n-2);
}
my @results;
foreach my $i (0..9) {
    push(@results, fib($i));
}
say "@results";`,
			ExpectedOutput: "0 1 1 2 3 5 8 13 21 34",
		},
		{
			Name: "Word frequency counter",
			Code: `my $text = "the quick brown fox jumps over the lazy dog the fox";
my @words = split(/\s+/, $text);
my %freq;
foreach my $word (@words) {
    $freq{$word}++;
}
my @sorted = sort { $freq{$b} <=> $freq{$a} } keys %freq;
foreach my $w (@sorted[0..2]) {
    say "$w: $freq{$w}";
}`,
			ExpectedOutput: "the: 3\nfox: 2\nquick: 1",
		},
		{
			Name: "Prime sieve",
			Code: `my @sieve = (0) x 30;
for my $i (2..29) {
    next if $sieve[$i];
    for (my $j = $i * 2; $j < 30; $j += $i) {
        $sieve[$j] = 1;
    }
}
my @primes;
for my $i (2..29) {
    push(@primes, $i) unless $sieve[$i];
}
say "@primes";`,
			ExpectedOutput: "2 3 5 7 11 13 17 19 23 29",
		},
		{
			Name: "Data processing pipeline",
			Code: `my @data = (
    "Alice,30,Engineer",
    "Bob,25,Designer",
    "Charlie,35,Manager"
);

my @parsed;
foreach my $line (@data) {
    my @fields = split(",", $line);
    push(@parsed, { name => $fields[0], age => $fields[1], role => $fields[2] });
}

my @sorted = sort { $a->{age} <=> $b->{age} } @parsed;
foreach my $person (@sorted) {
    say "$person->{name} ($person->{age}): $person->{role}";
}`,
			ExpectedOutput: "Bob (25): Designer\nAlice (30): Engineer\nCharlie (35): Manager",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

// ============================================================
// Main test runner
// ============================================================

func TestMain(m *testing.M) {
	// Change to project root
	os.Chdir("..")
	
	// Determine exe name based on OS
	exeName := "perlc"
	if os.PathSeparator == '\\' {
		exeName = "perlc.exe"
	}
	
	// Check if perlc exists
	if _, err := os.Stat("./" + exeName); os.IsNotExist(err) {
		// Try to build it
		fmt.Println("Building perlc...")
		cmd := exec.Command("go", "build", "-o", exeName, "./cmd/perlc")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to build perlc: %v\n", err)
			os.Exit(1)
		}
	}

	code := m.Run()
	
	// Cleanup
	os.Remove(exeName)
	
	os.Exit(code)
}
